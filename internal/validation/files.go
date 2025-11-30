package validation

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jasongoecke/go-veo3/pkg/config"
	"golang.org/x/image/webp"
)

// ValidateImageFile validates an image file for Veo compatibility
func ValidateImageFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("file not found: %s", path)
	}

	if info.Size() > config.MaxImageSize {
		return fmt.Errorf("image size %d exceeds limit of %d bytes", info.Size(), config.MaxImageSize)
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read magic bytes
	header := make([]byte, 512)
	if _, err := file.Read(header); err != nil {
		return fmt.Errorf("failed to read file header: %w", err)
	}

	// Reset file pointer
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		if _, err := png.DecodeConfig(file); err != nil {
			return fmt.Errorf("invalid PNG file: %w", err)
		}
	case ".jpg", ".jpeg":
		if _, err := jpeg.DecodeConfig(file); err != nil {
			return fmt.Errorf("invalid JPEG file: %w", err)
		}
	case ".webp":
		if _, err := webp.DecodeConfig(file); err != nil {
			return fmt.Errorf("invalid WebP file: %w", err)
		}
	default:
		return fmt.Errorf("unsupported image format: %s (must be .png, .jpg, .jpeg, or .webp)", ext)
	}

	return nil
}

// ValidateVideoFile validates a video file for Veo extension compatibility
func ValidateVideoFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("file not found: %s", path)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory: %s", path)
	}

	// Basic extension check
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".mp4" {
		return fmt.Errorf("unsupported video format: %s (must be .mp4)", ext)
	}

	// Note: Detailed video validation (metadata, duration) would typically require ffmpeg
	// or specific libraries. For now, we rely on API validation for deep checks.
	// We could add basic header checks here if needed.

	return nil
}

// ValidateImageDimensions checks if dimensions are compatible
func ValidateImageDimensions(path1, path2 string) error {
	img1, _, err := DecodeImageConfig(path1)
	if err != nil {
		return err
	}

	img2, _, err := DecodeImageConfig(path2)
	if err != nil {
		return err
	}

	if img1.Width != img2.Width || img1.Height != img2.Height {
		return fmt.Errorf("image dimensions mismatch: %dx%d vs %dx%d", img1.Width, img1.Height, img2.Width, img2.Height)
	}

	return nil
}

func DecodeImageConfig(path string) (image.Config, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return image.Config{}, "", err
	}
	defer file.Close()

	// Handle WebP specifically since image.DecodeConfig might not auto-detect it without import
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".webp" {
		config, err := webp.DecodeConfig(file)
		return config, ext, err
	}

	return image.DecodeConfig(file)
}

// ReadFileToBytes reads a file into a byte slice
func ReadFileToBytes(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ValidateImageFormat validates the image file format by extension
func ValidateImageFormat(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return nil
	default:
		return fmt.Errorf("unsupported image format: %s", ext)
	}
}

// ValidateImageSize validates the image file size
func ValidateImageSize(size int64) error {
	if size < 0 {
		return fmt.Errorf("invalid file size: %d", size)
	}

	if size == 0 {
		return fmt.Errorf("file is empty")
	}

	if size > config.MaxImageSize {
		sizeMB := float64(size) / (1024 * 1024)
		maxMB := float64(config.MaxImageSize) / (1024 * 1024)
		return fmt.Errorf("file exceeds maximum size: %.1f MB (maximum: %.1f MB)", sizeMB, maxMB)
	}

	return nil
}

// ValidateInterpolationImages validates that two images are compatible for interpolation
func ValidateInterpolationImages(firstPath, lastPath string) error {
	// Check that paths are not empty
	if firstPath == "" {
		return fmt.Errorf("first frame path cannot be empty")
	}
	if lastPath == "" {
		return fmt.Errorf("last frame path cannot be empty")
	}

	// Check that they're not the same file
	if firstPath == lastPath {
		return fmt.Errorf("first and last frame cannot be the same file")
	}

	// Validate both image files exist and are valid
	if err := ValidateImageFile(firstPath); err != nil {
		return fmt.Errorf("first frame validation failed: %w", err)
	}

	if err := ValidateImageFile(lastPath); err != nil {
		return fmt.Errorf("last frame validation failed: %w", err)
	}

	// Check dimensions compatibility
	config1, _, err := DecodeImageConfig(firstPath)
	if err != nil {
		return fmt.Errorf("cannot decode first frame: %w", err)
	}

	config2, _, err := DecodeImageConfig(lastPath)
	if err != nil {
		return fmt.Errorf("cannot decode last frame: %w", err)
	}

	if err := ValidateCompatibleDimensions(config1.Width, config1.Height, config2.Width, config2.Height); err != nil {
		return fmt.Errorf("frame dimension mismatch: %w", err)
	}

	return nil
}

// ValidateVideoFileForExtension validates a video file for extension functionality
func ValidateVideoFileForExtension(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("video path cannot be empty")
	}

	// Check if file exists and get info
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no such file: %s", filePath)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Check if it's a directory
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	// Check file size (empty file)
	if info.Size() == 0 {
		return fmt.Errorf("file is empty")
	}

	// Validate format (must be MP4)
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".mp4" {
		return fmt.Errorf("unsupported video format: %s (must be .mp4)", ext)
	}

	// Note: In a full implementation, we'd validate:
	// - Video is Veo-generated (metadata check)
	// - Video duration ≤ 141 seconds
	// - Video resolution and format compatibility
	// For now, basic file validation is sufficient

	return nil
}

// ValidateCompatibleDimensions validates that two images have compatible dimensions
func ValidateCompatibleDimensions(width1, height1, width2, height2 int) error {
	// Check for invalid dimensions
	if width1 <= 0 || height1 <= 0 || width2 <= 0 || height2 <= 0 {
		return fmt.Errorf("invalid dimensions: (%dx%d) and (%dx%d)", width1, height1, width2, height2)
	}

	// Check if dimensions match
	if width1 != width2 || height1 != height2 {
		return fmt.Errorf("dimensions must match: first image (%dx%d) != second image (%dx%d)",
			width1, height1, width2, height2)
	}

	return nil
}
