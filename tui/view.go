package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"kwatch/runner"
)

// View renders the main view
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	
	switch m.viewMode {
	case ViewMain:
		return m.renderMainView()
	case ViewHistory:
		return m.renderHistoryView()
	case ViewLogs:
		return m.renderLogsView()
	case ViewHelp:
		return m.renderHelpView()
	default:
		return m.renderMainView()
	}
}

// renderMainView renders the main dashboard view
func (m Model) renderMainView() string {
	header := m.renderHeader()
	commandTable := m.renderCommandTable()
	logsPanel := m.renderLogsPanel()
	statusBar := m.renderStatusBar()
	
	// Calculate available height for panels - ensure minimum space for header
	availableHeight := max(10, m.height - headerHeight - statusBarHeight - 4) // margins
	tableHeight := min(8, max(4, availableHeight/2))
	logsHeight := max(3, availableHeight - tableHeight)
	
	// Ensure we don't exceed available space
	if tableHeight + logsHeight > availableHeight {
		tableHeight = min(6, availableHeight/2)
		logsHeight = availableHeight - tableHeight
	}
	
	// Render panels with calculated heights
	commandTableStyled := panelStyle.Width(m.width - 4).Height(tableHeight).Render(commandTable)
	logsPanelStyled := panelStyle.Width(m.width - 4).Height(logsHeight).Render(logsPanel)
	
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		commandTableStyled,
		logsPanelStyled,
		statusBar,
	)
}

// renderHistoryView renders the command history view
func (m Model) renderHistoryView() string {
	header := m.renderHeader()
	historyTable := m.renderHistoryTable()
	statusBar := m.renderStatusBar()
	
	availableHeight := m.height - headerHeight - statusBarHeight - 2
	historyStyled := panelStyle.Width(m.width - 4).Height(availableHeight).Render(historyTable)
	
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		historyStyled,
		statusBar,
	)
}

// renderLogsView renders the logs view
func (m Model) renderLogsView() string {
	header := m.renderHeader()
	logsTable := m.renderDetailedLogs()
	statusBar := m.renderStatusBar()
	
	availableHeight := m.height - headerHeight - statusBarHeight - 2
	logsStyled := panelStyle.Width(m.width - 4).Height(availableHeight).Render(logsTable)
	
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		logsStyled,
		statusBar,
	)
}

// renderHelpView renders the help view
func (m Model) renderHelpView() string {
	header := m.renderHeader()
	help := m.renderHelp()
	statusBar := m.renderStatusBar()
	
	availableHeight := m.height - headerHeight - statusBarHeight - 2
	helpStyled := helpStyle.Width(m.width - 4).Height(availableHeight).Render(help)
	
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		helpStyled,
		statusBar,
	)
}

// renderHeader renders the main header
func (m Model) renderHeader() string {
	title := "KWatch - Development Monitor"
	
	// Current view indicator
	viewIndicator := ""
	switch m.viewMode {
	case ViewMain:
		viewIndicator = "Main"
	case ViewHistory:
		viewIndicator = "History"
	case ViewLogs:
		viewIndicator = "Logs"
	case ViewHelp:
		viewIndicator = "Help"
	}
	
	// Directory info
	dirInfo := fmt.Sprintf("Watching: %s", Truncate(m.watchDir, 40))
	
	// Status summary
	statusSummary := m.GetStatusSummary()
	
	// Error metrics
	errorCount, errorFiles := m.GetErrorMetrics()
	errorDisplay := ""
	if errorCount > 0 {
		if errorFiles > 0 {
			errorDisplay = fmt.Sprintf(" | %s %d err, %d files", 
				statusFailStyle.Render("✗"), errorCount, errorFiles)
		} else {
			errorDisplay = fmt.Sprintf(" | %s %d errors", 
				statusFailStyle.Render("✗"), errorCount)
		}
	}
	
	headerLeft := lipgloss.JoinHorizontal(lipgloss.Left,
		headerStyle.Render(title),
		normalTextStyle.Render(" | "),
		highlightStyle.Render(viewIndicator),
	)
	
	headerRight := lipgloss.JoinHorizontal(lipgloss.Right,
		dimTextStyle.Render(dirInfo),
		normalTextStyle.Render(" | "),
		GetStatusStyle(statusSummary == "All Passed", m.IsRunning()).Render(statusSummary),
		errorDisplay,
	)
	
	// Center the header content
	headerContent := lipgloss.JoinHorizontal(lipgloss.Left,
		headerLeft,
		strings.Repeat(" ", max(0, m.width-lipgloss.Width(headerLeft)-lipgloss.Width(headerRight))),
		headerRight,
	)
	
	return headerStyle.Width(m.width).Render(headerContent)
}

