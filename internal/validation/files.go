package validation

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/webp"
)

const (
	MaxImageSize     = 20 * 1024 * 1024 // 20MB
	MaxVideoDuration = 141              // seconds (approx)
)

// ValidateImageFile checks if the file exists, is readable, is within size limits, and is a valid image
func ValidateImageFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("image file not found: %s", path)
		}
		return fmt.Errorf("failed to check image file: %w", err)
	}

	if err := ValidateImageSize(info.Size()); err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read image file: %w", err)
	}

	// Validate image format by decoding config
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("invalid image format: %w (supported: jpeg, png, webp)", err)
	}

	if !isValidFormat(format) {
		return fmt.Errorf("unsupported image format: %s (supported: jpeg, png, webp)", format)
	}

	return nil
}

func isValidFormat(format string) bool {
	switch strings.ToLower(format) {
	case "jpeg", "jpg", "png", "webp":
		return true
	default:
		return false
	}
}

// ValidateImageFormat checks if the file has a supported extension (simple check)
func ValidateImageFormat(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return nil
	default:
		return fmt.Errorf("unsupported image format: %s", ext)
	}
}

// ValidateImageSize checks if file size is within limits
func ValidateImageSize(size int64) error {
	if size < 0 {
		return fmt.Errorf("invalid file size: %d bytes", size)
	}
	if size == 0 {
		return fmt.Errorf("image file is empty")
	}
	if size > MaxImageSize {
		return fmt.Errorf("image file too large: %d bytes (max %d bytes)", size, MaxImageSize)
	}
	return nil
}

// DecodeImageConfig decodes image configuration
func DecodeImageConfig(path string) (image.Config, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return image.Config{}, "", err
	}
	defer f.Close()
	return image.DecodeConfig(f)
}

// ValidateInterpolationImages checks two images for compatibility
func ValidateInterpolationImages(path1, path2 string) error {
	cfg1, _, err := DecodeImageConfig(path1)
	if err != nil {
		return fmt.Errorf("failed to decode first image: %w", err)
	}
	cfg2, _, err := DecodeImageConfig(path2)
	if err != nil {
		return fmt.Errorf("failed to decode second image: %w", err)
	}

	return ValidateCompatibleDimensions(cfg1.Width, cfg1.Height, cfg2.Width, cfg2.Height)
}

// ValidateCompatibleDimensions checks if dimensions match
func ValidateCompatibleDimensions(w1, h1, w2, h2 int) error {
	if w1 <= 0 || h1 <= 0 || w2 <= 0 || h2 <= 0 {
		return fmt.Errorf("invalid dimensions: %dx%d vs %dx%d", w1, h1, w2, h2)
	}
	if w1 != w2 || h1 != h2 {
		return fmt.Errorf("image dimensions mismatch: %dx%d vs %dx%d", w1, h1, w2, h2)
	}
	return nil
}

// ValidateVideoFileForExtension checks video file validity
func ValidateVideoFileForExtension(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	// Basic check
	if info.IsDir() {
		return fmt.Errorf("path is a directory")
	}
	// Check for empty file
	if info.Size() == 0 {
		return fmt.Errorf("video file is empty")
	}
	// Check extension
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".mp4" && ext != ".mov" {
		return fmt.Errorf("unsupported video format: %s", ext)
	}
	return nil
}
