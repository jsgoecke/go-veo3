package veo3

import (
	"fmt"
	"strings"
	"time"
)

// GenerationRequest represents a video generation job with all necessary parameters
type GenerationRequest struct {
	Prompt           string `json:"prompt" yaml:"prompt"`
	NegativePrompt   string `json:"negative_prompt,omitempty" yaml:"negative_prompt,omitempty"`
	Model            string `json:"model" yaml:"model"`
	AspectRatio      string `json:"aspect_ratio" yaml:"aspect_ratio"`
	Resolution       string `json:"resolution" yaml:"resolution"`
	DurationSeconds  int    `json:"duration_seconds" yaml:"duration_seconds"`
	Seed             *int   `json:"seed,omitempty" yaml:"seed,omitempty"`
	PersonGeneration string `json:"person_generation,omitempty" yaml:"person_generation,omitempty"`
}

// Validate validates the generation request parameters
func (r *GenerationRequest) Validate() error {
	// Validate prompt
	if err := validatePrompt(r.Prompt); err != nil {
		return err
	}

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

	return nil
}

// validatePrompt checks prompt requirements
func validatePrompt(prompt string) error {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	// Approximate token counting (1 token ~= 1 word for simple check)
	words := strings.Fields(prompt)
	if len(words) > 1024 {
		return fmt.Errorf("prompt exceeds 1024 tokens (approximately %d tokens)", len(words))
	}

	return nil
}

// validateModel checks if model is supported
func validateModel(model string) error {
	if model == "" {
		return fmt.Errorf("model cannot be empty")
	}

	_, exists := GetModel(model)
	if !exists {
		return fmt.Errorf("invalid model: %s", model)
	}

	return nil
}

// validateAspectRatio checks aspect ratio format
func validateAspectRatio(aspectRatio string) error {
	if aspectRatio == "" {
		return fmt.Errorf("aspect ratio cannot be empty")
	}

	if aspectRatio != "16:9" && aspectRatio != "9:16" {
		return fmt.Errorf("aspect ratio must be 16:9 or 9:16")
	}

	return nil
}

// validateDuration checks duration values
func validateDuration(duration int) error {
	if duration != 4 && duration != 6 && duration != 8 {
		return fmt.Errorf("duration must be 4, 6, or 8 seconds")
	}

	return nil
}

// validateResolutionAndDuration checks resolution and duration constraints
func validateResolutionAndDuration(resolution string, duration int) error {
	if resolution == "" {
		return fmt.Errorf("resolution cannot be empty")
	}

	if resolution != "720p" && resolution != "1080p" {
		return fmt.Errorf("resolution must be 720p or 1080p")
	}

	// 1080p requires 8 seconds duration
	if resolution == "1080p" && duration != 8 {
		return fmt.Errorf("1080p resolution requires 8 seconds duration")
	}

	return nil
}

// validatePersonGeneration checks person generation setting
func validatePersonGeneration(personGeneration string) error {
	if personGeneration == "" {
		return nil // Optional field
	}

	validValues := []string{"allow_all", "allow_adult", "dont_allow"}
	for _, valid := range validValues {
		if personGeneration == valid {
			return nil
		}
	}

	return fmt.Errorf("person_generation must be one of: allow_all, allow_adult, dont_allow")
}

// ImageRequest extends GenerationRequest for image-to-video generation
type ImageRequest struct {
	GenerationRequest
	ImagePath string `json:"image_path" yaml:"image_path"`
}

// InterpolationRequest extends GenerationRequest for frame interpolation
type InterpolationRequest struct {
	GenerationRequest
	FirstFramePath string `json:"first_frame_path" yaml:"first_frame_path"`
	LastFramePath  string `json:"last_frame_path" yaml:"last_frame_path"`
}

// ReferenceImageRequest extends GenerationRequest with reference images for guided generation
type ReferenceImageRequest struct {
	GenerationRequest
	ReferenceImagePaths []string `json:"reference_image_paths" yaml:"reference_image_paths"`
}

// ExtensionRequest extends GenerationRequest for video extension
type ExtensionRequest struct {
	VideoPath       string `json:"video_path" yaml:"video_path"`
	ExtensionPrompt string `json:"extension_prompt,omitempty" yaml:"extension_prompt,omitempty"`
	Model           string `json:"model" yaml:"model"`
}

// OperationStatus represents the current state of an operation
type OperationStatus string

const (
	StatusPending   OperationStatus = "PENDING"
	StatusRunning   OperationStatus = "RUNNING"
	StatusDone      OperationStatus = "DONE"
	StatusFailed    OperationStatus = "FAILED"
	StatusCancelled OperationStatus = "CANCELLED"
)

// Operation represents a long-running async operation tracked by Google API
type Operation struct {
	ID        string                 `json:"id"`
	Status    OperationStatus        `json:"status"`
	Progress  float64                `json:"progress,omitempty"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	VideoURI  string                 `json:"video_uri,omitempty"`
	Error     *OperationError        `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// OperationError represents error details for failed operations
type OperationError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Suggestion string                 `json:"suggestion,omitempty"`
}

// Error implements the error interface
func (e *OperationError) Error() string {
	if e.Suggestion != "" {
		return e.Message + ". " + e.Suggestion
	}
	return e.Message
}

// GeneratedVideo represents output video file with metadata
type GeneratedVideo struct {
	FilePath              string    `json:"file_path"`
	OperationID           string    `json:"operation_id"`
	Model                 string    `json:"model"`
	Prompt                string    `json:"prompt,omitempty"`
	DurationSeconds       int       `json:"duration_seconds"`
	Resolution            string    `json:"resolution"`
	AspectRatio           string    `json:"aspect_ratio"`
	FileSizeBytes         int64     `json:"file_size_bytes"`
	GenerationTimeSeconds int       `json:"generation_time_seconds"`
	CreatedAt             time.Time `json:"created_at"`
}
