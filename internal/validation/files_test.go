package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateImageSize(t *testing.T) {
	tests := []struct {
		name    string
		size    int64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid size",
			size:    1024 * 1024, // 1MB
			wantErr: false,
		},
		{
			name:    "negative size",
			size:    -1,
			wantErr: true,
			errMsg:  "invalid file size",
		},
		{
			name:    "zero size",
			size:    0,
			wantErr: true,
			errMsg:  "image file is empty",
		},
		{
			name:    "at max size",
			size:    MaxImageSize,
			wantErr: false,
		},
		{
			name:    "over max size",
			size:    MaxImageSize + 1,
			wantErr: true,
			errMsg:  "image file too large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageSize(tt.size)
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

func TestValidateImageFormat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"jpg extension", "image.jpg", false},
		{"jpeg extension", "image.jpeg", false},
		{"png extension", "image.png", false},
		{"webp extension", "image.webp", false},
		{"uppercase JPG", "IMAGE.JPG", false},
		{"invalid extension", "image.gif", true},
		{"no extension", "image", true},
		{"txt file", "image.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageFormat(tt.filename)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported image format")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		format string
		want   bool
	}{
		{"jpeg", true},
		{"jpg", true},
		{"png", true},
		{"webp", true},
		{"JPEG", true},
		{"PNG", true},
		{"gif", false},
		{"bmp", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := isValidFormat(tt.format)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateCompatibleDimensions(t *testing.T) {
	tests := []struct {
		name    string
		w1, h1  int
		w2, h2  int
		wantErr bool
		errMsg  string
	}{
		{
			name: "matching dimensions",
			w1:   1920, h1: 1080,
			w2: 1920, h2: 1080,
			wantErr: false,
		},
		{
			name: "different widths",
			w1:   1920, h1: 1080,
			w2: 1280, h2: 1080,
			wantErr: true,
			errMsg:  "image dimensions mismatch",
		},
		{
			name: "different heights",
			w1:   1920, h1: 1080,
			w2: 1920, h2: 720,
			wantErr: true,
			errMsg:  "image dimensions mismatch",
		},
		{
			name: "zero width",
			w1:   0, h1: 1080,
			w2: 1920, h2: 1080,
			wantErr: true,
			errMsg:  "invalid dimensions",
		},
		{
			name: "negative dimension",
			w1:   -1, h1: 1080,
			w2: 1920, h2: 1080,
			wantErr: true,
			errMsg:  "invalid dimensions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCompatibleDimensions(tt.w1, tt.h1, tt.w2, tt.h2)
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

func TestValidateVideoFileForExtension(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		setup   func() string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid mp4 file",
			setup: func() string {
				path := filepath.Join(tmpDir, "video.mp4")
				_ = os.WriteFile(path, []byte("fake video content"), 0600)
				return path
			},
			wantErr: false,
		},
		{
			name: "valid mov file",
			setup: func() string {
				path := filepath.Join(tmpDir, "video.mov")
				_ = os.WriteFile(path, []byte("fake video content"), 0600)
				return path
			},
			wantErr: false,
		},
		{
			name: "empty video file",
			setup: func() string {
				path := filepath.Join(tmpDir, "empty.mp4")
				_ = os.WriteFile(path, []byte{}, 0600)
				return path
			},
			wantErr: true,
			errMsg:  "video file is empty",
		},
		{
			name: "directory instead of file",
			setup: func() string {
				path := filepath.Join(tmpDir, "video_dir")
				_ = os.Mkdir(path, 0750)
				return path
			},
			wantErr: true,
			errMsg:  "path is a directory",
		},
		{
			name: "unsupported format",
			setup: func() string {
				path := filepath.Join(tmpDir, "video.avi")
				_ = os.WriteFile(path, []byte("content"), 0600)
				return path
			},
			wantErr: true,
			errMsg:  "unsupported video format",
		},
		{
			name: "non-existent file",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.mp4")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			err := ValidateVideoFileForExtension(path)
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

func TestConstants(t *testing.T) {
	assert.Equal(t, 20*1024*1024, MaxImageSize)
	assert.Equal(t, 141, MaxVideoDuration)
}

func TestValidateImageFile_NotFound(t *testing.T) {
	err := ValidateImageFile("/nonexistent/image.jpg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "image file not found")
}

func TestDecodeImageConfig_NonExistent(t *testing.T) {
	_, _, err := DecodeImageConfig("/nonexistent/image.jpg")
	assert.Error(t, err)
}

func TestValidateInterpolationImages_FirstImageError(t *testing.T) {
	err := ValidateInterpolationImages("/nonexistent1.jpg", "/nonexistent2.jpg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode first image")
}
