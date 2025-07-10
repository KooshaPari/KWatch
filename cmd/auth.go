package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	authStatus bool
	authJSON   bool
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Configure GitHub authentication",
	Long: `Configure GitHub authentication for monitoring GitHub Actions.

GitHub Actions monitoring requires a personal access token.
Set one of these environment variables:
- GITHUB_TOKEN
- GH_TOKEN

You'll need a GitHub personal access token with these scopes:
- repo (for private repositories)
- actions:read (for GitHub Actions)

Get your token at: https://github.com/settings/tokens

Examples:
  kwatch auth --status  # Check current authentication status
  export GITHUB_TOKEN=your_token_here`,
	Run: func(cmd *cobra.Command, args []string) {
		if authStatus {
			checkGitHubAuthStatus()
		} else {
			showAuthInstructions()
		}
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.Flags().BoolVarP(&authStatus, "status", "s", false, "Check current authentication status")
	authCmd.Flags().BoolVar(&authJSON, "json", false, "Output status in JSON format")
}

func checkGitHubAuthStatus() {
	if authJSON {
		showAuthStatusJSON()
		return
	}
	
	fmt.Println("🔍 GitHub Authentication Status")
	fmt.Println("===============================")
	fmt.Println()
	
	// Check environment variables
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
		fmt.Println()
	} else {
		fmt.Println("❌ No GitHub token found")
		fmt.Println("   💡 Set GITHUB_TOKEN or GH_TOKEN environment variable")
		fmt.Println()
	}
	
	// Check repository status
	checkRepositoryStatus()
	
	// Show usage examples
	fmt.Println("🧪 Test Commands:")
	fmt.Println("   kwatch run --command github    # Test GitHub Actions monitoring")
	fmt.Println("   kwatch master                   # Master view with GitHub Actions")
	fmt.Println()
}

func showAuthStatusJSON() {
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

func showAuthInstructions() {
	fmt.Println("🔐 GitHub Authentication Setup")
	fmt.Println("==============================")
	fmt.Println()
	
	// Check if token already exists
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
		fmt.Printf("✅ GitHub token already configured via %s\n", envSource)
		fmt.Println("   Run 'kwatch auth --status' for more details")
		return
	}
	
	fmt.Println("To use GitHub Actions monitoring, you need to set a GitHub token.")
	fmt.Println()
	fmt.Println("📋 Steps:")
	fmt.Println("1. Get a token at: https://github.com/settings/tokens")
	fmt.Println("2. Grant these scopes:")
	fmt.Println("   - repo (for private repositories)")
	fmt.Println("   - actions:read (for GitHub Actions)")
	fmt.Println("3. Set environment variable:")
	fmt.Println()
	fmt.Println("   # For bash/zsh (add to ~/.bashrc or ~/.zshrc):")
	fmt.Println("   export GITHUB_TOKEN=your_token_here")
	fmt.Println()
	fmt.Println("   # For current session only:")
	fmt.Println("   export GITHUB_TOKEN=your_token_here")
	fmt.Println()
	fmt.Println("4. Test the setup:")
	fmt.Println("   kwatch auth --status")
	fmt.Println("   kwatch run --command github")
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