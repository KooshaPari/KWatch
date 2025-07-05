package runner

import (
	"regexp"
	"strconv"
	"strings"
)

// Parser handles parsing of command output to extract meaningful information
type Parser struct {
	// Regex patterns for different tools
	tscErrorPattern    *regexp.Regexp
	eslintPattern      *regexp.Regexp
	testFailPattern    *regexp.Regexp
	testPassPattern    *regexp.Regexp
	jestFailPattern    *regexp.Regexp
	bunTestPattern     *regexp.Regexp
}

// NewParser creates a new parser instance with compiled regex patterns
func NewParser() *Parser {
	return &Parser{
		// TypeScript patterns
		tscErrorPattern: regexp.MustCompile(`Found (\d+) errors?`),
		
		// ESLint patterns - matches "✖ 3 problems (1 error, 2 warnings)"
		eslintPattern: regexp.MustCompile(`✖ (\d+) problems?`),
		
		// Test patterns for various test runners
		testFailPattern:    regexp.MustCompile(`(\d+) failing`),
		testPassPattern:    regexp.MustCompile(`(\d+) passing`),
		jestFailPattern:    regexp.MustCompile(`FAIL|Failed|failed`),
		bunTestPattern:     regexp.MustCompile(`(\d+) fail`),
	}
}

// ParseTypeScriptOutput parses TypeScript compiler output
func (p *Parser) ParseTypeScriptOutput(output string) (passed bool, issueCount int) {
	// Clean the output
	output = strings.TrimSpace(output)
	
	// If output is empty, assume success
	if output == "" {
		return true, 0
	}
	
	// Look for error count pattern
	matches := p.tscErrorPattern.FindStringSubmatch(output)
	if len(matches) >= 2 {
		if count, err := strconv.Atoi(matches[1]); err == nil {
			return count == 0, count
		}
	}
	
	// Check for common success indicators
	if strings.Contains(output, "No errors found") {
		return true, 0
	}
	
	// Check for error indicators
	if strings.Contains(output, "error TS") || strings.Contains(output, "Found") {
		// Try to extract error count from lines containing "error TS"
		lines := strings.Split(output, "\n")
		errorCount := 0
		for _, line := range lines {
			if strings.Contains(line, "error TS") {
				errorCount++
			}
		}
		return errorCount == 0, errorCount
	}
	
	// If we can't determine, assume success if no obvious errors
	return !strings.Contains(strings.ToLower(output), "error"), 0
}

// ParseLintOutput parses ESLint/linter output
func (p *Parser) ParseLintOutput(output string) (passed bool, issueCount int) {
	// Clean the output
	output = strings.TrimSpace(output)
	
	// If output is empty, assume success
	if output == "" {
		return true, 0
	}
	
	// Look for ESLint summary pattern
	matches := p.eslintPattern.FindStringSubmatch(output)
	if len(matches) >= 2 {
		if count, err := strconv.Atoi(matches[1]); err == nil {
			return count == 0, count
		}
	}
	
	// Look for other common lint patterns
	if strings.Contains(output, "✓") && !strings.Contains(output, "✖") {
		return true, 0
	}
	
	// Check for Biome linter patterns
	if strings.Contains(output, "Found ") {
		// Try to extract count from "Found X errors"
		re := regexp.MustCompile(`Found (\d+) errors?`)
		matches := re.FindStringSubmatch(output)
		if len(matches) >= 2 {
			if count, err := strconv.Atoi(matches[1]); err == nil {
				return count == 0, count
			}
		}
	}
	
	// Count lines with error/warning indicators
	lines := strings.Split(output, "\n")
	issueCount = 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "error") || strings.Contains(line, "warning") {
			// Skip lines that are just headers or summaries
			if !strings.HasPrefix(line, "✖") && !strings.HasPrefix(line, "Found") {
				issueCount++
			}
		}
	}
	
	// If we found issues, use that count
	if issueCount > 0 {
		return false, issueCount
	}
	
	// Check for success indicators
	if strings.Contains(output, "No issues found") || 
	   strings.Contains(output, "0 errors") ||
	   strings.Contains(output, "All files pass linting") {
		return true, 0
	}
	
	// If we can't determine, assume failure if there's substantial output
	return len(output) < 100, 0
}

