package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"kwatch/config"
	"kwatch/runner"
)

var (
	historyLimit  int
	historyFormat string
	historyFilter string
)

// historyResponse represents the JSON response for history command
type historyResponse struct {
	Directory string                `json:"directory"`
	Count     int                   `json:"count"`
	History   []runner.CommandResult `json:"history"`
}

var historyCmd = &cobra.Command{
	Use:   "history [directory]",
	Short: "Show command execution history",
	Long: `Show the history of command executions for the project.

The history includes all previous runs with timestamps, durations, and results.
You can filter by command type and limit the number of results shown.

Examples:
  kwatch history                           # Show all history
  kwatch history --limit 10                # Show last 10 entries
  kwatch history --filter tsc              # Show only TypeScript check history
  kwatch --dir /path/to/project history    # Show history for specific directory (flag)
  kwatch . history                         # Show history for current directory
  kwatch history --format table           # Show in table format
  kwatch history --format json            # Show in JSON format`,
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
		
		// For this demo, we'll run once to populate history
		// In a real implementation, history would be persistent
		ctx := context.Background()
		r.RunAll(ctx)

		// Get history
		history := r.GetHistory()

		// Filter history if requested
		if historyFilter != "" {
			history = filterHistory(history, historyFilter)
		}

		// Sort by timestamp (newest first)
		sort.Slice(history, func(i, j int) bool {
			return history[i].Timestamp.After(history[j].Timestamp)
		})

		// Apply limit
		if historyLimit > 0 && len(history) > historyLimit {
			history = history[:historyLimit]
		}

		// Output based on format
		switch historyFormat {
		case "json":
			outputHistoryJSON(absDir, history)
		case "table":
			outputHistoryTable(history)
		default:
			outputHistoryDefault(history)
		}
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "l", 0, "Limit number of history entries (0 for all)")
	historyCmd.Flags().StringVarP(&historyFormat, "format", "f", "default", "Output format (default, json, table)")
	historyCmd.Flags().StringVar(&historyFilter, "filter", "", "Filter by command type (tsc, lint, test)")
}

// filterHistory filters history entries by command type
func filterHistory(history []runner.CommandResult, filter string) []runner.CommandResult {
	var filtered []runner.CommandResult
	
	for _, entry := range history {
		// Match command type
		switch filter {
		case "tsc", "typescript":
			if strings.Contains(entry.Command, "tsc") {
				filtered = append(filtered, entry)
			}
		case "lint", "eslint":
			if strings.Contains(entry.Command, "lint") || strings.Contains(entry.Command, "eslint") {
				filtered = append(filtered, entry)
			}
		case "test":
			if strings.Contains(entry.Command, "test") {
				filtered = append(filtered, entry)
			}
		default:
			if strings.Contains(strings.ToLower(entry.Command), strings.ToLower(filter)) {
				filtered = append(filtered, entry)
			}
		}
	}
	
	return filtered
}

// outputHistoryJSON outputs history in JSON format
func outputHistoryJSON(directory string, history []runner.CommandResult) {
	response := historyResponse{
		Directory: directory,
		Count:     len(history),
		History:   history,
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonBytes))
}

// outputHistoryTable outputs history in table format
func outputHistoryTable(history []runner.CommandResult) {
	if len(history) == 0 {
		fmt.Println("No history entries found.")
		return
	}

	// Table header
	fmt.Printf("%-20s %-10s %-8s %-8s %-10s %s\n", "TIMESTAMP", "COMMAND", "PASSED", "ISSUES", "DURATION", "ERROR")
	fmt.Println(strings.Repeat("-", 80))

	// Table rows
	for _, entry := range history {
		timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
		command := getCommandTypeLabel(entry.Command)
		passed := "✓"
		if !entry.Passed {
			passed = "✗"
		}
		duration := formatDuration(entry.Duration)
		errorMsg := ""
		if entry.Error != "" {
			errorMsg = truncateString(entry.Error, 30)
		}

		fmt.Printf("%-20s %-10s %-8s %-8d %-10s %s\n", 
			timestamp, command, passed, entry.IssueCount, duration, errorMsg)
	}
}

// outputHistoryDefault outputs history in default format
func outputHistoryDefault(history []runner.CommandResult) {
	if len(history) == 0 {
		fmt.Println("No history entries found.")
		return
	}

	fmt.Printf("Command History (%d entries):\n\n", len(history))

	for i, entry := range history {
		status := "PASSED"
		if !entry.Passed {
			status = "FAILED"
		}

		fmt.Printf("%d. %s - %s (%s)\n", i+1, 
			getCommandTypeLabel(entry.Command), 
			status, 
			entry.Timestamp.Format("2006-01-02 15:04:05"))
		
		if entry.IssueCount > 0 {
			fmt.Printf("   Issues: %d\n", entry.IssueCount)
		}
		
		fmt.Printf("   Duration: %s\n", formatDuration(entry.Duration))
		
		if entry.Error != "" {
			fmt.Printf("   Error: %s\n", truncateString(entry.Error, 100))
		}
		
		fmt.Println()
	}
}

// getCommandTypeLabel returns a human-readable label for a command
func getCommandTypeLabel(command string) string {
	switch {
	case strings.Contains(command, "tsc"):
		return "TypeScript"
	case strings.Contains(command, "lint") || strings.Contains(command, "eslint"):
		return "Lint"
	case strings.Contains(command, "test"):
		return "Test"
	case strings.Contains(command, "github"):
		return "GitHub Actions"
	default:
		return command
	}
}

// truncateString truncates a string to the specified length
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}