// renderCommandTable renders the command status table
func (m Model) renderCommandTable() string {
	statuses := m.GetCurrentCommandStatuses()
	
	// Table header
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		tableHeaderStyle.Width(20).Render("Command"),
		tableHeaderStyle.Width(12).Render("Status"),
		tableHeaderStyle.Width(12).Render("Duration"),
		tableHeaderStyle.Width(20).Render("Last Run"),
		tableHeaderStyle.Width(12).Render("Count"),
	)
	
	// Table rows
	rows := make([]string, len(statuses))
	for i, status := range statuses {
		// Command name
		cmdName := string(status.Type)
		cmdStyle := GetCommandStyle(cmdName)
		
		// Status
		statusText := "Not Run"
		statusStyle := dimTextStyle
		if status.Running {
			statusText = "Running " + GetStatusIcon(false, true)
			statusStyle = GetStatusStyle(false, true)
		} else if status.Result != nil {
			statusText = GetStatusIcon(status.Result.Passed, false)
			if status.Result.Passed {
				statusText += " Passed"
			} else {
				statusText += " Failed"
			}
			statusStyle = GetStatusStyle(status.Result.Passed, false)
		}
		
		// Duration
		duration := "-"
		if status.Result != nil {
			duration = FormatDuration(status.Result.Duration.Milliseconds())
		}
		
		// Last run
		lastRun := "-"
		if !status.LastRun.IsZero() {
			lastRun = status.LastRun.Format("15:04:05")
		}
		
		// Count - show test results as PASS/TOTAL or errors/files
		count := "-"
		if status.Result != nil {
			if status.Type == runner.TestRunner {
				// For tests, show PASS/TOTAL format
				if status.Result.TotalTests > 0 {
					count = fmt.Sprintf("%d/%d", status.Result.PassedTests, status.Result.TotalTests)
				} else if status.Result.IssueCount > 0 {
					// Fallback for old test format
					count = fmt.Sprintf("%d", status.Result.IssueCount)
				} else {
					count = "0"
				}
			} else {
				// For other commands, show errors/files
				if status.Result.IssueCount > 0 {
					if status.Result.FileCount > 0 {
						count = fmt.Sprintf("%d/%d", status.Result.IssueCount, status.Result.FileCount)
					} else {
						count = fmt.Sprintf("%d", status.Result.IssueCount)
					}
				} else {
					count = "0"
				}
			}
		}
		
		// Row style
		rowStyle := tableCellStyle
		if m.viewMode == ViewMain && i == m.selectedRow {
			rowStyle = selectedRowStyle
		}
		
		row := lipgloss.JoinHorizontal(lipgloss.Left,
			rowStyle.Width(20).Render(cmdStyle.Render(cmdName)),
			rowStyle.Width(12).Render(statusStyle.Render(statusText)),
			rowStyle.Width(12).Render(duration),
			rowStyle.Width(20).Render(lastRun),
			rowStyle.Width(12).Render(count),
		)
		
		rows[i] = row
	}
	
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.JoinVertical(lipgloss.Left, rows...),
	)
}

// renderLogsPanel renders the activity logs panel
func (m Model) renderLogsPanel() string {
	// Calculate max logs to show based on available height
	maxLogsToShow := min(8, max(3, (m.height-headerHeight-statusBarHeight-8)/2))
	logs := m.GetRecentLogs(maxLogsToShow)
	
	if len(logs) == 0 {
		return dimTextStyle.Render("No activity yet...")
	}
	
	// Calculate available width for messages (subtract timestamp and separator)
	availableWidth := max(40, m.width-20)
	
	logLines := make([]string, len(logs))
	for i, log := range logs {
		timestamp := logTimestampStyle.Render(log.Timestamp.Format("15:04:05"))
		
		var message string
		switch log.Type {
		case LogFileChange:
			fileName := Truncate(log.File, 30)
			message = lipgloss.JoinHorizontal(lipgloss.Left,
				logFileStyle.Render(fileName),
				normalTextStyle.Render(" - "),
				logActionStyle.Render(log.Action),
			)
		case LogCommandStart, LogCommandEnd:
			truncatedMessage := Truncate(log.Message, availableWidth-15)
			message = lipgloss.JoinHorizontal(lipgloss.Left,
				GetCommandStyle(log.Action).Render(log.Action),
				normalTextStyle.Render(" - "),
				normalTextStyle.Render(truncatedMessage),
			)
		case LogError:
			truncatedMessage := Truncate(log.Message, availableWidth-5)
			message = statusFailStyle.Render(truncatedMessage)
		case LogWarning:
			truncatedMessage := Truncate(log.Message, availableWidth-5)
			message = statusRunningStyle.Render(truncatedMessage)
		default:
			truncatedMessage := Truncate(log.Message, availableWidth-5)
			message = normalTextStyle.Render(truncatedMessage)
		}
		
		logLines[i] = lipgloss.JoinHorizontal(lipgloss.Left,
			timestamp,
			normalTextStyle.Render(" │ "),
			message,
		)
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, logLines...)
}

