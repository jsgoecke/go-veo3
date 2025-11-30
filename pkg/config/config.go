package config

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
