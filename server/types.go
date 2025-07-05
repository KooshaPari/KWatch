package server

import (
	"time"
)

// Config represents the configuration for the HTTP server
type Config struct {
	// Host specifies the host to bind the server to
	Host string
	// Port specifies the port to bind the server to
	Port int
	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout time.Duration
	// WriteTimeout is the maximum duration before timing out writes
	WriteTimeout time.Duration
	// IdleTimeout is the maximum amount of time to wait for the next request
	IdleTimeout time.Duration
	// AuthToken is the simple authentication token for protected endpoints
	AuthToken string
	// EnableCORS enables CORS support for web-based agents
	EnableCORS bool
	// AllowedOrigins specifies allowed origins for CORS requests
	AllowedOrigins []string
	// WorkingDir is the directory being monitored
	WorkingDir string
}

// DefaultConfig returns a default configuration for the server
func DefaultConfig() *Config {
	return &Config{
		Host:           "localhost",
		Port:           3737,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		AuthToken:      "",
		EnableCORS:     true,
		AllowedOrigins: []string{"*"},
		WorkingDir:     ".",
	}
}

// StatusResponse represents the full status response
type StatusResponse struct {
	Status    string                       `json:"status"`
	Directory string                       `json:"directory"`
	Timestamp string                       `json:"timestamp"`
	Commands  map[string]CommandResult     `json:"commands"`
	Server    ServerInfo                   `json:"server"`
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	Passed     bool   `json:"passed"`
	IssueCount int    `json:"issue_count"`
	Duration   string `json:"duration"`
	LastRun    string `json:"last_run"`
}

// ServerInfo contains basic server information
type ServerInfo struct {
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
	GoVersion string `json:"go_version"`
}

// QuickResponse represents a minimal response for ultra-fast endpoints
type QuickResponse struct {
	Status string `json:"status"`
}

// HealthResponse represents the health check response with system info
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Directory string            `json:"directory"`
	System    map[string]string `json:"system"`
	Uptime    string            `json:"uptime"`
}

// MetricsResponse represents the metrics response for monitoring
type MetricsResponse struct {
	Timestamp      string             `json:"timestamp"`
	Server         ServerMetrics      `json:"server"`
	Commands       CommandMetrics     `json:"commands"`
	Performance    PerformanceMetrics `json:"performance"`
}

// ServerMetrics contains server-specific metrics
type ServerMetrics struct {
	RequestCount     int64             `json:"request_count"`
	RequestsByPath   map[string]int64  `json:"requests_by_path"`
	AverageResponse  float64           `json:"average_response_ms"`
	ErrorCount       int64             `json:"error_count"`
	UptimeSeconds    float64           `json:"uptime_seconds"`
}

// CommandMetrics contains command execution metrics
type CommandMetrics struct {
	TotalRuns     int64   `json:"total_runs"`
	SuccessRate   float64 `json:"success_rate"`
	FailureCount  int64   `json:"failure_count"`
	AverageTime   float64 `json:"average_time_ms"`
}

// PerformanceMetrics contains performance-related metrics
type PerformanceMetrics struct {
	MemoryUsageMB    float64 `json:"memory_usage_mb"`
	CPUUsagePercent  float64 `json:"cpu_usage_percent"`
	GoroutineCount   int     `json:"goroutine_count"`
	GCPauseTimeMS    float64 `json:"gc_pause_time_ms"`
}

// RunResponse represents the response from triggering a manual run
type RunResponse struct {
	Status      string                   `json:"status"`
	Timestamp   string                   `json:"timestamp"`
	Triggered   bool                     `json:"triggered"`
	Results     map[string]CommandResult `json:"results,omitempty"`
	Message     string                   `json:"message,omitempty"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error     string `json:"error"`
	Timestamp string `json:"timestamp"`
	Code      int    `json:"code"`
}

// Runner interface defines the methods needed to interact with the command runner
type Runner interface {
	RunAll(ctx interface{}) map[string]CommandResult
	GetHistory() []interface{}
	GetMetrics() CommandMetrics
}

// Server represents the HTTP server instance
type Server struct {
	config     *Config
	runner     Runner
	startTime  time.Time
	metrics    *ServerMetrics
}