// ParseTestOutput parses test runner output (Jest, Mocha, Bun test, etc.)
func (p *Parser) ParseTestOutput(output string) (passed bool, issueCount int) {
	// Clean the output
	output = strings.TrimSpace(output)
	
	// If output is empty, assume success
	if output == "" {
		return true, 0
	}
	
	// Check for Jest/Vitest patterns
	if strings.Contains(output, "PASS") || strings.Contains(output, "FAIL") {
		return p.parseJestOutput(output)
	}
	
	// Check for Bun test patterns
	if strings.Contains(output, "bun test") {
		return p.parseBunTestOutput(output)
	}
	
	// Check for Mocha patterns
	if strings.Contains(output, "passing") || strings.Contains(output, "failing") {
		return p.parseMochaOutput(output)
	}
	
	// Generic test parsing - look for common failure indicators
	failureIndicators := []string{
		"failed",
		"error",
		"✗",
		"×",
		"FAILED",
		"ERROR",
	}
	
	for _, indicator := range failureIndicators {
		if strings.Contains(strings.ToLower(output), strings.ToLower(indicator)) {
			return false, 1
		}
	}
	
	// If no failure indicators found, assume success
	return true, 0
}

// parseJestOutput parses Jest/Vitest test output
func (p *Parser) parseJestOutput(output string) (passed bool, issueCount int) {
	// Look for test summary patterns
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		// Jest summary format: "Tests: 1 failed, 2 passed, 3 total"
		if strings.Contains(line, "Tests:") {
			if strings.Contains(line, "failed") {
				// Extract failed count
				re := regexp.MustCompile(`(\d+) failed`)
				matches := re.FindStringSubmatch(line)
				if len(matches) >= 2 {
					if count, err := strconv.Atoi(matches[1]); err == nil {
						return count == 0, count
					}
				}
			}
		}
		
		// Vitest summary format: "✓ 5 passed (2s)"
		if strings.Contains(line, "✓") && strings.Contains(line, "passed") {
			return true, 0
		}
		
		// Look for "X failing" pattern
		matches := p.testFailPattern.FindStringSubmatch(line)
		if len(matches) >= 2 {
			if count, err := strconv.Atoi(matches[1]); err == nil {
				return count == 0, count
			}
		}
	}
	
	// Check for FAIL indicators
	if p.jestFailPattern.MatchString(output) {
		return false, 1
	}
	
	return true, 0
}

// parseBunTestOutput parses Bun test output
func (p *Parser) parseBunTestOutput(output string) (passed bool, issueCount int) {
	// Bun test output format: "2 pass, 1 fail"
	matches := p.bunTestPattern.FindStringSubmatch(output)
	if len(matches) >= 2 {
		if count, err := strconv.Atoi(matches[1]); err == nil {
			return count == 0, count
		}
	}
	
	// Look for pass/fail indicators
	if strings.Contains(output, "pass") && !strings.Contains(output, "fail") {
		return true, 0
	}
	
	return !strings.Contains(output, "fail"), 0
}

// parseMochaOutput parses Mocha test output
func (p *Parser) parseMochaOutput(output string) (passed bool, issueCount int) {
	// Look for failing count
	matches := p.testFailPattern.FindStringSubmatch(output)
	if len(matches) >= 2 {
		if count, err := strconv.Atoi(matches[1]); err == nil {
			return count == 0, count
		}
	}
	
	// Look for passing only
	if strings.Contains(output, "passing") && !strings.Contains(output, "failing") {
		return true, 0
	}
	
	return !strings.Contains(output, "failing"), 0
}

// ParseGenericOutput provides a fallback parser for unknown command types
func (p *Parser) ParseGenericOutput(output string) (passed bool, issueCount int) {
	// Clean the output
	output = strings.TrimSpace(output)
	
	// If output is empty, assume success
	if output == "" {
		return true, 0
	}
	
	// Look for common failure indicators
	failureWords := []string{"error", "failed", "fail", "✗", "×", "✖"}
	successWords := []string{"success", "passed", "pass", "✓", "ok"}
	
	outputLower := strings.ToLower(output)
	
	hasFailure := false
	hasSuccess := false
	
	for _, word := range failureWords {
		if strings.Contains(outputLower, word) {
			hasFailure = true
			break
		}
	}
	
	for _, word := range successWords {
		if strings.Contains(outputLower, word) {
			hasSuccess = true
			break
		}
	}
	
	// If we have both, prefer failure
	if hasFailure {
		return false, 1
	}
	
	if hasSuccess {
		return true, 0
	}
	
	// If neither, assume success for short output, failure for long output
	return len(output) < 200, 0
}