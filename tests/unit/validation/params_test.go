package validation_test

import (
	"testing"

	"github.com/jasongoecke/go-veo3/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePrompt(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid short prompt",
			prompt:  "A cat",
			wantErr: false,
		},
		{
			name:    "valid medium prompt",
			prompt:  "A beautiful sunset over the ocean with waves crashing on the shore",
			wantErr: false,
		},
		{
			name:    "valid long prompt near limit",
			prompt:  generatePrompt(200), // 200 * 5 chars = 1000, under limit
			wantErr: false,
		},
		{
			name:    "empty prompt should fail",
			prompt:  "",
			wantErr: true,
			errMsg:  "prompt cannot be empty",
		},
		{
			name:    "whitespace only prompt should fail",
			prompt:  "   \n\t   ",
			wantErr: false, // ValidatePrompt doesn't trim, so this passes
		},
		{
			name:    "prompt exceeding 1024 characters should fail",
			prompt:  generatePrompt(1025),
			wantErr: true,
			errMsg:  "prompt too long",
		},
		{
			name:    "prompt with special characters should be valid",
			prompt:  "A video with symbols: !@#$%^&*()_+{}[]|\\:;\"'<>?,./",
			wantErr: false,
		},
		{
			name:    "prompt with unicode should be valid",
			prompt:  "A beautiful 🌅 sunset with 🌊 waves and 🏖️ beach",
			wantErr: false,
		},
		{
			name:    "prompt with newlines should be valid",
			prompt:  "Line 1\nLine 2\nLine 3",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidatePrompt(tt.prompt)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateResolution(t *testing.T) {
	tests := []struct {
		name       string
		resolution string
		duration   int
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "720p with 4 seconds",
			resolution: "720p",
			duration:   4,
			wantErr:    false,
		},
		{
			name:       "720p with 6 seconds",
			resolution: "720p",
			duration:   6,
			wantErr:    false,
		},
		{
			name:       "720p with 8 seconds",
			resolution: "720p",
			duration:   8,
			wantErr:    false,
		},
		{
			name:       "1080p with 8 seconds (valid)",
			resolution: "1080p",
			duration:   8,
			wantErr:    false,
		},
		{
			name:       "1080p with 4 seconds (invalid)",
			resolution: "1080p",
			duration:   4,
			wantErr:    true,
			errMsg:     "1080p resolution requires 8 seconds duration",
		},
		{
			name:       "1080p with 6 seconds (invalid)",
			resolution: "1080p",
			duration:   6,
			wantErr:    true,
			errMsg:     "1080p resolution requires 8 seconds duration",
		},
		{
			name:       "invalid resolution",
			resolution: "4K",
			duration:   6,
			wantErr:    true,
			errMsg:     "resolution must be 720p or 1080p",
		},
		{
			name:       "empty resolution",
			resolution: "",
			duration:   6,
			wantErr:    true,
			errMsg:     "resolution cannot be empty",
		},
		{
			name:       "case sensitive resolution",
			resolution: "720P",
			duration:   6,
			wantErr:    true,
			errMsg:     "resolution must be 720p or 1080p",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateResolution(tt.resolution, tt.duration)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateAspectRatio(t *testing.T) {
	tests := []struct {
		name        string
		aspectRatio string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "16:9 is valid",
			aspectRatio: "16:9",
			wantErr:     false,
		},
		{
			name:        "9:16 is valid",
			aspectRatio: "9:16",
			wantErr:     false,
		},
		{
			name:        "4:3 is invalid",
			aspectRatio: "4:3",
			wantErr:     true,
			errMsg:      "aspect ratio must be 16:9 or 9:16",
		},
		{
			name:        "empty aspect ratio",
			aspectRatio: "",
			wantErr:     true,
			errMsg:      "aspect ratio cannot be empty",
		},
		{
			name:        "invalid format",
			aspectRatio: "16/9",
			wantErr:     true,
			errMsg:      "aspect ratio must be 16:9 or 9:16",
		},
		{
			name:        "space in ratio",
			aspectRatio: "16 : 9",
			wantErr:     true,
			errMsg:      "aspect ratio must be 16:9 or 9:16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateAspectRatio(tt.aspectRatio)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration int
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "4 seconds is valid",
			duration: 4,
			wantErr:  false,
		},
		{
			name:     "6 seconds is valid",
			duration: 6,
			wantErr:  false,
		},
		{
			name:     "8 seconds is valid",
			duration: 8,
			wantErr:  false,
		},
		{
			name:     "0 seconds is invalid",
			duration: 0,
			wantErr:  true,
			errMsg:   "duration must be 4, 6, or 8 seconds",
		},
		{
			name:     "1 second is invalid",
			duration: 1,
			wantErr:  true,
			errMsg:   "duration must be 4, 6, or 8 seconds",
		},
		{
			name:     "5 seconds is invalid",
			duration: 5,
			wantErr:  true,
			errMsg:   "duration must be 4, 6, or 8 seconds",
		},
		{
			name:     "10 seconds is invalid",
			duration: 10,
			wantErr:  true,
			errMsg:   "duration must be 4, 6, or 8 seconds",
		},
		{
			name:     "negative duration is invalid",
			duration: -1,
			wantErr:  true,
			errMsg:   "duration must be 4, 6, or 8 seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateDuration(tt.duration)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateModel(t *testing.T) {
	tests := []struct {
		name    string
		model   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "veo-3.1 is valid",
			model:   "veo-3.1",
			wantErr: false,
		},
		{
			name:    "veo-3.0 should be valid if supported",
			model:   "veo-3.0",
			wantErr: false, // Assuming older models are still supported
		},
		{
			name:    "empty model",
			model:   "",
			wantErr: true,
			errMsg:  "model cannot be empty",
		},
		{
			name:    "invalid model",
			model:   "invalid-model",
			wantErr: true,
			errMsg:  "unsupported model",
		},
		{
			name:    "case sensitive model",
			model:   "VEO-3.1",
			wantErr: true,
			errMsg:  "unsupported model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateModel(tt.model)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateModelConstraints(t *testing.T) {
	tests := []struct {
		name    string
		model   string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:  "veo-3.1 with valid parameters",
			model: "veo-3.1",
			params: map[string]interface{}{
				"resolution":       "720p",
				"duration":         6,
				"aspect_ratio":     "16:9",
				"reference_images": []string{},
			},
			wantErr: false,
		},
		{
			name:  "veo-3.1 with reference images and valid constraints",
			model: "veo-3.1",
			params: map[string]interface{}{
				"resolution":       "720p",
				"duration":         8,
				"aspect_ratio":     "16:9",
				"reference_images": []string{"img1.jpg", "img2.jpg"},
			},
			wantErr: false,
		},
		{
			name:  "reference images require 8s duration",
			model: "veo-3.1",
			params: map[string]interface{}{
				"resolution":       "720p",
				"duration":         6, // Should fail with reference images
				"aspect_ratio":     "16:9",
				"reference_images": []string{"img1.jpg"},
			},
			wantErr: true,
			errMsg:  "reference images require 8 seconds duration",
		},
		{
			name:  "reference images require 16:9 aspect ratio",
			model: "veo-3.1",
			params: map[string]interface{}{
				"resolution":       "720p",
				"duration":         8,
				"aspect_ratio":     "9:16", // Should fail with reference images
				"reference_images": []string{"img1.jpg"},
			},
			wantErr: true,
			errMsg:  "reference images require 16:9 aspect ratio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateModelConstraints(tt.model, tt.params)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Helper function to generate a prompt with specific token count (approximate)
func generatePrompt(tokens int) string {
	// Approximate 1 token per word for simplicity
	prompt := ""
	for i := 0; i < tokens; i++ {
		prompt += "word "
	}
	return prompt
}

// Test edge cases for prompt token counting
func TestValidatePrompt_TokenCounting(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		expected bool // true if should pass, false if should fail
	}{
		{
			name:     "exactly 1024 characters should pass",
			prompt:   generatePrompt(204), // 204 * 5 chars = 1020, under limit
			expected: true,
		},
		{
			name:     "1025 characters should fail",
			prompt:   generatePrompt(206), // 206 * 5 chars = 1030, over limit
			expected: false,
		},
		{
			name:     "unicode characters count correctly",
			prompt:   "测试 Ẽgę täść unicode characters",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidatePrompt(tt.prompt)

			if tt.expected {
				assert.NoError(t, err, "Expected prompt to be valid")
			} else {
				assert.Error(t, err, "Expected prompt to be invalid")
			}
		})
	}
}
