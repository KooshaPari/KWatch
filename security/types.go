package security

import (
	"time"
)

// SecurityFinding represents a security issue found in the codebase
type SecurityFinding struct {
	ID          string    `json:"id"`
	File        string    `json:"file"`
	Line        int       `json:"line"`
	Column      int       `json:"column"`
	Type        string    `json:"type"`        // "api_key", "password", "jwt_token", etc.
	Severity    string    `json:"severity"`    // "critical", "high", "medium", "low"
	Message     string    `json:"message"`
	Context     string    `json:"context"`     // surrounding code lines
	Value       string    `json:"value"`       // masked secret value
	RawValue    string    `json:"-"`           // actual secret (not exposed in JSON)
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`      // "active", "resolved", "ignored"
	Rule        string    `json:"rule"`        // which detection rule triggered
	Confidence  float64   `json:"confidence"`  // confidence score 0.0-1.0
}

// SecurityScanResult represents the result of a security scan
type SecurityScanResult struct {
	Findings     []SecurityFinding `json:"findings"`
	FilesScanned int               `json:"files_scanned"`
	Duration     time.Duration     `json:"duration"`
	Timestamp    time.Time         `json:"timestamp"`
	ScanType     string            `json:"scan_type"` // "full", "incremental", "file"
}

// SecurityPattern represents a detection pattern for secrets
type SecurityPattern struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Pattern     string  `json:"pattern"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Enabled     bool    `json:"enabled"`
}

// SecurityConfig represents the security scanner configuration
type SecurityConfig struct {
	Patterns         []SecurityPattern `json:"patterns"`
	ExcludedPaths    []string          `json:"excluded_paths"`
	ExcludedFiles    []string          `json:"excluded_files"`
	MaxFileSize      int64             `json:"max_file_size"`      // in bytes
	ContextLines     int               `json:"context_lines"`      // lines of context to capture
	EnabledSeverity  []string          `json:"enabled_severity"`   // which severities to report
	HistoricalScan   bool              `json:"historical_scan"`    // scan git history
	MaxHistoryDepth  int               `json:"max_history_depth"`  // max commits to scan
	RespectGitignore bool              `json:"respect_gitignore"`  // respect .gitignore patterns
	DefaultScanMode  string            `json:"default_scan_mode"`  // default scan mode
}

// ScanOptions represents options for a security scan
type ScanOptions struct {
	Paths           []string `json:"paths"`
	IncludeHistory  bool     `json:"include_history"`
	MaxDepth        int      `json:"max_depth"`
	FilePatterns    []string `json:"file_patterns"`
	ExcludePatterns []string `json:"exclude_patterns"`
	ScanMode        string   `json:"scan_mode"`        // risky, tracked, staged, modified, comprehensive
	RespectGitignore bool    `json:"respect_gitignore"` // whether to respect .gitignore patterns
}

// SecurityStats represents statistics about security findings
type SecurityStats struct {
	TotalFindings    int            `json:"total_findings"`
	FindingsBySeverity map[string]int `json:"findings_by_severity"`
	FindingsByType   map[string]int `json:"findings_by_type"`
	FilesWithIssues  int            `json:"files_with_issues"`
	LastScanTime     time.Time      `json:"last_scan_time"`
}

// SecurityDatabase interface for storing and retrieving security findings
type SecurityDatabase interface {
	SaveFinding(finding SecurityFinding) error
	GetFindings(filters map[string]interface{}) ([]SecurityFinding, error)
	GetFindingByID(id string) (*SecurityFinding, error)
	UpdateFindingStatus(id string, status string) error
	DeleteFinding(id string) error
	GetStats() (*SecurityStats, error)
	Close() error
}

// SecurityScanner interface for scanning files and directories
type SecurityScanner interface {
	ScanFile(filepath string) (*SecurityScanResult, error)
	ScanDirectory(dirpath string, options ScanOptions) (*SecurityScanResult, error)
	ScanGitHistory(repoPath string, maxDepth int) (*SecurityScanResult, error)
	LoadConfig(configPath string) error
	GetConfig() *SecurityConfig
	AddPattern(pattern SecurityPattern) error
	RemovePattern(name string) error
}
