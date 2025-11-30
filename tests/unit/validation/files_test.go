package validation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jasongoecke/go-veo3/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateImageFile(t *testing.T) {
	// Create temporary test files
	tempDir := t.TempDir()

	// Create test files
	jpegFile := filepath.Join(tempDir, "test.jpg")
	pngFile := filepath.Join(tempDir, "test.png")
	webpFile := filepath.Join(tempDir, "test.webp")
	gifFile := filepath.Join(tempDir, "test.gif")
	emptyFile := filepath.Join(tempDir, "empty.jpg")
	largeFile := filepath.Join(tempDir, "large.jpg")

	// Create files with different sizes
	err := os.WriteFile(jpegFile, []byte("fake-jpeg-content"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(pngFile, []byte("fake-png-content"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(webpFile, []byte("fake-webp-content"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(gifFile, []byte("fake-gif-content"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(emptyFile, []byte{}, 0644)
	require.NoError(t, err)

	// Create a large file (simulate 21MB)
	largeContent := make([]byte, 21*1024*1024)
	err = os.WriteFile(largeFile, largeContent, 0644)
	require.NoError(t, err)

	tests := []struct {
		name     string
		filePath string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid JPEG file",
			filePath: jpegFile,
			wantErr:  false,
		},
		{
			name:     "valid PNG file",
			filePath: pngFile,
			wantErr:  false,
		},
		{
			name:     "valid WebP file",
			filePath: webpFile,
			wantErr:  false,
		},
		{
			name:     "unsupported GIF file",
			filePath: gifFile,
			wantErr:  true,
			errMsg:   "unsupported image format",
		},
		{
			name:     "empty file",
			filePath: emptyFile,
			wantErr:  true,
			errMsg:   "file is empty",
		},
		{
			name:     "file too large",
			filePath: largeFile,
			wantErr:  true,
			errMsg:   "exceeds maximum size",
		},
		{
			name:     "non-existent file",
			filePath: filepath.Join(tempDir, "nonexistent.jpg"),
			wantErr:  true,
			errMsg:   "no such file",
		},
		{
			name:     "directory path",
			filePath: tempDir,
			wantErr:  true,
			errMsg:   "is a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateImageFile(tt.filePath)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateImageFormat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "JPEG extension (.jpg)",
			filename: "test.jpg",
			wantErr:  false,
		},
		{
			name:     "JPEG extension (.jpeg)",
			filename: "test.jpeg",
			wantErr:  false,
		},
		{
			name:     "JPEG extension (uppercase)",
			filename: "test.JPG",
			wantErr:  false,
		},
		{
			name:     "PNG extension",
			filename: "test.png",
			wantErr:  false,
		},
		{
			name:     "PNG extension (uppercase)",
			filename: "test.PNG",
			wantErr:  false,
		},
		{
			name:     "WebP extension",
			filename: "test.webp",
			wantErr:  false,
		},
		{
			name:     "WebP extension (uppercase)",
			filename: "test.WEBP",
			wantErr:  false,
		},
		{
			name:     "GIF not supported",
			filename: "test.gif",
			wantErr:  true,
			errMsg:   "unsupported image format",
		},
		{
			name:     "BMP not supported",
			filename: "test.bmp",
			wantErr:  true,
			errMsg:   "unsupported image format",
		},
		{
			name:     "TIFF not supported",
			filename: "test.tiff",
			wantErr:  true,
			errMsg:   "unsupported image format",
		},
		{
			name:     "SVG not supported",
			filename: "test.svg",
			wantErr:  true,
			errMsg:   "unsupported image format",
		},
		{
			name:     "no extension",
			filename: "test",
			wantErr:  true,
			errMsg:   "unsupported image format",
		},
		{
			name:     "multiple extensions",
			filename: "test.jpg.txt",
			wantErr:  true,
			errMsg:   "unsupported image format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateImageFormat(tt.filename)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateImageSize(t *testing.T) {
	tests := []struct {
		name    string
		size    int64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "1KB file valid",
			size:    1024,
			wantErr: false,
		},
		{
			name:    "1MB file valid",
			size:    1 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "10MB file valid",
			size:    10 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "20MB file valid (at limit)",
			size:    20 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "20MB + 1 byte invalid",
			size:    20*1024*1024 + 1,
			wantErr: true,
			errMsg:  "exceeds maximum size",
		},
		{
			name:    "21MB file invalid",
			size:    21 * 1024 * 1024,
			wantErr: true,
			errMsg:  "exceeds maximum size",
		},
		{
			name:    "100MB file invalid",
			size:    100 * 1024 * 1024,
			wantErr: true,
			errMsg:  "exceeds maximum size",
		},
		{
			name:    "0 byte file invalid",
			size:    0,
			wantErr: true,
			errMsg:  "file is empty",
		},
		{
			name:    "negative size invalid",
			size:    -1,
			wantErr: true,
			errMsg:  "invalid file size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateImageSize(tt.size)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateCompatibleDimensions(t *testing.T) {
	// Test for image compatibility checks (for interpolation)
	tests := []struct {
		name    string
		width1  int
		height1 int
		width2  int
		height2 int
		wantErr bool
		errMsg  string
	}{
		{
			name:   "identical dimensions",
			width1: 1920, height1: 1080,
			width2: 1920, height2: 1080,
			wantErr: false,
		},
		{
			name:   "different width",
			width1: 1920, height1: 1080,
			width2: 1280, height2: 1080,
			wantErr: true,
			errMsg:  "dimensions must match",
		},
		{
			name:   "different height",
			width1: 1920, height1: 1080,
			width2: 1920, height2: 720,
			wantErr: true,
			errMsg:  "dimensions must match",
		},
		{
			name:   "completely different",
			width1: 1920, height1: 1080,
			width2: 640, height2: 480,
			wantErr: true,
			errMsg:  "dimensions must match",
		},
		{
			name:   "zero dimensions invalid",
			width1: 0, height1: 0,
			width2: 1920, height2: 1080,
			wantErr: true,
			errMsg:  "invalid dimensions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateCompatibleDimensions(
				tt.width1, tt.height1,
				tt.width2, tt.height2,
			)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
