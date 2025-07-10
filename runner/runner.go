package runner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
	
	"kwatch/config"
)

// Runner manages command execution and history
type Runner struct {
	config       RunnerConfig
	history      *ResultHistory
	parser       *Parser
	mutex        sync.RWMutex
	kwatchConfig *config.Config
	githubClient *GitHubClient
}

// NewRunner creates a new runner instance
func NewRunner(config RunnerConfig, kwatchConfig *config.Config) *Runner {
	runner := &Runner{
		config:       config,
		history:      &ResultHistory{},
		parser:       NewParser(),
		kwatchConfig: kwatchConfig,
	}
	
	// Initialize GitHub client if possible
	if config.WorkingDir != "" {
		if githubClient, err := GitHubFromRepository(config.WorkingDir); err == nil {
			runner.githubClient = githubClient
		}
	}
	
	return runner
}

// RunCommand executes a single command and returns the result
func (r *Runner) RunCommand(ctx context.Context, command Command) CommandResult {
	// Handle GitHub Actions commands differently
	if command.Type == GitHubActions {
		return r.runGitHubCommand(ctx, command)
	}
	
	start := time.Now()
	result := CommandResult{
		Command:   command.Command,
		Timestamp: start,
	}

	// Create command context with timeout
	timeout := command.Timeout
	if timeout == 0 {
		timeout = r.config.DefaultTimeout
	}
	
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute command
	cmd := exec.CommandContext(cmdCtx, command.Command, command.Args...)
	if r.config.WorkingDir != "" {
		cmd.Dir = r.config.WorkingDir
	}

	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(start)
	result.Output = string(output)

	if err != nil {
		result.Error = err.Error()
	}
	
	// Parse output based on command type
	if command.Type == TestRunner {
		testResult := r.parser.ParseTestOutput(result.Output)
		result.Passed = testResult.Passed
		result.IssueCount = testResult.FailedTests
		result.TotalTests = testResult.TotalTests
		result.PassedTests = testResult.PassedTests
		result.FailedTests = testResult.FailedTests
	} else {
		passed, issueCount := r.parseCommandOutput(command.Type, result.Output)
		result.Passed = passed
		result.IssueCount = issueCount
	}
	
	// For lint commands, try to extract file count
	if command.Type == LintCheck {
		result.FileCount = r.extractFileCount(result.Output)
	} else {
		result.FileCount = 0
	}

	// Add to history
	r.history.Add(result)

	return result
}

// runGitHubCommand handles GitHub Actions command execution
func (r *Runner) runGitHubCommand(ctx context.Context, command Command) CommandResult {
	if r.githubClient == nil {
		return CommandResult{
			Command:   command.Command,
			Timestamp: time.Now(),
			Error:     "GitHub client not initialized - no GitHub repository detected or token missing",
			Duration:  0,
		}
	}
	
	result, err := r.githubClient.CheckWorkflowStatus(ctx)
	if err != nil {
		result.Error = err.Error()
	}
	
	// Add to history
	r.history.Add(result)
	
	return result
}

// RunAll executes all configured commands
func (r *Runner) RunAll(ctx context.Context) map[CommandType]CommandResult {
	commands := r.getDefaultCommands()
	results := make(map[CommandType]CommandResult)
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	for cmdType, cmd := range commands {
		wg.Add(1)
		go func(ct CommandType, c Command) {
			defer wg.Done()
			result := r.RunCommand(ctx, c)
			mu.Lock()
			results[ct] = result
			mu.Unlock()
		}(cmdType, cmd)
	}
	
	wg.Wait()
	return results
}

// GetLatestResults returns the latest results for each command type
func (r *Runner) GetLatestResults() map[CommandType]CommandResult {
	return r.history.GetLatest()
}

// GetHistory returns the full command history
func (r *Runner) GetHistory() []CommandResult {
	return r.history.GetAll()
}

// ClearHistory clears the command history
func (r *Runner) ClearHistory() {
	r.history.Clear()
}

