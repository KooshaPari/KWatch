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
	masterWatchDirs []string
	masterFormat    string
	masterWatch     bool
)

// WatchedDirectory represents a directory being monitored
type WatchedDirectory struct {
	Path     string                       `json:"path"`
	Name     string                       `json:"name"`
	Commands map[string]DirectoryCommand `json:"commands"`
	LastRun  time.Time                   `json:"last_run"`
	Error    string                      `json:"error,omitempty"`
}

// DirectoryCommand represents a command result for a directory
type DirectoryCommand struct {
	Passed     bool          `json:"passed"`
	IssueCount int           `json:"issue_count"`
	Duration   time.Duration `json:"duration"`
	LastRun    time.Time     `json:"last_run"`
	Error      string        `json:"error,omitempty"`
}

// MasterStatus represents the overall status of all watched directories
type MasterStatus struct {
	Timestamp    time.Time                   `json:"timestamp"`
	Directories  map[string]WatchedDirectory `json:"directories"`
	Summary      MasterSummary               `json:"summary"`
}

// MasterSummary provides overall statistics
type MasterSummary struct {
	TotalDirectories int `json:"total_directories"`
	PassedDirectories int `json:"passed_directories"`
	FailedDirectories int `json:"failed_directories"`
	TotalCommands    int `json:"total_commands"`
	PassedCommands   int `json:"passed_commands"`
	FailedCommands   int `json:"failed_commands"`
}

var masterCmd = &cobra.Command{
	Use:   "master [directories...]",
	Short: "Master KWatch - monitor multiple directories from one interface",
	Long: `Master KWatch allows you to monitor multiple project directories from a single interface.
	
This command provides a consolidated view of all your projects with a matrix display showing
the status of each command (TypeScript, Lint, Test, GitHub Actions) across all directories.

Examples:
  kwatch master                                    # Auto-discover projects
  kwatch master /path/to/proj1 /path/to/proj2     # Monitor specific directories
  kwatch master --format matrix                   # Matrix format output
  kwatch master --watch                           # Continuous monitoring mode
  kwatch master --format json                     # JSON output for automation`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var dirs []string
		
		if len(args) > 0 {
			// Use provided directories
			dirs = args
		} else {
			// Auto-discover directories with kwatch configurations
			discovered, err := discoverKWatchDirectories()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error discovering directories: %v\n", err)
				os.Exit(1)
			}
			dirs = discovered
		}
		
		if len(dirs) == 0 {
			fmt.Fprintf(os.Stderr, "No directories found to monitor\n")
			fmt.Fprintf(os.Stderr, "Use 'kwatch master /path/to/project1 /path/to/project2' to specify directories\n")
			os.Exit(1)
		}
		
		if masterWatch {
			// Continuous monitoring mode
			runMasterWatch(dirs)
		} else {
			// Single run mode
			runMasterSingle(dirs)
		}
	},
}

func init() {
	rootCmd.AddCommand(masterCmd)
	masterCmd.Flags().StringSliceVarP(&masterWatchDirs, "dirs", "D", nil, "Additional directories to monitor")
	masterCmd.Flags().StringVarP(&masterFormat, "format", "f", "matrix", "Output format (matrix, json, compact)")
	masterCmd.Flags().BoolVarP(&masterWatch, "watch", "w", false, "Continuous monitoring mode")
}

// discoverKWatchDirectories finds directories with kwatch configurations
func discoverKWatchDirectories() ([]string, error) {
	var dirs []string
	
	// Start from current directory and look for subdirectories with .kwatch
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}
		
		// Skip hidden directories and common non-project directories
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || 
			info.Name() == "node_modules" || info.Name() == "vendor") {
			return filepath.SkipDir
		}
		
		// Check if this directory has kwatch config
		if info.IsDir() && config.ConfigExists(path) {
			absPath, err := filepath.Abs(path)
			if err == nil {
				dirs = append(dirs, absPath)
			}
		}
		
		return nil
	})
	
	return dirs, err
}

