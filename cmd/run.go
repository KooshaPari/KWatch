package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"kwatch/config"
	"kwatch/runner"
)

var (
	runCommand string
	runVerbose bool
	runFormat  string
)

// runResponse represents the JSON response for run command
type runResponse struct {
	Directory string                             `json:"directory"`
	Timestamp string                             `json:"timestamp"`
	Summary   runSummary                         `json:"summary"`
	Results   map[string]runCommandResult        `json:"results"`
}

// runSummary provides a summary of the run
type runSummary struct {
	Total    int `json:"total"`
	Passed   int `json:"passed"`
	Failed   int `json:"failed"`
	Duration string `json:"duration"`
}

// runCommandResult represents a command result in the run response
type runCommandResult struct {
	Command    string `json:"command"`
	Passed     bool   `json:"passed"`
	IssueCount int    `json:"issue_count"`
	Duration   string `json:"duration"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
}

var runCmd = &cobra.Command{
	Use:   "run [directory]",
	Short: "Force manual execution of commands",
	Long: `Force a manual execution of all configured commands or a specific command.

This command runs the build tools (TypeScript, linting, tests) and reports the results.
Use this to trigger execution on demand rather than waiting for file changes.

Examples:
  kwatch run                           # Run all commands
  kwatch run --command tsc             # Run only TypeScript check
  kwatch run --command lint            # Run only linting
  kwatch --dir /path/to/project run    # Run in specific directory (flag)
  kwatch . run                         # Run in current directory
  kwatch run --verbose                 # Show detailed output
  kwatch run --format json            # Output results as JSON`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := getWorkingDirectory(args)

		absDir, err := filepath.Abs(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving directory: %v\n", err)
			os.Exit(1)
		}

		// Check if directory exists
		if _, err := os.Stat(absDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Directory does not exist: %s\n", absDir)
			os.Exit(1)
		}

		// Load kwatch configuration
		kwatchConfig, err := config.Load(absDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading kwatch config: %v\n", err)
			os.Exit(1)
		}

		// Create runner configuration
		runnerConfig := runner.RunnerConfig{
			DefaultTimeout: 30 * time.Second,
			MaxParallel:    kwatchConfig.MaxParallel,
			WorkingDir:     absDir,
		}

		r := runner.NewRunner(runnerConfig, kwatchConfig)
		ctx := context.Background()
		start := time.Now()

		var results map[runner.CommandType]runner.CommandResult

		if runCommand != "" {
			// Run specific command
			results = runSpecificCommand(ctx, r, runCommand)
		} else {
			// Run all commands
			results = r.RunAll(ctx)
		}

		totalDuration := time.Since(start)

		// Output results based on format
		switch runFormat {
		case "json":
			outputRunJSON(absDir, results, totalDuration)
		case "compact":
			outputRunCompact(results)
		default:
			outputRunDefault(results, totalDuration)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&runCommand, "command", "c", "", "Run specific command (tsc, lint, test)")
	runCmd.Flags().BoolVarP(&runVerbose, "verbose", "v", false, "Show verbose output including command output")
	runCmd.Flags().StringVarP(&runFormat, "format", "f", "default", "Output format (default, json, compact)")
}

// runSpecificCommand runs a specific command type
func runSpecificCommand(ctx context.Context, r *runner.Runner, cmdType string) map[runner.CommandType]runner.CommandResult {
	results := make(map[runner.CommandType]runner.CommandResult)

	// Map command string to command type
	var targetType runner.CommandType
	var cmd runner.Command

	switch strings.ToLower(cmdType) {
	case "tsc", "typescript":
		targetType = runner.TypescriptCheck
		cmd = runner.Command{
			Type:    runner.TypescriptCheck,
			Command: "npx",
			Args:    []string{"tsc", "--noEmit"},
			Timeout: 30 * time.Second,
		}
	case "lint", "eslint":
		targetType = runner.LintCheck
		cmd = runner.Command{
			Type:    runner.LintCheck,
			Command: "npx",
			Args:    []string{"eslint", ".", "--ext", ".ts,.tsx,.js,.jsx"},
			Timeout: 30 * time.Second,
		}
	case "test":
		targetType = runner.TestRunner
		cmd = runner.Command{
			Type:    runner.TestRunner,
			Command: "npm",
			Args:    []string{"test"},
			Timeout: 60 * time.Second,
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command type: %s\n", cmdType)
		fmt.Fprintf(os.Stderr, "Available commands: tsc, lint, test\n")
		os.Exit(1)
	}

	// Run the specific command
	result := r.RunCommand(ctx, cmd)
	results[targetType] = result

	return results
}

// outputRunJSON outputs run results in JSON format
func outputRunJSON(directory string, results map[runner.CommandType]runner.CommandResult, totalDuration time.Duration) {
	response := runResponse{
		Directory: directory,
		Timestamp: time.Now().Format(time.RFC3339),
		Results:   make(map[string]runCommandResult),
	}

	// Calculate summary
	total := len(results)
	passed := 0
	failed := 0

	cmdNames := map[runner.CommandType]string{
		runner.TypescriptCheck: "tsc",
		runner.LintCheck:       "lint",
		runner.TestRunner:      "test",
	}

	for cmdType, result := range results {
		cmdName := cmdNames[cmdType]
		if cmdName == "" {
			cmdName = string(cmdType)
		}

		runResult := runCommandResult{
			Command:    result.Command,
			Passed:     result.Passed,
			IssueCount: result.IssueCount,
			Duration:   formatDuration(result.Duration),
		}

		if runVerbose {
			runResult.Output = result.Output
			runResult.Error = result.Error
		}

		response.Results[cmdName] = runResult

		if result.Passed {
			passed++
		} else {
			failed++
		}
	}

	response.Summary = runSummary{
		Total:    total,
		Passed:   passed,
		Failed:   failed,
		Duration: formatDuration(totalDuration),
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonBytes))
}

// outputRunCompact outputs run results in compact format
func outputRunCompact(results map[runner.CommandType]runner.CommandResult) {
	compact := runner.FormatCompactStatus(results)
	fmt.Println(compact)
}

// outputRunDefault outputs run results in default format
func outputRunDefault(results map[runner.CommandType]runner.CommandResult, totalDuration time.Duration) {
	fmt.Printf("Running commands...\n\n")

	total := len(results)
	passed := 0
	failed := 0

	// Display results for each command
	for _, result := range results {
		cmdName := getCommandTypeLabel(result.Command)
		
		status := "✓ PASSED"
		if !result.Passed {
			status = "✗ FAILED"
			failed++
		} else {
			passed++
		}

		fmt.Printf("%s: %s", cmdName, status)
		if result.IssueCount > 0 {
			fmt.Printf(" (%d issues)", result.IssueCount)
		}
		fmt.Printf(" in %s\n", formatDuration(result.Duration))

		if runVerbose && result.Output != "" {
			fmt.Printf("  Output: %s\n", truncateString(result.Output, 200))
		}

		if result.Error != "" {
			fmt.Printf("  Error: %s\n", truncateString(result.Error, 200))
		}
	}

	// Display summary
	fmt.Printf("\nSummary: %d/%d passed", passed, total)
	if failed > 0 {
		fmt.Printf(", %d failed", failed)
	}
	fmt.Printf(" (completed in %s)\n", formatDuration(totalDuration))

	// Exit with error code if any command failed
	if failed > 0 {
		os.Exit(1)
	}
}