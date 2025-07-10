package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"kwatch/runner"
)

var (
	authStatus bool
	authClear  bool
	authInit   bool
	authJSON   bool
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Secure GitHub authentication management",
	Long: `Secure GitHub authentication management with encrypted token storage.

This command provides a secure way to store and manage your GitHub personal access token.
The token is encrypted using AES-256-GCM and stored locally, eliminating the need to 
store tokens in environment variables or shell profiles.

Security Features:
• AES-256-GCM encryption
• System-specific encryption keys
• Secure file permissions (600)
• No plaintext storage
• Automatic token detection

Token Requirements:
• GitHub personal access token (classic or fine-grained)
• Scopes: repo (for private repos) + actions:read (for GitHub Actions)
• Get your token at: https://github.com/settings/tokens

Examples:
  kwatch auth --init           # Securely setup new token
  kwatch auth --status         # Check authentication status
  kwatch auth --clear          # Remove stored token
  kwatch auth --status --json  # JSON status output`,
	Run: func(cmd *cobra.Command, args []string) {
		store := runner.NewSecureTokenStore()
		
		switch {
		case authInit:
			if err := store.InitSecureToken(); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to initialize token: %v\n", err)
				os.Exit(1)
			}
		case authClear:
			clearStoredToken(store)
		case authStatus:
			showAuthStatus(store)
		default:
			// Default behavior - init if no token exists, otherwise show status
			if store.HasStoredToken() {
				showAuthStatus(store)
			} else {
				fmt.Println("🔐 No secure token found. Initializing setup...")
				fmt.Println()
				if err := store.InitSecureToken(); err != nil {
					fmt.Fprintf(os.Stderr, "❌ Failed to initialize token: %v\n", err)
					os.Exit(1)
				}
			}
		}
	},
}

func init() {
	// Temporarily disabled auth command to fix hanging issues
	// rootCmd.AddCommand(authCmd)
	// authCmd.Flags().BoolVarP(&authStatus, "status", "s", false, "Check current authentication status")
	// authCmd.Flags().BoolVarP(&authClear, "clear", "c", false, "Remove stored encrypted token")
	// authCmd.Flags().BoolVarP(&authInit, "init", "i", false, "Initialize new encrypted token (force)")
	// authCmd.Flags().BoolVarP(&authJSON, "json", "j", false, "Output status in JSON format")
}

func showAuthStatus(store *runner.SecureTokenStore) {
	if authJSON {
		showAuthStatusJSON(store)
		return
	}
	
	fmt.Println("🔍 GitHub Authentication Status")
	fmt.Println("===============================")
	fmt.Println()
	
	// Check environment variables first
	var envToken string
	var envSource string
	
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		envToken = t
		envSource = "GITHUB_TOKEN"
	} else if t := os.Getenv("GH_TOKEN"); t != "" {
		envToken = t
		envSource = "GH_TOKEN"
	}
	
	if envToken != "" {
		fmt.Printf("✅ Environment token found via %s\n", envSource)
		if len(envToken) >= 12 {
			fmt.Printf("   Token: %s...%s\n", envToken[:8], envToken[len(envToken)-4:])
		}
		fmt.Println("   📝 Note: Environment tokens take precedence over stored tokens")
		fmt.Println()
	}
	
	// Check stored token
	if store.HasStoredToken() {
		fmt.Println("🔒 Encrypted Token Store")
		
		status, err := store.GetTokenStatus()
		if err != nil {
			fmt.Printf("❌ Error getting token status: %v\n", err)
			return
		}
		
		if valid, ok := status["valid"].(bool); ok && valid {
			fmt.Println("✅ Encrypted token found and valid")
			
			if preview, ok := status["token_preview"].(string); ok {
				fmt.Printf("   Token: %s\n", preview)
			}
			
			if tokenType, ok := status["token_type"].(string); ok {
				fmt.Printf("   Type: %s\n", strings.ReplaceAll(tokenType, "_", " "))
			}
			
			if created, ok := status["created"]; ok {
				fmt.Printf("   Created: %v\n", created)
			}
			
			if perms, ok := status["permissions"].(string); ok {
				fmt.Printf("   Permissions: %s\n", perms)
			}
		} else {
			fmt.Println("❌ Encrypted token found but invalid")
			if errStr, ok := status["decrypt_error"].(string); ok {
				fmt.Printf("   Error: %s\n", errStr)
			}
			fmt.Println("   💡 Run 'kwatch auth --clear && kwatch auth --init' to reset")
		}
		
		if configDir, ok := status["config_dir"].(string); ok {
			fmt.Printf("   📁 Storage: %s\n", configDir)
		}
		fmt.Println()
	} else {
		fmt.Println("❌ No encrypted token stored")
		fmt.Println("   💡 Run 'kwatch auth --init' to setup secure token storage")
		fmt.Println()
	}
	
	// Check repository status
	checkRepositoryStatus()
	
	// Show usage examples
	fmt.Println("🧪 Test Commands:")
	fmt.Println("   kwatch run --command github    # Test GitHub Actions monitoring")
	fmt.Println("   kwatch master                   # Master view with GitHub Actions")
	fmt.Println()
	
	// Show management commands
	fmt.Println("🔧 Management Commands:")
	fmt.Println("   kwatch auth --init              # Setup new token")
	fmt.Println("   kwatch auth --clear             # Remove stored token")
	fmt.Println("   kwatch auth --status --json     # JSON status output")
}

