package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// New creates a new HTTP server instance
func New(config *Config, runner Runner) *Server {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &Server{
		config:    config,
		runner:    runner,
		startTime: time.Now(),
		metrics: &ServerMetrics{
			RequestsByPath: make(map[string]int64),
		},
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Set up routes
	mux := s.setupRoutes()
	
	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:      mux,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}
	
	// Log server start
	log.Printf("Starting kwatch HTTP server on %s:%d", s.config.Host, s.config.Port)
	log.Printf("Monitoring directory: %s", s.config.WorkingDir)
	if s.config.AuthToken != "" {
		log.Printf("Authentication enabled")
	}
	if s.config.EnableCORS {
		log.Printf("CORS enabled for origins: %v", s.config.AllowedOrigins)
	}
	
	// Print available endpoints
	s.printEndpoints()
	
	// Start server
	return server.ListenAndServe()
}

// StartWithContext starts the HTTP server with context support for graceful shutdown
func (s *Server) StartWithContext(ctx context.Context) error {
	// Set up routes
	mux := s.setupRoutes()
	
	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:      mux,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}
	
	// Log server start
	log.Printf("Starting kwatch HTTP server on %s:%d", s.config.Host, s.config.Port)
	log.Printf("Monitoring directory: %s", s.config.WorkingDir)
	if s.config.AuthToken != "" {
		log.Printf("Authentication enabled")
	}
	if s.config.EnableCORS {
		log.Printf("CORS enabled for origins: %v", s.config.AllowedOrigins)
	}
	
	// Print available endpoints
	s.printEndpoints()
	
	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()
	
	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		log.Println("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-serverErr:
		return err
	}
}

// setupRoutes configures the HTTP routes and middleware
func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()
	
	// Define allowed methods for each endpoint
	allowedMethods := map[string][]string{
		"/":               {"GET"},
		"/status":         {"GET"},
		"/status/compact": {"GET"},
		"/quick":          {"GET"},
		"/run":            {"POST"},
		"/metrics":        {"GET"},
		"/health":         {"GET"},
		"/history":        {"GET"},
		"/ping":           {"GET"},
		"/favicon.ico":    {"GET"},
	}
	
	// Register handlers
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/status/compact", s.handleStatusCompact)
	mux.HandleFunc("/quick", s.handleQuick)
	mux.HandleFunc("/run", s.handleRun)
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/history", s.handleHistory)
	mux.HandleFunc("/ping", s.handlePing)
	mux.HandleFunc("/favicon.ico", s.handleFavicon)
	
	// Wrap with middleware (order matters - last added is first executed)
	var handler http.Handler = mux
	
	// Add method validation middleware
	handler = s.methodMiddleware(allowedMethods)(handler)
	
	// Add security headers and fast response middleware
	handler = s.fastResponseMiddleware(handler)
	
	// Add metrics middleware
	handler = s.metricsMiddleware(handler)
	
	// Add authentication middleware
	handler = s.authMiddleware(handler)
	
	// Add CORS middleware
	handler = s.corsMiddleware(handler)
	
	// Add logging middleware (outermost)
	handler = s.logMiddleware(handler)
	
	return handler
}

// printEndpoints prints available endpoints to the console
func (s *Server) printEndpoints() {
	baseURL := fmt.Sprintf("http://%s:%d", s.config.Host, s.config.Port)
	
	log.Printf("Available endpoints:")
	log.Printf("  GET  %s/            - API information", baseURL)
	log.Printf("  GET  %s/status      - Quick status check (JSON)", baseURL)
	log.Printf("  GET  %s/status/compact - Single-line status", baseURL)
	log.Printf("  GET  %s/quick       - Ultra-fast health check", baseURL)
	log.Printf("  POST %s/run         - Trigger manual run", baseURL)
	log.Printf("  GET  %s/metrics     - Basic metrics", baseURL)
	log.Printf("  GET  %s/health      - Health check with system info", baseURL)
	log.Printf("  GET  %s/history     - Command execution history", baseURL)
	log.Printf("  GET  %s/ping        - Simple ping", baseURL)
}

// GetConfig returns the server configuration
func (s *Server) GetConfig() *Config {
	return s.config
}

// GetMetrics returns the current server metrics
func (s *Server) GetMetrics() *ServerMetrics {
	return s.metrics
}

// GetUptime returns the server uptime
func (s *Server) GetUptime() time.Duration {
	return time.Since(s.startTime)
}

// IsHealthy checks if the server is healthy
func (s *Server) IsHealthy() bool {
	// Basic health check - could be extended with more sophisticated checks
	return true
}

// SetRunner allows changing the runner after server creation
func (s *Server) SetRunner(runner Runner) {
	s.runner = runner
}

// UpdateConfig updates the server configuration
// Note: Some config changes may require server restart
func (s *Server) UpdateConfig(config *Config) {
	s.config = config
}

// ListenAndServe is a convenience method that starts the server
func ListenAndServe(addr string, runner Runner) error {
	config := DefaultConfig()
	
	// Parse address
	if addr != "" {
		// Simple parsing - could be enhanced for more complex addresses
		config.Host = "localhost"
		config.Port = 3737
	}
	
	server := New(config, runner)
	return server.Start()
}

// QuickServer creates and starts a server with minimal configuration
func QuickServer(port int, runner Runner) error {
	config := DefaultConfig()
	config.Port = port
	
	server := New(config, runner)
	return server.Start()
}

// SecureServer creates and starts a server with authentication enabled
func SecureServer(port int, authToken string, runner Runner) error {
	config := DefaultConfig()
	config.Port = port
	config.AuthToken = authToken
	
	server := New(config, runner)
	return server.Start()
}

// DevServer creates and starts a development server with CORS enabled
func DevServer(port int, runner Runner) error {
	config := DefaultConfig()
	config.Port = port
	config.EnableCORS = true
	config.AllowedOrigins = []string{"*"}
	
	server := New(config, runner)
	return server.Start()
}

// ProductionServer creates and starts a production server with security features
func ProductionServer(port int, authToken string, allowedOrigins []string, runner Runner) error {
	config := DefaultConfig()
	config.Port = port
	config.AuthToken = authToken
	config.EnableCORS = true
	config.AllowedOrigins = allowedOrigins
	config.ReadTimeout = 10 * time.Second
	config.WriteTimeout = 10 * time.Second
	config.IdleTimeout = 30 * time.Second
	
	server := New(config, runner)
	return server.Start()
}