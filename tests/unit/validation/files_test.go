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
	// Use testdata directory with actual valid image files
	testDataDir := "testdata"

	// Create test files if they don't exist
	jpegFile := filepath.Join(testDataDir, "test.jpg")
	pngFile := filepath.Join(testDataDir, "test.png")
	webpFile := filepath.Join(testDataDir, "test.webp")
	gifFile := filepath.Join(testDataDir, "test.gif")
	emptyFile := filepath.Join(testDataDir, "empty.jpg")
	largeFile := filepath.Join(testDataDir, "21mb.jpg")

	// Ensure empty file exists
	if _, err := os.Stat(emptyFile); os.IsNotExist(err) {
		err = os.WriteFile(emptyFile, []byte{}, 0600)
		require.NoError(t, err)
	}

	// Ensure test.gif exists (for unsupported format test)
	// The gif should already exist in testdata from fixtures

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
			errMsg:   "invalid image format",
		},
		{
			name:     "empty file",
			filePath: emptyFile,
			wantErr:  true,
			errMsg:   "image file is empty",
		},
		{
			name:     "file too large",
			filePath: largeFile,
			wantErr:  true,
			errMsg:   "image file too large",
		},
		{
			name:     "non-existent file",
			filePath: filepath.Join(testDataDir, "nonexistent.jpg"),
			wantErr:  true,
			errMsg:   "image file not found",
		},
		{
			name:     "directory path",
			filePath: testDataDir,
			wantErr:  true,
			errMsg:   "failed to read image file",
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

// Additional tests for interpolation-specific validation
func TestValidateInterpolationImages(t *testing.T) {
	// Use testdata directory with actual valid image files
	testDataDir := "testdata"

	// Use real test image files
	image1 := filepath.Join(testDataDir, "test.jpg")
	image2 := filepath.Join(testDataDir, "test2.jpg")

	tests := []struct {
		name      string
		firstPath string
		lastPath  string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "both images exist and have compatible formats",
			firstPath: image1,
			lastPath:  image2,
			wantErr:   false,
		},
		{
			name:      "first image doesn't exist",
			firstPath: filepath.Join(testDataDir, "nonexistent1.jpg"),
			lastPath:  image1,
			wantErr:   true,
			errMsg:    "failed to decode first image",
		},
		{
			name:      "second image doesn't exist",
			firstPath: image1,
			lastPath:  filepath.Join(testDataDir, "nonexistent2.jpg"),
			wantErr:   true,
			errMsg:    "failed to decode second image",
		},
		{
			name:      "identical file paths",
			firstPath: image1,
			lastPath:  image1,
			wantErr:   false, // ValidateInterpolationImages doesn't check for same file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateInterpolationImages(tt.firstPath, tt.lastPath)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateVideoFileForExtension(t *testing.T) {
	// Tests for video extension functionality (User Story 5)
	tempDir := t.TempDir()

	// Create mock video files
	validVideo := filepath.Join(tempDir, "valid.mp4")
	invalidVideo := filepath.Join(tempDir, "invalid.avi")
	emptyVideo := filepath.Join(tempDir, "empty.mp4")
	largeVideo := filepath.Join(tempDir, "large.mp4")

	// Create test files
	err := os.WriteFile(validVideo, []byte("fake-mp4-content"), 0600)
	require.NoError(t, err)

	err = os.WriteFile(invalidVideo, []byte("fake-avi-content"), 0600)
	require.NoError(t, err)

	err = os.WriteFile(emptyVideo, []byte{}, 0600)
	require.NoError(t, err)

	// Create large video file (simulate video over max length)
	largeContent := make([]byte, 1024*1024) // 1MB as placeholder
	err = os.WriteFile(largeVideo, largeContent, 0600)
	require.NoError(t, err)

	tests := []struct {
		name     string
		filePath string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid MP4 file",
			filePath: validVideo,
			wantErr:  false,
		},
		{
			name:     "invalid AVI format",
			filePath: invalidVideo,
			wantErr:  true,
			errMsg:   "unsupported video format",
		},
		{
			name:     "empty video file",
			filePath: emptyVideo,
			wantErr:  true,
			errMsg:   "video file is empty",
		},
		{
			name:     "non-existent video",
			filePath: filepath.Join(tempDir, "nonexistent.mp4"),
			wantErr:  true,
			errMsg:   "no such file or directory",
		},
		{
			name:     "directory instead of file",
			filePath: tempDir,
			wantErr:  true,
			errMsg:   "is a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateVideoFileForExtension(tt.filePath)

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
			errMsg:  "image file too large",
		},
		{
			name:    "21MB file invalid",
			size:    21 * 1024 * 1024,
			wantErr: true,
			errMsg:  "image file too large",
		},
		{
			name:    "100MB file invalid",
			size:    100 * 1024 * 1024,
			wantErr: true,
			errMsg:  "image file too large",
		},
		{
			name:    "0 byte file invalid",
			size:    0,
			wantErr: true,
			errMsg:  "image file is empty",
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
			errMsg:  "image dimensions mismatch",
		},
		{
			name:   "different height",
			width1: 1920, height1: 1080,
			width2: 1920, height2: 720,
			wantErr: true,
			errMsg:  "image dimensions mismatch",
		},
		{
			name:   "completely different",
			width1: 1920, height1: 1080,
			width2: 640, height2: 480,
			wantErr: true,
			errMsg:  "image dimensions mismatch",
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
