package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePrompt(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty prompt",
			prompt:  "",
			wantErr: true,
			errMsg:  "prompt cannot be empty",
		},
		{
			name:    "valid prompt",
			prompt:  "A beautiful sunset over the ocean",
			wantErr: false,
		},
		{
			name:    "prompt at max length",
			prompt:  strings.Repeat("a", 1024),
			wantErr: false,
		},
		{
			name:    "prompt too long",
			prompt:  strings.Repeat("a", 1025),
			wantErr: true,
			errMsg:  "prompt too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePrompt(tt.prompt)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateResolution(t *testing.T) {
	tests := []struct {
		name     string
		res      string
		duration []int
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "empty resolution",
			res:     "",
			wantErr: true,
			errMsg:  "resolution cannot be empty",
		},
		{
			name:    "valid 720p",
			res:     "720p",
			wantErr: false,
		},
		{
			name:     "valid 1080p with 8s",
			res:      "1080p",
			duration: []int{8},
			wantErr:  false,
		},
		{
			name:     "1080p with wrong duration",
			res:      "1080p",
			duration: []int{4},
			wantErr:  true,
			errMsg:   "1080p resolution requires 8 seconds duration",
		},
		{
			name:    "invalid resolution",
			res:     "4K",
			wantErr: true,
			errMsg:  "resolution must be 720p or 1080p",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResolution(tt.res, tt.duration...)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAspectRatio(t *testing.T) {
	tests := []struct {
		name    string
		ratio   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty aspect ratio",
			ratio:   "",
			wantErr: true,
			errMsg:  "aspect ratio cannot be empty",
		},
		{
			name:    "valid 16:9",
			ratio:   "16:9",
			wantErr: false,
		},
		{
			name:    "valid 9:16",
			ratio:   "9:16",
			wantErr: false,
		},
		{
			name:    "invalid ratio",
			ratio:   "4:3",
			wantErr: true,
			errMsg:  "aspect ratio must be 16:9 or 9:16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAspectRatio(tt.ratio)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDuration(t *testing.T) {
	tests := []struct {
		name    string
		seconds int
		wantErr bool
	}{
		{"valid 4 seconds", 4, false},
		{"valid 6 seconds", 6, false},
		{"valid 8 seconds", 8, false},
		{"invalid 3 seconds", 3, true},
		{"invalid 5 seconds", 5, true},
		{"invalid 10 seconds", 10, true},
		{"invalid 0 seconds", 0, true},
		{"invalid negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDuration(tt.seconds)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "duration must be 4, 6, or 8 seconds")
			} else {
				assert.NoError(t, err)
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
			name:    "empty model",
			model:   "",
			wantErr: true,
			errMsg:  "model cannot be empty",
		},
		{
			name:    "valid veo-3.1 model",
			model:   "veo-3.1-generate-preview",
			wantErr: false,
		},
		{
			name:    "valid veo-3.0 model",
			model:   "veo-3.0-test",
			wantErr: false,
		},
		{
			name:    "valid veo-2.0 model",
			model:   "veo-2.0-model",
			wantErr: false,
		},
		{
			name:    "invalid model",
			model:   "invalid-model",
			wantErr: true,
			errMsg:  "unsupported model",
		},
		{
			name:    "partial match not valid",
			model:   "veo",
			wantErr: true,
			errMsg:  "unsupported model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModel(tt.model)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
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
			name:    "invalid model",
			model:   "invalid",
			params:  map[string]interface{}{},
			wantErr: true,
			errMsg:  "unsupported model",
		},
		{
			name:    "valid model without constraints",
			model:   "veo-3.1-generate-preview",
			params:  map[string]interface{}{},
			wantErr: false,
		},
		{
			name:  "reference images with wrong duration",
			model: "veo-3.1-generate-preview",
			params: map[string]interface{}{
				"reference_images": []string{"img1.jpg"},
				"duration":         4,
			},
			wantErr: true,
			errMsg:  "reference images require 8 seconds duration",
		},
		{
			name:  "reference images with wrong aspect ratio",
			model: "veo-3.1-generate-preview",
			params: map[string]interface{}{
				"reference_images": []string{"img1.jpg"},
				"duration":         8,
				"aspect_ratio":     "9:16",
			},
			wantErr: true,
			errMsg:  "reference images require 16:9 aspect ratio",
		},
		{
			name:  "reference images with correct constraints",
			model: "veo-3.1-generate-preview",
			params: map[string]interface{}{
				"reference_images": []string{"img1.jpg"},
				"duration":         8,
				"aspect_ratio":     "16:9",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModelConstraints(tt.model, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
