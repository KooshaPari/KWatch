package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"kwatch/tui"
)

// Global flags
var (
	globalDir string
)

var rootCmd = &cobra.Command{
	Use:   "kwatch [directory]",
	Short: "Monitor project build status with TUI panel",
	Long:  `kwatch monitors TypeScript/JavaScript projects and provides real-time build status through a TUI panel and HTTP API for AI agents.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := getWorkingDirectory(args)
		
		absDir, err := filepath.Abs(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving directory: %v\n", err)
			os.Exit(1)
		}
		
		// Start TUI panel
		startTUI(absDir)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringVarP(&globalDir, "dir", "d", "", "Directory to monitor (default: current directory)")
}

// getWorkingDirectory determines the working directory from args and flags
func getWorkingDirectory(args []string) string {
	// Priority: --dir flag > positional argument > current directory
	if globalDir != "" {
		return globalDir
	}
	if len(args) > 0 {
		return args[0]
	}
	return "."
}

func startTUI(dir string) {
	if err := tui.StartTUI(dir); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting TUI: %v\n", err)
		os.Exit(1)
	}
}