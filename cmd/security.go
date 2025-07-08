package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"kwatch/security"
)

var (
	securityOutputFormat string
	securitySeverity     []string
	securityIncludeHistory bool
	securityMaxDepth     int
	securityConfigFile   string
	securityDatabaseFile string
	securityScanMode     string
	securityRespectGitignore bool
)

var securityCmd = &cobra.Command{
	Use:   "security [path]",
	Short: "Run security scans to detect secrets and vulnerabilities",
	Long: `Run security scans on files and directories to detect:
- API keys and tokens
- Database connection strings
- Private keys and certificates
- Passwords and secrets
- JWT tokens
- Webhook URLs

Examples:
  kwatch security .                    # Scan risky files (tracked + untracked non-ignored)
  kwatch security src/                 # Scan src directory
  kwatch security --format json       # Output in JSON format
  kwatch security --severity critical # Only show critical issues
  kwatch security --mode staged       # Only scan staged files
  kwatch security --mode tracked      # Only scan git-tracked files
  kwatch security --mode comprehensive # Scan all files including ignored
  kwatch security --no-gitignore      # Don't respect .gitignore patterns`,
	Args: cobra.MaximumNArgs(1),
	Run:  runSecurityScan,
}

var securityListCmd = &cobra.Command{
	Use:   "list",
	Short: "List security findings",
	Long:  "List all security findings from previous scans",
	Run:   runSecurityList,
}

var securityStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show security statistics",
	Long:  "Show statistics about security findings",
	Run:   runSecurityStats,
}

var securityResolveCmd = &cobra.Command{
	Use:   "resolve [finding-id]",
	Short: "Mark a security finding as resolved",
	Long:  "Mark a specific security finding as resolved",
	Args:  cobra.ExactArgs(1),
	Run:   runSecurityResolve,
}

var securityIgnoreCmd = &cobra.Command{
	Use:   "ignore [finding-id]",
	Short: "Mark a security finding as ignored",
	Long:  "Mark a specific security finding as ignored",
	Args:  cobra.ExactArgs(1),
	Run:   runSecurityIgnore,
}

func init() {
	rootCmd.AddCommand(securityCmd)
	securityCmd.AddCommand(securityListCmd)
	securityCmd.AddCommand(securityStatsCmd)
	securityCmd.AddCommand(securityResolveCmd)
	securityCmd.AddCommand(securityIgnoreCmd)

	// Security scan flags
	securityCmd.Flags().StringVarP(&securityOutputFormat, "format", "f", "table", "Output format (table, json, csv)")
	securityCmd.Flags().StringSliceVarP(&securitySeverity, "severity", "s", []string{}, "Filter by severity (critical, high, medium, low)")
	securityCmd.Flags().StringVarP(&securityScanMode, "mode", "m", "risky", "Scan mode (risky, tracked, staged, modified, comprehensive)")
	securityCmd.Flags().BoolVar(&securityRespectGitignore, "gitignore", true, "Respect .gitignore patterns")
	securityCmd.Flags().BoolVar(&securityIncludeHistory, "history", false, "Include git history scan")
	securityCmd.Flags().IntVar(&securityMaxDepth, "max-depth", 100, "Maximum git history depth to scan")
	securityCmd.Flags().StringVar(&securityConfigFile, "config", "", "Security configuration file")
	securityCmd.Flags().StringVar(&securityDatabaseFile, "database", ".security-findings.json", "Security findings database file")

	// Security list flags
	securityListCmd.Flags().StringVarP(&securityOutputFormat, "format", "f", "table", "Output format (table, json, csv)")
	securityListCmd.Flags().StringSliceVarP(&securitySeverity, "severity", "s", []string{}, "Filter by severity")
	securityListCmd.Flags().StringVar(&securityDatabaseFile, "database", ".security-findings.json", "Security findings database file")

	// Security stats flags
	securityStatsCmd.Flags().StringVarP(&securityOutputFormat, "format", "f", "table", "Output format (table, json)")
	securityStatsCmd.Flags().StringVar(&securityDatabaseFile, "database", ".security-findings.json", "Security findings database file")

	// Security resolve/ignore flags
	securityResolveCmd.Flags().StringVar(&securityDatabaseFile, "database", ".security-findings.json", "Security findings database file")
	securityIgnoreCmd.Flags().StringVar(&securityDatabaseFile, "database", ".security-findings.json", "Security findings database file")
}

