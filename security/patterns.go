package security

import (
	"regexp"
)

// DefaultSecurityPatterns returns a list of default security patterns for secret detection
func DefaultSecurityPatterns() []SecurityPattern {
	return []SecurityPattern{
		// AWS Secrets
		{
			Name:        "aws_access_key",
			Type:        "aws_access_key",
			Pattern:     `(?i)aws[_\-\s]*access[_\-\s]*key[_\-\s]*id[_\-\s]*[=:]\s*["\']?([A-Z0-9]{20})["\']?`,
			Severity:    "critical",
			Description: "AWS Access Key ID detected",
			Confidence:  0.9,
			Enabled:     true,
		},
		{
			Name:        "aws_secret_key",
			Type:        "aws_secret_key",
			Pattern:     `(?i)aws[_\-\s]*secret[_\-\s]*access[_\-\s]*key[_\-\s]*[=:]\s*["\']?([A-Za-z0-9/+=]{40})["\']?`,
			Severity:    "critical",
			Description: "AWS Secret Access Key detected",
			Confidence:  0.9,
			Enabled:     true,
		},
		{
			Name:        "aws_session_token",
			Type:        "aws_session_token",
			Pattern:     `(?i)aws[_\-\s]*session[_\-\s]*token[_\-\s]*[=:]\s*["\']?([A-Za-z0-9/+=]{100,})["\']?`,
			Severity:    "high",
			Description: "AWS Session Token detected",
			Confidence:  0.8,
			Enabled:     true,
		},

		// GitHub Tokens
		{
			Name:        "github_token",
			Type:        "github_token",
			Pattern:     `(?i)github[_\-\s]*token[_\-\s]*[=:]\s*["\']?(ghp_[A-Za-z0-9]{36}|gho_[A-Za-z0-9]{36}|ghu_[A-Za-z0-9]{36}|ghs_[A-Za-z0-9]{36}|ghr_[A-Za-z0-9]{36})["\']?`,
			Severity:    "critical",
			Description: "GitHub Personal Access Token detected",
			Confidence:  0.95,
			Enabled:     true,
		},
		{
			Name:        "github_oauth",
			Type:        "github_oauth",
			Pattern:     `(?i)github[_\-\s]*oauth[_\-\s]*[=:]\s*["\']?([a-f0-9]{40})["\']?`,
			Severity:    "high",
			Description: "GitHub OAuth Token detected",
			Confidence:  0.8,
			Enabled:     true,
		},

		// Google API Keys
		{
			Name:        "google_api_key",
			Type:        "google_api_key",
			Pattern:     `(?i)google[_\-\s]*api[_\-\s]*key[_\-\s]*[=:]\s*["\']?(AIza[0-9A-Za-z\\-_]{35})["\']?`,
			Severity:    "high",
			Description: "Google API Key detected",
			Confidence:  0.9,
			Enabled:     true,
		},
		{
			Name:        "google_oauth",
			Type:        "google_oauth",
			Pattern:     `(?i)google[_\-\s]*oauth[_\-\s]*[=:]\s*["\']?([0-9]+-[0-9A-Za-z_]{32}\.apps\.googleusercontent\.com)["\']?`,
			Severity:    "high",
			Description: "Google OAuth Client ID detected",
			Confidence:  0.9,
			Enabled:     true,
		},

		// JWT Tokens
		{
			Name:        "jwt_token",
			Type:        "jwt_token",
			Pattern:     `(?i)jwt[_\-\s]*token[_\-\s]*[=:]\s*["\']?(eyJ[A-Za-z0-9_\/+-]*\.eyJ[A-Za-z0-9_\/+-]*\.[A-Za-z0-9_\/+-]*)["\']?`,
			Severity:    "medium",
			Description: "JWT Token detected",
			Confidence:  0.8,
			Enabled:     true,
		},

		// Database Connection Strings
		{
			Name:        "postgres_connection",
			Type:        "database_connection",
			Pattern:     `(?i)postgres(?:ql)?://[a-zA-Z0-9_\-\.]+:[a-zA-Z0-9_\-\.@]+@[a-zA-Z0-9_\-\.]+:[0-9]+/[a-zA-Z0-9_\-\.]+`,
			Severity:    "critical",
			Description: "PostgreSQL connection string with credentials detected",
			Confidence:  0.9,
			Enabled:     true,
		},
		{
			Name:        "mysql_connection",
			Type:        "database_connection",
			Pattern:     `(?i)mysql://[a-zA-Z0-9_\-\.]+:[a-zA-Z0-9_\-\.@]+@[a-zA-Z0-9_\-\.]+:[0-9]+/[a-zA-Z0-9_\-\.]+`,
			Severity:    "critical",
			Description: "MySQL connection string with credentials detected",
			Confidence:  0.9,
			Enabled:     true,
		},
		{
			Name:        "mongodb_connection",
			Type:        "database_connection",
			Pattern:     `(?i)mongodb(\+srv)?://[a-zA-Z0-9_\-\.]+:[a-zA-Z0-9_\-\.@]+@[a-zA-Z0-9_\-\.]+/[a-zA-Z0-9_\-\.]+`,
			Severity:    "critical",
			Description: "MongoDB connection string with credentials detected",
			Confidence:  0.9,
			Enabled:     true,
		},

		// Private Keys
		{
			Name:        "rsa_private_key",
			Type:        "private_key",
			Pattern:     `-----BEGIN RSA PRIVATE KEY-----`,
			Severity:    "critical",
			Description: "RSA Private Key detected",
			Confidence:  1.0,
			Enabled:     true,
		},
		{
			Name:        "openssh_private_key",
			Type:        "private_key",
			Pattern:     `-----BEGIN OPENSSH PRIVATE KEY-----`,
			Severity:    "critical",
			Description: "OpenSSH Private Key detected",
			Confidence:  1.0,
			Enabled:     true,
		},
		{
			Name:        "ec_private_key",
			Type:        "private_key",
			Pattern:     `-----BEGIN EC PRIVATE KEY-----`,
			Severity:    "critical",
			Description: "EC Private Key detected",
			Confidence:  1.0,
			Enabled:     true,
		},

		// API Keys (Generic)
		{
			Name:        "generic_api_key",
			Type:        "api_key",
			Pattern:     `(?i)api[_\-\s]*key[_\-\s]*[=:]\s*["\']?([a-zA-Z0-9_\-]{32,})["\']?`,
			Severity:    "medium",
			Description: "Generic API Key detected",
			Confidence:  0.6,
			Enabled:     true,
		},
		{
			Name:        "generic_secret",
			Type:        "secret",
			Pattern:     `(?i)secret[_\-\s]*[=:]\s*["\']?([a-zA-Z0-9_\-]{16,})["\']?`,
			Severity:    "medium",
			Description: "Generic Secret detected",
			Confidence:  0.5,
			Enabled:     true,
		},

		// Passwords
		{
			Name:        "password_assignment",
			Type:        "password",
			Pattern:     `(?i)password[_\-\s]*[=:]\s*["\']?([a-zA-Z0-9_\-@#$%^&*!]{8,})["\']?`,
			Severity:    "medium",
			Description: "Password assignment detected",
			Confidence:  0.6,
			Enabled:     true,
		},

		// Slack Tokens
		{
			Name:        "slack_token",
			Type:        "slack_token",
			Pattern:     `(?i)slack[_\-\s]*token[_\-\s]*[=:]\s*["\']?(xox[bpoa]-[0-9]{12}-[0-9]{12}-[0-9]{12}-[a-z0-9]{32})["\']?`,
			Severity:    "high",
			Description: "Slack Token detected",
			Confidence:  0.9,
			Enabled:     true,
		},

		// Discord Tokens
		{
			Name:        "discord_token",
			Type:        "discord_token",
			Pattern:     `(?i)discord[_\-\s]*token[_\-\s]*[=:]\s*["\']?([MN][A-Za-z\d]{23}\.[\w-]{6}\.[\w-]{27})["\']?`,
			Severity:    "high",
			Description: "Discord Bot Token detected",
			Confidence:  0.9,
			Enabled:     true,
		},

		// Webhook URLs
		{
			Name:        "webhook_url",
			Type:        "webhook",
			Pattern:     `(?i)webhook[_\-\s]*url[_\-\s]*[=:]\s*["\']?(https://[a-zA-Z0-9\-\.]+/[a-zA-Z0-9\-_/]+)["\']?`,
			Severity:    "medium",
			Description: "Webhook URL detected",
			Confidence:  0.7,
			Enabled:     true,
		},

		// Email Credentials
		{
			Name:        "smtp_password",
			Type:        "email_credential",
			Pattern:     `(?i)smtp[_\-\s]*password[_\-\s]*[=:]\s*["\']?([a-zA-Z0-9_\-@#$%^&*!]{8,})["\']?`,
			Severity:    "medium",
			Description: "SMTP Password detected",
			Confidence:  0.7,
			Enabled:     true,
		},
	}
}

// CompilePatterns compiles regex patterns for efficient matching
func CompilePatterns(patterns []SecurityPattern) (map[string]*regexp.Regexp, error) {
	compiled := make(map[string]*regexp.Regexp)
	
	for _, pattern := range patterns {
		if !pattern.Enabled {
			continue
		}
		
		regex, err := regexp.Compile(pattern.Pattern)
		if err != nil {
			return nil, err
		}
		
		compiled[pattern.Name] = regex
	}
	
	return compiled, nil
}
