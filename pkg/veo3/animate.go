package veo3

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jasongoecke/go-veo3/internal/validation"
)

// Validate validates the image request parameters
func (r *ImageRequest) Validate() error {
	// For image-to-video, prompt is optional, so validate other parameters manually
	// instead of calling GenerationRequest.Validate()

	// Validate model
	if err := validateModel(r.Model); err != nil {
		return err
	}

	// Validate aspect ratio
	if err := validateAspectRatio(r.AspectRatio); err != nil {
		return err
	}

	// Validate resolution and duration combination
	if err := validateResolutionAndDuration(r.Resolution, r.DurationSeconds); err != nil {
		return err
	}

	// Validate duration
	if err := validateDuration(r.DurationSeconds); err != nil {
		return err
	}

	// Validate person generation setting
	if err := validatePersonGeneration(r.PersonGeneration); err != nil {
		return err
	}

	// Validate prompt (optional for image-to-video, but if provided, must be valid)
	if r.Prompt != "" {
		if err := validatePrompt(r.Prompt); err != nil {
			return err
		}
	}

	// Validate image path
	if r.ImagePath == "" {
		return fmt.Errorf("image path cannot be empty")
	}

	// Validate image file
	if err := validation.ValidateImageFile(r.ImagePath); err != nil {
		return err
	}

	return nil
}

// EncodeImageToBase64 encodes an image file to base64 for API submission
func EncodeImageToBase64(imagePath string) (string, error) {
	if imagePath == "" {
		return "", fmt.Errorf("image path cannot be empty")
	}

	// Check if file exists and is not a directory
	info, err := os.Stat(imagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no such file: %s", imagePath)
		}
		return "", fmt.Errorf("cannot access file: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", imagePath)
	}

	// Read file contents
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %w", err)
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, nil
}

// ValidateImageFormat validates the image file format by extension
func ValidateImageFormat(filename string) error {
	return validation.ValidateImageFormat(filename)
}

// ValidateImageSize validates the image file size
func ValidateImageSize(size int64) error {
	return validation.ValidateImageSize(size)
}

// BuildImageToVideoPayload builds the API payload for image-to-video generation
func BuildImageToVideoPayload(request *ImageRequest) (map[string]interface{}, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Validate request first
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Encode image to base64
	encodedImage, err := EncodeImageToBase64(request.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	// Build the payload structure
	payload := map[string]interface{}{
		"model": request.Model,
		"inputImage": map[string]interface{}{
			"gcsUri": "", // Will be set after upload to GCS
			"data":   encodedImage,
		},
		"parameters": map[string]interface{}{
			"resolution":  request.Resolution,
			"duration":    fmt.Sprintf("%ds", request.DurationSeconds),
			"aspectRatio": request.AspectRatio,
		},
	}

	// Add optional prompt if provided
	if request.Prompt != "" {
		payload["prompt"] = request.Prompt
	}

	// Add optional negative prompt if provided
	if request.NegativePrompt != "" {
		payload["negativePrompt"] = request.NegativePrompt
	}

	// Add optional seed if provided
	if request.Seed != nil {
		payload["seed"] = *request.Seed
	}

	// Add optional person generation setting if provided
	if request.PersonGeneration != "" {
		payload["personGeneration"] = request.PersonGeneration
	}

	return payload, nil
}

// AnimateImage generates a video from an input image
func (c *Client) AnimateImage(ctx context.Context, req *ImageRequest) (*Operation, error) {
	// Validate the request first
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// For mock/test implementation, check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// This is a placeholder implementation
	// The actual implementation would:
	// 1. Upload image to Google Cloud Storage
	// 2. Make HTTP request to the Veo API image-to-video endpoint
	// 3. Return the operation for polling

	// Extract filename for operation ID
	filename := filepath.Base(req.ImagePath)
	filenameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	op := &Operation{
		ID:     "operations/animate-" + filenameWithoutExt + "-" + generateID(),
		Status: StatusPending,
		Metadata: map[string]interface{}{
			"model":            req.Model,
			"prompt":           req.Prompt,
			"image_path":       req.ImagePath,
			"resolution":       req.Resolution,
			"duration_seconds": req.DurationSeconds,
			"aspect_ratio":     req.AspectRatio,
		},
	}

	return op, nil
}

// GetImageInfo returns information about an image file
func GetImageInfo(imagePath string) (*ImageInfo, error) {
	// Validate file exists
	info, err := os.Stat(imagePath)
	if err != nil {
		return nil, fmt.Errorf("cannot access image file: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", imagePath)
	}

	// Get image dimensions
	config, format, err := validation.DecodeImageConfig(imagePath)
	if err != nil {
		return nil, fmt.Errorf("cannot decode image: %w", err)
	}

	imageInfo := &ImageInfo{
		FilePath:    imagePath,
		Format:      format,
		Width:       config.Width,
		Height:      config.Height,
		SizeBytes:   info.Size(),
		AspectRatio: calculateAspectRatio(config.Width, config.Height),
	}

	return imageInfo, nil
}

// ImageInfo holds metadata about an image file
type ImageInfo struct {
	FilePath    string `json:"file_path"`
	Format      string `json:"format"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	SizeBytes   int64  `json:"size_bytes"`
	AspectRatio string `json:"aspect_ratio"`
}

// calculateAspectRatio calculates the closest standard aspect ratio
func calculateAspectRatio(width, height int) string {
	ratio := float64(width) / float64(height)

	// Check against standard ratios with some tolerance
	ratio16_9 := 16.0 / 9.0 // ~1.778
	ratio9_16 := 9.0 / 16.0 // ~0.563

	if abs(ratio-ratio16_9) < 0.1 {
		return "16:9"
	} else if abs(ratio-ratio9_16) < 0.1 {
		return "9:16"
	}

	// Return calculated ratio
	return fmt.Sprintf("%.2f:1", ratio)
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