func runSecurityScan(cmd *cobra.Command, args []string) {
	// Determine scan path
	scanPath := "."
	if len(args) > 0 {
		scanPath = args[0]
	}

	// Get absolute path
	absPath, err := filepath.Abs(scanPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	db := security.NewFileDatabase(securityDatabaseFile)
	defer db.Close()

	// Initialize scanner
	scanner := security.NewScanner(db)

	// Load custom config if specified
	if securityConfigFile != "" {
		if err := scanner.LoadConfig(securityConfigFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
	}

	// Prepare scan options
	options := security.ScanOptions{
		Paths:            []string{absPath},
		IncludeHistory:   securityIncludeHistory,
		MaxDepth:         securityMaxDepth,
		ScanMode:         securityScanMode,
		RespectGitignore: securityRespectGitignore,
	}

	// Run the scan
	fmt.Printf("🔍 Scanning %s for security issues...\n", absPath)
	
	var result *security.SecurityScanResult
	
	// Check if path is a file or directory
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing path: %v\n", err)
		os.Exit(1)
	}

	if fileInfo.IsDir() {
		result, err = scanner.ScanDirectory(absPath, options)
	} else {
		result, err = scanner.ScanFile(absPath)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during scan: %v\n", err)
		os.Exit(1)
	}

	// Filter by severity if specified
	if len(securitySeverity) > 0 {
		result.Findings = filterBySeverity(result.Findings, securitySeverity)
	}

	// Output results
	outputSecurityResults(result)

	// Exit with error code if critical or high severity issues found
	if hasCriticalIssues(result.Findings) {
		os.Exit(1)
	}
}

func runSecurityList(cmd *cobra.Command, args []string) {
	// Initialize database
	db := security.NewFileDatabase(securityDatabaseFile)
	defer db.Close()

	// Prepare filters
	filters := make(map[string]interface{})
	if len(securitySeverity) > 0 {
		// For simplicity, just use the first severity filter
		filters["severity"] = securitySeverity[0]
	}

	// Get findings
	findings, err := db.GetFindings(filters)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error retrieving findings: %v\n", err)
		os.Exit(1)
	}

	// Create result structure for consistent output
	result := &security.SecurityScanResult{
		Findings:     findings,
		FilesScanned: 0,
		Duration:     0,
		Timestamp:    time.Now(),
		ScanType:     "list",
	}

	outputSecurityResults(result)
}

func runSecurityStats(cmd *cobra.Command, args []string) {
	// Initialize database
	db := security.NewFileDatabase(securityDatabaseFile)
	defer db.Close()

	// Get statistics
	stats, err := db.GetStats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error retrieving stats: %v\n", err)
		os.Exit(1)
	}

	// Output statistics
	outputSecurityStats(stats)
}

func runSecurityResolve(cmd *cobra.Command, args []string) {
	findingID := args[0]

	// Initialize database
	db := security.NewFileDatabase(securityDatabaseFile)
	defer db.Close()

	// Update finding status
	if err := db.UpdateFindingStatus(findingID, "resolved"); err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving finding: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Finding %s marked as resolved\n", findingID)
}

