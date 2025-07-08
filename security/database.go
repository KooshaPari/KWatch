package security

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

// MemoryDatabase implements SecurityDatabase interface using in-memory storage
type MemoryDatabase struct {
	findings map[string]SecurityFinding
	mutex    sync.RWMutex
	filePath string // Optional file path for persistence
}

// NewMemoryDatabase creates a new in-memory database
func NewMemoryDatabase(filePath string) *MemoryDatabase {
	db := &MemoryDatabase{
		findings: make(map[string]SecurityFinding),
		filePath: filePath,
	}
	
	// Load existing data if file exists
	if filePath != "" {
		db.loadFromFile()
	}
	
	return db
}

// SaveFinding saves a security finding to the database
func (db *MemoryDatabase) SaveFinding(finding SecurityFinding) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	db.findings[finding.ID] = finding
	
	// Persist to file if configured
	if db.filePath != "" {
		return db.saveToFile()
	}
	
	return nil
}

// GetFindings retrieves security findings based on filters
func (db *MemoryDatabase) GetFindings(filters map[string]interface{}) ([]SecurityFinding, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	
	var results []SecurityFinding
	
	for _, finding := range db.findings {
		if db.matchesFilters(finding, filters) {
			results = append(results, finding)
		}
	}
	
	return results, nil
}

// GetFindingByID retrieves a specific finding by ID
func (db *MemoryDatabase) GetFindingByID(id string) (*SecurityFinding, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	
	finding, exists := db.findings[id]
	if !exists {
		return nil, fmt.Errorf("finding with ID %s not found", id)
	}
	
	return &finding, nil
}

// UpdateFindingStatus updates the status of a finding
func (db *MemoryDatabase) UpdateFindingStatus(id string, status string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	finding, exists := db.findings[id]
	if !exists {
		return fmt.Errorf("finding with ID %s not found", id)
	}
	
	finding.Status = status
	db.findings[id] = finding
	
	// Persist to file if configured
	if db.filePath != "" {
		return db.saveToFile()
	}
	
	return nil
}

// DeleteFinding removes a finding from the database
func (db *MemoryDatabase) DeleteFinding(id string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	delete(db.findings, id)
	
	// Persist to file if configured
	if db.filePath != "" {
		return db.saveToFile()
	}
	
	return nil
}

// GetStats returns statistics about security findings
func (db *MemoryDatabase) GetStats() (*SecurityStats, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	
	stats := &SecurityStats{
		FindingsBySeverity: make(map[string]int),
		FindingsByType:     make(map[string]int),
	}
	
	filesWithIssues := make(map[string]bool)
	var lastScanTime time.Time
	
	for _, finding := range db.findings {
		stats.TotalFindings++
		stats.FindingsBySeverity[finding.Severity]++
		stats.FindingsByType[finding.Type]++
		filesWithIssues[finding.File] = true
		
		if finding.Timestamp.After(lastScanTime) {
			lastScanTime = finding.Timestamp
		}
	}
	
	stats.FilesWithIssues = len(filesWithIssues)
	stats.LastScanTime = lastScanTime
	
	return stats, nil
}

// Close closes the database connection
func (db *MemoryDatabase) Close() error {
	// Save to file one last time if configured
	if db.filePath != "" {
		return db.saveToFile()
	}
	return nil
}

// Helper methods

func (db *MemoryDatabase) matchesFilters(finding SecurityFinding, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "severity":
			if finding.Severity != value.(string) {
				return false
			}
		case "type":
			if finding.Type != value.(string) {
				return false
			}
		case "status":
			if finding.Status != value.(string) {
				return false
			}
		case "file":
			if finding.File != value.(string) {
				return false
			}
		case "min_confidence":
			if finding.Confidence < value.(float64) {
				return false
			}
		}
	}
	return true
}

func (db *MemoryDatabase) saveToFile() error {
	// Convert findings map to slice for JSON serialization
	var findingsList []SecurityFinding
	for _, finding := range db.findings {
		findingsList = append(findingsList, finding)
	}
	
	data, err := json.MarshalIndent(findingsList, "", "  ")
	if err != nil {
		return err
	}
	
	return ioutil.WriteFile(db.filePath, data, 0644)
}

func (db *MemoryDatabase) loadFromFile() error {
	if _, err := os.Stat(db.filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, start with empty database
	}
	
	data, err := ioutil.ReadFile(db.filePath)
	if err != nil {
		return err
	}
	
	var findingsList []SecurityFinding
	if err := json.Unmarshal(data, &findingsList); err != nil {
		return err
	}
	
	// Convert slice back to map
	for _, finding := range findingsList {
		db.findings[finding.ID] = finding
	}
	
	return nil
}

// FileDatabase implements SecurityDatabase interface using file-based storage
type FileDatabase struct {
	*MemoryDatabase
}

// NewFileDatabase creates a new file-based database
func NewFileDatabase(filePath string) *FileDatabase {
	return &FileDatabase{
		MemoryDatabase: NewMemoryDatabase(filePath),
	}
}
