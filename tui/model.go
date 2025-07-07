package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"kwatch/config"
	"kwatch/runner"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	ViewMain ViewMode = iota
	ViewHistory
	ViewLogs
	ViewHelp
)

// Model represents the application state
type Model struct {
	// Core state
	ready      bool
	width      int
	height     int
	viewMode   ViewMode
	
	// Directory and configuration
	watchDir   string
	serverPort int
	
	// Command execution state
	history    *runner.ResultHistory
	running    map[runner.CommandType]bool
	lastRun    time.Time
	runner     *runner.Runner
	kwatchConfig *config.Config
	
	// UI state
	selectedRow int
	scrollOffset int
	
	// Logs and activities
	logs       []LogEntry
	maxLogs    int
	
	// Status
	watcherActive bool
	serverActive  bool
	
	// Error state
	error string
}

// LogEntry represents a log entry for the activity panel
type LogEntry struct {
	Timestamp time.Time
	Type      LogType
	Message   string
	File      string
	Action    string
}

// LogType represents the type of log entry
type LogType int

const (
	LogInfo LogType = iota
	LogWarning
	LogError
	LogFileChange
	LogCommandStart
	LogCommandEnd
)

// CommandStatus represents the current status of a command
type CommandStatus struct {
	Type      runner.CommandType
	Running   bool
	LastRun   time.Time
	Result    *runner.CommandResult
}

// NewModel creates a new model instance
func NewModel(watchDir string) Model {
	// Load kwatch configuration
	kwatchConfig, err := config.Load(watchDir)
	if err != nil {
		// Fall back to default config if loading fails
		kwatchConfig = config.DefaultConfig()
	}
	
	// Create runner configuration
	runnerConfig := runner.RunnerConfig{
		DefaultTimeout: 30 * time.Second,
		MaxParallel:    kwatchConfig.MaxParallel,
		WorkingDir:     watchDir,
	}
	
	// Create runner instance
	r := runner.NewRunner(runnerConfig, kwatchConfig)
	
	return Model{
		ready:        false,
		width:        80,
		height:       24,
		viewMode:     ViewMain,
		watchDir:     watchDir,
		serverPort:   8080,
		history:      &runner.ResultHistory{},
		running:      make(map[runner.CommandType]bool),
		lastRun:      time.Now(),
		runner:       r,
		kwatchConfig: kwatchConfig,
		logs:         make([]LogEntry, 0),
		maxLogs:      100,
		watcherActive: false,
		serverActive:  false,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tick(),
		tea.EnterAltScreen,
		// Initial log entry
		tea.Cmd(func() tea.Msg {
			return statusUpdateMsg{
				watcherActive: false, // Will be set to true by runInitialCommands
				serverActive:  false,
			}
		}),
	)
}

// UpdateSize updates the model dimensions
func (m *Model) UpdateSize(width, height int) {
	m.width = width
	m.height = height
	m.ready = true
}

// AddLog adds a new log entry
func (m *Model) AddLog(logType LogType, message, file, action string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Type:      logType,
		Message:   message,
		File:      file,
		Action:    action,
	}
	
	m.logs = append(m.logs, entry)
	
	// Keep only the last maxLogs entries (more aggressive truncation)
	if len(m.logs) > m.maxLogs {
		m.logs = m.logs[len(m.logs)-m.maxLogs:]
	}
	
	// Additional cleanup - if logs are still too many for UI, trim more aggressively
	if len(m.logs) > 50 {
		// Keep only the most recent 50 logs when UI might be overwhelmed
		m.logs = m.logs[len(m.logs)-50:]
	}
}

// GetCurrentCommandStatuses returns the current status of all commands
func (m *Model) GetCurrentCommandStatuses() []CommandStatus {
	latest := m.history.GetLatest()
	statuses := make([]CommandStatus, 0, 3)
	
	commandTypes := []runner.CommandType{
		runner.TypescriptCheck,
		runner.LintCheck,
		runner.TestRunner,
	}
	
	for _, cmdType := range commandTypes {
		status := CommandStatus{
			Type:    cmdType,
			Running: m.running[cmdType],
		}
		
		if result, exists := latest[cmdType]; exists {
			status.Result = &result
			status.LastRun = result.Timestamp
		}
		
		statuses = append(statuses, status)
	}
	
	return statuses
}

// GetRecentLogs returns the most recent log entries
func (m *Model) GetRecentLogs(count int) []LogEntry {
	if len(m.logs) == 0 {
		return []LogEntry{}
	}
	
	start := len(m.logs) - count
	if start < 0 {
		start = 0
	}
	
	return m.logs[start:]
}

// GetHistoryForView returns command history formatted for display
func (m *Model) GetHistoryForView() []runner.CommandResult {
	results := m.history.GetAll()
	
	// Sort by timestamp (most recent first)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Timestamp.Before(results[j].Timestamp) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	
	return results
}