func runSecurityIgnore(cmd *cobra.Command, args []string) {
	findingID := args[0]

	// Initialize database
	db := security.NewFileDatabase(securityDatabaseFile)
	defer db.Close()

	// Update finding status
	if err := db.UpdateFindingStatus(findingID, "ignored"); err != nil {
		fmt.Fprintf(os.Stderr, "Error ignoring finding: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🙈 Finding %s marked as ignored\n", findingID)
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

func hasCriticalIssues(findings []security.SecurityFinding) bool {
	for _, finding := range findings {
		if finding.Severity == "critical" || finding.Severity == "high" {
			return true
		}
	}
	return false
}

func outputSecurityResults(result *security.SecurityScanResult) {
	switch securityOutputFormat {
	case "json":
		outputJSON(result)
	case "csv":
		outputCSV(result)
	default:
		outputTable(result)
	}
}

func outputSecurityStats(stats *security.SecurityStats) {
	switch securityOutputFormat {
	case "json":
		data, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(data))
	default:
		fmt.Printf("📊 Security Statistics\n")
		fmt.Printf("═══════════════════════\n")
		fmt.Printf("Total Findings: %d\n", stats.TotalFindings)
		fmt.Printf("Files with Issues: %d\n", stats.FilesWithIssues)
		fmt.Printf("Last Scan: %s\n", stats.LastScanTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("\nBy Severity:\n")
		for severity, count := range stats.FindingsBySeverity {
			fmt.Printf("  %s: %d\n", severity, count)
		}
		fmt.Printf("\nBy Type:\n")
		for findingType, count := range stats.FindingsByType {
			fmt.Printf("  %s: %d\n", findingType, count)
		}
	}
}

func outputJSON(result *security.SecurityScanResult) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func outputCSV(result *security.SecurityScanResult) {
	fmt.Println("ID,File,Line,Column,Type,Severity,Message,Status,Confidence")
	for _, finding := range result.Findings {
		fmt.Printf("%s,%s,%d,%d,%s,%s,%s,%s,%.2f\n",
			finding.ID, finding.File, finding.Line, finding.Column,
			finding.Type, finding.Severity, finding.Message,
			finding.Status, finding.Confidence)
	}
}

func outputTable(result *security.SecurityScanResult) {
	if len(result.Findings) == 0 {
		fmt.Printf("✅ No security issues found!\n")
		fmt.Printf("📁 Files scanned: %d\n", result.FilesScanned)
		fmt.Printf("⏱️  Duration: %v\n", result.Duration)
		return
	}

	fmt.Printf("🚨 Security Issues Found: %d\n", len(result.Findings))
	fmt.Printf("📁 Files scanned: %d\n", result.FilesScanned)
	fmt.Printf("⏱️  Duration: %v\n", result.Duration)
	fmt.Printf("\n")

	// Group by severity
	severityGroups := make(map[string][]security.SecurityFinding)
	for _, finding := range result.Findings {
		severityGroups[finding.Severity] = append(severityGroups[finding.Severity], finding)
	}

	// Output by severity (critical first)
	severityOrder := []string{"critical", "high", "medium", "low"}
	for _, severity := range severityOrder {
		findings := severityGroups[severity]
		if len(findings) == 0 {
			continue
		}

		fmt.Printf("%s %s Issues (%d)\n", getSeverityIcon(severity), strings.ToUpper(severity), len(findings))
		fmt.Printf("═══════════════════════════════════════\n")

		for _, finding := range findings {
			fmt.Printf("📄 %s:%d:%d\n", finding.File, finding.Line, finding.Column)
			fmt.Printf("🔍 %s (%s)\n", finding.Message, finding.Type)
			fmt.Printf("🆔 %s\n", finding.ID)
			fmt.Printf("💯 Confidence: %.0f%%\n", finding.Confidence*100)
			fmt.Printf("🔒 Value: %s\n", finding.Value)
			fmt.Printf("\n")
		}
	}

	fmt.Printf("💡 Use 'kwatch security resolve <id>' to mark issues as resolved\n")
	fmt.Printf("💡 Use 'kwatch security ignore <id>' to ignore false positives\n")
}

func getSeverityIcon(severity string) string {
	switch severity {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🔵"
	default:
		return "⚪"
	}
}
