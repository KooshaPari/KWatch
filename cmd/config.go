package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"kwatch/config"
)

var (
	configForce bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage kwatch configuration",
	Long:  `Manage kwatch configuration files and settings.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize kwatch configuration",
	Long: `Initialize kwatch configuration in the specified directory.

Creates a .kwatch/kwatch.yaml configuration file with default settings.
Use --force to overwrite existing configuration.

Examples:
  kwatch config init                # Initialize config in current directory
  kwatch config init /path/to/proj  # Initialize config in specific directory
  kwatch config init --force        # Overwrite existing config`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := getWorkingDirectory(args)
		
		absDir, err := filepath.Abs(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving directory: %v\n", err)
			os.Exit(1)
		}
		
		// Check if directory exists
		if _, err := os.Stat(absDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Directory does not exist: %s\n", absDir)
			os.Exit(1)
		}
		
		// Check if config already exists
		if config.ConfigExists(absDir) && !configForce {
			fmt.Fprintf(os.Stderr, "Configuration already exists in %s\n", absDir)
			fmt.Fprintf(os.Stderr, "Use --force to overwrite existing configuration\n")
			os.Exit(1)
		}
		
		// Initialize config
		if err := initializeConfig(absDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✓ Configuration initialized in %s\n", filepath.Join(absDir, ".kwatch", "kwatch.yaml"))
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show [directory]",
	Short: "Show current configuration",
	Long: `Show the current configuration for the specified directory.

Displays the effective configuration, including default values if no 
configuration file exists.

Examples:
  kwatch config show                # Show config for current directory
  kwatch config show /path/to/proj  # Show config for specific directory`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := getWorkingDirectory(args)
		
		absDir, err := filepath.Abs(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving directory: %v\n", err)
			os.Exit(1)
		}
		
		// Load config
		cfg, err := config.Load(absDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		
		// Display config
		displayConfig(cfg, absDir)
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit [directory]",
	Short: "Interactively edit configuration",
	Long: `Interactively edit the configuration for the specified directory.

Provides a guided interface to modify configuration settings.
If no configuration exists, creates a new one.

Examples:
  kwatch config edit                # Edit config for current directory
  kwatch config edit /path/to/proj  # Edit config for specific directory`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := getWorkingDirectory(args)
		
		absDir, err := filepath.Abs(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving directory: %v\n", err)
			os.Exit(1)
		}
		
		// Load existing config or create default
		cfg, err := config.Load(absDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		
		// Interactive edit
		if err := editConfigInteractive(cfg, absDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error editing config: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	
	configInitCmd.Flags().BoolVarP(&configForce, "force", "f", false, "Overwrite existing configuration")
}

// initializeConfig creates a default configuration file
func initializeConfig(dir string) error {
	cfg := config.DefaultConfig()
	return cfg.Save(dir)
}

// displayConfig displays the current configuration
func displayConfig(cfg *config.Config, dir string) {
	configPath := filepath.Join(dir, ".kwatch", "kwatch.yaml")
	
	fmt.Printf("Configuration for: %s\n", dir)
	if config.ConfigExists(dir) {
		fmt.Printf("Config file: %s\n", configPath)
	} else {
		fmt.Printf("Using default configuration (no config file found)\n")
	}
	fmt.Println()
	
	fmt.Printf("Default Timeout: %s\n", cfg.DefaultTimeout)
	fmt.Printf("Max Parallel: %d\n", cfg.MaxParallel)
	fmt.Println()
	
	fmt.Println("Commands:")
	for name, cmd := range cfg.Commands {
		status := "enabled"
		if !cmd.Enabled {
			status = "disabled"
		}
		
		fmt.Printf("  %s (%s):\n", name, status)
		fmt.Printf("    Command: %s %s\n", cmd.Command, strings.Join(cmd.Args, " "))
		fmt.Printf("    Timeout: %s\n", cmd.Timeout)
		fmt.Println()
	}
}

// editConfigInteractive provides an interactive configuration editor
func editConfigInteractive(cfg *config.Config, dir string) error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Printf("Interactive Configuration Editor\n")
	fmt.Printf("Directory: %s\n", dir)
	fmt.Println()
	
	// Edit default timeout
	fmt.Printf("Default timeout [%s]: ", cfg.DefaultTimeout)
	if input, err := reader.ReadString('\n'); err == nil {
		input = strings.TrimSpace(input)
		if input != "" {
			cfg.DefaultTimeout = input
		}
	}
	
	// Edit max parallel
	fmt.Printf("Max parallel commands [%d]: ", cfg.MaxParallel)
	if input, err := reader.ReadString('\n'); err == nil {
		input = strings.TrimSpace(input)
		if input != "" {
			if val, err := strconv.Atoi(input); err == nil && val > 0 {
				cfg.MaxParallel = val
			}
		}
	}
	
	fmt.Println()
	
	// Edit commands
	for name, cmd := range cfg.Commands {
		fmt.Printf("Configure command: %s\n", name)
		
		// Enable/disable
		enabledStr := "y"
		if !cmd.Enabled {
			enabledStr = "n"
		}
		fmt.Printf("  Enabled [%s]: ", enabledStr)
		if input, err := reader.ReadString('\n'); err == nil {
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "" {
				cmd.Enabled = input == "y" || input == "yes" || input == "true"
			}
		}
		
		if cmd.Enabled {
			// Edit command
			fmt.Printf("  Command [%s]: ", cmd.Command)
			if input, err := reader.ReadString('\n'); err == nil {
				input = strings.TrimSpace(input)
				if input != "" {
					cmd.Command = input
				}
			}
			
			// Edit args
			argsStr := strings.Join(cmd.Args, " ")
			fmt.Printf("  Arguments [%s]: ", argsStr)
			if input, err := reader.ReadString('\n'); err == nil {
				input = strings.TrimSpace(input)
				if input != "" {
					cmd.Args = strings.Fields(input)
				}
			}
			
			// Edit timeout
			fmt.Printf("  Timeout [%s]: ", cmd.Timeout)
			if input, err := reader.ReadString('\n'); err == nil {
				input = strings.TrimSpace(input)
				if input != "" {
					cmd.Timeout = input
				}
			}
		}
		
		cfg.Commands[name] = cmd
		fmt.Println()
	}
	
	// Validate and save
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	
	if err := cfg.Save(dir); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	
	fmt.Printf("✓ Configuration saved to %s\n", filepath.Join(dir, ".kwatch", "kwatch.yaml"))
	return nil
}