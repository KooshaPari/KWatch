package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"kwatch/runner"
)

var (
	compactFlag bool
)

// statusResponse represents the JSON response format for status command
type statusResponse struct {
	Directory string                             `json:"directory"`
	Timestamp string                             `json:"timestamp"`
	Commands  map[string]statusCommandResult    `json:"commands"`
}

// statusCommandResult represents a command result in the status response
type statusCommandResult struct {
	Passed     bool   `json:"passed"`
	IssueCount int    `json:"issue_count"`
	Duration   string `json:"duration"`
}

var statusCmd = &cobra.Command{
	Use:   "status [directory]",
	Short: "Get current build status",
	Long: `Get the current build status of the project.
	
By default, outputs detailed JSON status information.
Use --compact for a one-line status summary.

Examples:
  kwatch status                    # Status for current directory
  kwatch status /path/to/project   # Status for specific directory
  kwatch --dir /path/to/project status # Status for specific directory (flag)
  kwatch status --compact          # Compact one-line output
  kwatch . status                  # Status for current directory`,
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

		// Create runner configuration
		config := runner.RunnerConfig{
			DefaultTimeout: 30 * time.Second,
			MaxParallel:    3,
			WorkingDir:     absDir,
		}

		r := runner.NewRunner(config)
		ctx := context.Background()

		// Run all commands
		results := r.RunAll(ctx)

		if compactFlag {
			// Output compact status
			compact := runner.FormatCompactStatus(results)
			fmt.Println(compact)
		} else {
			// Output detailed JSON status
			response := statusResponse{
				Directory: absDir,
				Timestamp: time.Now().Format(time.RFC3339),
				Commands:  make(map[string]statusCommandResult),
			}

			// Convert results to response format
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

				response.Commands[cmdName] = statusCommandResult{
					Passed:     result.Passed,
					IssueCount: result.IssueCount,
					Duration:   formatDuration(result.Duration),
				}
			}

			// Output JSON
			jsonBytes, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(string(jsonBytes))
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVarP(&compactFlag, "compact", "c", false, "Output compact one-line status")
}

// formatDuration formats a duration to a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d)/float64(time.Millisecond))
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}