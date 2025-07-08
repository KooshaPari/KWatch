package security

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Scanner implements the SecurityScanner interface
type Scanner struct {
	config           *SecurityConfig
	patterns         []SecurityPattern
	compiledPatterns map[string]*regexp.Regexp
	database         SecurityDatabase
}

// NewScanner creates a new security scanner instance
func NewScanner(db SecurityDatabase) *Scanner {
	scanner := &Scanner{
		database: db,
		config:   DefaultConfig(),
		patterns: DefaultSecurityPatterns(),
	}
	
	// Compile patterns
	compiled, err := CompilePatterns(scanner.patterns)
	if err != nil {
		// Log error but continue with empty patterns
		compiled = make(map[string]*regexp.Regexp)
	}
	scanner.compiledPatterns = compiled
	
	return scanner
}

// DefaultConfig returns default security configuration
func DefaultConfig() *SecurityConfig {
	return &SecurityConfig{
		Patterns:        DefaultSecurityPatterns(),
		ExcludedPaths:   []string{"node_modules", ".git", "vendor", "dist", "build"},
		ExcludedFiles:   []string{"*.log", "*.tmp", "*.cache", ".security-findings.json", "security-config.json"},
		MaxFileSize:     10 * 1024 * 1024, // 10MB
		ContextLines:    3,
		EnabledSeverity: []string{"critical", "high", "medium", "low"},
		HistoricalScan:  false,
		MaxHistoryDepth: 100,
		RespectGitignore: true,
		DefaultScanMode: "risky",
	}
}

// LoadConfig loads configuration from a file
func (s *Scanner) LoadConfig(configPath string) error {
	// TODO: Implement config file loading
	return nil
}

// GetConfig returns the current configuration
func (s *Scanner) GetConfig() *SecurityConfig {
	return s.config
}

// AddPattern adds a new security pattern
func (s *Scanner) AddPattern(pattern SecurityPattern) error {
	s.patterns = append(s.patterns, pattern)
	
	// Recompile patterns
	compiled, err := CompilePatterns(s.patterns)
	if err != nil {
		return err
	}
	s.compiledPatterns = compiled
	
	return nil
}

// RemovePattern removes a security pattern by name
func (s *Scanner) RemovePattern(name string) error {
	for i, pattern := range s.patterns {
		if pattern.Name == name {
			s.patterns = append(s.patterns[:i], s.patterns[i+1:]...)
			break
		}
	}
	
	// Recompile patterns
	compiled, err := CompilePatterns(s.patterns)
	if err != nil {
		return err
	}
	s.compiledPatterns = compiled
	
	return nil
}

// ScanFile scans a single file for security issues
func (s *Scanner) ScanFile(filePath string) (*SecurityScanResult, error) {
	startTime := time.Now()

	// Check if file should be excluded
	if s.shouldExcludeFile(filePath) {
		return &SecurityScanResult{
			Findings:     []SecurityFinding{},
			FilesScanned: 0,
			Duration:     time.Since(startTime),
			Timestamp:    startTime,
			ScanType:     "file",
		}, nil
	}

	// Check file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	if fileInfo.Size() > s.config.MaxFileSize {
		return &SecurityScanResult{
			Findings:     []SecurityFinding{},
			FilesScanned: 0,
			Duration:     time.Since(startTime),
			Timestamp:    startTime,
			ScanType:     "file",
		}, nil
	}

	// Read file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Scan content for patterns
	findings := s.scanContent(string(content), filePath)
	
	// Save findings to database
	for _, finding := range findings {
		if err := s.database.SaveFinding(finding); err != nil {
			// Log error but continue
		}
	}
	
	return &SecurityScanResult{
		Findings:     findings,
		FilesScanned: 1,
		Duration:     time.Since(startTime),
		Timestamp:    startTime,
		ScanType:     "file",
	}, nil
}

