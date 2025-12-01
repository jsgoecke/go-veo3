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

// Validate validates the extension request parameters
func (r *ExtensionRequest) Validate() error {
	// Validate video path
	if r.VideoPath == "" {
		return fmt.Errorf("video path cannot be empty")
	}

	// Validate model
	if r.Model == "" {
		return fmt.Errorf("model cannot be empty")
	}

	if err := validateModel(r.Model); err != nil {
		return err
	}

	// Validate model supports extension
	if err := ValidateModelForExtension(r.Model); err != nil {
		return err
	}

	// Validate extension prompt (optional, but if provided, must be valid)
	if r.ExtensionPrompt != "" {
		if err := validatePrompt(r.ExtensionPrompt); err != nil {
			return err
		}
	}

	// Validate video file
	if err := validation.ValidateVideoFileForExtension(r.VideoPath); err != nil {
		return err
	}

	return nil
}

// ValidateVideoForExtension validates a video file for extension with duration check
func ValidateVideoForExtension(videoPath string, durationSeconds int) error {
	// Basic file validation
	if err := validation.ValidateVideoFileForExtension(videoPath); err != nil {
		return err
	}

	// Validate duration constraints
	if durationSeconds <= 0 {
		return fmt.Errorf("invalid video duration: %d seconds", durationSeconds)
	}

	if durationSeconds > 141 { // Max video length per API docs
		return fmt.Errorf("video exceeds maximum duration: %d seconds (maximum: 141 seconds)", durationSeconds)
	}

	return nil
}

// EncodeVideoToBase64 encodes a video file to base64 for API submission
func EncodeVideoToBase64(videoPath string) (string, error) {
	if videoPath == "" {
		return "", fmt.Errorf("video path cannot be empty")
	}

	// Check if file exists and is not a directory
	info, err := os.Stat(videoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no such file: %s", videoPath)
		}
		return "", fmt.Errorf("cannot access file: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", videoPath)
	}

	if info.Size() == 0 {
		return "", fmt.Errorf("file is empty: %s", videoPath)
	}

	// Read file contents
	data, err := os.ReadFile(videoPath)
	if err != nil {
		return "", fmt.Errorf("failed to read video file: %w", err)
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, nil
}

// BuildExtensionPayload builds the API payload for video extension
func BuildExtensionPayload(request *ExtensionRequest) (map[string]interface{}, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Validate request first
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Encode video to base64
	encodedVideo, err := EncodeVideoToBase64(request.VideoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to encode video: %w", err)
	}

	// Build the payload structure
	payload := map[string]interface{}{
		"model": request.Model,
		"inputVideo": map[string]interface{}{
			"gcsUri": "", // Will be set after upload to GCS
			"data":   encodedVideo,
		},
		"parameters": map[string]interface{}{
			"maxExtensionSeconds": 7, // Maximum extension length
		},
	}

	// Add extension prompt if provided
	if request.ExtensionPrompt != "" {
		payload["prompt"] = request.ExtensionPrompt
	}

	return payload, nil
}

// ExtendVideo extends an existing Veo-generated video
func (c *Client) ExtendVideo(ctx context.Context, req *ExtensionRequest) (*Operation, error) {
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
	// 1. Upload video to Google Cloud Storage
	// 2. Make HTTP request to the Veo API video extension endpoint
	// 3. Return the operation for polling

	// Extract filename for operation ID
	filename := filepath.Base(req.VideoPath)
	filenameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	op := &Operation{
		ID:     "operations/extend-" + filenameWithoutExt + "-" + generateID(),
		Status: StatusPending,
		Metadata: map[string]interface{}{
			"model":            req.Model,
			"extension_prompt": req.ExtensionPrompt,
			"video_path":       req.VideoPath,
			"operation_type":   "extension",
		},
	}

	return op, nil
}

// GetVideoInfo returns information about a video file
func GetVideoInfo(videoPath string) (*VideoInfo, error) {
	// Validate file exists
	info, err := os.Stat(videoPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access video file: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", videoPath)
	}

	// Basic validation
	if err := validation.ValidateVideoFileForExtension(videoPath); err != nil {
		return nil, err
	}

	videoInfo := &VideoInfo{
		FilePath:  videoPath,
		Format:    "mp4", // Based on extension validation
		SizeBytes: info.Size(),
	}

	// Note: In a full implementation, we'd use ffmpeg or similar to get:
	// - Actual video duration
	// - Resolution and frame rate
	// - Codec information
	// - Whether it's Veo-generated (metadata check)

	return videoInfo, nil
}

// VideoInfo holds metadata about a video file
type VideoInfo struct {
	FilePath        string `json:"file_path"`
	Format          string `json:"format"`
	SizeBytes       int64  `json:"size_bytes"`
	DurationSeconds int    `json:"duration_seconds,omitempty"`
	Resolution      string `json:"resolution,omitempty"`
	FrameRate       int    `json:"frame_rate,omitempty"`
	VeoGenerated    bool   `json:"veo_generated,omitempty"`
}
