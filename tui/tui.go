package tui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/term"
)

// TUI represents the main TUI application
type TUI struct {
	program   *tea.Program
	model     Model
	watcher   *fsnotify.Watcher
	watchDir  string
	logFile   *os.File
}

// NewTUI creates a new TUI instance
func NewTUI(watchDir string) (*TUI, error) {
	// Resolve absolute path
	absDir, err := filepath.Abs(watchDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directory: %w", err)
	}
	
	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", absDir)
	}
	
	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}
	
	// Create model
	model := NewModel(absDir)
	
	// Create TUI
	tui := &TUI{
		model:    model,
		watcher:  watcher,
		watchDir: absDir,
	}
	
	// Setup logging
	if err := tui.setupLogging(); err != nil {
		return nil, fmt.Errorf("failed to setup logging: %w", err)
	}
	
	return tui, nil
}

// Start starts the TUI application
func (t *TUI) Start() error {
	// Check if we're in a compatible terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("not running in an interactive terminal\nTry running 'kwatch . daemon' for background mode or 'kwatch . status' for status checking")
	}
	
	// Initialize bubbletea program with fallback options
	t.program = tea.NewProgram(
		t.model,
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stderr),
	)
	
	// Start file watcher
	if err := t.startFileWatcher(); err != nil {
		return fmt.Errorf("failed to start file watcher: %w", err)
	}
	
	// Run initial commands and set status
	go t.runInitialCommands()
	
	// Start the program
	_, err := t.program.Run()
	if err != nil {
		return fmt.Errorf("failed to run TUI: %w\nTry running 'kwatch . status' instead for non-interactive monitoring", err)
	}
	
	return nil
}

// Stop stops the TUI application
func (t *TUI) Stop() error {
	if t.watcher != nil {
		t.watcher.Close()
	}
	
	if t.logFile != nil {
		t.logFile.Close()
	}
	
	if t.program != nil {
		t.program.Kill()
	}
	
	return nil
}

// setupLogging sets up logging for the TUI
func (t *TUI) setupLogging() error {
	// Create logs directory if it doesn't exist
	logDir := filepath.Join(t.watchDir, ".kwatch")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	
	// Create log file
	logFile := filepath.Join(logDir, "kwatch.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	
	t.logFile = file
	return nil
}

// startFileWatcher starts the file system watcher
func (t *TUI) startFileWatcher() error {
	// Watch the main directory
	if err := t.watcher.Add(t.watchDir); err != nil {
		return fmt.Errorf("failed to watch directory %s: %w", t.watchDir, err)
	}
	
	// Watch common source directories
	watchDirs := []string{
		"src",
		"lib",
		"components",
		"pages",
		"utils",
		"types",
		"hooks",
		"services",
		"api",
		"styles",
		"public",
		"tests",
		"__tests__",
		"test",
		"spec",
	}
	
	for _, dir := range watchDirs {
		dirPath := filepath.Join(t.watchDir, dir)
		if _, err := os.Stat(dirPath); err == nil {
			if err := t.addWatchRecursive(dirPath); err != nil {
				// Log error but continue
				t.logError(fmt.Sprintf("Failed to watch directory %s: %v", dirPath, err))
			}
		}
	}
	
	// Start watching in a goroutine
	go t.watchFiles()
	
	return nil
}

// addWatchRecursive adds watches recursively
func (t *TUI) addWatchRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip hidden directories and files
		if info.IsDir() && (info.Name()[0] == '.' || info.Name() == "node_modules") {
			return filepath.SkipDir
		}
		
		// Only watch directories
		if info.IsDir() {
			return t.watcher.Add(path)
		}
		
		return nil
	})
}

// watchFiles processes file system events with debouncing
func (t *TUI) watchFiles() {
	var lastEventTime time.Time
	debounceDelay := 2 * time.Second
	
	for {
		select {
		case event, ok := <-t.watcher.Events:
			if !ok {
				// Watcher channel closed, notify that watcher stopped
				if t.program != nil {
					t.program.Send(statusUpdateMsg{
						watcherActive: false,
						serverActive:  false,
					})
				}
				return
			}
			
			// Skip ignored events (like chmod)
			if t.shouldIgnoreEvent(event.Op) {
				continue
			}
			
			// Filter relevant file types
			if !t.isRelevantFile(event.Name) {
				continue
			}
			
			// Debounce events - ignore if too soon after last event
			now := time.Now()
			if now.Sub(lastEventTime) < debounceDelay {
				continue
			}
			lastEventTime = now
			
			action := t.getFileAction(event.Op)
			
			// Send file change message to the program
			if t.program != nil {
				t.program.Send(fileChangeMsg{
					file:   event.Name,
					action: action,
				})
			}
			
			// Log the file change
			t.logFileChange(event.Name, action)
		
		case err, ok := <-t.watcher.Errors:
			if !ok {
				// Watcher error channel closed, notify that watcher stopped
				if t.program != nil {
					t.program.Send(statusUpdateMsg{
						watcherActive: false,
						serverActive:  false,
					})
				}
				return
			}
			
			// Send error message to the program
			if t.program != nil {
				t.program.Send(errorMsg{
					err: fmt.Sprintf("File watcher error: %v", err),
				})
			}
			
			t.logError(fmt.Sprintf("File watcher error: %v", err))
		}
	}
}

