package validation

import (
	"fmt"
)

// ValidatePrompt checks if the prompt is valid
func ValidatePrompt(prompt string) error {
	if len(prompt) == 0 {
		return fmt.Errorf("prompt cannot be empty")
	}
	if len(prompt) > 1024 { // Assuming 1024 chars approx to tokens for now, though spec says tokens
		return fmt.Errorf("prompt too long: %d characters (max 1024)", len(prompt))
	}
	return nil
}

// ValidateResolution checks if the resolution is supported and compatible with duration
func ValidateResolution(resolution string, duration ...int) error {
	if resolution == "" {
		return fmt.Errorf("resolution cannot be empty")
	}

	switch resolution {
	case "720p":
		return nil
	case "1080p":
		// 1080p requires 8 seconds duration
		if len(duration) > 0 && duration[0] != 8 {
			return fmt.Errorf("1080p resolution requires 8 seconds duration")
		}
		return nil
	default:
		return fmt.Errorf("resolution must be 720p or 1080p")
	}
}

// ValidateAspectRatio checks if the aspect ratio is supported
func ValidateAspectRatio(ratio string) error {
	if ratio == "" {
		return fmt.Errorf("aspect ratio cannot be empty")
	}

	switch ratio {
	case "16:9", "9:16":
		return nil
	default:
		return fmt.Errorf("aspect ratio must be 16:9 or 9:16")
	}
}

// ValidateDuration checks if the duration is supported
func ValidateDuration(seconds int) error {
	switch seconds {
	case 4, 6, 8:
		return nil
	default:
		return fmt.Errorf("duration must be 4, 6, or 8 seconds")
	}
}

// ValidateModel checks if the model ID is valid
func ValidateModel(model string) error {
	if model == "" {
		return fmt.Errorf("model cannot be empty")
	}

	// Check against known model prefixes
	validPrefixes := []string{"veo-3.1", "veo-3.0", "veo-2.0"}
	for _, prefix := range validPrefixes {
		if len(model) >= len(prefix) && model[:len(prefix)] == prefix {
			return nil
		}
	}

	return fmt.Errorf("unsupported model: %s", model)
}

// ValidateModelConstraints checks if parameters are compatible with the model
func ValidateModelConstraints(model string, params map[string]interface{}) error {
	// First validate the model itself
	if err := ValidateModel(model); err != nil {
		return err
	}

	// Check reference image constraints
	if refImages, ok := params["reference_images"].([]string); ok && len(refImages) > 0 {
		// Reference images have specific requirements
		if duration, ok := params["duration"].(int); ok && duration != 8 {
			return fmt.Errorf("reference images require 8 seconds duration")
		}

		if aspectRatio, ok := params["aspect_ratio"].(string); ok && aspectRatio != "16:9" {
			return fmt.Errorf("reference images require 16:9 aspect ratio")
		}
	}

	return nil
}
