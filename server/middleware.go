package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// corsMiddleware handles CORS headers for web-based agents
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.config.EnableCORS {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range s.config.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authMiddleware handles simple token-based authentication
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication if no token is configured
		if s.config.AuthToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip authentication for public endpoints
		if s.isPublicEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Check for token in Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.writeErrorResponse(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Expected format: "Bearer <token>" or just "<token>"
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			// No "Bearer " prefix, use the header value directly
			token = authHeader
		}

		if token != s.config.AuthToken {
			s.writeErrorResponse(w, "Invalid authorization token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// metricsMiddleware tracks request metrics
func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Increment request count
		s.metrics.RequestCount++
		
		// Track requests by path
		if s.metrics.RequestsByPath == nil {
			s.metrics.RequestsByPath = make(map[string]int64)
		}
		s.metrics.RequestsByPath[r.URL.Path]++

		// Create a response writer wrapper to capture status code
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrappedWriter, r)

		// Calculate response time
		duration := time.Since(start)
		
		// Update average response time (simple moving average)
		if s.metrics.AverageResponse == 0 {
			s.metrics.AverageResponse = float64(duration.Nanoseconds()) / 1e6
		} else {
			s.metrics.AverageResponse = (s.metrics.AverageResponse + float64(duration.Nanoseconds())/1e6) / 2
		}

		// Track errors
		if wrappedWriter.statusCode >= 400 {
			s.metrics.ErrorCount++
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status codes
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// isPublicEndpoint checks if an endpoint should be publicly accessible
func (s *Server) isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/quick",
		"/health",
		"/status",
		"/status/compact",
		"/metrics",
	}

	for _, endpoint := range publicEndpoints {
		if path == endpoint {
			return true
		}
	}

	return false
}

// writeErrorResponse writes a standardized error response
func (s *Server) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := ErrorResponse{
		Error:     message,
		Timestamp: time.Now().Format(time.RFC3339),
		Code:      statusCode,
	}
	
	json.NewEncoder(w).Encode(response)
}

// logMiddleware provides basic request logging
func (s *Server) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer wrapper to capture status code
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrappedWriter, r)

		// Log the request
		duration := time.Since(start)
		fmt.Printf("[%s] %s %s %d %v\n", 
			start.Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			wrappedWriter.statusCode,
			duration,
		)
	})
}

// fastResponseMiddleware adds cache headers for fast responses
func (s *Server) fastResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add headers for fast responses
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		
		// For ultra-fast endpoints, set minimal headers
		if r.URL.Path == "/quick" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}

		next.ServeHTTP(w, r)
	})
}

// methodMiddleware ensures only allowed HTTP methods are used
func (s *Server) methodMiddleware(allowedMethods map[string][]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			method := r.Method
			
			if allowedMethods, exists := allowedMethods[path]; exists {
				methodAllowed := false
				for _, allowedMethod := range allowedMethods {
					if method == allowedMethod {
						methodAllowed = true
						break
					}
				}
				
				if !methodAllowed {
					s.writeErrorResponse(w, fmt.Sprintf("Method %s not allowed for %s", method, path), http.StatusMethodNotAllowed)
					return
				}
			}
			
			next.ServeHTTP(w, r)
		})
	}
}