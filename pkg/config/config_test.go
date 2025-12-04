package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfiguration_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Configuration
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			config: Configuration{
				DefaultModel:        "veo-3.1-generate-preview",
				DefaultResolution:   "720p",
				DefaultAspectRatio:  "16:9",
				DefaultDuration:     8,
				OutputDirectory:     "./output",
				PollIntervalSeconds: 10,
			},
			wantErr: false,
		},
		{
			name: "empty default model is valid",
			config: Configuration{
				DefaultModel:        "",
				DefaultResolution:   "720p",
				DefaultAspectRatio:  "16:9",
				DefaultDuration:     8,
				OutputDirectory:     "./output",
				PollIntervalSeconds: 10,
			},
			wantErr: false,
		},
		{
			name: "invalid default model",
			config: Configuration{
				DefaultModel:        "non-existent-model",
				DefaultResolution:   "720p",
				DefaultAspectRatio:  "16:9",
				DefaultDuration:     8,
				OutputDirectory:     "./output",
				PollIntervalSeconds: 10,
			},
			wantErr: true,
			errMsg:  "invalid default_model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfiguration_Fields(t *testing.T) {
	config := Configuration{
		APIKey:              "test-api-key",
		APIKeyEnv:           "GEMINI_API_KEY",
		DefaultModel:        "veo-3.1-generate-preview",
		DefaultResolution:   "1080p",
		DefaultAspectRatio:  "9:16",
		DefaultDuration:     6,
		OutputDirectory:     "/tmp/videos",
		PollIntervalSeconds: 15,
		ConfigVersion:       "1.0",
	}

	assert.Equal(t, "test-api-key", config.APIKey)
	assert.Equal(t, "GEMINI_API_KEY", config.APIKeyEnv)
	assert.Equal(t, "veo-3.1-generate-preview", config.DefaultModel)
	assert.Equal(t, "1080p", config.DefaultResolution)
	assert.Equal(t, "9:16", config.DefaultAspectRatio)
	assert.Equal(t, 6, config.DefaultDuration)
	assert.Equal(t, "/tmp/videos", config.OutputDirectory)
	assert.Equal(t, 15, config.PollIntervalSeconds)
	assert.Equal(t, "1.0", config.ConfigVersion)
}
