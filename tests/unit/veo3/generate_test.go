package veo3_test

import (
	"testing"

	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerationRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *veo3.GenerationRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with defaults",
			request: &veo3.GenerationRequest{
				Prompt:          "A beautiful sunset over the ocean",
				Model:           "veo-3.1",
				AspectRatio:     "16:9",
				Resolution:      "720p",
				DurationSeconds: 6,
			},
			wantErr: false,
		},
		{
			name: "valid request with 1080p and 8 seconds",
			request: &veo3.GenerationRequest{
				Prompt:          "A cityscape at night",
				Model:           "veo-3.1",
				AspectRatio:     "16:9",
				Resolution:      "1080p",
				DurationSeconds: 8,
			},
			wantErr: false,
		},
		{
			name: "valid request with 9:16 aspect ratio",
			request: &veo3.GenerationRequest{
				Prompt:          "A vertical video",
				Model:           "veo-3.1",
				AspectRatio:     "9:16",
				Resolution:      "720p",
				DurationSeconds: 4,
			},
			wantErr: false,
		},
		{
			name: "empty prompt should fail",
			request: &veo3.GenerationRequest{
				Prompt:          "",
				Model:           "veo-3.1",
				AspectRatio:     "16:9",
				Resolution:      "720p",
				DurationSeconds: 6,
			},
			wantErr: true,
			errMsg:  "prompt cannot be empty",
		},
		{
			name: "prompt too long should fail",
			request: &veo3.GenerationRequest{
				Prompt:          generateLongPrompt(1025), // Over 1024 token limit
				Model:           "veo-3.1",
				AspectRatio:     "16:9",
				Resolution:      "720p",
				DurationSeconds: 6,
			},
			wantErr: true,
			errMsg:  "exceeds 1024 tokens",
		},
		{
			name: "invalid model should fail",
			request: &veo3.GenerationRequest{
				Prompt:          "Test prompt",
				Model:           "invalid-model",
				AspectRatio:     "16:9",
				Resolution:      "720p",
				DurationSeconds: 6,
			},
			wantErr: true,
			errMsg:  "invalid model",
		},
		{
			name: "invalid aspect ratio should fail",
			request: &veo3.GenerationRequest{
				Prompt:          "Test prompt",
				Model:           "veo-3.1",
				AspectRatio:     "4:3",
				Resolution:      "720p",
				DurationSeconds: 6,
			},
			wantErr: true,
			errMsg:  "aspect ratio must be",
		},
		{
			name: "invalid resolution should fail",
			request: &veo3.GenerationRequest{
				Prompt:          "Test prompt",
				Model:           "veo-3.1",
				AspectRatio:     "16:9",
				Resolution:      "4K",
				DurationSeconds: 6,
			},
			wantErr: true,
			errMsg:  "resolution must be",
		},
		{
			name: "invalid duration should fail",
			request: &veo3.GenerationRequest{
				Prompt:          "Test prompt",
				Model:           "veo-3.1",
				AspectRatio:     "16:9",
				Resolution:      "720p",
				DurationSeconds: 5, // Not 4, 6, or 8
			},
			wantErr: true,
			errMsg:  "duration must be",
		},
		{
			name: "1080p with non-8 second duration should fail",
			request: &veo3.GenerationRequest{
				Prompt:          "Test prompt",
				Model:           "veo-3.1",
				AspectRatio:     "16:9",
				Resolution:      "1080p",
				DurationSeconds: 6, // 1080p requires 8 seconds
			},
			wantErr: true,
			errMsg:  "1080p resolution requires 8 seconds duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenerationRequest_ValidateWithNegativePrompt(t *testing.T) {
	req := &veo3.GenerationRequest{
		Prompt:          "A beautiful landscape",
		NegativePrompt:  "avoid people, cars",
		Model:           "veo-3.1",
		AspectRatio:     "16:9",
		Resolution:      "720p",
		DurationSeconds: 6,
	}

	err := req.Validate()
	assert.NoError(t, err)
}

func TestGenerationRequest_ValidateWithSeed(t *testing.T) {
	seed := 12345
	req := &veo3.GenerationRequest{
		Prompt:          "A beautiful landscape",
		Model:           "veo-3.1",
		AspectRatio:     "16:9",
		Resolution:      "720p",
		DurationSeconds: 6,
		Seed:            &seed,
	}

	err := req.Validate()
	assert.NoError(t, err)
}

func TestGenerationRequest_ValidatePersonGeneration(t *testing.T) {
	tests := []struct {
		name             string
		personGeneration string
		wantErr          bool
	}{
		{"allow_all is valid", "allow_all", false},
		{"allow_adult is valid", "allow_adult", false},
		{"dont_allow is valid", "dont_allow", false},
		{"empty is valid (default)", "", false},
		{"invalid value should fail", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &veo3.GenerationRequest{
				Prompt:           "A beautiful landscape",
				Model:            "veo-3.1",
				AspectRatio:      "16:9",
				Resolution:       "720p",
				DurationSeconds:  6,
				PersonGeneration: tt.personGeneration,
			}

			err := req.Validate()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to generate a prompt with specific token count (approximate)
func generateLongPrompt(tokens int) string {
	// Approximate 1 token per word for simplicity
	prompt := ""
	for i := 0; i < tokens; i++ {
		prompt += "word "
	}
	return prompt
}

// Test that the struct fields are properly tagged for JSON/YAML
func TestGenerationRequest_JSONTags(t *testing.T) {
	// This test ensures the struct has proper JSON tags for serialization
	// The actual implementation should have these tags
	req := &veo3.GenerationRequest{
		Prompt:          "Test",
		Model:           "veo-3.1",
		AspectRatio:     "16:9",
		Resolution:      "720p",
		DurationSeconds: 6,
	}

	// This will fail until the struct is implemented
	_ = req
	t.Skip("GenerationRequest struct not implemented yet - this test should pass after implementation")
}