func showAuthStatusJSON(store *runner.SecureTokenStore) {
	result := make(map[string]interface{})
	
	// Environment token info
	envInfo := make(map[string]interface{})
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		envInfo["found"] = true
		envInfo["source"] = "GITHUB_TOKEN"
		if len(token) >= 12 {
			envInfo["preview"] = token[:8] + "..." + token[len(token)-4:]
		}
	} else if token := os.Getenv("GH_TOKEN"); token != "" {
		envInfo["found"] = true
		envInfo["source"] = "GH_TOKEN"
		if len(token) >= 12 {
			envInfo["preview"] = token[:8] + "..." + token[len(token)-4:]
		}
	} else {
		envInfo["found"] = false
	}
	result["environment"] = envInfo
	
	// Stored token info
	if status, err := store.GetTokenStatus(); err == nil {
		result["stored"] = status
	} else {
		result["stored"] = map[string]interface{}{
			"error": err.Error(),
		}
	}
	
	// Repository info
	wd, _ := os.Getwd()
	repoInfo := map[string]interface{}{
		"current_directory": wd,
		"is_git_repo":       isGitRepository(),
		"has_github_remote": hasGitHubRemote(),
	}
	result["repository"] = repoInfo
	
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(string(jsonBytes))
}

func clearStoredToken(store *runner.SecureTokenStore) {
	if !store.HasStoredToken() {
		fmt.Println("❌ No stored token to clear")
		return
	}
	
	fmt.Println("🗑️  Clear Stored Token")
	fmt.Println("====================")
	fmt.Println()
	fmt.Println("⚠️  This will permanently delete your encrypted GitHub token.")
	fmt.Print("Are you sure? (y/N): ")
	
	var response string
	fmt.Scanln(&response)
	
	if response != "y" && response != "Y" && response != "yes" {
		fmt.Println("❌ Cancelled.")
		return
	}
	
	if err := store.ClearStoredToken(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to clear token: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("✅ Encrypted token cleared successfully")
	fmt.Println("💡 Run 'kwatch auth --init' to setup a new token")
}

func checkRepositoryStatus() {
	wd, _ := os.Getwd()
	fmt.Printf("📂 Repository Status (Current: %s)\n", wd)
	
	if isGitRepository() {
		fmt.Println("✅ Git repository detected")
		
		if hasGitHubRemote() {
			fmt.Println("✅ GitHub remote detected")
			
			// Try to show remote URL
			if configData, err := os.ReadFile(".git/config"); err == nil {
				content := string(configData)
				lines := strings.Split(content, "\n")
				for i, line := range lines {
					if strings.Contains(line, `[remote "origin"]`) && i+1 < len(lines) {
						urlLine := strings.TrimSpace(lines[i+1])
						if strings.HasPrefix(urlLine, "url = ") {
							url := strings.TrimPrefix(urlLine, "url = ")
							fmt.Printf("   Remote: %s\n", url)
						}
						break
					}
				}
			}
		} else {
			fmt.Println("⚠️  No GitHub remote found")
			fmt.Println("   💡 This directory won't support GitHub Actions monitoring")
		}
	} else {
		fmt.Println("⚠️  Not a git repository")
		fmt.Println("   💡 GitHub Actions monitoring only works in git repositories")
	}
	fmt.Println()
}

func isGitRepository() bool {
	_, err := os.Stat(".git")
	return err == nil
}

func hasGitHubRemote() bool {
	if !isGitRepository() {
		return false
	}
	
	configData, err := os.ReadFile(".git/config")
	if err != nil {
		return false
	}
	
	return strings.Contains(string(configData), "github.com")
}