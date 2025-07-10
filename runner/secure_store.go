package runner

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/term"
)

const (
	secureConfigDir  = ".kwatch"
	tokenFileName    = "secure_token.enc"
	saltFileName     = "token.salt"
)

// SecureTokenStore handles encrypted storage of GitHub tokens
type SecureTokenStore struct {
	configDir string
}

// NewSecureTokenStore creates a new secure token store
func NewSecureTokenStore() *SecureTokenStore {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, secureConfigDir)
	
	return &SecureTokenStore{
		configDir: configDir,
	}
}

// InitSecureToken prompts for and securely stores a GitHub token
func (s *SecureTokenStore) InitSecureToken() error {
	fmt.Println("🔐 Secure GitHub Token Setup")
	fmt.Println("=============================")
	fmt.Println()
	fmt.Println("This will securely encrypt and store your GitHub token locally.")
	fmt.Println("The token will be encrypted using your system's unique identifier.")
	fmt.Println()
	
	// Check if token already exists
	if s.HasStoredToken() {
		fmt.Println("⚠️  An encrypted token already exists.")
		fmt.Print("Do you want to replace it? (y/N): ")
		
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println("Keeping existing token.")
			return nil
		}
	}
	
	// Get token securely
	fmt.Print("Enter your GitHub token (input will be hidden): ")
	tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	fmt.Println() // New line after hidden input
	
	token := string(tokenBytes)
	if token == "" {
		return fmt.Errorf("no token provided")
	}
	
	// Validate token format
	if !isValidGitHubToken(token) {
		fmt.Println("⚠️  Warning: Token doesn't appear to be a valid GitHub token")
		fmt.Print("Continue anyway? (y/N): ")
		
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" {
			return fmt.Errorf("token validation failed")
		}
	}
	
	// Store token securely
	if err := s.StoreToken(token); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}
	
	fmt.Println("✅ Token encrypted and stored securely!")
	fmt.Printf("📁 Location: %s\n", s.getTokenPath())
	fmt.Println()
	fmt.Println("🔒 Security Notes:")
	fmt.Println("  • Token is encrypted using AES-256-GCM")
	fmt.Println("  • Encryption key derived from system-specific data")
	fmt.Println("  • Only accessible by your user account")
	fmt.Println("  • No token stored in shell profile or environment")
	fmt.Println()
	fmt.Println("🧪 Test with: kwatch run --command github")
	
	return nil
}

// StoreToken encrypts and stores a GitHub token
func (s *SecureTokenStore) StoreToken(token string) error {
	// Ensure config directory exists
	if err := os.MkdirAll(s.configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Generate or get encryption key
	key, err := s.getOrCreateKey()
	if err != nil {
		return fmt.Errorf("failed to get encryption key: %w", err)
	}
	
	// Encrypt token
	encryptedToken, err := s.encrypt(token, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt token: %w", err)
	}
	
	// Store encrypted token
	tokenPath := s.getTokenPath()
	if err := os.WriteFile(tokenPath, []byte(encryptedToken), 0600); err != nil {
		return fmt.Errorf("failed to write encrypted token: %w", err)
	}
	
	return nil
}

// GetToken retrieves and decrypts the stored GitHub token
func (s *SecureTokenStore) GetToken() (string, error) {
	if !s.HasStoredToken() {
		return "", fmt.Errorf("no stored token found")
	}
	
	// Get encryption key
	key, err := s.getOrCreateKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %w", err)
	}
	
	// Read encrypted token
	tokenPath := s.getTokenPath()
	encryptedData, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("failed to read encrypted token: %w", err)
	}
	
	// Decrypt token
	token, err := s.decrypt(string(encryptedData), key)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt token: %w", err)
	}
	
	return token, nil
}

// HasStoredToken checks if an encrypted token exists
func (s *SecureTokenStore) HasStoredToken() bool {
	tokenPath := s.getTokenPath()
	_, err := os.Stat(tokenPath)
	return err == nil
}

// ClearStoredToken removes the encrypted token
func (s *SecureTokenStore) ClearStoredToken() error {
	tokenPath := s.getTokenPath()
	saltPath := s.getSaltPath()
	
	// Remove token file
	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token file: %w", err)
	}
	
	// Remove salt file
	if err := os.Remove(saltPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove salt file: %w", err)
	}
	
	return nil
}

