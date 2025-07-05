package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"
)

// handleStatus handles GET /status - Quick status check (JSON)
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	// Fast response - use minimal processing
	ctx := context.Background()
	
	// Get results from runner
	results := s.runner.RunAll(ctx)
	
	// Build response
	response := StatusResponse{
		Status:    "ok",
		Directory: s.config.WorkingDir,
		Timestamp: time.Now().Format(time.RFC3339),
		Commands:  results,
		Server: ServerInfo{
			Version:   "1.0.0",
			Uptime:    time.Since(s.startTime).String(),
			GoVersion: runtime.Version(),
		},
	}
	
	// Check if any command failed
	for _, result := range results {
		if !result.Passed {
			response.Status = "error"
			break
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStatusCompact handles GET /status/compact - Single-line status for shell integration
func (s *Server) handleStatusCompact(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	results := s.runner.RunAll(ctx)
	
	// Build compact status string
	compact := s.buildCompactStatus(results)
	
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(compact))
}

// handleQuick handles GET /quick - Ultra-fast health check (just "ok")
func (s *Server) handleQuick(w http.ResponseWriter, r *http.Request) {
	// Ultra-fast response - minimal processing
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("ok"))
}

// handleRun handles POST /run - Trigger manual run (returns immediately)
func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	
	// Start run asynchronously for immediate response
	go func() {
		s.runner.RunAll(ctx)
	}()
	
	response := RunResponse{
		Status:    "triggered",
		Timestamp: time.Now().Format(time.RFC3339),
		Triggered: true,
		Message:   "Manual run triggered in background",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMetrics handles GET /metrics - Basic metrics for monitoring
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// Get runtime metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	// Get GC stats
	var gcStats debug.GCStats
	debug.ReadGCStats(&gcStats)
	
	// Calculate uptime
	uptime := time.Since(s.startTime)
	
	// Build metrics response
	response := MetricsResponse{
		Timestamp: time.Now().Format(time.RFC3339),
		Server: ServerMetrics{
			RequestCount:     s.metrics.RequestCount,
			RequestsByPath:   s.metrics.RequestsByPath,
			AverageResponse:  s.metrics.AverageResponse,
			ErrorCount:       s.metrics.ErrorCount,
			UptimeSeconds:    uptime.Seconds(),
		},
		Commands: s.runner.GetMetrics(),
		Performance: PerformanceMetrics{
			MemoryUsageMB:    float64(memStats.Alloc) / 1024 / 1024,
			CPUUsagePercent:  0.0, // Would need additional CPU monitoring
			GoroutineCount:   runtime.NumGoroutine(),
			GCPauseTimeMS:    float64(gcStats.PauseTotal.Nanoseconds()) / 1e6,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth handles GET /health - Health check with system info
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Get system information
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	systemInfo := map[string]string{
		"go_version":     runtime.Version(),
		"goroutines":     fmt.Sprintf("%d", runtime.NumGoroutine()),
		"memory_alloc":   fmt.Sprintf("%.2f MB", float64(memStats.Alloc)/1024/1024),
		"memory_sys":     fmt.Sprintf("%.2f MB", float64(memStats.Sys)/1024/1024),
		"gc_runs":        fmt.Sprintf("%d", memStats.NumGC),
		"cpu_cores":      fmt.Sprintf("%d", runtime.NumCPU()),
		"os":             runtime.GOOS,
		"arch":           runtime.GOARCH,
	}
	
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Directory: s.config.WorkingDir,
		System:    systemInfo,
		Uptime:    time.Since(s.startTime).String(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHistory handles GET /history - Command execution history
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	history := s.runner.GetHistory()
	
	response := map[string]interface{}{
		"history":   history,
		"count":     len(history),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleNotFound handles 404 errors
func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	s.writeErrorResponse(w, fmt.Sprintf("Endpoint not found: %s", r.URL.Path), http.StatusNotFound)
}

// handleMethodNotAllowed handles 405 errors
func (s *Server) handleMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	s.writeErrorResponse(w, fmt.Sprintf("Method %s not allowed for %s", r.Method, r.URL.Path), http.StatusMethodNotAllowed)
}

// buildCompactStatus builds a compact status string for shell integration
func (s *Server) buildCompactStatus(results map[string]CommandResult) string {
	if len(results) == 0 {
		return "NO_COMMANDS"
	}
	
	var status []string
	var failed []string
	
	for name, result := range results {
		if result.Passed {
			status = append(status, fmt.Sprintf("%s:✓", name))
		} else {
			status = append(status, fmt.Sprintf("%s:✗", name))
			failed = append(failed, name)
		}
	}
	
	// Build compact string
	compact := ""
	if len(failed) == 0 {
		compact = "ALL_PASSED"
	} else {
		compact = fmt.Sprintf("FAILED:%s", fmt.Sprintf("%v", failed))
	}
	
	return compact
}

// handlePing handles GET /ping - Simple ping endpoint
func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("pong"))
}

// handleRoot handles GET / - Root endpoint with API information
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	apiInfo := map[string]interface{}{
		"name":        "kwatch-server",
		"version":     "1.0.0",
		"description": "HTTP server for AI agent polling",
		"uptime":      time.Since(s.startTime).String(),
		"endpoints": map[string]string{
			"GET /":               "API information",
			"GET /status":         "Quick status check (JSON)",
			"GET /status/compact": "Single-line status for shell integration",
			"GET /quick":          "Ultra-fast health check",
			"POST /run":           "Trigger manual run",
			"GET /metrics":        "Basic metrics for monitoring",
			"GET /health":         "Health check with system info",
			"GET /history":        "Command execution history",
			"GET /ping":           "Simple ping endpoint",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiInfo)
}

// handleFavicon handles GET /favicon.ico - Prevent 404s from browsers
func (s *Server) handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// validateRequestMethod validates that the request method is allowed for the endpoint
func (s *Server) validateRequestMethod(w http.ResponseWriter, r *http.Request, allowedMethods []string) bool {
	for _, method := range allowedMethods {
		if r.Method == method {
			return true
		}
	}
	
	s.writeErrorResponse(w, fmt.Sprintf("Method %s not allowed", r.Method), http.StatusMethodNotAllowed)
	return false
}

// addSecurityHeaders adds security headers to responses
func (s *Server) addSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Server", "kwatch-server/1.0.0")
}