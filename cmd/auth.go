package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	authStatus bool
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Configure GitHub authentication",
	Long: `Configure GitHub authentication for monitoring GitHub Actions.

This command helps you set up a GitHub personal access token for monitoring
GitHub Actions workflows. The token will be stored securely and used for
API requests to GitHub.

You'll need a GitHub personal access token with these scopes:
- repo (for private repositories)
- actions:read (for GitHub Actions)

Get your token at: https://github.com/settings/tokens

Examples:
  kwatch auth           # Interactive setup
  kwatch auth --status  # Check current authentication status`,
	Run: func(cmd *cobra.Command, args []string) {
		if authStatus {
			checkGitHubAuthStatus()
		} else {
			setupGitHubAuth()
		}
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.Flags().BoolVarP(&authStatus, "status", "s", false, "Check current authentication status")
}

func setupGitHubAuth() {
	fmt.Println("🔐 GitHub Authentication Setup")
	fmt.Println("===============================")
	fmt.Println()
	
	// Check if token already exists
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		fmt.Println("✅ GitHub token is already configured via GITHUB_TOKEN environment variable")
		fmt.Printf("Token: %s...%s\n", token[:8], token[len(token)-4:])
		fmt.Println()
		
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Do you want to update it? (y/N): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		
		if response != "y" && response != "yes" {
			fmt.Println("Keeping existing token.")
			return
		}
	}
	
	fmt.Println("To get a GitHub token:")
	fmt.Println("1. Go to https://github.com/settings/tokens")
	fmt.Println("2. Click 'Generate new token (classic)'")
	fmt.Println("3. Select these scopes:")
	fmt.Println("   - repo (for private repositories)")
	fmt.Println("   - actions:read (for GitHub Actions)")
	fmt.Println("4. Copy the generated token")
	fmt.Println()
	
	// Prompt for token
	fmt.Print("Enter your GitHub token (input will be hidden): ")
	tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading token: %v\n", err)
		os.Exit(1)
	}
	
	token := strings.TrimSpace(string(tokenBytes))
	fmt.Println() // New line after hidden input
	
	if token == "" {
		fmt.Println("❌ No token provided. Exiting.")
		os.Exit(1)
	}
	
	// Validate token format (basic check)
	if !strings.HasPrefix(token, "ghp_") && !strings.HasPrefix(token, "github_pat_") {
		fmt.Println("⚠️  Warning: Token doesn't look like a GitHub personal access token")
		fmt.Println("   Expected format: ghp_... or github_pat_...")
		fmt.Println()
		
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Continue anyway? (y/N): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		
		if response != "y" && response != "yes" {
			fmt.Println("❌ Cancelled.")
			os.Exit(1)
		}
	}
	
	// Show setup instructions
	showSetupInstructions(token)
}

func showSetupInstructions(token string) {
	fmt.Println()
	fmt.Println("🎯 Setup Instructions")
	fmt.Println("=====================")
	fmt.Println()
	
	// Detect shell
	shell := os.Getenv("SHELL")
	configFile := ""
	
	switch {
	case strings.Contains(shell, "zsh"):
		configFile = "~/.zshrc"
	case strings.Contains(shell, "bash"):
		configFile = "~/.bashrc"
	case strings.Contains(shell, "fish"):
		configFile = "~/.config/fish/config.fish"
	default:
		configFile = "~/.profile"
	}
	
	fmt.Printf("Add this line to your %s:\n", configFile)
	fmt.Println()
	
	if strings.Contains(shell, "fish") {
		fmt.Printf("set -Ux GITHUB_TOKEN \"%s\"\n", token)
	} else {
		fmt.Printf("export GITHUB_TOKEN=\"%s\"\n", token)
	}
	
	fmt.Println()
	fmt.Println("Then reload your shell:")
	
	if strings.Contains(shell, "fish") {
		fmt.Println("exec fish")
	} else {
		fmt.Printf("source %s\n", configFile)
	}
	
	fmt.Println()
	fmt.Println("Or set it temporarily for this session:")
	fmt.Printf("export GITHUB_TOKEN=\"%s\"\n", token)
	fmt.Println()
	
	fmt.Println("✅ Once set, kwatch will automatically detect GitHub repositories")
	fmt.Println("   and monitor GitHub Actions in the TUI and master view.")
	fmt.Println()
	fmt.Println("Test it with: kwatch run --command github")
}

func checkGitHubAuthStatus() {
	fmt.Println("🔍 GitHub Authentication Status")
	fmt.Println("===============================")
	fmt.Println()
	
	// Check environment variables
	var token string
	var tokenSource string
	
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		token = t
		tokenSource = "GITHUB_TOKEN"
	} else if t := os.Getenv("GH_TOKEN"); t != "" {
		token = t
		tokenSource = "GH_TOKEN"
	}
	
	if token == "" {
		fmt.Println("❌ No GitHub token found")
		fmt.Println("   Checked environment variables: GITHUB_TOKEN, GH_TOKEN")
		fmt.Println()
		fmt.Println("💡 Run 'kwatch auth' to set up authentication")
		return
	}
	
	// Show token info (masked)
	fmt.Printf("✅ GitHub token found via %s\n", tokenSource)
	if len(token) >= 12 {
		fmt.Printf("   Token: %s...%s\n", token[:8], token[len(token)-4:])
	} else {
		fmt.Printf("   Token: %s\n", strings.Repeat("*", len(token)))
	}
	
	// Check if we're in a git repository
	fmt.Println()
	wd, _ := os.Getwd()
	fmt.Printf("📂 Current directory: %s\n", wd)
	
	// Try to detect GitHub repo
	if _, err := os.Stat(".git"); err == nil {
		fmt.Println("✅ Git repository detected")
		
		// Try to parse git config for GitHub remote
		if configData, err := os.ReadFile(".git/config"); err == nil {
			content := string(configData)
			if strings.Contains(content, "github.com") {
				fmt.Println("✅ GitHub remote detected")
				
				// Try to extract repo info
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
			} else {
				fmt.Println("⚠️  No GitHub remote found in git config")
			}
		}
	} else {
		fmt.Println("⚠️  Not a git repository")
	}
	
	fmt.Println()
	fmt.Println("🧪 Test GitHub Actions monitoring:")
	fmt.Println("   kwatch run --command github")
	fmt.Println("   kwatch master")
}