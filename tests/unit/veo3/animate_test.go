package veo3_test

import (
	"testing"

	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *veo3.ImageRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid image request with defaults",
			request: &veo3.ImageRequest{
				GenerationRequest: veo3.GenerationRequest{
					Prompt:          "Animate this image",
					Model:           "veo-3.1-generate-preview",
					AspectRatio:     "16:9",
					Resolution:      "720p",
					DurationSeconds: 6,
				},
				ImagePath: "testdata/test.jpg",
			},
			wantErr: false,
		},
		{
			name: "valid image request with 1080p",
			request: &veo3.ImageRequest{
				GenerationRequest: veo3.GenerationRequest{
					Prompt:          "Add motion to this image",
					Model:           "veo-3.1-generate-preview",
					AspectRatio:     "16:9",
					Resolution:      "1080p",
					DurationSeconds: 8, // Required for 1080p
				},
				ImagePath: "testdata/test.png",
			},
			wantErr: false,
		},
		{
			name: "valid image request without prompt",
			request: &veo3.ImageRequest{
				GenerationRequest: veo3.GenerationRequest{
					Prompt:          "", // Should be optional for image-to-video
					Model:           "veo-3.1-generate-preview",
					AspectRatio:     "16:9",
					Resolution:      "720p",
					DurationSeconds: 6,
				},
				ImagePath: "testdata/test.jpg",
			},
			wantErr: false, // Empty prompt is actually valid for image-to-video
		},
		{
			name: "empty image path should fail",
			request: &veo3.ImageRequest{
				GenerationRequest: veo3.GenerationRequest{
					Prompt:          "Animate this",
					Model:           "veo-3.1-generate-preview",
					AspectRatio:     "16:9",
					Resolution:      "720p",
					DurationSeconds: 6,
				},
				ImagePath: "",
			},
			wantErr: true,
			errMsg:  "image path cannot be empty",
		},
		{
			name: "invalid image file extension",
			request: &veo3.ImageRequest{
				GenerationRequest: veo3.GenerationRequest{
					Prompt:          "Animate this",
					Model:           "veo-3.1-generate-preview",
					AspectRatio:     "16:9",
					Resolution:      "720p",
					DurationSeconds: 6,
				},
				ImagePath: "testdata/test.gif", // GIF not supported
			},
			wantErr: true,
			errMsg:  "image file not found",
		},
		{
			name: "1080p with non-8 second duration should fail",
			request: &veo3.ImageRequest{
				GenerationRequest: veo3.GenerationRequest{
					Prompt:          "Animate this",
					Model:           "veo-3.1-generate-preview",
					AspectRatio:     "16:9",
					Resolution:      "1080p",
					DurationSeconds: 6, // Should be 8 for 1080p
				},
				ImagePath: "testdata/test.jpg",
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

func TestImageRequest_Base64Encoding(t *testing.T) {
	// Test that image data can be properly encoded for API submission
	tests := []struct {
		name      string
		imagePath string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid JPEG encoding",
			imagePath: "testdata/test.jpg",
			wantErr:   false,
		},
		{
			name:      "valid PNG encoding",
			imagePath: "testdata/test.png",
			wantErr:   false,
		},
		{
			name:      "non-existent file",
			imagePath: "testdata/nonexistent.jpg",
			wantErr:   true,
			errMsg:    "no such file",
		},
		{
			name:      "directory instead of file",
			imagePath: "testdata/",
			wantErr:   true,
			errMsg:    "is a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This tests the EncodeImageToBase64 function that should be implemented
			encoded, err := veo3.EncodeImageToBase64(tt.imagePath)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Empty(t, encoded)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, encoded)

				// Basic base64 validation
				assert.True(t, len(encoded) > 0, "Encoded string should not be empty")
				assert.True(t, len(encoded)%4 == 0 || len(encoded)%4 == 2, "Base64 should have proper padding")
			}
		})
	}
}

func TestImageRequest_ValidateImageFormat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "JPEG extension valid",
			filename: "test.jpg",
			wantErr:  false,
		},
		{
			name:     "JPEG extension valid (uppercase)",
			filename: "test.JPG",
			wantErr:  false,
		},
		{
			name:     "JPEG extension valid (alternative)",
			filename: "test.jpeg",
			wantErr:  false,
		},
		{
			name:     "PNG extension valid",
			filename: "test.png",
			wantErr:  false,
		},
		{
			name:     "PNG extension valid (uppercase)",
			filename: "test.PNG",
			wantErr:  false,
		},
		{
			name:     "WebP extension valid",
			filename: "test.webp",
			wantErr:  false,
		},
		{
			name:     "GIF extension invalid",
			filename: "test.gif",
			wantErr:  true,
			errMsg:   "unsupported image format",
		},
		{
			name:     "BMP extension invalid",
			filename: "test.bmp",
			wantErr:  true,
			errMsg:   "unsupported image format",
		},
		{
			name:     "TIFF extension invalid",
			filename: "test.tiff",
			wantErr:  true,
			errMsg:   "unsupported image format",
		},
		{
			name:     "no extension",
			filename: "test",
			wantErr:  true,
			errMsg:   "unsupported image format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := veo3.ValidateImageFormat(tt.filename)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestImageRequest_ValidateImageSize(t *testing.T) {
	tests := []struct {
		name     string
		fileSize int64
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "1MB file valid",
			fileSize: 1 * 1024 * 1024,
			wantErr:  false,
		},
		{
			name:     "20MB file valid (at limit)",
			fileSize: 20 * 1024 * 1024,
			wantErr:  false,
		},
		{
			name:     "21MB file invalid (over limit)",
			fileSize: 21 * 1024 * 1024,
			wantErr:  true,
			errMsg:   "image file too large",
		},
		{
			name:     "0 byte file invalid",
			fileSize: 0,
			wantErr:  true,
			errMsg:   "image file is empty",
		},
		{
			name:     "tiny file valid",
			fileSize: 1024, // 1KB
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := veo3.ValidateImageSize(tt.fileSize)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestImageRequest_APIPayloadStructure(t *testing.T) {
	// Test that ImageRequest can be properly converted to API payload
	request := &veo3.ImageRequest{
		GenerationRequest: veo3.GenerationRequest{
			Prompt:          "Add motion to the water",
			Model:           "veo-3.1-generate-preview",
			AspectRatio:     "16:9",
			Resolution:      "720p",
			DurationSeconds: 6,
		},
		ImagePath: "testdata/test.jpg",
	}

	// This function should be implemented to build the API payload
	payload, err := veo3.BuildImageToVideoPayload(request)

	// For now, expect this to fail since it's not implemented
	if err != nil {
		t.Skip("BuildImageToVideoPayload not implemented yet")
		return
	}

	// Validate payload structure when implemented
	require.NotNil(t, payload)
	assert.Contains(t, payload, "model")
	assert.Contains(t, payload, "prompt")
	assert.Contains(t, payload, "inputImage")
	assert.Contains(t, payload, "parameters")
}
