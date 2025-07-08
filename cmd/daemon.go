package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"kwatch/config"
	"kwatch/runner"
	"kwatch/server"
)

var (
	daemonPort int
	daemonHost string
)

// daemonServer represents the HTTP server for daemon mode
type daemonServer struct {
	runner  *runner.Runner
	workDir string
	server  *http.Server
}

// daemonStatusResponse represents the daemon status response
type daemonStatusResponse struct {
	Status    string                             `json:"status"`
	Directory string                             `json:"directory"`
	Timestamp string                             `json:"timestamp"`
	Commands  map[string]statusCommandResult    `json:"commands"`
}

var daemonCmd = &cobra.Command{
	Use:   "daemon [directory]",
	Short: "Start background daemon server",
	Long: `Start kwatch in daemon mode, providing an HTTP API for status monitoring.

The daemon runs in the background and provides endpoints for:
- GET /status - Get current build status (JSON)
- GET /status/compact - Get compact one-line status
- POST /run - Force a manual run of all commands
- GET /history - Get command execution history
- POST /security/scan - Run security scan
- GET /security/findings - List security findings
- GET /security/stats - Security statistics
- POST /security/resolve/{id} - Mark finding as resolved
- POST /security/ignore/{id} - Mark finding as ignored

Examples:
  kwatch daemon                        # Start daemon on port 3737
  kwatch daemon --port 8080            # Start daemon on port 8080
  kwatch daemon --host 0.0.0.0        # Bind to all interfaces
  kwatch daemon /path/to/project       # Monitor specific directory
  kwatch --dir /path/to/project daemon # Monitor specific directory (flag)
  kwatch . daemon                      # Monitor current directory
  kwatch daemon .                      # Monitor current directory`,
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
		
		// Create daemon server
		daemon := &daemonServer{
			runner:  r,
			workDir: absDir,
		}

		// Set up HTTP server
		addr := fmt.Sprintf("%s:%d", daemonHost, daemonPort)
		daemon.server = &http.Server{
			Addr:         addr,
			Handler:      daemon.setupRoutes(),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		}

		// Handle graceful shutdown
		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			fmt.Println("\nShutting down daemon...")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := daemon.server.Shutdown(ctx); err != nil {
				log.Printf("Error shutting down server: %v\n", err)
			}
		}()

		// Start server
		fmt.Printf("\n=== KWatch Daemon Starting ===\n")
		fmt.Printf("Monitoring directory: %s\n", absDir)
		fmt.Printf("Server address: %s\n", addr)
		fmt.Printf("\nAvailable endpoints:\n")
		fmt.Printf("  GET  http://%s/status\n", addr)
		fmt.Printf("  GET  http://%s/status/compact\n", addr)
		fmt.Printf("  POST http://%s/run\n", addr)
		fmt.Printf("  GET  http://%s/history\n", addr)
		fmt.Printf("  GET  http://%s/health\n", addr)
		fmt.Printf("\nPress Ctrl+C to stop the daemon\n")
		fmt.Printf("===============================\n\n")

		if err := daemon.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().IntVarP(&daemonPort, "port", "p", 3737, "Port to bind the daemon server")
	daemonCmd.Flags().StringVarP(&daemonHost, "host", "H", "localhost", "Host to bind the daemon server")
}

// setupRoutes configures the HTTP routes for the daemon
func (d *daemonServer) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Status endpoint (JSON)
	mux.HandleFunc("/status", d.handleStatus)

	// Status endpoint (compact)
	mux.HandleFunc("/status/compact", d.handleStatusCompact)

	// Manual run endpoint
	mux.HandleFunc("/run", d.handleRun)

	// History endpoint
	mux.HandleFunc("/history", d.handleHistory)

	// Health check endpoint
	mux.HandleFunc("/health", d.handleHealth)

	// Security endpoints
	securityAPI := server.NewSecurityAPI(".security-findings.json")
	mux.HandleFunc("/security/scan", securityAPI.HandleSecurityScan)
	mux.HandleFunc("/security/findings", securityAPI.HandleSecurityFindings)
	mux.HandleFunc("/security/findings/", securityAPI.HandleSecurityFinding)
	mux.HandleFunc("/security/stats", securityAPI.HandleSecurityStats)
	mux.HandleFunc("/security/resolve/", securityAPI.HandleSecurityResolve)
	mux.HandleFunc("/security/ignore/", securityAPI.HandleSecurityIgnore)

	return mux
}

// handleStatus handles GET /status
func (d *daemonServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	results := d.runner.RunAll(ctx)

	response := daemonStatusResponse{
		Status:    "ok",
		Directory: d.workDir,
		Timestamp: time.Now().Format(time.RFC3339),
		Commands:  make(map[string]statusCommandResult),
	}

	// Convert results to response format
	cmdNames := map[runner.CommandType]string{
		runner.TypescriptCheck: "tsc",
		runner.LintCheck:       "lint",
		runner.TestRunner:      "test",
	}

	for cmdType, result := range results {
		cmdName := cmdNames[cmdType]
		if cmdName == "" {
			cmdName = string(cmdType)
		}

		response.Commands[cmdName] = statusCommandResult{
			Passed:     result.Passed,
			IssueCount: result.IssueCount,
			Duration:   formatDuration(result.Duration),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStatusCompact handles GET /status/compact
func (d *daemonServer) handleStatusCompact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	results := d.runner.RunAll(ctx)
	compact := runner.FormatCompactStatus(results)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(compact))
}

// handleRun handles POST /run
func (d *daemonServer) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	results := d.runner.RunAll(ctx)

	response := map[string]interface{}{
		"status":    "completed",
		"timestamp": time.Now().Format(time.RFC3339),
		"results":   results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHistory handles GET /history
func (d *daemonServer) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	history := d.runner.GetHistory()
	
	response := map[string]interface{}{
		"history": history,
		"count":   len(history),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth handles GET /health
func (d *daemonServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"directory": d.workDir,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}