// renderHistoryTable renders the command history table
func (m Model) renderHistoryTable() string {
	history := m.GetHistoryForView()
	
	if len(history) == 0 {
		return dimTextStyle.Render("No command history available...")
	}
	
	// Table header
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		tableHeaderStyle.Width(16).Render("Command"),
		tableHeaderStyle.Width(12).Render("Status"),
		tableHeaderStyle.Width(12).Render("Duration"),
		tableHeaderStyle.Width(20).Render("Timestamp"),
		tableHeaderStyle.Width(20).Render("Output"),
	)
	
	// Limit to visible rows
	visibleRows := min(len(history), m.height-10)
	rows := make([]string, visibleRows)
	
	for i := 0; i < visibleRows; i++ {
		result := history[i]
		
		// Command
		cmdType := getCommandType(result.Command)
		cmdStyle := GetCommandStyle(string(cmdType))
		
		// Status
		statusText := GetStatusIcon(result.Passed, false)
		if result.Passed {
			statusText += " Passed"
		} else {
			statusText += " Failed"
		}
		statusStyle := GetStatusStyle(result.Passed, false)
		
		// Duration
		duration := FormatDuration(result.Duration.Milliseconds())
		
		// Timestamp
		timestamp := result.Timestamp.Format("15:04:05")
		
		// Output (truncated)
		output := strings.ReplaceAll(result.Output, "\n", " ")
		output = Truncate(output, 18)
		
		// Row style
		rowStyle := tableCellStyle
		if m.viewMode == ViewHistory && i == m.selectedRow {
			rowStyle = selectedRowStyle
		}
		
		row := lipgloss.JoinHorizontal(lipgloss.Left,
			rowStyle.Width(16).Render(cmdStyle.Render(string(cmdType))),
			rowStyle.Width(12).Render(statusStyle.Render(statusText)),
			rowStyle.Width(12).Render(duration),
			rowStyle.Width(20).Render(timestamp),
			rowStyle.Width(20).Render(output),
		)
		
		rows[i] = row
	}
	
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.JoinVertical(lipgloss.Left, rows...),
	)
}

// renderDetailedLogs renders detailed logs view
func (m Model) renderDetailedLogs() string {
	if len(m.logs) == 0 {
		return dimTextStyle.Render("No logs available...")
	}
	
	// Show logs in reverse order (newest first) - limit to prevent UI overflow
	maxVisibleLogs := min(len(m.logs), min(30, m.height-8))
	visibleLogs := maxVisibleLogs
	startIdx := max(0, len(m.logs)-visibleLogs)
	
	logLines := make([]string, visibleLogs)
	for i := 0; i < visibleLogs; i++ {
		logIdx := startIdx + i
		log := m.logs[logIdx]
		
		timestamp := logTimestampStyle.Render(log.Timestamp.Format("15:04:05.000"))
		
		typeStr := ""
		switch log.Type {
		case LogInfo:
			typeStr = normalTextStyle.Render("INFO")
		case LogWarning:
			typeStr = statusRunningStyle.Render("WARN")
		case LogError:
			typeStr = statusFailStyle.Render("ERROR")
		case LogFileChange:
			typeStr = logFileStyle.Render("FILE")
		case LogCommandStart:
			typeStr = GetCommandStyle(log.Action).Render("START")
		case LogCommandEnd:
			typeStr = GetCommandStyle(log.Action).Render("END")
		}
		
		message := log.Message
		if log.File != "" {
			message = fmt.Sprintf("%s: %s", log.File, message)
		}
		
		line := lipgloss.JoinHorizontal(lipgloss.Left,
			timestamp,
			normalTextStyle.Render(" │ "),
			typeStr,
			normalTextStyle.Render(" │ "),
			normalTextStyle.Render(message),
		)
		
		logLines[i] = line
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, logLines...)
}

