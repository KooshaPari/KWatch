package tui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"kwatch/runner"
)

// Messages for the update loop
type (
	// Window size message
	windowSizeMsg struct {
		width  int
		height int
	}
	
	// Tick message for regular updates
	tickMsg time.Time
	
	// Command result message
	commandResultMsg struct {
		result runner.CommandResult
	}
	
	// Command start message
	commandStartMsg struct {
		cmdType runner.CommandType
	}
	
	// File change message
	fileChangeMsg struct {
		file   string
		action string
	}
	
	// Status update message
	statusUpdateMsg struct {
		watcherActive bool
		serverActive  bool
	}
	
	// Error message
	errorMsg struct {
		err string
	}
	
	// Refresh message
	refreshMsg struct{}
)

// Update handles all messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	
	// Handle window size changes
	case tea.WindowSizeMsg:
		m.UpdateSize(msg.Width, msg.Height)
		return m, nil
	
	// Handle keyboard input
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	
	// Handle regular tick updates
	case tickMsg:
		return m, tick()
	
	// Handle command results
	case commandResultMsg:
		m.AddCommandResult(msg.result)
		return m, nil
	
	// Handle command start
	case commandStartMsg:
		m.SetCommandRunning(msg.cmdType, true)
		return m, nil
	
	// Handle file changes
	case fileChangeMsg:
		m.AddLog(LogFileChange, "File changed", msg.file, msg.action)
		// Only run commands if not already running
		if !m.IsAnyCommandRunning() {
			return m, m.runCommandsOnChange()
		}
		return m, nil
	
	// Handle status updates
	case statusUpdateMsg:
		m.SetWatcherActive(msg.watcherActive)
		m.SetServerActive(msg.serverActive)
		return m, nil
	
	// Handle errors
	case errorMsg:
		m.SetError(msg.err)
		return m, nil
	
	// Handle refresh
	case refreshMsg:
		m.ClearError()
		return m, m.runAllCommands()
	
	default:
		return m, nil
	}
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	
	// Quit
	case "q", "ctrl+c":
		return m, tea.Quit
	
	// Refresh / Manual run
	case "r":
		m.ClearError()
		m.AddLog(LogInfo, "Manual refresh triggered", "", "refresh")
		return m, m.runAllCommands()
	
	// Show status
	case "s":
		m.AddLog(LogInfo, "Status check", "", "status")
		return m, m.checkStatus()
	
	// Show help
	case "h":
		m.viewMode = ViewHelp
		return m, nil
	
	// View navigation
	case "1":
		m.viewMode = ViewMain
		m.selectedRow = 0
		return m, nil
	
	case "2":
		m.viewMode = ViewHistory
		m.selectedRow = 0
		return m, nil
	
	case "3":
		m.viewMode = ViewLogs
		m.selectedRow = 0
		return m, nil
	
	// Navigation
	case "up", "k":
		m.NavigateUp()
		return m, nil
	
	case "down", "j":
		m.NavigateDown()
		return m, nil
	
	// Enter to view details
	case "enter":
		return m.handleEnterKey()
	
	// Escape to go back
	case "esc":
		if m.viewMode != ViewMain {
			m.viewMode = ViewMain
			m.selectedRow = 0
		}
		return m, nil
	
	// Clear error
	case "c":
		if m.HasError() {
			m.ClearError()
		}
		return m, nil
	
	default:
		return m, nil
	}
}

// handleEnterKey handles the enter key press
func (m Model) handleEnterKey() (tea.Model, tea.Cmd) {
	switch m.viewMode {
	case ViewMain:
		// Run selected command
		statuses := m.GetCurrentCommandStatuses()
		if m.selectedRow >= 0 && m.selectedRow < len(statuses) {
			cmdType := statuses[m.selectedRow].Type
			return m, m.runSpecificCommand(cmdType)
		}
	
	case ViewHistory:
		// Show detailed history for selected item
		history := m.GetHistoryForView()
		if m.selectedRow >= 0 && m.selectedRow < len(history) {
			result := history[m.selectedRow]
			m.AddLog(LogInfo, fmt.Sprintf("History details: %s", result.Output), "", "detail")
		}
	
	case ViewLogs:
		// Show log details or clear logs
		if len(m.logs) > 0 {
			m.AddLog(LogInfo, "Logs cleared", "", "clear")
			m.logs = []LogEntry{}
		}
	}
	
	return m, nil
}