// SetCommandRunning sets the running state for a command
func (m *Model) SetCommandRunning(cmdType runner.CommandType, running bool) {
	m.running[cmdType] = running
	
	if running {
		m.AddLog(LogCommandStart, "Command started", "", string(cmdType))
	} else {
		m.AddLog(LogCommandEnd, "Command completed", "", string(cmdType))
	}
}

// IsAnyCommandRunning checks if any commands are currently running
func (m *Model) IsAnyCommandRunning() bool {
	for _, running := range m.running {
		if running {
			return true
		}
	}
	return false
}

// AddCommandResult adds a command result to the history
func (m *Model) AddCommandResult(result runner.CommandResult) {
	m.history.Add(result)
	m.SetCommandRunning(getCommandType(result.Command), false)
	
	status := "PASSED"
	if !result.Passed {
		status = "FAILED"
	}
	
	m.AddLog(LogCommandEnd, "Command "+status, "", result.Command)
}

// getCommandType determines command type from command string
func getCommandType(command string) runner.CommandType {
	switch {
	case command == "tsc" || command == "typescript":
		return runner.TypescriptCheck
	case command == "lint":
		return runner.LintCheck
	case command == "test":
		return runner.TestRunner
	default:
		return runner.CommandType(command)
	}
}

// NavigateUp moves selection up
func (m *Model) NavigateUp() {
	if m.selectedRow > 0 {
		m.selectedRow--
	}
}

// NavigateDown moves selection down
func (m *Model) NavigateDown() {
	maxRows := m.getMaxRows()
	if m.selectedRow < maxRows-1 {
		m.selectedRow++
	}
}

// getMaxRows returns the maximum number of rows for the current view
func (m *Model) getMaxRows() int {
	switch m.viewMode {
	case ViewMain:
		return 3 // Three command types
	case ViewHistory:
		return len(m.GetHistoryForView())
	case ViewLogs:
		return len(m.logs)
	default:
		return 0
	}
}

// GetStatusSummary returns a summary of current status
func (m *Model) GetStatusSummary() string {
	statuses := m.GetCurrentCommandStatuses()
	passed := 0
	failed := 0
	running := 0
	
	for _, status := range statuses {
		if status.Running {
			running++
		} else if status.Result != nil {
			if status.Result.Passed {
				passed++
			} else {
				failed++
			}
		}
	}
	
	if running > 0 {
		return "Running"
	}
	if failed > 0 {
		return "Failed"
	}
	if passed == len(statuses) {
		return "All Passed"
	}
	
	return "Ready"
}

// GetErrorMetrics returns total error count and error file count for failed commands
func (m *Model) GetErrorMetrics() (int, int) {
	statuses := m.GetCurrentCommandStatuses()
	totalErrors := 0
	errorFiles := 0
	
	for _, status := range statuses {
		if status.Result != nil && !status.Result.Passed {
			if status.Type == runner.TestRunner {
				// For tests, count failed tests as errors
				totalErrors += status.Result.FailedTests
			} else {
				// For other commands, count issues as errors
				totalErrors += status.Result.IssueCount
				if status.Result.FileCount > 0 {
					errorFiles += status.Result.FileCount
				}
			}
		}
	}
	
	return totalErrors, errorFiles
}

// IsRunning returns true if any command is currently running
func (m *Model) IsRunning() bool {
	for _, running := range m.running {
		if running {
			return true
		}
	}
	return false
}

// SetWatcherActive sets the watcher active state
func (m *Model) SetWatcherActive(active bool) {
	// Only log if status actually changed
	if m.watcherActive != active {
		m.watcherActive = active
		if active {
			m.AddLog(LogInfo, "File watcher started", m.watchDir, "watch")
		} else {
			m.AddLog(LogInfo, "File watcher stopped", m.watchDir, "stop")
		}
	} else if active && len(m.logs) == 0 {
		// Add initial startup log if no logs exist yet
		m.AddLog(LogInfo, "KWatch TUI started - monitoring "+m.watchDir, "", "start")
	}
}

// SetServerActive sets the server active state
func (m *Model) SetServerActive(active bool) {
	// Only log if status actually changed
	if m.serverActive != active {
		m.serverActive = active
		if active {
			m.AddLog(LogInfo, "Server started", "", "server")
		} else {
			m.AddLog(LogInfo, "Server stopped", "", "server")
		}
	}
}

// SetError sets an error message
func (m *Model) SetError(err string) {
	m.error = err
	if err != "" {
		m.AddLog(LogError, err, "", "error")
	}
}

// ClearError clears the error message
func (m *Model) ClearError() {
	m.error = ""
}

// HasError returns true if there's an error
func (m *Model) HasError() bool {
	return m.error != ""
}

// GetError returns the current error message
func (m *Model) GetError() string {
	return m.error
}