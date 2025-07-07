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
	if m.runner == nil {
		return nil
	}
	
	// Create individual commands for each enabled command type
	var cmds []tea.Cmd
	enabledCommands := m.kwatchConfig.GetEnabledCommands()
	
	for name := range enabledCommands {
		var cmdType runner.CommandType
		switch name {
		case "typescript":
			cmdType = runner.TypescriptCheck
		case "lint":
			cmdType = runner.LintCheck
		case "test":
			cmdType = runner.TestRunner
		default:
			cmdType = runner.CommandType(name)
		}
		
		// Create a command to run this specific command type
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
	if m.runner == nil {
		return nil
	}
	
	return tea.Cmd(func() tea.Msg {
		// Find the command configuration for this type
		enabledCommands := m.kwatchConfig.GetEnabledCommands()
		
		var configCmd *runner.Command
		for name, cmd := range enabledCommands {
			var mappedType runner.CommandType
			switch name {
			case "typescript":
				mappedType = runner.TypescriptCheck
			case "lint":
				mappedType = runner.LintCheck
			case "test":
				mappedType = runner.TestRunner
			default:
				mappedType = runner.CommandType(name)
			}
			
			if mappedType == cmdType {
				timeout := m.kwatchConfig.GetTimeout(name)
				configCmd = &runner.Command{
					Type:    cmdType,
					Command: cmd.Command,
					Args:    cmd.Args,
					Timeout: timeout,
				}
				break
			}
		}
		
		if configCmd == nil {
			// Fallback for unknown command types
			return commandResultMsg{
				result: runner.CommandResult{
					Command:   string(cmdType),
					Passed:    false,
					Output:    "Command not found in configuration",
					Timestamp: time.Now(),
					Error:     "Command not configured",
				},
			}
		}
		
		// Execute the command using the runner
		ctx := context.Background()
		result := m.runner.RunCommand(ctx, *configCmd)
		
		// Send the result
		return commandResultMsg{result: result}
	})
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