// runAllCommands runs all configured commands
func (m Model) runAllCommands() tea.Cmd {
	// Mark all commands as starting
	commands := []runner.CommandType{
		runner.TypescriptCheck,
		runner.LintCheck, 
		runner.TestRunner,
	}
	
	// Create batch of start messages and command executions
	var cmds []tea.Cmd
	for _, cmdType := range commands {
		// Capture cmdType in closure properly
		ct := cmdType
		cmds = append(cmds, 
			tea.Cmd(func() tea.Msg {
				return commandStartMsg{cmdType: ct}
			}),
			m.runSpecificCommand(ct),
		)
	}
	
	return tea.Batch(cmds...)
}

// runCommandsOnChange runs commands when files change
func (m Model) runCommandsOnChange() tea.Cmd {
	// Run TypeScript check and lint on most file changes
	// Only run tests if test files changed
	return tea.Batch(
		m.runSpecificCommand(runner.TypescriptCheck),
		m.runSpecificCommand(runner.LintCheck),
	)
}

// runSpecificCommand runs a specific command type
func (m Model) runSpecificCommand(cmdType runner.CommandType) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Execute the command (don't send start message from here - it creates wrong program)
		result := executeCommand(cmdType, m.watchDir)
		
		// Send the result
		return commandResultMsg{result: result}
	})
}

// executeCommand executes a command and returns the result
func executeCommand(cmdType runner.CommandType, workDir string) runner.CommandResult {
	startTime := time.Now()
	
	// Create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	var cmd *exec.Cmd
	var cmdString string
	
	switch cmdType {
	case runner.TypescriptCheck:
		cmd = exec.CommandContext(ctx, "npx", "tsc", "--noEmit")
		cmdString = "tsc"
	case runner.LintCheck:
		cmd = exec.CommandContext(ctx, "npm", "run", "lint")
		cmdString = "lint"
	case runner.TestRunner:
		cmd = exec.CommandContext(ctx, "npm", "run", "test")
		cmdString = "test"
	default:
		return runner.CommandResult{
			Command:    string(cmdType),
			Passed:     false,
			IssueCount: 0,
			Output:     "Unknown command type",
			Duration:   0,
			Timestamp:  startTime,
			Error:      "Unknown command type",
		}
	}
	
	cmd.Dir = workDir
	
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)
	
	result := runner.CommandResult{
		Command:    cmdString,
		Passed:     err == nil,
		IssueCount: 0,
		Output:     string(output),
		Duration:   duration,
		Timestamp:  startTime,
	}
	
	if err != nil {
		result.Error = err.Error()
	}
	
	// Parse issue count and file count from output
	result.IssueCount, result.FileCount = parseIssueAndFileCount(cmdString, string(output))
	
	return result
}