// runMasterSingle runs a single scan of all directories
func runMasterSingle(dirs []string) {
	status := scanDirectories(dirs)
	
	switch masterFormat {
	case "json":
		outputMasterJSON(status)
	case "compact":
		outputMasterCompact(status)
	case "matrix":
		outputMasterMatrix(status)
	default:
		outputMasterMatrix(status)
	}
}

// runMasterWatch runs continuous monitoring
func runMasterWatch(dirs []string) {
	fmt.Printf("Master KWatch - Monitoring %d directories\n", len(dirs))
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	// Initial scan
	runMasterSingle(dirs)
	
	for {
		select {
		case <-ticker.C:
			fmt.Print("\033[H\033[2J") // Clear screen
			fmt.Printf("Master KWatch - Last updated: %s\n", time.Now().Format("15:04:05"))
			runMasterSingle(dirs)
		}
	}
}

// scanDirectories scans all directories and returns consolidated status
func scanDirectories(dirs []string) MasterStatus {
	status := MasterStatus{
		Timestamp:   time.Now(),
		Directories: make(map[string]WatchedDirectory),
	}
	
	ctx := context.Background()
	
	for _, dir := range dirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		
		dirName := filepath.Base(absDir)
		watched := WatchedDirectory{
			Path:     absDir,
			Name:     dirName,
			Commands: make(map[string]DirectoryCommand),
			LastRun:  time.Now(),
		}
		
		// Load configuration
		kwatchConfig, err := config.Load(absDir)
		if err != nil {
			watched.Error = err.Error()
			status.Directories[dirName] = watched
			continue
		}
		
		// Create runner
		runnerConfig := runner.RunnerConfig{
			DefaultTimeout: 30 * time.Second,
			MaxParallel:    kwatchConfig.MaxParallel,
			WorkingDir:     absDir,
		}
		
		r := runner.NewRunner(runnerConfig, kwatchConfig)
		
		// Run all commands
		results := r.RunAll(ctx)
		
		// Convert results to directory commands
		cmdNames := map[runner.CommandType]string{
			runner.TypescriptCheck: "tsc",
			runner.LintCheck:       "lint",
			runner.TestRunner:      "test",
			runner.GitHubActions:   "github",
		}
		
		for cmdType, result := range results {
			cmdName := cmdNames[cmdType]
			if cmdName == "" {
				cmdName = string(cmdType)
			}
			
			watched.Commands[cmdName] = DirectoryCommand{
				Passed:     result.Passed,
				IssueCount: result.IssueCount,
				Duration:   result.Duration,
				LastRun:    result.Timestamp,
				Error:      result.Error,
			}
		}
		
		status.Directories[dirName] = watched
	}
	
	// Calculate summary
	status.Summary = calculateMasterSummary(status.Directories)
	
	return status
}

// calculateMasterSummary calculates overall statistics
func calculateMasterSummary(directories map[string]WatchedDirectory) MasterSummary {
	summary := MasterSummary{}
	
	summary.TotalDirectories = len(directories)
	
	for _, dir := range directories {
		if dir.Error != "" {
			summary.FailedDirectories++
			continue
		}
		
		dirPassed := true
		for _, cmd := range dir.Commands {
			summary.TotalCommands++
			if cmd.Passed {
				summary.PassedCommands++
			} else {
				summary.FailedCommands++
				dirPassed = false
			}
		}
		
		if dirPassed {
			summary.PassedDirectories++
		} else {
			summary.FailedDirectories++
		}
	}
	
	return summary
}

