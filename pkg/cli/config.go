package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jasongoecke/go-veo3/internal/format"
	"github.com/jasongoecke/go-veo3/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newConfigCmd creates the config command group
func newConfigCmd() *cobra.Command {
	// Create fresh command instance to avoid flag redefinition in tests
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration and preferences",
		Long: `Manage CLI configuration including API credentials, defaults, and preferences.

Configuration is stored in YAML format with secure file permissions.
The config file location follows XDG Base Directory specification:
- Linux/macOS: ~/.config/veo3/config.yaml
- Windows: %APPDATA%\veo3\config.yaml`,
		Example: `  # Initialize configuration interactively
  veo3 config init

  # Set API key
  veo3 config set api-key YOUR_API_KEY

  # Show current configuration (sensitive data masked)
  veo3 config show

  # Reset configuration to defaults
  veo3 config reset`,
	}

	// Add subcommands
	configCmd.AddCommand(newConfigInitCmd())
	configCmd.AddCommand(newConfigSetCmd())
	configCmd.AddCommand(newConfigGetCmd())
	configCmd.AddCommand(newConfigShowCmd())
	configCmd.AddCommand(newConfigResetCmd())

	return configCmd
}

// newConfigInitCmd creates the 'config init' command
func newConfigInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration interactively",
		Long: `Initialize configuration by prompting for required settings.

This command walks you through setting up your API key, default preferences,
and output directory. It creates the config file with secure permissions.`,
		Example: `  # Interactive setup
  veo3 config init

  # Non-interactive setup with provided values
  veo3 config init --api-key YOUR_API_KEY --output ./videos/`,
		RunE: runConfigInit,
	}

	cmd.Flags().String("api-key", "", "API key (skips interactive prompt)")
	cmd.Flags().String("output", "", "Default output directory")
	cmd.Flags().String("model", "", "Default model")
	cmd.Flags().String("resolution", "", "Default resolution")
	cmd.Flags().Bool("force", false, "Overwrite existing config file")

	return cmd
}

// newConfigSetCmd creates the 'config set' command
func newConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Long: `Set a specific configuration key to a value.

Available configuration keys:
- api-key: Google Gemini API key
- default-model: Default model for generations
- default-resolution: Default resolution (720p or 1080p)
- default-duration: Default duration (4, 6, or 8)
- default-aspect-ratio: Default aspect ratio (16:9 or 9:16)
- output-directory: Default output directory for videos
- poll-interval: Status polling interval in seconds`,
		Example: `  # Set API key
  veo3 config set api-key YOUR_API_KEY

  # Set default resolution
  veo3 config set default-resolution 1080p

  # Set output directory
  veo3 config set output-directory ./my-videos/`,
		Args: cobra.ExactArgs(2),
		RunE: runConfigSet,
	}

	return cmd
}

// newConfigGetCmd creates the 'config get' command
func newConfigGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get a configuration value",
		Long: `Get the value of a specific configuration key.

For sensitive values like API keys, only the last 4 characters are shown
unless --show-sensitive is used.`,
		Example: `  # Get API key (masked)
  veo3 config get api-key

  # Get default model
  veo3 config get default-model

  # Get all values in JSON
  veo3 config get --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: runConfigGet,
	}

	cmd.Flags().Bool("show-sensitive", false, "Show full sensitive values (use with caution)")

	return cmd
}

// newConfigShowCmd creates the 'config show' command
func newConfigShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long: `Show the current configuration with all keys and values.

Sensitive values like API keys are masked by default. Use --show-sensitive
to display full values (use with caution in shared environments).`,
		Example: `  # Show configuration (sensitive data masked)
  veo3 config show

  # Show configuration in JSON format
  veo3 config show --json

  # Show with sensitive data (use with caution)
  veo3 config show --show-sensitive`,
		RunE: runConfigShow,
	}

	cmd.Flags().Bool("show-sensitive", false, "Show full sensitive values (use with caution)")
	cmd.Flags().Bool("pretty", false, "Pretty-print JSON output (with --json)")

	return cmd
}

// newConfigResetCmd creates the 'config reset' command
func newConfigResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		Long: `Reset configuration to default values.