// GetTokenStatus returns information about the stored token
func (s *SecureTokenStore) GetTokenStatus() (map[string]interface{}, error) {
	status := make(map[string]interface{})
	
	status["has_stored_token"] = s.HasStoredToken()
	status["config_dir"] = s.configDir
	status["token_path"] = s.getTokenPath()
	
	if s.HasStoredToken() {
		// Try to decrypt and validate
		token, err := s.GetToken()
		if err != nil {
			status["decrypt_error"] = err.Error()
			status["valid"] = false
		} else {
			status["valid"] = true
			status["token_length"] = len(token)
			if len(token) >= 12 {
				status["token_preview"] = token[:8] + "..." + token[len(token)-4:]
			}
			status["token_type"] = getTokenType(token)
		}
		
		// File info
		if info, err := os.Stat(s.getTokenPath()); err == nil {
			status["created"] = info.ModTime()
			status["permissions"] = info.Mode().String()
		}
	}
	
	return status, nil
}

// getTokenPath returns the path to the encrypted token file
func (s *SecureTokenStore) getTokenPath() string {
	return filepath.Join(s.configDir, tokenFileName)
}

// getSaltPath returns the path to the salt file
func (s *SecureTokenStore) getSaltPath() string {
	return filepath.Join(s.configDir, saltFileName)
}

// getOrCreateKey generates or retrieves the encryption key
func (s *SecureTokenStore) getOrCreateKey() ([]byte, error) {
	saltPath := s.getSaltPath()
	
	var salt []byte
	
	// Try to read existing salt
	if _, err := os.Stat(saltPath); err == nil {
		var err error
		salt, err = os.ReadFile(saltPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read salt: %w", err)
		}
	} else {
		// Generate new salt
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, fmt.Errorf("failed to generate salt: %w", err)
		}
		
		// Store salt
		if err := os.WriteFile(saltPath, salt, 0600); err != nil {
			return nil, fmt.Errorf("failed to store salt: %w", err)
		}
	}
	
	// Derive key from system-specific data + salt
	keyMaterial := s.getSystemKeyMaterial()
	hasher := sha256.New()
	hasher.Write(keyMaterial)
	hasher.Write(salt)
	
	return hasher.Sum(nil), nil
}

// getSystemKeyMaterial generates system-specific key material
func (s *SecureTokenStore) getSystemKeyMaterial() []byte {
	hasher := sha256.New()
	
	// Add various system-specific identifiers
	hasher.Write([]byte(runtime.GOOS))
	hasher.Write([]byte(runtime.GOARCH))
	
	// Add username
	if user := os.Getenv("USER"); user != "" {
		hasher.Write([]byte(user))
	}
	if user := os.Getenv("USERNAME"); user != "" {
		hasher.Write([]byte(user))
	}
	
	// Add home directory path
	if home, err := os.UserHomeDir(); err == nil {
		hasher.Write([]byte(home))
	}
	
	// Add hostname if available
	if hostname, err := os.Hostname(); err == nil {
		hasher.Write([]byte(hostname))
	}
	
	return hasher.Sum(nil)
}

// encrypt encrypts plaintext using AES-GCM
func (s *SecureTokenStore) encrypt(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts ciphertext using AES-GCM
func (s *SecureTokenStore) decrypt(ciphertext string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	
	nonce, ciphertext_bytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext_bytes, nil)
	if err != nil {
		return "", err
	}
	
	return string(plaintext), nil
}

// isValidGitHubToken performs basic GitHub token validation
func isValidGitHubToken(token string) bool {
	// GitHub personal access tokens start with these prefixes
	validPrefixes := []string{"ghp_", "github_pat_", "gho_", "ghu_", "ghs_", "ghr_"}
	
	for _, prefix := range validPrefixes {
		if len(token) > len(prefix) && token[:len(prefix)] == prefix {
			return true
		}
	}
	
	return false
}

// getTokenType identifies the type of GitHub token
func getTokenType(token string) string {
	if len(token) < 4 {
		return "unknown"
	}
	
	switch token[:4] {
	case "ghp_":
		return "personal_access_token"
	case "gho_":
		return "oauth_token"
	case "ghu_":
		return "user_token"
	case "ghs_":
		return "server_token"
	case "ghr_":
		return "refresh_token"
	default:
		if token[:11] == "github_pat_" {
			return "fine_grained_personal_access_token"
		}
		return "unknown"
	}
}