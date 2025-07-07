package tui

import (
	"fmt"
	
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	// Primary colors
	primaryColor     = lipgloss.Color("#646cff")
	secondaryColor   = lipgloss.Color("#42b883")
	accentColor      = lipgloss.Color("#ffd700")
	
	// Status colors
	successColor     = lipgloss.Color("#00ff00")
	errorColor       = lipgloss.Color("#ff0000")
	warningColor     = lipgloss.Color("#ffff00")
	runningColor     = lipgloss.Color("#ffa500")
	
	// UI colors
	borderColor      = lipgloss.Color("#444444")
	headerColor      = lipgloss.Color("#ffffff")
	textColor        = lipgloss.Color("#cccccc")
	dimTextColor     = lipgloss.Color("#888888")
	bgColor          = lipgloss.Color("#1a1a1a")
	selectedBgColor  = lipgloss.Color("#2a2a2a")
)

// Base styles
var (
	// Border styles
	borderStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2)

	// Header styles
	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(headerColor).
		Background(primaryColor).
		Padding(0, 1).
		MarginBottom(1)

	// Status styles
	statusPassStyle = lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true)

	statusFailStyle = lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true)

	statusRunningStyle = lipgloss.NewStyle().
		Foreground(runningColor).
		Bold(true)

	// Text styles
	normalTextStyle = lipgloss.NewStyle().
		Foreground(textColor)

	dimTextStyle = lipgloss.NewStyle().
		Foreground(dimTextColor)

	highlightStyle = lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true)

	// Panel styles
	panelStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		MarginBottom(1)

	focusedPanelStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		MarginBottom(1)

	// Table styles
	tableHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(headerColor).
		Background(borderColor).
		Padding(0, 1)

	tableCellStyle = lipgloss.NewStyle().
		Foreground(textColor).
		Padding(0, 1)

	selectedRowStyle = lipgloss.NewStyle().
		Background(selectedBgColor).
		Foreground(headerColor)

	// Status bar styles
	statusBarStyle = lipgloss.NewStyle().
		Background(borderColor).
		Foreground(headerColor).
		Padding(0, 1).
		Bold(true)

	// Help styles
	helpStyle = lipgloss.NewStyle().
		Foreground(dimTextColor).
		Background(bgColor).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor)

	helpKeyStyle = lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true)

	helpDescStyle = lipgloss.NewStyle().
		Foreground(textColor)

	// Log styles
	logEntryStyle = lipgloss.NewStyle().
		Foreground(textColor).
		MarginBottom(0)

	logTimestampStyle = lipgloss.NewStyle().
		Foreground(dimTextColor).
		Bold(true)

	logFileStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)

	logActionStyle = lipgloss.NewStyle().
		Foreground(secondaryColor)

	// Command specific styles
	commandTSCStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#007acc")).
		Bold(true)

	commandLintStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4b32c3")).
		Bold(true)

	commandTestStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#97ca00")).
		Bold(true)
)

// Panel dimensions and layout
const (
	minWidth       = 80
	minHeight      = 24
	statusBarHeight = 1
	headerHeight    = 3
	helpPanelHeight = 10
)

// GetStatusStyle returns appropriate style for command status
func GetStatusStyle(passed bool, running bool) lipgloss.Style {
	if running {
		return statusRunningStyle
	}
	if passed {
		return statusPassStyle
	}
	return statusFailStyle
}

// GetStatusIcon returns appropriate icon for command status
func GetStatusIcon(passed bool, running bool) string {
	if running {
		return "⟳"
	}
	if passed {
		return "✓"
	}
	return "✗"
}

// GetCommandStyle returns appropriate style for command type
func GetCommandStyle(commandType string) lipgloss.Style {
	switch commandType {
	case "typescript":
		return commandTSCStyle
	case "lint":
		return commandLintStyle
	case "test":
		return commandTestStyle
	default:
		return normalTextStyle
	}
}

// FormatDuration formats duration for display
func FormatDuration(d int64) string {
	if d < 1000 {
		return lipgloss.NewStyle().Foreground(successColor).Render(fmt.Sprintf("%dms", d))
	}
	if d < 10000 {
		return lipgloss.NewStyle().Foreground(warningColor).Render(fmt.Sprintf("%.1fs", float64(d)/1000))
	}
	return lipgloss.NewStyle().Foreground(errorColor).Render(fmt.Sprintf("%.1fs", float64(d)/1000))
}

// Center centers text within given width
func Center(text string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(text)
}

// Truncate truncates text to fit within given width
func Truncate(text string, width int) string {
	if len(text) <= width {
		return text
	}
	if width <= 3 {
		return text[:width]
	}
	return text[:width-3] + "..."
}