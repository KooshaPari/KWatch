package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"kwatch/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp [directory]",
	Short: "Start Model Context Protocol (MCP) server",
	Long: `Start kwatch as an MCP server providing build monitoring capabilities to AI assistants.

The MCP server exposes kwatch's build monitoring functionality through the Model Context Protocol,
allowing AI assistants like Claude to monitor project build status, run commands, and access history.

The server communicates over stdio using JSON-RPC 2.0, making it compatible with standard MCP clients.

Available MCP Tools:
- get_build_status: Get current build status (TypeScript, linting, tests)
- run_commands: Execute build commands manually
- get_command_history: Get command execution history

Examples:
  kwatch mcp                           # Start MCP server for current directory
  kwatch mcp /path/to/project          # Start MCP server for specific directory
  kwatch --dir /path/to/project mcp    # Start MCP server using flag

Configuration:
Add this to your MCP client config (e.g., Claude Desktop):
{
  "kwatch": {
    "command": "kwatch",
    "args": ["mcp", "/path/to/your/project"]
  }
}`,
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

		// Log to stderr since stdout is used for MCP communication
		fmt.Fprintf(os.Stderr, "Starting KWatch MCP server for directory: %s\n", absDir)
		fmt.Fprintf(os.Stderr, "Protocol: JSON-RPC 2.0 over stdio\n")
		fmt.Fprintf(os.Stderr, "Available tools: get_build_status, run_commands, get_command_history\n")

		// Create and start MCP server
		server := mcp.NewMCPServer(absDir)
		
		// Start server (blocks until stdin closes)
		if err := server.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}