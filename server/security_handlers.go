package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"kwatch/security"
)

// SecurityAPI provides HTTP handlers for security operations
type SecurityAPI struct {
	scanner  security.SecurityScanner
	database security.SecurityDatabase
}

// NewSecurityAPI creates a new security API instance
func NewSecurityAPI(dbPath string) *SecurityAPI {
	db := security.NewFileDatabase(dbPath)
	scanner := security.NewScanner(db)
	
	return &SecurityAPI{
		scanner:  scanner,
		database: db,
	}
}

// SecurityScanRequest represents a security scan request
type SecurityScanRequest struct {
	Path             string   `json:"path"`
	IncludeHistory   bool     `json:"include_history"`
	MaxDepth         int      `json:"max_depth"`
	Severity         []string `json:"severity"`
	ScanMode         string   `json:"scan_mode"`         // risky, tracked, staged, modified, comprehensive
	RespectGitignore bool     `json:"respect_gitignore"` // whether to respect .gitignore patterns
}

// SecurityFindingResponse represents a security finding response
type SecurityFindingResponse struct {
	*security.SecurityFinding
	ContextLines []string `json:"context_lines"`
}

// SecurityScanResponse represents a security scan response
type SecurityScanResponse struct {
	Findings     []SecurityFindingResponse `json:"findings"`
	FilesScanned int                       `json:"files_scanned"`
	Duration     string                    `json:"duration"`
	Timestamp    string                    `json:"timestamp"`
	ScanType     string                    `json:"scan_type"`
	Summary      SecuritySummary           `json:"summary"`
}

// SecuritySummary provides a summary of security findings
type SecuritySummary struct {
	TotalFindings      int            `json:"total_findings"`
	FindingsBySeverity map[string]int `json:"findings_by_severity"`
	FindingsByType     map[string]int `json:"findings_by_type"`
	CriticalCount      int            `json:"critical_count"`
	HighCount          int            `json:"high_count"`
	MediumCount        int            `json:"medium_count"`
	LowCount           int            `json:"low_count"`
}

// HandleSecurityScan handles POST /security/scan requests
func (api *SecurityAPI) HandleSecurityScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SecurityScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Default values
	if req.Path == "" {
		req.Path = "."
	}
	if req.MaxDepth == 0 {
		req.MaxDepth = 100
	}
	if req.ScanMode == "" {
		req.ScanMode = "risky"
	}

	// Get absolute path
	absPath, err := filepath.Abs(req.Path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid path: %v", err), http.StatusBadRequest)
		return
	}

	// Prepare scan options
	options := security.ScanOptions{
		Paths:            []string{absPath},
		IncludeHistory:   req.IncludeHistory,
		MaxDepth:         req.MaxDepth,
		ScanMode:         req.ScanMode,
		RespectGitignore: req.RespectGitignore,
	}

	// Run the scan
	var result *security.SecurityScanResult
	
	// Check if path is a file or directory
	if strings.Contains(absPath, ".") && !strings.HasSuffix(absPath, "/") {
		result, err = api.scanner.ScanFile(absPath)
	} else {
		result, err = api.scanner.ScanDirectory(absPath, options)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Scan failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Filter by severity if specified
	if len(req.Severity) > 0 {
		result.Findings = filterBySeverity(result.Findings, req.Severity)
	}

	// Convert to response format
	response := convertToSecurityResponse(result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSecurityFindings handles GET /security/findings requests
func (api *SecurityAPI) HandleSecurityFindings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	filters := make(map[string]interface{})

	if severity := query.Get("severity"); severity != "" {
		filters["severity"] = severity
	}
	if findingType := query.Get("type"); findingType != "" {
		filters["type"] = findingType
	}
	if status := query.Get("status"); status != "" {
		filters["status"] = status
	}
	if file := query.Get("file"); file != "" {
		filters["file"] = file
	}
	if minConfidence := query.Get("min_confidence"); minConfidence != "" {
		if conf, err := strconv.ParseFloat(minConfidence, 64); err == nil {
			filters["min_confidence"] = conf
		}
	}

	// Get findings from database
	findings, err := api.database.GetFindings(filters)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve findings: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var responseFindings []SecurityFindingResponse
	for _, finding := range findings {
		responseFindings = append(responseFindings, SecurityFindingResponse{
			SecurityFinding: &finding,
			ContextLines:    strings.Split(finding.Context, "\n"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseFindings)
}

// HandleSecurityFinding handles GET /security/findings/{id} requests
func (api *SecurityAPI) HandleSecurityFinding(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract finding ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/security/findings/")
	if path == "" {
		http.Error(w, "Finding ID required", http.StatusBadRequest)
		return
	}

	// Get finding from database
	finding, err := api.database.GetFindingByID(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Finding not found: %v", err), http.StatusNotFound)
		return
	}

	// Convert to response format
	response := SecurityFindingResponse{
		SecurityFinding: finding,
		ContextLines:    strings.Split(finding.Context, "\n"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSecurityStats handles GET /security/stats requests
func (api *SecurityAPI) HandleSecurityStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get statistics from database
	stats, err := api.database.GetStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve stats: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleSecurityResolve handles POST /security/resolve/{id} requests
func (api *SecurityAPI) HandleSecurityResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract finding ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/security/resolve/")
	if path == "" {
		http.Error(w, "Finding ID required", http.StatusBadRequest)
		return
	}

	// Update finding status
	if err := api.database.UpdateFindingStatus(path, "resolved"); err != nil {
		http.Error(w, fmt.Sprintf("Failed to resolve finding: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Finding %s marked as resolved", path),
	})
}

// HandleSecurityIgnore handles POST /security/ignore/{id} requests
func (api *SecurityAPI) HandleSecurityIgnore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract finding ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/security/ignore/")
	if path == "" {
		http.Error(w, "Finding ID required", http.StatusBadRequest)
		return
	}

	// Update finding status
	if err := api.database.UpdateFindingStatus(path, "ignored"); err != nil {
		http.Error(w, fmt.Sprintf("Failed to ignore finding: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Finding %s marked as ignored", path),
	})
}

// Helper functions

func filterBySeverity(findings []security.SecurityFinding, severities []string) []security.SecurityFinding {
	severityMap := make(map[string]bool)
	for _, s := range severities {
		severityMap[s] = true
	}

	var filtered []security.SecurityFinding
	for _, finding := range findings {
		if severityMap[finding.Severity] {
			filtered = append(filtered, finding)
		}
	}

	return filtered
}

func convertToSecurityResponse(result *security.SecurityScanResult) SecurityScanResponse {
	// Convert findings
	var responseFindings []SecurityFindingResponse
	for _, finding := range result.Findings {
		responseFindings = append(responseFindings, SecurityFindingResponse{
			SecurityFinding: &finding,
			ContextLines:    strings.Split(finding.Context, "\n"),
		})
	}

	// Calculate summary
	summary := SecuritySummary{
		TotalFindings:      len(result.Findings),
		FindingsBySeverity: make(map[string]int),
		FindingsByType:     make(map[string]int),
	}

	for _, finding := range result.Findings {
		summary.FindingsBySeverity[finding.Severity]++
		summary.FindingsByType[finding.Type]++

		switch finding.Severity {
		case "critical":
			summary.CriticalCount++
		case "high":
			summary.HighCount++
		case "medium":
			summary.MediumCount++
		case "low":
			summary.LowCount++
		}
	}

	return SecurityScanResponse{
		Findings:     responseFindings,
		FilesScanned: result.FilesScanned,
		Duration:     result.Duration.String(),
		Timestamp:    result.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		ScanType:     result.ScanType,
		Summary:      summary,
	}
}