// outputMasterJSON outputs the status as JSON
func outputMasterJSON(status MasterStatus) {
	jsonBytes, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

// outputMasterCompact outputs a compact one-line status for each directory
func outputMasterCompact(status MasterStatus) {
	// Sort directories by name
	var dirNames []string
	for name := range status.Directories {
		dirNames = append(dirNames, name)
	}
	sort.Strings(dirNames)
	
	for _, name := range dirNames {
		dir := status.Directories[name]
		if dir.Error != "" {
			fmt.Printf("%s: ERROR - %s\n", name, dir.Error)
			continue
		}
		
		var parts []string
		cmdOrder := []string{"tsc", "lint", "test", "github"}
		
		for _, cmdName := range cmdOrder {
			if cmd, exists := dir.Commands[cmdName]; exists {
				symbol := "✓"
				if !cmd.Passed {
					symbol = "✗"
				}
				parts = append(parts, fmt.Sprintf("%s:%s", strings.ToUpper(cmdName), symbol))
			}
		}
		
		fmt.Printf("%s: %s\n", name, strings.Join(parts, " "))
	}
}

// outputMasterMatrix outputs a matrix view as requested
func outputMasterMatrix(status MasterStatus) {
	// Sort directories by name
	var dirNames []string
	for name := range status.Directories {
		dirNames = append(dirNames, name)
	}
	sort.Strings(dirNames)
	
	// Determine which commands exist across all directories
	allCommands := make(map[string]bool)
	for _, dir := range status.Directories {
		for cmdName := range dir.Commands {
			allCommands[cmdName] = true
		}
	}
	
	// Command order preference
	cmdOrder := []string{"tsc", "lint", "test", "github"}
	var commands []string
	for _, cmd := range cmdOrder {
		if allCommands[cmd] {
			commands = append(commands, cmd)
		}
	}
	
	// Add any other commands not in the preferred order
	for cmd := range allCommands {
		found := false
		for _, orderedCmd := range commands {
			if cmd == orderedCmd {
				found = true
				break
			}
		}
		if !found {
			commands = append(commands, cmd)
		}
	}
	
	// Print header
	fmt.Printf("Master KWatch Status - %s\n", status.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Directories: %d | Commands: %d | Passed: %d | Failed: %d\n", 
		status.Summary.TotalDirectories, 
		status.Summary.TotalCommands,
		status.Summary.PassedCommands, 
		status.Summary.FailedCommands)
	fmt.Println()
	
	// Print matrix header
	fmt.Printf("%-20s", "DIRECTORY")
	for _, cmd := range commands {
		fmt.Printf("%-12s", strings.ToUpper(cmd))
	}
	fmt.Println()
	
	// Print separator
	fmt.Printf("%-20s", strings.Repeat("-", 20))
	for range commands {
		fmt.Printf("%-12s", strings.Repeat("-", 12))
	}
	fmt.Println()
	
	// Print each directory row
	for _, name := range dirNames {
		dir := status.Directories[name]
		
		// Directory name (truncated if needed)
		dirDisplay := name
		if len(dirDisplay) > 18 {
			dirDisplay = dirDisplay[:15] + "..."
		}
		fmt.Printf("%-20s", dirDisplay)
		
		if dir.Error != "" {
			// Show error for all commands
			for range commands {
				fmt.Printf("%-12s", "ERROR")
			}
		} else {
			// Show status for each command
			for _, cmdName := range commands {
				if cmd, exists := dir.Commands[cmdName]; exists {
					var status string
					if cmd.Passed {
						if cmd.IssueCount == 0 {
							status = "✓"
						} else {
							status = fmt.Sprintf("✓(%d)", cmd.IssueCount)
						}
					} else {
						if cmd.Error != "" {
							status = "ERR"
						} else {
							status = fmt.Sprintf("✗(%d)", cmd.IssueCount)
						}
					}
					fmt.Printf("%-12s", status)
				} else {
					fmt.Printf("%-12s", "-")
				}
			}
		}
		fmt.Println()
	}
	
	fmt.Println()
	fmt.Printf("Legend: ✓ = Passed, ✗ = Failed, ERR = Error, (-) = Not applicable\n")
	fmt.Printf("Numbers in parentheses show issue count\n")
}