This removes all custom settings and restores the configuration to its
default state. The API key is cleared and must be set again.`,
		Example: `  # Reset with confirmation prompt
  veo3 config reset

  # Reset without confirmation
  veo3 config reset --force`,
		RunE: runConfigReset,
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}

// Command implementations

func runConfigInit(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	jsonFormat := viper.GetBool("json")
	out := cmd.OutOrStdout()

	// Get config file path from viper (respects --config flag)
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		// If no config file specified, get it from root flag
		configPath, _ = cmd.Flags().GetString("config")
	}

	// Check if config already exists
	manager := config.NewManager(configPath)
	if !force {
		if _, err := manager.Load(); err == nil {
			return fmt.Errorf("configuration already exists (use --force to overwrite)")
		}
	}

	// Get provided values or prompt for them
	apiKey, _ := cmd.Flags().GetString("api-key")
	outputDir, _ := cmd.Flags().GetString("output")
	model, _ := cmd.Flags().GetString("model")
	resolution, _ := cmd.Flags().GetString("resolution")

	// Interactive prompts if values not provided and not in JSON mode
	if !jsonFormat && apiKey == "" {
		_, _ = fmt.Fprint(out, "Enter your Google Gemini API key: ")
		_, _ = fmt.Scanln(&apiKey)
	}

	if !jsonFormat && outputDir == "" {
		_, _ = fmt.Fprint(out, "Enter default output directory (default: current directory): ")
		_, _ = fmt.Scanln(&outputDir)
		if outputDir == "" {
			outputDir = "."
		}
	}

	// Create configuration with defaults
	cfg := &config.Configuration{
		APIKey:              apiKey,
		DefaultModel:        config.DefaultModel,
		DefaultResolution:   config.DefaultResolution,
		DefaultAspectRatio:  config.DefaultAspectRatio,
		DefaultDuration:     config.DefaultDuration,
		OutputDirectory:     outputDir,
		PollIntervalSeconds: config.DefaultPollInterval,
		ConfigVersion:       config.DefaultConfigVersion,
	}

	// Override with provided values
	if model != "" {
		cfg.DefaultModel = model
	}
	if resolution != "" {
		cfg.DefaultResolution = resolution
	}

	// Save configuration
	err := manager.Save(cfg)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	if jsonFormat {
		result := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"message":     "Configuration initialized successfully",
				"config_file": manager.ConfigFile(),
			},
		}
		jsonOutput, _ := format.FormatGenericJSON(result)
		_, _ = fmt.Fprintln(out, jsonOutput)
	} else {
		_, _ = fmt.Fprintf(out, "✓ Configuration initialized successfully\n")
		_, _ = fmt.Fprintf(out, "Config file: %s\n", manager.ConfigFile())
		_, _ = fmt.Fprintf(out, "\nNext steps:\n")
		_, _ = fmt.Fprintf(out, "1. Set your API key: veo3 config set api-key YOUR_API_KEY\n")
		_, _ = fmt.Fprintf(out, "2. Generate your first video: veo3 generate --prompt 'A beautiful sunset'\n")
	}

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]
	jsonFormat := viper.GetBool("json")
	out := cmd.OutOrStdout()

	// Get config file path from viper (respects --config flag)
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		configPath, _ = cmd.Root().PersistentFlags().GetString("config")
	}

	// Load existing configuration
	manager := config.NewManager(configPath)
	cfg, err := manager.Load()
	if err != nil {
		// Create new config if none exists
		cfg = &config.Configuration{
			DefaultModel:        config.DefaultModel,
			DefaultResolution:   config.DefaultResolution,
			DefaultAspectRatio:  config.DefaultAspectRatio,
			DefaultDuration:     config.DefaultDuration,
			OutputDirectory:     ".",
			PollIntervalSeconds: config.DefaultPollInterval,
			ConfigVersion:       config.DefaultConfigVersion,
		}
	}

	// Set the value
	switch strings.ToLower(key) {
	case "api-key", "api_key":
		cfg.APIKey = value
	case "default-model", "default_model":
		cfg.DefaultModel = value
	case "default-resolution", "default_resolution":
		cfg.DefaultResolution = value
	case "default-duration", "default_duration":
		if duration, err := strconv.Atoi(value); err == nil {
			cfg.DefaultDuration = duration
		} else {
			return fmt.Errorf("invalid duration value: %s (must be a number)", value)
		}
	case "default-aspect-ratio", "default_aspect_ratio":
		cfg.DefaultAspectRatio = value
	case "output-directory", "output_directory":
		cfg.OutputDirectory = value
	case "poll-interval", "poll_interval":
		if interval, err := strconv.Atoi(value); err == nil {
			cfg.PollIntervalSeconds = interval
		} else {
			return fmt.Errorf("invalid poll interval value: %s (must be a number)", value)
		}
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	// Save configuration
	err = manager.Save(cfg)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	if jsonFormat {
		result := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"key":     key,
				"message": "Configuration updated successfully",
			},
		}
		jsonOutput, _ := format.FormatGenericJSON(result)
		_, _ = fmt.Fprintln(out, jsonOutput)
	} else {
		_, _ = fmt.Fprintf(out, "✓ Configuration updated: %s\n", key)
	}

	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	showSensitive, _ := cmd.Flags().GetBool("show-sensitive")
	jsonFormat := viper.GetBool("json")
	out := cmd.OutOrStdout()

	// Get config file path from viper (respects --config flag)
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		configPath, _ = cmd.Root().PersistentFlags().GetString("config")
	}

	// Load configuration
	manager := config.NewManager(configPath)
	cfg, err := manager.Load()
	if err != nil {
		return fmt.Errorf("no configuration found (run 'veo3 config init' first)")
	}

	if len(args) == 0 {
		// Show all values
		return runConfigShow(cmd, args)
	}

	key := args[0]
	var value string

	switch strings.ToLower(key) {
	case "api-key", "api_key":
		value = cfg.APIKey
		if !showSensitive && value != "" {
			value = maskSensitiveValue(value)
		}
	case "default-model", "default_model":
		value = cfg.DefaultModel
	case "default-resolution", "default_resolution":
		value = cfg.DefaultResolution
	case "default-duration", "default_duration":
		value = fmt.Sprintf("%d", cfg.DefaultDuration)
	case "default-aspect-ratio", "default_aspect_ratio":
		value = cfg.DefaultAspectRatio
	case "output-directory", "output_directory":
		value = cfg.OutputDirectory
	case "poll-interval", "poll_interval":
		value = fmt.Sprintf("%d", cfg.PollIntervalSeconds)
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	if jsonFormat {
		result := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"key":   key,
				"value": value,
			},
		}
		jsonOutput, _ := format.FormatGenericJSON(result)
		_, _ = fmt.Fprintln(out, jsonOutput)
	} else {
		_, _ = fmt.Fprintf(out, "%s: %s\n", key, value)
	}

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	showSensitive, _ := cmd.Flags().GetBool("show-sensitive")
	jsonFormat := viper.GetBool("json")
	out := cmd.OutOrStdout()
	// pretty, _ := cmd.Flags().GetBool("pretty") // TODO: Use for formatted output

	// Get config file path from viper (respects --config flag)
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		configPath, _ = cmd.Root().PersistentFlags().GetString("config")
	}

	// Load configuration
	manager := config.NewManager(configPath)
	cfg, err := manager.Load()
	if err != nil {
		return fmt.Errorf("no configuration found (run 'veo3 config init' first)")
	}

	if jsonFormat {
		// Create config copy for output (mask sensitive if needed)
		outputCfg := *cfg
		if !showSensitive && outputCfg.APIKey != "" {
			outputCfg.APIKey = maskSensitiveValue(outputCfg.APIKey)
		}

		jsonOutput, err := format.FormatConfigJSON(outputCfg)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(out, jsonOutput)
	} else {
		// Human-readable format
		_, _ = fmt.Fprintf(out, "Configuration File: %s\n\n", manager.ConfigFile())

		apiKey := cfg.APIKey
		if !showSensitive && apiKey != "" {
			apiKey = maskSensitiveValue(apiKey)
		}

		_, _ = fmt.Fprintf(out, "API Key: %s\n", apiKey)
		_, _ = fmt.Fprintf(out, "Default Model: %s\n", cfg.DefaultModel)
		_, _ = fmt.Fprintf(out, "Default Resolution: %s\n", cfg.DefaultResolution)
		_, _ = fmt.Fprintf(out, "Default Duration: %ds\n", cfg.DefaultDuration)
		_, _ = fmt.Fprintf(out, "Default Aspect Ratio: %s\n", cfg.DefaultAspectRatio)
		_, _ = fmt.Fprintf(out, "Output Directory: %s\n", cfg.OutputDirectory)
		_, _ = fmt.Fprintf(out, "Poll Interval: %ds\n", cfg.PollIntervalSeconds)
		_, _ = fmt.Fprintf(out, "Config Version: %s\n", cfg.ConfigVersion)
	}

	return nil
}

func runConfigReset(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	jsonFormat := viper.GetBool("json")
	out := cmd.OutOrStdout()

	// Get config file path from viper (respects --config flag)
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		configPath, _ = cmd.Root().PersistentFlags().GetString("config")
	}

	// Confirmation prompt
	if !force && !jsonFormat {
		_, _ = fmt.Fprint(out, "Reset configuration to defaults? This will clear your API key. (y/N): ")
		var response string
		_, _ = fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			_, _ = fmt.Fprintln(out, "Reset cancelled")
			return nil
		}
	}

	// Create default configuration
	cfg := &config.Configuration{
		APIKey:              "", // Cleared
		DefaultModel:        config.DefaultModel,
		DefaultResolution:   config.DefaultResolution,
		DefaultAspectRatio:  config.DefaultAspectRatio,
		DefaultDuration:     config.DefaultDuration,
		OutputDirectory:     ".",
		PollIntervalSeconds: config.DefaultPollInterval,
		ConfigVersion:       config.DefaultConfigVersion,
	}

	// Save configuration
	manager := config.NewManager(configPath)
	err := manager.Save(cfg)
	if err != nil {
		return fmt.Errorf("failed to reset configuration: %w", err)
	}

	if jsonFormat {
		result := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"message": "Configuration reset to defaults",
			},
		}
		jsonOutput, _ := format.FormatGenericJSON(result)
		_, _ = fmt.Fprintln(out, jsonOutput)
	} else {
		_, _ = fmt.Fprintf(out, "✓ Configuration reset to defaults\n")
		_, _ = fmt.Fprintf(out, "Remember to set your API key: veo3 config set api-key YOUR_API_KEY\n")
	}

	return nil
}

// Helper functions

func maskSensitiveValue(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return "****" + value[len(value)-4:]
}