// parseIssueAndFileCount extracts issue count and file count from command output
func parseIssueAndFileCount(command, output string) (int, int) {
	switch command {
	case "tsc":
		// Count "error TS" occurrences and unique files
		issueCount := 0
		fileMap := make(map[string]bool)
		lines := strings.Split(output, "\n")
		
		for _, line := range lines {
			if strings.Contains(line, "error TS") {
				issueCount++
				// Extract file path (format: "file.ts(line,col): error TS...")
				parts := strings.Split(line, "(")
				if len(parts) > 0 {
					fileName := strings.TrimSpace(parts[0])
					if fileName != "" {
						fileMap[fileName] = true
					}
				}
			}
		}
		return issueCount, len(fileMap)
		
	case "lint":
		// Parse ESLint output for problems and files
		issueCount := 0
		fileMap := make(map[string]bool)
		lines := strings.Split(output, "\n")
		
		for _, line := range lines {
			// Count individual error/warning lines
			if strings.Contains(line, "error") || strings.Contains(line, "warning") {
				// Lines like "  8:40  warning  Unexpected any..."
				parts := strings.Fields(line)
				if len(parts) >= 2 && strings.Contains(parts[1], ":") {
					issueCount++
				}
			}
			// Count files with issues (lines starting with file path)
			if strings.HasPrefix(line, "/") && strings.Contains(line, ".ts") {
				fileMap[line] = true
			}
		}
		
		// If we found issues but no files, assume 1 file
		if issueCount > 0 && len(fileMap) == 0 {
			return issueCount, 1
		}
		
		return issueCount, len(fileMap)
		
	case "test":
		// Count failed tests and test files
		issueCount := 0
		fileCount := 0
		
		if strings.Contains(output, "No tests found") {
			return 0, 0
		}
		
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "FAIL") || strings.Contains(line, "failed") {
				issueCount++
			}
			if strings.Contains(line, ".test.") || strings.Contains(line, ".spec.") {
				fileCount++
			}
		}
		
		// If we found failures but no test files, assume 1 file
		if issueCount > 0 && fileCount == 0 {
			fileCount = 1
		}
		
		return issueCount, fileCount
		
	default:
		return 0, 0
	}
}

// checkStatus checks the current status of watcher and server
func (m Model) checkStatus() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Check actual watcher and server status
		return statusUpdateMsg{
			watcherActive: m.watcherActive, // Use current watcher status
			serverActive:  m.serverActive,  // Use current server status
		}
	})
}

// tick creates a regular tick for updates
func tick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// refreshCmd creates a refresh command
func refreshCmd() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		return refreshMsg{}
	})
}

// fileWatchCmd creates a file watch command (placeholder)
func fileWatchCmd(watchDir string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// This would be replaced with actual file watching logic
		// For now, return a placeholder
		return fileChangeMsg{
			file:   "example.ts",
			action: "modified",
		}
	})
}

// startFileWatcher starts the file watcher
func (m Model) startFileWatcher() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Start file watcher
		go func() {
			// This would implement actual file watching using fsnotify
			// For now, we'll simulate periodic file changes
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			
			for {
				select {
				case <-ticker.C:
					// Simulate file change
					tea.NewProgram(nil).Send(fileChangeMsg{
						file:   "src/example.ts",
						action: "modified",
					})
				}
			}
		}()
		
		return statusUpdateMsg{
			watcherActive: true,
			serverActive:  false,
		}
	})
}

// Utility functions

// isValidCommand checks if a command is valid/available
func isValidCommand(cmdType runner.CommandType) bool {
	switch cmdType {
	case runner.TypescriptCheck:
		return commandExists("npx")
	case runner.LintCheck:
		return commandExists("npm")
	case runner.TestRunner:
		return commandExists("npm")
	default:
		return false
	}
}

// commandExists checks if a command exists in PATH
func commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// parseCommandOutput parses command output for additional information
func parseCommandOutput(cmdType runner.CommandType, output string) (count int, summary string) {
	switch cmdType {
	case runner.TypescriptCheck:
		// Parse TypeScript output
		lines := strings.Split(output, "\n")
		errorCount := 0
		for _, line := range lines {
			if strings.Contains(line, "error TS") {
				errorCount++
			}
		}
		return errorCount, fmt.Sprintf("%d errors", errorCount)
	
	case runner.LintCheck:
		// Parse ESLint output
		lines := strings.Split(output, "\n")
		problemCount := 0
		for _, line := range lines {
			if strings.Contains(line, "problem") {
				problemCount++
			}
		}
		return problemCount, fmt.Sprintf("%d problems", problemCount)
	
	case runner.TestRunner:
		// Parse test output
		lines := strings.Split(output, "\n")
		testCount := 0
		for _, line := range lines {
			if strings.Contains(line, "test") {
				testCount++
			}
		}
		return testCount, fmt.Sprintf("%d tests", testCount)
	
	default:
		return 0, "Unknown"
	}
}

// formatError formats error messages for display
func formatError(err error) string {
	if err == nil {
		return ""
	}
	
	errStr := err.Error()
	// Truncate very long errors
	if len(errStr) > 100 {
		errStr = errStr[:97] + "..."
	}
	
	return errStr
}