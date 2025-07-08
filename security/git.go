package security

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitRepository provides git-related functionality for security scanning
type GitRepository struct {
	rootPath string
	isGitRepo bool
}

// NewGitRepository creates a new GitRepository instance
func NewGitRepository(path string) *GitRepository {
	repo := &GitRepository{
		rootPath: path,
	}
	repo.isGitRepo = repo.checkIsGitRepository()
	return repo
}

// IsGitRepository checks if the current directory is a git repository
func (g *GitRepository) IsGitRepository() bool {
	return g.isGitRepo
}

// checkIsGitRepository checks if we're in a git repository
func (g *GitRepository) checkIsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = g.rootPath
	err := cmd.Run()
	return err == nil
}

// GetTrackedFiles returns all files tracked by git
func (g *GitRepository) GetTrackedFiles() ([]string, error) {
	if !g.isGitRepo {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "ls-files")
	cmd.Dir = g.rootPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get tracked files: %w", err)
	}

	var files []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		file := strings.TrimSpace(scanner.Text())
		if file != "" {
			// Convert to absolute path
			absPath := filepath.Join(g.rootPath, file)
			files = append(files, absPath)
		}
	}

	return files, nil
}

// GetStagedFiles returns all files currently staged for commit
func (g *GitRepository) GetStagedFiles() ([]string, error) {
	if !g.isGitRepo {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = g.rootPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}

	var files []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		file := strings.TrimSpace(scanner.Text())
		if file != "" {
			// Convert to absolute path
			absPath := filepath.Join(g.rootPath, file)
			files = append(files, absPath)
		}
	}

	return files, nil
}

// IsIgnored checks if a file is ignored by git
func (g *GitRepository) IsIgnored(filePath string) bool {
	if !g.isGitRepo {
		return false
	}

	// Convert to relative path from git root
	relPath, err := filepath.Rel(g.rootPath, filePath)
	if err != nil {
		return false
	}

	cmd := exec.Command("git", "check-ignore", relPath)
	cmd.Dir = g.rootPath
	err = cmd.Run()
	// git check-ignore returns 0 if file is ignored, 1 if not ignored
	return err == nil
}

// GetUntrackedFiles returns untracked files that are not ignored
func (g *GitRepository) GetUntrackedFiles() ([]string, error) {
	if !g.isGitRepo {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = g.rootPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get untracked files: %w", err)
	}

	var files []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		file := strings.TrimSpace(scanner.Text())
		if file != "" {
			// Convert to absolute path
			absPath := filepath.Join(g.rootPath, file)
			files = append(files, absPath)
		}
	}

	return files, nil
}

// GetRiskyFiles returns files that could potentially be committed (tracked + untracked non-ignored)
func (g *GitRepository) GetRiskyFiles() ([]string, error) {
	var allFiles []string

	// Get tracked files
	trackedFiles, err := g.GetTrackedFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get tracked files: %w", err)
	}
	allFiles = append(allFiles, trackedFiles...)

	// Get untracked files that aren't ignored
	untrackedFiles, err := g.GetUntrackedFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get untracked files: %w", err)
	}
	allFiles = append(allFiles, untrackedFiles...)

	return allFiles, nil
}

// GetModifiedFiles returns files that have been modified (staged + unstaged changes)
func (g *GitRepository) GetModifiedFiles() ([]string, error) {
	if !g.isGitRepo {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	cmd.Dir = g.rootPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get modified files: %w", err)
	}

	var files []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		file := strings.TrimSpace(scanner.Text())
		if file != "" {
			// Convert to absolute path
			absPath := filepath.Join(g.rootPath, file)
			files = append(files, absPath)
		}
	}

	// Also get staged files
	stagedFiles, err := g.GetStagedFiles()
	if err == nil {
		files = append(files, stagedFiles...)
	}

	// Remove duplicates
	fileMap := make(map[string]bool)
	var uniqueFiles []string
	for _, file := range files {
		if !fileMap[file] {
			fileMap[file] = true
			uniqueFiles = append(uniqueFiles, file)
		}
	}

	return uniqueFiles, nil
}

// ScanMode represents different scanning modes
type ScanMode string

const (
	ScanModeRisky        ScanMode = "risky"        // Tracked + untracked non-ignored files (default)
	ScanModeTracked      ScanMode = "tracked"      // Only git-tracked files
	ScanModeStaged       ScanMode = "staged"       // Only staged files
	ScanModeModified     ScanMode = "modified"     // Only modified files (staged + unstaged)
	ScanModeComprehensive ScanMode = "comprehensive" // All files including ignored
)

// GetFilesForScanMode returns files based on the specified scan mode
func (g *GitRepository) GetFilesForScanMode(mode ScanMode) ([]string, error) {
	switch mode {
	case ScanModeRisky:
		return g.GetRiskyFiles()
	case ScanModeTracked:
		return g.GetTrackedFiles()
	case ScanModeStaged:
		return g.GetStagedFiles()
	case ScanModeModified:
		return g.GetModifiedFiles()
	case ScanModeComprehensive:
		// Fall back to directory walking for comprehensive scan
		return nil, nil
	default:
		return g.GetRiskyFiles()
	}
}
