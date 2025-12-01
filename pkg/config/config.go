package config

import (
	"fmt"

	"github.com/jasongoecke/go-veo3/pkg/veo3"
)

// Configuration User settings and preferences.
type Configuration struct {
	APIKey              string `yaml:"api_key,omitempty" json:"-"`
	APIKeyEnv           string `yaml:"api_key_env,omitempty" json:"api_key_env,omitempty"`
	DefaultModel        string `yaml:"default_model" json:"default_model"`
	DefaultResolution   string `yaml:"default_resolution" json:"default_resolution"`
	DefaultAspectRatio  string `yaml:"default_aspect_ratio" json:"default_aspect_ratio"`
	DefaultDuration     int    `yaml:"default_duration" json:"default_duration"`
	OutputDirectory     string `yaml:"output_directory" json:"output_directory"`
	PollIntervalSeconds int    `yaml:"poll_interval_seconds" json:"poll_interval_seconds"`
	ConfigVersion       string `yaml:"version" json:"version"`
}

// Validate checks if the configuration values are valid
func (c *Configuration) Validate() error {
	// Validate default model exists in registry
	if c.DefaultModel != "" {
		if _, exists := veo3.GetModel(c.DefaultModel); !exists {
			return fmt.Errorf("invalid default_model '%s': model not found in registry. Use 'veo3 models list' to see available models", c.DefaultModel)
		}
	}

	// Additional validation can be added here for other fields

	return nil
}