// getDefaultCommands returns the configured commands to run
func (r *Runner) getDefaultCommands() map[CommandType]Command {
	commands := make(map[CommandType]Command)
	
	// Use kwatch config if available, otherwise fall back to hardcoded defaults
	if r.kwatchConfig != nil {
		enabledCommands := r.kwatchConfig.GetEnabledCommands()
		
		for name, configCmd := range enabledCommands {
			var cmdType CommandType
			switch name {
			case "typescript":
				cmdType = TypescriptCheck
			case "lint":
				cmdType = LintCheck
			case "test":
				cmdType = TestRunner
			case "github_actions":
				cmdType = GitHubActions
			default:
				// For custom commands, use the name as the type
				cmdType = CommandType(name)
			}
			
			// Get timeout for this command
			timeout := r.kwatchConfig.GetTimeout(name)
			
			commands[cmdType] = Command{
				Type:    cmdType,
				Command: configCmd.Command,
				Args:    configCmd.Args,
				Timeout: timeout,
			}
		}
	} else {
		// Fallback to hardcoded defaults
		commands = map[CommandType]Command{
			TypescriptCheck: {
				Type:    TypescriptCheck,
				Command: "npx",
				Args:    []string{"tsc", "--noEmit"},
				Timeout: 30 * time.Second,
			},
			LintCheck: {
				Type:    LintCheck,
				Command: "npx",
				Args:    []string{"eslint", ".", "--ext", ".ts,.tsx,.js,.jsx"},
				Timeout: 30 * time.Second,
			},
			TestRunner: {
				Type:    TestRunner,
				Command: "npm",
				Args:    []string{"test"},
				Timeout: 60 * time.Second,
			},
		}
	}
	
	// Always add GitHub Actions if client is available
	if r.githubClient != nil {
		commands[GitHubActions] = Command{
			Type:    GitHubActions,
			Command: "github_actions",
			Args:    []string{},
			Timeout: 30 * time.Second,
		}
	}
	
	return commands
}


// FormatCompactStatus formats results as a compact one-line status
func FormatCompactStatus(results map[CommandType]CommandResult) string {
	var parts []string
	
	// Order: TSC, LINT, TEST, GITHUB
	types := []CommandType{TypescriptCheck, LintCheck, TestRunner, GitHubActions}
	labels := map[CommandType]string{
		TypescriptCheck: "TSC",
		LintCheck:       "LINT",
		TestRunner:      "TEST",
		GitHubActions:   "GH",
	}
	
	for _, cmdType := range types {
		if result, exists := results[cmdType]; exists {
			symbol := "✓"
			if !result.Passed {
				symbol = "✗"
			}
			
			if cmdType == TestRunner {
				// For tests, show PASS/TOTAL format
				if result.TotalTests > 0 {
					parts = append(parts, fmt.Sprintf("%s:%s%d/%d", labels[cmdType], symbol, result.PassedTests, result.TotalTests))
				} else {
					parts = append(parts, fmt.Sprintf("%s:%s%d", labels[cmdType], symbol, result.IssueCount))
				}
			} else if cmdType == GitHubActions {
				// For GitHub Actions, show number of failed jobs
				if len(result.JobResults) > 0 {
					failedJobs := 0
					for _, job := range result.JobResults {
						if job.Conclusion == "failure" || job.Conclusion == "cancelled" || job.Conclusion == "timed_out" {
							failedJobs++
						}
					}
					parts = append(parts, fmt.Sprintf("%s:%s%d/%d", labels[cmdType], symbol, failedJobs, len(result.JobResults)))
				} else {
					parts = append(parts, fmt.Sprintf("%s:%s", labels[cmdType], symbol))
				}
			} else if result.IssueCount > 0 && result.FileCount > 0 {
				parts = append(parts, fmt.Sprintf("%s:%s%d/%d", labels[cmdType], symbol, result.IssueCount, result.FileCount))
			} else {
				parts = append(parts, fmt.Sprintf("%s:%s%d", labels[cmdType], symbol, result.IssueCount))
			}
		}
	}
	
	return strings.Join(parts, " ")
}

// parseCommandOutput parses command output based on command type
func (r *Runner) parseCommandOutput(cmdType CommandType, output string) (bool, int) {
	switch cmdType {
	case TypescriptCheck:
		return r.parser.ParseTypeScriptOutput(output)
	case LintCheck:
		return r.parser.ParseLintOutput(output)
	default:
		return r.parser.ParseGenericOutput(output)
	}
}

// extractFileCount extracts the number of files with issues from ESLint output
func (r *Runner) extractFileCount(output string) int {
	// Count unique file paths in ESLint output
	lines := strings.Split(output, "\n")
	fileMap := make(map[string]bool)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// ESLint file paths start with / or ./ and contain .ts, .js, etc.
		if (strings.HasPrefix(line, "/") || strings.HasPrefix(line, "./")) && 
		   (strings.Contains(line, ".ts") || strings.Contains(line, ".js") || 
		    strings.Contains(line, ".tsx") || strings.Contains(line, ".jsx")) {
			// Extract just the file path (before any spaces/colons)
			parts := strings.Fields(line)
			if len(parts) > 0 {
				fileMap[parts[0]] = true
			}
		}
	}
	
	return len(fileMap)
}