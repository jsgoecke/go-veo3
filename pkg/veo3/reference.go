package veo3

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jasongoecke/go-veo3/internal/validation"
)

// Validate validates the reference image request parameters
func (r *ReferenceImageRequest) Validate() error {
	// Validate the base generation request (prompt is required for reference images)
	if err := r.GenerationRequest.Validate(); err != nil {
		return err
	}

	// Reference images have specific constraints
	if r.DurationSeconds != 8 {
		return fmt.Errorf("reference images require 8 seconds duration")
	}

	if r.AspectRatio != "16:9" {
		return fmt.Errorf("reference images require 16:9 aspect ratio")
	}

	// Validate model supports reference images
	if err := ValidateModelForReferenceImages(r.Model, len(r.ReferenceImagePaths)); err != nil {
		return err
	}

	// Validate reference image count
	if err := ValidateReferenceImageCount(len(r.ReferenceImagePaths)); err != nil {
		return err
	}

	// Validate each reference image path
	for i, imagePath := range r.ReferenceImagePaths {
		if imagePath == "" {
			return fmt.Errorf("reference image path cannot be empty (index %d)", i)
		}

		// Validate image file
		if err := validation.ValidateImageFile(imagePath); err != nil {
			return fmt.Errorf("reference image %d validation failed: %w", i+1, err)
		}
	}

	return nil
}

// ValidateReferenceImageCount validates the number of reference images
func ValidateReferenceImageCount(count int) error {
	if count < 1 {
		return fmt.Errorf("at least 1 reference image required")
	}

	if count > 3 {
		return fmt.Errorf("maximum 3 reference images allowed")
	}

	return nil
}

// EncodeReferenceImagesToBase64 encodes multiple reference images to base64
func EncodeReferenceImagesToBase64(imagePaths []string) ([]string, error) {
	if len(imagePaths) == 0 {
		return nil, fmt.Errorf("at least 1 reference image required")
	}

	if len(imagePaths) > 3 {
		return nil, fmt.Errorf("maximum 3 reference images allowed")
	}

	encodedImages := make([]string, len(imagePaths))

	for i, imagePath := range imagePaths {
		if imagePath == "" {
			return nil, fmt.Errorf("reference image path cannot be empty (index %d)", i)
		}

		encoded, err := EncodeImageToBase64(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to encode reference image %d: %w", i+1, err)
		}

		encodedImages[i] = encoded
	}

	return encodedImages, nil
}

// BuildReferenceImagePayload builds the API payload for reference image-guided generation
func BuildReferenceImagePayload(request *ReferenceImageRequest) (map[string]interface{}, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Validate request first
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Encode all reference images to base64
	encodedImages, err := EncodeReferenceImagesToBase64(request.ReferenceImagePaths)
	if err != nil {
		return nil, fmt.Errorf("failed to encode reference images: %w", err)
	}

	// Build reference images array for API
	referenceImages := make([]interface{}, len(encodedImages))
	for i, encoded := range encodedImages {
		referenceImages[i] = map[string]interface{}{
			"gcsUri": "", // Will be set after upload to GCS
			"data":   encoded,
		}
	}

	// Build the payload structure
	payload := map[string]interface{}{
		"model":           request.Model,
		"prompt":          request.Prompt,
		"referenceImages": referenceImages,
		"parameters": map[string]interface{}{
			"resolution":  request.Resolution,
			"duration":    "8s",   // Fixed for reference images
			"aspectRatio": "16:9", // Fixed for reference images
		},
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

// GenerateWithReferenceImages generates a video using reference images for style guidance
func (c *Client) GenerateWithReferenceImages(ctx context.Context, req *ReferenceImageRequest) (*Operation, error) {
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
	// 1. Upload all reference images to Google Cloud Storage
	// 2. Make HTTP request to the Veo API reference image endpoint
	// 3. Return the operation for polling

	// Create operation ID based on reference images
	refImageNames := make([]string, len(req.ReferenceImagePaths))
	for i, path := range req.ReferenceImagePaths {
		filename := filepath.Base(path)
		refImageNames[i] = strings.TrimSuffix(filename, filepath.Ext(filename))
	}
	refNamesStr := strings.Join(refImageNames, "-")

	op := &Operation{
		ID:     "operations/reference-" + refNamesStr + "-" + generateID(),
		Status: StatusPending,
		Metadata: map[string]interface{}{
			"model":                 req.Model,
			"prompt":                req.Prompt,
			"reference_image_paths": req.ReferenceImagePaths,
			"reference_image_count": len(req.ReferenceImagePaths),
			"resolution":            req.Resolution,
			"duration_seconds":      req.DurationSeconds,
			"aspect_ratio":          req.AspectRatio,
		},
	}

	return op, nil
}

// GetReferenceImageInfo returns information about reference images
func GetReferenceImageInfo(imagePaths []string) (*ReferenceImageInfo, error) {
	if len(imagePaths) == 0 {
		return nil, fmt.Errorf("at least 1 reference image required")
	}

	if len(imagePaths) > 3 {
		return nil, fmt.Errorf("maximum 3 reference images allowed")
	}

	imageInfos := make([]*ImageInfo, len(imagePaths))
	totalSize := int64(0)

	for i, imagePath := range imagePaths {
		if imagePath == "" {
			return nil, fmt.Errorf("reference image path cannot be empty (index %d)", i)
		}

		info, err := GetImageInfo(imagePath)
		if err != nil {
			return nil, fmt.Errorf("cannot get info for reference image %d: %w", i+1, err)
		}

		imageInfos[i] = info
		totalSize += info.SizeBytes
	}

	referenceInfo := &ReferenceImageInfo{
		Images:     imageInfos,
		Count:      len(imagePaths),
		TotalSize:  totalSize,
		Compatible: true, // Assume compatible after validation
	}

	return referenceInfo, nil
}

// ReferenceImageInfo holds metadata about reference images
type ReferenceImageInfo struct {
	Images     []*ImageInfo `json:"images"`
	Count      int          `json:"count"`
	TotalSize  int64        `json:"total_size_bytes"`
	Compatible bool         `json:"compatible"`
}
