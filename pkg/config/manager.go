package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Manager manages configuration loading and access
type Manager struct {
	configPath string
}

// NewManager creates a new configuration manager
func NewManager(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

// Load reads the configuration from file or environment
func (m *Manager) Load() (*Configuration, error) {
	if m.configPath != "" {
		viper.SetConfigFile(m.configPath)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}

		configDir := filepath.Join(home, ".config", "veo3")
		configPath := filepath.Join(configDir, "config.yaml")
		viper.SetConfigFile(configPath)
	}

	viper.SetEnvPrefix("VEO3")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("default_model", DefaultModel)
	viper.SetDefault("default_resolution", DefaultResolution)
	viper.SetDefault("default_aspect_ratio", DefaultAspectRatio)
	viper.SetDefault("default_duration", DefaultDuration)
	viper.SetDefault("poll_interval_seconds", DefaultPollInterval)
	viper.SetDefault("output_directory", ".")

	// If config file exists, read it
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is okay, use defaults
	}

	var cfg Configuration
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Manual bind for API Key Env if needed, or rely on automatic env
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("GEMINI_API_KEY")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to the config file
func (m *Manager) Save(cfg *Configuration) error {
	configPath := m.getConfigPath()

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	viper.Set("api_key", cfg.APIKey)
	viper.Set("default_model", cfg.DefaultModel)
	viper.Set("default_resolution", cfg.DefaultResolution)
	viper.Set("default_aspect_ratio", cfg.DefaultAspectRatio)
	viper.Set("default_duration", cfg.DefaultDuration)
	viper.Set("poll_interval_seconds", cfg.PollIntervalSeconds)
	viper.Set("output_directory", cfg.OutputDirectory)

	return viper.WriteConfigAs(configPath)
}

// ConfigFile returns the path to the configuration file
func (m *Manager) ConfigFile() string {
	return m.getConfigPath()
}

// getConfigPath returns the configuration file path
func (m *Manager) getConfigPath() string {
	if m.configPath != "" {
		return m.configPath
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".config", "veo3", "config.yaml")
}

// Load reads the configuration from file or environment (package-level convenience function)
func Load() (*Configuration, error) {
	m := NewManager("")
	return m.Load()
}