// ScanDirectory scans a directory for security issues
func (s *Scanner) ScanDirectory(dirpath string, options ScanOptions) (*SecurityScanResult, error) {
	startTime := time.Now()
	var allFindings []SecurityFinding
	filesScanned := 0

	// Initialize git repository
	gitRepo := NewGitRepository(dirpath)

	// Determine scan mode
	scanMode := ScanMode(options.ScanMode)
	if scanMode == "" {
		scanMode = ScanMode(s.config.DefaultScanMode)
	}

	var filesToScan []string
	var err error

	// Get files based on scan mode and git awareness
	if gitRepo.IsGitRepository() && s.config.RespectGitignore && scanMode != ScanModeComprehensive {
		filesToScan, err = gitRepo.GetFilesForScanMode(scanMode)
		if err != nil {
			return nil, fmt.Errorf("failed to get files for scan mode %s: %w", scanMode, err)
		}
	} else {
		// Fall back to directory walking for non-git repos or comprehensive scans
		err = filepath.Walk(dirpath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continue on errors
			}

			if info.IsDir() {
				return nil
			}

			// For comprehensive scans or non-git repos, check exclusions
			if !s.config.RespectGitignore || scanMode == ScanModeComprehensive {
				if s.shouldExcludeFile(path) {
					return nil
				}
			}

			filesToScan = append(filesToScan, path)
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	// Scan each file
	for _, filePath := range filesToScan {
		// Additional exclusion check for git-aware scans
		if gitRepo.IsGitRepository() && s.config.RespectGitignore && scanMode != ScanModeComprehensive {
			if s.shouldExcludeFile(filePath) {
				continue
			}
		}

		result, err := s.ScanFile(filePath)
		if err != nil {
			continue // Continue on errors
		}

		if result.FilesScanned > 0 {
			filesScanned++
			allFindings = append(allFindings, result.Findings...)
		}
	}

	return &SecurityScanResult{
		Findings:     allFindings,
		FilesScanned: filesScanned,
		Duration:     time.Since(startTime),
		Timestamp:    startTime,
		ScanType:     fmt.Sprintf("directory-%s", scanMode),
	}, nil
}

// ScanGitHistory scans git history for security issues
func (s *Scanner) ScanGitHistory(repoPath string, maxDepth int) (*SecurityScanResult, error) {
	startTime := time.Now()
	
	// TODO: Implement git history scanning
	// This would involve:
	// 1. Using git log to get commit history
	// 2. For each commit, get the diff
	// 3. Scan the diff content for secrets
	// 4. Track when secrets were introduced/removed
	
	return &SecurityScanResult{
		Findings:     []SecurityFinding{},
		FilesScanned: 0,
		Duration:     time.Since(startTime),
		Timestamp:    startTime,
		ScanType:     "history",
	}, nil
}

// scanContent scans file content for security patterns
func (s *Scanner) scanContent(content, filePath string) []SecurityFinding {
	var findings []SecurityFinding
	lines := strings.Split(content, "\n")

	for patternName, regex := range s.compiledPatterns {
		pattern := s.getPatternByName(patternName)
		if pattern == nil {
			continue
		}

		// Check if severity is enabled
		if !s.isSeverityEnabled(pattern.Severity) {
			continue
		}

		for lineNum, line := range lines {
			matches := regex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					finding := SecurityFinding{
						ID:          generateFindingID(filePath, lineNum, patternName),
						File:        filePath,
						Line:        lineNum + 1,
						Column:      strings.Index(line, match[1]) + 1,
						Type:        pattern.Type,
						Severity:    pattern.Severity,
						Message:     pattern.Description,
						Context:     s.getContext(lines, lineNum),
						Value:       s.maskSecret(match[1]),
						RawValue:    match[1],
						Timestamp:   time.Now(),
						Status:      "active",
						Rule:        pattern.Name,
						Confidence:  pattern.Confidence,
					}
					findings = append(findings, finding)
				}
			}
		}
	}

	return findings
}

// Helper functions

func (s *Scanner) shouldExcludeFile(filePath string) bool {
	// Check excluded paths
	for _, excludedPath := range s.config.ExcludedPaths {
		if strings.Contains(filePath, excludedPath) {
			return true
		}
	}

	// Check excluded file patterns
	for _, pattern := range s.config.ExcludedFiles {
		matched, _ := filepath.Match(pattern, filepath.Base(filePath))
		if matched {
			return true
		}
	}

	return false
}

func (s *Scanner) getPatternByName(name string) *SecurityPattern {
	for _, pattern := range s.patterns {
		if pattern.Name == name {
			return &pattern
		}
	}
	return nil
}

func (s *Scanner) isSeverityEnabled(severity string) bool {
	for _, enabled := range s.config.EnabledSeverity {
		if enabled == severity {
			return true
		}
	}
	return false
}

func (s *Scanner) getContext(lines []string, lineNum int) string {
	start := lineNum - s.config.ContextLines
	end := lineNum + s.config.ContextLines + 1
	
	if start < 0 {
		start = 0
	}
	if end > len(lines) {
		end = len(lines)
	}
	
	contextLines := lines[start:end]
	return strings.Join(contextLines, "\n")
}

func (s *Scanner) maskSecret(secret string) string {
	if len(secret) <= 8 {
		return strings.Repeat("*", len(secret))
	}
	
	// Show first 4 and last 4 characters
	return secret[:4] + strings.Repeat("*", len(secret)-8) + secret[len(secret)-4:]
}

func generateFindingID(filePath string, lineNum int, patternName string) string {
	data := fmt.Sprintf("%s:%d:%s", filePath, lineNum, patternName)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)[:16]
}
