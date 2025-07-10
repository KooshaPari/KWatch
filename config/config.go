package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

// Config represents the kwatch configuration
type Config struct {
	DefaultTimeout string             `yaml:"defaultTimeout"`
	MaxParallel    int               `yaml:"maxParallel"`
	Commands       map[string]Command `yaml:"commands"`
}

// Command represents a single command configuration
type Command struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	Timeout string   `yaml:"timeout"`
	Enabled bool     `yaml:"enabled"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultTimeout: "30s",
		MaxParallel:    3,
		Commands: map[string]Command{
			"typescript": {
				Command: "npx",
				Args:    []string{"tsc", "--noEmit"},
				Timeout: "30s",
				Enabled: true,
			},
			"lint": {
				Command: "npx",
				Args:    []string{"eslint", ".", "--ext", ".ts,.tsx,.js,.jsx"},
				Timeout: "30s",
				Enabled: true,
			},
			"test": {
				Command: "npm",
				Args:    []string{"test"},
				Timeout: "60s",
				Enabled: true,
			},
			"github_actions": {
				Command: "github_actions",
				Args:    []string{},
				Timeout: "30s",
				Enabled: false, // Disabled by default, will be auto-enabled if GitHub repo detected
			},
		},
	}
}

// Load loads configuration from the specified directory
func Load(dir string) (*Config, error) {
	configPath := filepath.Join(dir, ".kwatch", "kwatch.yaml")
	
	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	return &config, nil
}

// Save saves the configuration to the specified directory
func (c *Config) Save(dir string) error {
	configDir := filepath.Join(dir, ".kwatch")
	configPath := filepath.Join(configDir, "kwatch.yaml")
	
	// Create .kwatch directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Marshal config to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate default timeout
	if _, err := time.ParseDuration(c.DefaultTimeout); err != nil {
		return fmt.Errorf("invalid defaultTimeout: %w", err)
	}
	
	// Validate max parallel
	if c.MaxParallel < 1 {
		return fmt.Errorf("maxParallel must be at least 1")
	}
	
	// Validate commands
	for name, cmd := range c.Commands {
		if cmd.Command == "" {
			return fmt.Errorf("command %s: command field is required", name)
		}
		
		if cmd.Timeout != "" {
			if _, err := time.ParseDuration(cmd.Timeout); err != nil {
				return fmt.Errorf("command %s: invalid timeout: %w", name, err)
			}
		}
	}
	
	return nil
}

// GetTimeout returns the timeout for a command, falling back to default
func (c *Config) GetTimeout(cmdName string) time.Duration {
	cmd, exists := c.Commands[cmdName]
	if !exists {
		// Parse default timeout
		if duration, err := time.ParseDuration(c.DefaultTimeout); err == nil {
			return duration
		}
		return 30 * time.Second
	}
	
	// Use command-specific timeout if set
	if cmd.Timeout != "" {
		if duration, err := time.ParseDuration(cmd.Timeout); err == nil {
			return duration
		}
	}
	
	// Fall back to default timeout
	if duration, err := time.ParseDuration(c.DefaultTimeout); err == nil {
		return duration
	}
	
	// Final fallback
	return 30 * time.Second
}

// GetEnabledCommands returns only the enabled commands
func (c *Config) GetEnabledCommands() map[string]Command {
	enabled := make(map[string]Command)
	for name, cmd := range c.Commands {
		if cmd.Enabled {
			enabled[name] = cmd
		}
	}
	return enabled
}

// ConfigExists checks if a config file exists in the specified directory
func ConfigExists(dir string) bool {
	configPath := filepath.Join(dir, ".kwatch", "kwatch.yaml")
	_, err := os.Stat(configPath)
	return err == nil
}