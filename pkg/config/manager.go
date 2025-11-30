package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Manager handles configuration loading and saving
type Manager struct {
	ConfigFile string
}

// NewManager creates a new configuration manager
func NewManager(configFile string) *Manager {
	return &Manager{
		ConfigFile: configFile,
	}
}

// Load loads the configuration from file
func (m *Manager) Load() (*Configuration, error) {
	if m.ConfigFile != "" {
		viper.SetConfigFile(m.ConfigFile)
	}

	// Attempt to read config, but don't error if file doesn't exist (use defaults)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Configuration
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// Set defaults if not present
	if config.DefaultModel == "" {
		config.DefaultModel = DefaultModel
	}
	if config.DefaultResolution == "" {
		config.DefaultResolution = DefaultResolution
	}
	if config.DefaultAspectRatio == "" {
		config.DefaultAspectRatio = DefaultAspectRatio
	}
	if config.DefaultDuration == 0 {
		config.DefaultDuration = DefaultDuration
	}
	if config.PollIntervalSeconds == 0 {
		config.PollIntervalSeconds = DefaultPollInterval
	}
	if config.ConfigVersion == "" {
		config.ConfigVersion = DefaultConfigVersion
	}

	return &config, nil
}

// Save saves the configuration to file
func (m *Manager) Save(config *Configuration) error {
	if m.ConfigFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configDir := filepath.Join(home, ".config", "veo3")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return err
		}
		m.ConfigFile = filepath.Join(configDir, "config.yaml")
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	// 0600 permissions for security
	return os.WriteFile(m.ConfigFile, data, 0600)
}