// renderHelp renders the help view
func (m Model) renderHelp() string {
	helpText := []string{
		helpKeyStyle.Render("KWATCH - Development Monitor Help"),
		"",
		lipgloss.JoinHorizontal(lipgloss.Left, helpKeyStyle.Render("q"), helpDescStyle.Render("           Quit application")),
		lipgloss.JoinHorizontal(lipgloss.Left, helpKeyStyle.Render("r"), helpDescStyle.Render("           Refresh / Manual run")),
		lipgloss.JoinHorizontal(lipgloss.Left, helpKeyStyle.Render("s"), helpDescStyle.Render("           Show status")),
		lipgloss.JoinHorizontal(lipgloss.Left, helpKeyStyle.Render("h"), helpDescStyle.Render("           Show this help")),
		lipgloss.JoinHorizontal(lipgloss.Left, helpKeyStyle.Render("1"), helpDescStyle.Render("           Main view")),
		lipgloss.JoinHorizontal(lipgloss.Left, helpKeyStyle.Render("2"), helpDescStyle.Render("           History view")),
		lipgloss.JoinHorizontal(lipgloss.Left, helpKeyStyle.Render("3"), helpDescStyle.Render("           Logs view")),
		lipgloss.JoinHorizontal(lipgloss.Left, helpKeyStyle.Render("↑/↓"), helpDescStyle.Render("         Navigate up/down")),
		lipgloss.JoinHorizontal(lipgloss.Left, helpKeyStyle.Render("Enter"), helpDescStyle.Render("       View details")),
		lipgloss.JoinHorizontal(lipgloss.Left, helpKeyStyle.Render("Esc"), helpDescStyle.Render("         Back to main view")),
		"",
		helpDescStyle.Render("COMMANDS:"),
		lipgloss.JoinHorizontal(lipgloss.Left, commandTSCStyle.Render("typescript"), helpDescStyle.Render("   TypeScript compilation check")),
		lipgloss.JoinHorizontal(lipgloss.Left, commandLintStyle.Render("lint"), helpDescStyle.Render("        Code linting and formatting")),
		lipgloss.JoinHorizontal(lipgloss.Left, commandTestStyle.Render("test"), helpDescStyle.Render("        Test suite execution")),
		"",
		helpDescStyle.Render("STATUS INDICATORS:"),
		lipgloss.JoinHorizontal(lipgloss.Left, statusPassStyle.Render("✓"), helpDescStyle.Render("           Passed")),
		lipgloss.JoinHorizontal(lipgloss.Left, statusFailStyle.Render("✗"), helpDescStyle.Render("           Failed")),
		lipgloss.JoinHorizontal(lipgloss.Left, statusRunningStyle.Render("⟳"), helpDescStyle.Render("           Running")),
		"",
		helpDescStyle.Render("The monitor watches your project files and automatically runs"),
		helpDescStyle.Render("the configured commands when changes are detected."),
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, helpText...)
}

// renderStatusBar renders the bottom status bar
func (m Model) renderStatusBar() string {
	left := ""
	if m.watcherActive {
		left = statusPassStyle.Render("● Watching")
	} else {
		left = statusFailStyle.Render("● Stopped")
	}
	
	if m.serverActive {
		left += normalTextStyle.Render(" | ") + statusPassStyle.Render("● Server")
	} else {
		left += normalTextStyle.Render(" | ") + dimTextStyle.Render("● Server N/A")
	}
	
	// View navigation hints
	center := dimTextStyle.Render("1:Main 2:History 3:Logs h:Help q:Quit")
	
	// Current time
	right := dimTextStyle.Render(time.Now().Format("15:04:05"))
	
	// Error display
	if m.HasError() {
		right = statusFailStyle.Render("Error: " + m.GetError())
	}
	
	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	centerWidth := lipgloss.Width(center)
	rightWidth := lipgloss.Width(right)
	
	totalContentWidth := leftWidth + centerWidth + rightWidth
	if totalContentWidth >= m.width {
		// Truncate center if needed
		availableCenter := m.width - leftWidth - rightWidth - 2
		if availableCenter > 0 {
			center = dimTextStyle.Render(Truncate(center, availableCenter))
		} else {
			center = ""
		}
	}
	
	// Calculate final spacing
	centerWidth = lipgloss.Width(center)
	spacing1 := (m.width - leftWidth - centerWidth - rightWidth) / 2
	spacing2 := m.width - leftWidth - centerWidth - rightWidth - spacing1
	
	statusContent := lipgloss.JoinHorizontal(lipgloss.Left,
		left,
		strings.Repeat(" ", max(0, spacing1)),
		center,
		strings.Repeat(" ", max(0, spacing2)),
		right,
	)
	
	return statusBarStyle.Width(m.width).Render(statusContent)
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}