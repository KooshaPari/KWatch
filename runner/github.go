package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GitHubClient handles GitHub API interactions
type GitHubClient struct {
	config     GitHubConfig
	httpClient *http.Client
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient(config GitHubConfig) *GitHubClient {
	return &GitHubClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GitHubFromRepository creates a GitHub client by detecting repository info
func GitHubFromRepository(workingDir string) (*GitHubClient, error) {
	config, err := detectGitHubConfig(workingDir)
	if err != nil {
		return nil, err
	}
	
	return NewGitHubClient(config), nil
}

// detectGitHubConfig attempts to detect GitHub repository configuration
func detectGitHubConfig(workingDir string) (GitHubConfig, error) {
	config := GitHubConfig{}
	
	// Try to read from git remote
	gitDir := filepath.Join(workingDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		if remoteConfig, err := parseGitRemote(gitDir); err == nil {
			config.Owner = remoteConfig.Owner
			config.Repo = remoteConfig.Repo
		}
	}
	
	// Try to get token from environment first
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		config.Token = token
	} else if token := os.Getenv("GH_TOKEN"); token != "" {
		config.Token = token
	} else {
		// Try to get token from secure store (with safe error handling)
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Silently recover from any secure store panics
				}
			}()
			store := NewSecureTokenStore()
			if store != nil && store.HasStoredToken() {
				if token, err := store.GetToken(); err == nil && token != "" {
					config.Token = token
				}
			}
		}()
	}
	
	// Set default branch if not specified
	if config.Branch == "" {
		config.Branch = "main"
	}
	
	// Validate required fields
	if config.Owner == "" || config.Repo == "" {
		return config, fmt.Errorf("could not detect GitHub repository owner/name")
	}
	
	return config, nil
}

// GitRemoteConfig represents parsed git remote configuration
type GitRemoteConfig struct {
	Owner string
	Repo  string
}

// parseGitRemote parses git remote configuration to extract GitHub info
func parseGitRemote(gitDir string) (GitRemoteConfig, error) {
	configFile := filepath.Join(gitDir, "config")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return GitRemoteConfig{}, err
	}
	
	content := string(data)
	lines := strings.Split(content, "\n")
	
	var inRemoteOrigin bool
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if line == `[remote "origin"]` {
			inRemoteOrigin = true
			continue
		}
		
		if strings.HasPrefix(line, "[") && line != `[remote "origin"]` {
			inRemoteOrigin = false
			continue
		}
		
		if inRemoteOrigin && strings.HasPrefix(line, "url = ") {
			url := strings.TrimPrefix(line, "url = ")
			return parseGitHubURL(url)
		}
	}
	
	return GitRemoteConfig{}, fmt.Errorf("GitHub remote origin not found")
}

// parseGitHubURL parses a GitHub URL to extract owner and repo
func parseGitHubURL(url string) (GitRemoteConfig, error) {
	// Handle SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		path := strings.TrimPrefix(url, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) == 2 {
			return GitRemoteConfig{Owner: parts[0], Repo: parts[1]}, nil
		}
	}
	
	// Handle HTTPS format: https://github.com/owner/repo.git
	if strings.HasPrefix(url, "https://github.com/") {
		path := strings.TrimPrefix(url, "https://github.com/")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) == 2 {
			return GitRemoteConfig{Owner: parts[0], Repo: parts[1]}, nil
		}
	}
	
	return GitRemoteConfig{}, fmt.Errorf("unsupported GitHub URL format: %s", url)
}

// GetLatestWorkflowRuns fetches the latest workflow runs for the repository
func (gc *GitHubClient) GetLatestWorkflowRuns(ctx context.Context) ([]WorkflowRun, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs?per_page=10", 
		gc.config.Owner, gc.config.Repo)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add authorization header if token is available
	if gc.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+gc.config.Token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "kwatch/1.0")
	
	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}
	
	var response struct {
		WorkflowRuns []WorkflowRun `json:"workflow_runs"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return response.WorkflowRuns, nil
}

// GetWorkflowJobs fetches jobs for a specific workflow run
func (gc *GitHubClient) GetWorkflowJobs(ctx context.Context, runID int64) ([]GitHubActionJob, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs/%d/jobs", 
		gc.config.Owner, gc.config.Repo, runID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add authorization header if token is available
	if gc.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+gc.config.Token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "kwatch/1.0")
	
	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}
	
	var response struct {
		Jobs []GitHubActionJob `json:"jobs"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return response.Jobs, nil
}

// CheckWorkflowStatus fetches the latest workflow status and returns a CommandResult
func (gc *GitHubClient) CheckWorkflowStatus(ctx context.Context) (CommandResult, error) {
	start := time.Now()
	result := CommandResult{
		Command:   "github_actions",
		Timestamp: start,
	}
	
	// Get latest workflow runs
	runs, err := gc.GetLatestWorkflowRuns(ctx)
	if err != nil {
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result, nil
	}
	
	if len(runs) == 0 {
		result.Passed = true
		result.Output = "No workflow runs found"
		result.Duration = time.Since(start)
		return result, nil
	}
	
	// Use the latest run for the main branch or current branch
	var latestRun WorkflowRun
	for _, run := range runs {
		if run.HeadBranch == gc.config.Branch || 
		   (gc.config.Branch == "main" && (run.HeadBranch == "main" || run.HeadBranch == "master")) {
			latestRun = run
			break
		}
	}
	
	// If no run found for target branch, use the most recent
	if latestRun.ID == 0 && len(runs) > 0 {
		latestRun = runs[0]
	}
	
	result.WorkflowName = latestRun.Name
	result.RunID = latestRun.ID
	result.WorkflowStatus = latestRun.Status
	
	// Get jobs for this run
	jobs, err := gc.GetWorkflowJobs(ctx, latestRun.ID)
	if err != nil {
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result, nil
	}
	
	result.JobResults = jobs
	
	// Calculate status based on workflow conclusion
	switch latestRun.Conclusion {
	case "success":
		result.Passed = true
		result.IssueCount = 0
	case "failure", "cancelled", "timed_out":
		result.Passed = false
		// Count failed jobs as issues
		failedJobs := 0
		for _, job := range jobs {
			if job.Conclusion == "failure" || job.Conclusion == "cancelled" || job.Conclusion == "timed_out" {
				failedJobs++
			}
		}
		result.IssueCount = failedJobs
	case "":
		// Still running
		result.Passed = true // Don't mark as failed while running
		result.IssueCount = 0
	default:
		result.Passed = false
		result.IssueCount = 1
	}
	
	// Format output summary
	summary := fmt.Sprintf("Workflow: %s\nStatus: %s", latestRun.Name, latestRun.Status)
	if latestRun.Conclusion != "" {
		summary += fmt.Sprintf("\nConclusion: %s", latestRun.Conclusion)
	}
	summary += fmt.Sprintf("\nJobs: %d", len(jobs))
	
	result.Output = summary
	result.Duration = time.Since(start)
	
	return result, nil
}