// isRelevantFile checks if a file change is relevant for monitoring
func (t *TUI) isRelevantFile(filename string) bool {
	// Ignore hidden files and directories
	if strings.Contains(filename, "/.") {
		return false
	}
	
	// Ignore common build/temp directories
	ignoreDirs := []string{
		"node_modules/", "dist/", "build/", ".next/", ".nuxt/",
		"coverage/", ".nyc_output/", ".cache/", ".tmp/", "tmp/",
		".kwatch/", ".git/", ".vscode/", ".idea/",
		"__pycache__/", ".pytest_cache/",
	}
	
	for _, ignoreDir := range ignoreDirs {
		if strings.Contains(filename, ignoreDir) {
			return false
		}
	}
	
	// Ignore temp/log files
	ignoreExts := []string{
		".log", ".tmp", ".temp", ".cache", ".pid", ".lock",
		".swp", ".swo", ".DS_Store", ".env.local",
	}
	
	ext := filepath.Ext(filename)
	for _, ignoreExt := range ignoreExts {
		if ext == ignoreExt {
			return false
		}
	}
	
	// Only watch source files
	relevantExts := []string{
		".ts", ".tsx", ".js", ".jsx",
		".json", ".yaml", ".yml",
		".css", ".scss", ".sass", ".less",
		".html", ".htm", ".vue",
		".md", ".mdx",
		".graphql", ".gql",
		".prisma", ".proto",
	}
	
	for _, relevantExt := range relevantExts {
		if ext == relevantExt {
			return true
		}
	}
	
	// Check for specific config filenames
	base := filepath.Base(filename)
	relevantFiles := []string{
		"package.json", "tsconfig.json", "jsconfig.json",
		".eslintrc.js", ".eslintrc.json", ".prettierrc",
		"jest.config.js", "vite.config.js", "webpack.config.js",
		"next.config.js", "tailwind.config.js",
	}
	
	for _, relevantFile := range relevantFiles {
		if base == relevantFile {
			return true
		}
	}
	
	return false
}

// getFileAction converts fsnotify operation to readable action
func (t *TUI) getFileAction(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create != 0:
		return "created"
	case op&fsnotify.Write != 0:
		return "modified"
	case op&fsnotify.Remove != 0:
		return "deleted"
	case op&fsnotify.Rename != 0:
		return "renamed"
	case op&fsnotify.Chmod != 0:
		return "chmod"
	default:
		return "changed"
	}
}

// shouldIgnoreEvent checks if we should ignore this file event
func (t *TUI) shouldIgnoreEvent(op fsnotify.Op) bool {
	// Ignore chmod events to prevent loops
	return op&fsnotify.Chmod != 0
}

// runInitialCommands runs the initial set of commands
func (t *TUI) runInitialCommands() {
	// Wait a bit for the UI to initialize
	time.Sleep(500 * time.Millisecond)
	
	// Send initial status - watcher is active
	if t.program != nil {
		t.program.Send(statusUpdateMsg{
			watcherActive: true,
			serverActive:  false,
		})
	}
	
	// Wait a bit more then send refresh command to trigger initial run
	time.Sleep(100 * time.Millisecond)
	if t.program != nil {
		t.program.Send(refreshMsg{})
	}
}

// logFileChange logs a file change event
func (t *TUI) logFileChange(filename, action string) {
	if t.logFile == nil {
		return
	}
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("%s [FILE] %s: %s\n", timestamp, action, filename)
	
	if _, err := t.logFile.WriteString(logEntry); err != nil {
		log.Printf("Failed to write to log file: %v", err)
	}
}

// logError logs an error event
func (t *TUI) logError(message string) {
	if t.logFile == nil {
		return
	}
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("%s [ERROR] %s\n", timestamp, message)
	
	if _, err := t.logFile.WriteString(logEntry); err != nil {
		log.Printf("Failed to write to log file: %v", err)
	}
}

// StartTUI is the main entry point for starting the TUI
func StartTUI(watchDir string) error {
	// Create TUI
	tui, err := NewTUI(watchDir)
	if err != nil {
		return fmt.Errorf("failed to create TUI: %w", err)
	}
	
	// Setup cleanup on exit
	defer func() {
		if err := tui.Stop(); err != nil {
			log.Printf("Error stopping TUI: %v", err)
		}
	}()
	
	// Start TUI
	if err := tui.Start(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}
	
	return nil
}

// RunWithConfig runs the TUI with a specific configuration
func RunWithConfig(config Config) error {
	// Create TUI with config
	tui, err := NewTUI(config.WatchDir)
	if err != nil {
		return fmt.Errorf("failed to create TUI: %w", err)
	}
	
	// Apply configuration
	tui.model.serverPort = config.ServerPort
	if config.MaxLogs > 0 {
		tui.model.maxLogs = config.MaxLogs
	}
	
	// Setup cleanup on exit
	defer func() {
		if err := tui.Stop(); err != nil {
			log.Printf("Error stopping TUI: %v", err)
		}
	}()
	
	// Start TUI
	if err := tui.Start(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}
	
	return nil
}

// Config represents TUI configuration
type Config struct {
	WatchDir   string
	ServerPort int
	MaxLogs    int
	LogLevel   string
}

// DefaultConfig returns default configuration
func DefaultConfig(watchDir string) Config {
	return Config{
		WatchDir:   watchDir,
		ServerPort: 8080,
		MaxLogs:    1000,
		LogLevel:   "info",
	}
}