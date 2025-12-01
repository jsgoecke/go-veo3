package veo3

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jasongoecke/go-veo3/internal/validation"
)

// Validate validates the interpolation request parameters
func (r *InterpolationRequest) Validate() error {
	// For interpolation, prompt is optional, so validate other parameters manually

	// Validate model
	if err := validateModel(r.Model); err != nil {
		return err
	}

	// Validate model supports interpolation
	if err := ValidateModelForInterpolation(r.Model); err != nil {
		return err
	}

	// Interpolation has specific constraints: 8s duration and 16:9 aspect ratio
	if r.DurationSeconds != 8 {
		return fmt.Errorf("interpolation requires 8 seconds duration")
	}

	if r.AspectRatio != "16:9" {
		return fmt.Errorf("interpolation requires 16:9 aspect ratio")
	}

	// Validate resolution
	if r.Resolution != "720p" && r.Resolution != "1080p" {
		return fmt.Errorf("resolution must be 720p or 1080p")
	}

	// Validate person generation setting
	if err := validatePersonGeneration(r.PersonGeneration); err != nil {
		return err
	}

	// Validate prompt (optional for interpolation, but if provided, must be valid)
	if r.Prompt != "" {
		if err := validatePrompt(r.Prompt); err != nil {
			return err
		}
	}

	// Validate frame paths
	if r.FirstFramePath == "" {
		return fmt.Errorf("first frame path cannot be empty")
	}

	if r.LastFramePath == "" {
		return fmt.Errorf("last frame path cannot be empty")
	}

	if r.FirstFramePath == r.LastFramePath {
		return fmt.Errorf("first and last frame cannot be the same file")
	}

	// Validate both images and compatibility
	if err := ValidateCompatibleImages(r.FirstFramePath, r.LastFramePath); err != nil {
		return err
	}

	return nil
}

// ValidateCompatibleImages validates that two images are compatible for interpolation
func ValidateCompatibleImages(firstPath, lastPath string) error {
	return validation.ValidateInterpolationImages(firstPath, lastPath)
}

// ValidateModelForInterpolation checks if a model supports frame interpolation
func ValidateModelForInterpolation(modelID string) error {
	model, exists := GetModel(modelID)
	if !exists {
		return &OperationError{
			Code:    "INVALID_MODEL",
			Message: "Unknown model: " + modelID,
		}
	}

	// For now, assume only Veo 3.1 models support interpolation
	if !strings.HasPrefix(model.Version, "3.1") {
		return &OperationError{
			Code:       "INTERPOLATION_NOT_SUPPORTED",
			Message:    "Model " + modelID + " does not support frame interpolation",
			Suggestion: "Use a Veo 3.1 model for frame interpolation support",
		}
	}

	return nil
}

// EncodeBothImagesToBase64 encodes both interpolation frames to base64
func EncodeBothImagesToBase64(firstPath, lastPath string) (string, string, error) {
	// Encode first frame
	encoded1, err := EncodeImageToBase64(firstPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode first frame: %w", err)
	}

	// Encode last frame
	encoded2, err := EncodeImageToBase64(lastPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode last frame: %w", err)
	}

	return encoded1, encoded2, nil
}

// BuildInterpolationPayload builds the API payload for frame interpolation
func BuildInterpolationPayload(request *InterpolationRequest) (map[string]interface{}, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Validate request first
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Encode both images to base64
	firstEncoded, lastEncoded, err := EncodeBothImagesToBase64(request.FirstFramePath, request.LastFramePath)
	if err != nil {
		return nil, fmt.Errorf("failed to encode images: %w", err)
	}

	// Build the payload structure
	payload := map[string]interface{}{
		"model": request.Model,
		"firstFrame": map[string]interface{}{
			"gcsUri": "", // Will be set after upload to GCS
			"data":   firstEncoded,
		},
		"lastFrame": map[string]interface{}{
			"gcsUri": "", // Will be set after upload to GCS
			"data":   lastEncoded,
		},
		"parameters": map[string]interface{}{
			"resolution":  request.Resolution,
			"duration":    "8s",   // Fixed for interpolation
			"aspectRatio": "16:9", // Fixed for interpolation
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

// InterpolateFrames generates a video by interpolating between two frames
func (c *Client) InterpolateFrames(ctx context.Context, req *InterpolationRequest) (*Operation, error) {
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
	// 1. Upload both images to Google Cloud Storage
	// 2. Make HTTP request to the Veo API interpolation endpoint
	// 3. Return the operation for polling

	// Extract filenames for operation ID
	firstName := filepath.Base(req.FirstFramePath)
	lastName := filepath.Base(req.LastFramePath)
	firstNameWithoutExt := strings.TrimSuffix(firstName, filepath.Ext(firstName))
	lastNameWithoutExt := strings.TrimSuffix(lastName, filepath.Ext(lastName))

	op := &Operation{
		ID:     "operations/interpolate-" + firstNameWithoutExt + "-to-" + lastNameWithoutExt + "-" + generateID(),
		Status: StatusPending,
		Metadata: map[string]interface{}{
			"model":            req.Model,
			"prompt":           req.Prompt,
			"first_frame_path": req.FirstFramePath,
			"last_frame_path":  req.LastFramePath,
			"resolution":       req.Resolution,
			"duration_seconds": req.DurationSeconds,
			"aspect_ratio":     req.AspectRatio,
		},
	}

	return op, nil
}

// GetInterpolationInfo returns information about an interpolation request
func GetInterpolationInfo(firstPath, lastPath string) (*InterpolationInfo, error) {
	// Get info about both images
	firstInfo, err := GetImageInfo(firstPath)
	if err != nil {
		return nil, fmt.Errorf("cannot get first frame info: %w", err)
	}

	lastInfo, err := GetImageInfo(lastPath)
	if err != nil {
		return nil, fmt.Errorf("cannot get last frame info: %w", err)
	}

	// Validate compatibility
	if err := ValidateCompatibleImages(firstPath, lastPath); err != nil {
		return nil, err
	}

	interpolationInfo := &InterpolationInfo{
		FirstFrame: firstInfo,
		LastFrame:  lastInfo,
		Compatible: true,
		TotalSize:  firstInfo.SizeBytes + lastInfo.SizeBytes,
	}

	return interpolationInfo, nil
}

// InterpolationInfo holds metadata about an interpolation request
type InterpolationInfo struct {
	FirstFrame *ImageInfo `json:"first_frame"`
	LastFrame  *ImageInfo `json:"last_frame"`
	Compatible bool       `json:"compatible"`
	TotalSize  int64      `json:"total_size_bytes"`
}
