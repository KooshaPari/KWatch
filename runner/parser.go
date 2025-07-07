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

// TestResult represents detailed test execution results
type TestResult struct {
	Passed      bool
	TotalTests  int
	PassedTests int
	FailedTests int
}

// ParseTestOutput parses test runner output (Jest, Mocha, Bun test, etc.)
func (p *Parser) ParseTestOutput(output string) TestResult {
	// Clean the output
	output = strings.TrimSpace(output)
	
	// If output is empty, assume success
	if output == "" {
		return TestResult{Passed: true, TotalTests: 0, PassedTests: 0, FailedTests: 0}
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
			return TestResult{Passed: false, TotalTests: 1, PassedTests: 0, FailedTests: 1}
		}
	}
	
	// If no failure indicators found, assume success
	return TestResult{Passed: true, TotalTests: 1, PassedTests: 1, FailedTests: 0}
}

// parseJestOutput parses Jest/Vitest test output
func (p *Parser) parseJestOutput(output string) TestResult {
	result := TestResult{Passed: true, TotalTests: 0, PassedTests: 0, FailedTests: 0}
	
	// Look for test summary patterns
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		// Jest summary format: "Tests: 1 failed, 2 passed, 3 total"
		if strings.Contains(line, "Tests:") {
			// Extract failed count
			if strings.Contains(line, "failed") {
				re := regexp.MustCompile(`(\d+) failed`)
				matches := re.FindStringSubmatch(line)
				if len(matches) >= 2 {
					if count, err := strconv.Atoi(matches[1]); err == nil {
						result.FailedTests = count
					}
				}
			}
			
			// Extract passed count
			if strings.Contains(line, "passed") {
				re := regexp.MustCompile(`(\d+) passed`)
				matches := re.FindStringSubmatch(line)
				if len(matches) >= 2 {
					if count, err := strconv.Atoi(matches[1]); err == nil {
						result.PassedTests = count
					}
				}
			}
			
			// Extract total count
			if strings.Contains(line, "total") {
				re := regexp.MustCompile(`(\d+) total`)
				matches := re.FindStringSubmatch(line)
				if len(matches) >= 2 {
					if count, err := strconv.Atoi(matches[1]); err == nil {
						result.TotalTests = count
					}
				}
			}
		}
		
		// Vitest summary format: "✓ 5 passed (2s)"
		if strings.Contains(line, "✓") && strings.Contains(line, "passed") {
			re := regexp.MustCompile(`✓ (\d+) passed`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 2 {
				if count, err := strconv.Atoi(matches[1]); err == nil {
					result.PassedTests = count
					result.TotalTests = count
				}
			}
		}
		
		// Look for "X failing" pattern
		matches := p.testFailPattern.FindStringSubmatch(line)
		if len(matches) >= 2 {
			if count, err := strconv.Atoi(matches[1]); err == nil {
				result.FailedTests = count
			}
		}
		
		// Look for "X passing" pattern
		matches = p.testPassPattern.FindStringSubmatch(line)
		if len(matches) >= 2 {
			if count, err := strconv.Atoi(matches[1]); err == nil {
				result.PassedTests = count
			}
		}
	}
	
	// Calculate total if not found but we have pass/fail counts
	if result.TotalTests == 0 && (result.PassedTests > 0 || result.FailedTests > 0) {
		result.TotalTests = result.PassedTests + result.FailedTests
	}
	
	// Check for FAIL indicators
	if p.jestFailPattern.MatchString(output) && result.FailedTests == 0 {
		result.FailedTests = 1
		result.TotalTests = 1
	}
	
	result.Passed = result.FailedTests == 0
	return result
}

// parseBunTestOutput parses Bun test output
func (p *Parser) parseBunTestOutput(output string) TestResult {
	result := TestResult{Passed: true, TotalTests: 0, PassedTests: 0, FailedTests: 0}
	
	// Bun test output format: "2 pass, 1 fail"
	failMatches := p.bunTestPattern.FindStringSubmatch(output)
	if len(failMatches) >= 2 {
		if count, err := strconv.Atoi(failMatches[1]); err == nil {
			result.FailedTests = count
		}
	}
	
	// Look for pass count
	passRe := regexp.MustCompile(`(\d+) pass`)
	passMatches := passRe.FindStringSubmatch(output)
	if len(passMatches) >= 2 {
		if count, err := strconv.Atoi(passMatches[1]); err == nil {
			result.PassedTests = count
		}
	}
	
	// Calculate total
	result.TotalTests = result.PassedTests + result.FailedTests
	
	// Look for pass/fail indicators if no specific counts found
	if result.TotalTests == 0 {
		if strings.Contains(output, "pass") && !strings.Contains(output, "fail") {
			result.PassedTests = 1
			result.TotalTests = 1
		} else if strings.Contains(output, "fail") {
			result.FailedTests = 1
			result.TotalTests = 1
		}
	}
	
	result.Passed = result.FailedTests == 0
	return result
}

// parseMochaOutput parses Mocha test output
func (p *Parser) parseMochaOutput(output string) TestResult {
	result := TestResult{Passed: true, TotalTests: 0, PassedTests: 0, FailedTests: 0}
	
	// Look for failing count
	failMatches := p.testFailPattern.FindStringSubmatch(output)
	if len(failMatches) >= 2 {
		if count, err := strconv.Atoi(failMatches[1]); err == nil {
			result.FailedTests = count
		}
	}
	
	// Look for passing count
	passMatches := p.testPassPattern.FindStringSubmatch(output)
	if len(passMatches) >= 2 {
		if count, err := strconv.Atoi(passMatches[1]); err == nil {
			result.PassedTests = count
		}
	}
	
	// Calculate total
	result.TotalTests = result.PassedTests + result.FailedTests
	
	// Look for passing only if no specific counts found
	if result.TotalTests == 0 {
		if strings.Contains(output, "passing") && !strings.Contains(output, "failing") {
			result.PassedTests = 1
			result.TotalTests = 1
		} else if strings.Contains(output, "failing") {
			result.FailedTests = 1
			result.TotalTests = 1
		}
	}
	
	result.Passed = result.FailedTests == 0
	return result
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