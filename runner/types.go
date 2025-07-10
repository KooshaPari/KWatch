package runner

import (
	"strings"
	"sync"
	"time"
)

// CommandResult represents the result of a command execution
type CommandResult struct {
	Command    string        `json:"command"`
	Passed     bool          `json:"passed"`
	IssueCount int           `json:"issue_count"`
	FileCount  int           `json:"file_count"`
	Output     string        `json:"output"`
	Duration   time.Duration `json:"duration"`
	Timestamp  time.Time     `json:"timestamp"`
	Error      string        `json:"error,omitempty"`
	// Test-specific fields
	TotalTests   int `json:"total_tests,omitempty"`
	PassedTests  int `json:"passed_tests,omitempty"`
	FailedTests  int `json:"failed_tests,omitempty"`
	// GitHub Actions specific fields
	WorkflowName    string              `json:"workflow_name,omitempty"`
	RunID          int64               `json:"run_id,omitempty"`
	WorkflowStatus string              `json:"workflow_status,omitempty"`
	JobResults     []GitHubActionJob   `json:"job_results,omitempty"`
}

// RunResult represents the result of running multiple commands
type RunResult struct {
	Timestamp time.Time                `json:"timestamp"`
	Commands  map[string]CommandResult `json:"commands"`
	Duration  time.Duration            `json:"duration"`
}

// CommandType represents the type of command being executed
type CommandType string

const (
	TypescriptCheck CommandType = "typescript"
	LintCheck       CommandType = "lint"
	TestRunner      CommandType = "test"
	GitHubActions   CommandType = "github_actions"
)

// Command represents a command to be executed
type Command struct {
	Type    CommandType `json:"type"`
	Command string      `json:"command"`
	Args    []string    `json:"args"`
	Timeout time.Duration `json:"timeout"`
}

// RunnerConfig holds configuration for the command runner
type RunnerConfig struct {
	DefaultTimeout time.Duration `json:"default_timeout"`
	MaxParallel    int           `json:"max_parallel"`
	WorkingDir     string        `json:"working_dir"`
}

// ResultHistory stores command execution history
type ResultHistory struct {
	Results []CommandResult `json:"results"`
	mutex   sync.RWMutex
}

// Add adds a result to the history
func (h *ResultHistory) Add(result CommandResult) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.Results = append(h.Results, result)
}

// GetLatest returns the latest results for each command type
func (h *ResultHistory) GetLatest() map[CommandType]CommandResult {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	latest := make(map[CommandType]CommandResult)
	for _, result := range h.Results {
		cmdType := getCommandType(result.Command)
		if existing, exists := latest[cmdType]; !exists || result.Timestamp.After(existing.Timestamp) {
			latest[cmdType] = result
		}
	}
	return latest
}

// GetAll returns all results
func (h *ResultHistory) GetAll() []CommandResult {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	results := make([]CommandResult, len(h.Results))
	copy(results, h.Results)
	return results
}

// Clear clears all results
func (h *ResultHistory) Clear() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.Results = nil
}

// getCommandType determines command type from command string
func getCommandType(command string) CommandType {
	switch {
	case strings.Contains(command, "tsc"):
		return TypescriptCheck
	case strings.Contains(command, "lint"):
		return LintCheck
	case strings.Contains(command, "test"):
		return TestRunner
	case strings.Contains(command, "github"):
		return GitHubActions
	default:
		return CommandType(command)
	}
}

// GitHubActionJob represents a single job in a GitHub Actions workflow
type GitHubActionJob struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	StartedAt  string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
}

// GitHubConfig represents GitHub API configuration
type GitHubConfig struct {
	Owner      string `json:"owner"`
	Repo       string `json:"repo"`
	Token      string `json:"token,omitempty"`
	Branch     string `json:"branch,omitempty"`
}

// WorkflowRun represents a GitHub Actions workflow run
type WorkflowRun struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	HeadBranch string `json:"head_branch"`
	HeadSHA    string `json:"head_sha"`
}