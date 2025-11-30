package validation

import (
	"fmt"
	"strings"

	"github.com/jasongoecke/go-veo3/pkg/config"
)

// ValidatePrompt validates the generation prompt
func ValidatePrompt(prompt string) error {
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	// Approximate token count check (1 token ~= 1 word for simplicity)
	words := strings.Fields(prompt)
	if len(words) > config.MaxPromptLength {
		return fmt.Errorf("prompt exceeds 1024 tokens (approximately %d tokens)", len(words))
	}

	return nil
}

// ValidateResolution validates the output resolution with duration constraints
func ValidateResolution(resolution string, duration int) error {
	if resolution == "" {
		return fmt.Errorf("resolution cannot be empty")
	}

	switch resolution {
	case "720p", "1080p":
		// 1080p requires 8 seconds duration
		if resolution == "1080p" && duration != 8 {
			return fmt.Errorf("1080p resolution requires 8 seconds duration")
		}
		return nil
	default:
		return fmt.Errorf("resolution must be 720p or 1080p")
	}
}

// ValidateAspectRatio validates the aspect ratio
func ValidateAspectRatio(aspectRatio string) error {
	if aspectRatio == "" {
		return fmt.Errorf("aspect ratio cannot be empty")
	}

	switch aspectRatio {
	case "16:9", "9:16":
		return nil
	default:
		return fmt.Errorf("aspect ratio must be 16:9 or 9:16")
	}
}

// ValidateDuration validates the video duration
func ValidateDuration(duration int) error {
	switch duration {
	case 4, 6, 8:
		return nil
	default:
		return fmt.Errorf("duration must be 4, 6, or 8 seconds")
	}
}

// ValidatePersonGeneration validates the person generation safety setting
func ValidatePersonGeneration(setting string) error {
	if setting == "" {
		return nil // Optional
	}
	switch setting {
	case "allow_all", "allow_adult", "dont_allow":
		return nil
	default:
		return fmt.Errorf("invalid person generation setting: %s", setting)
	}
}

// ValidateModel validates the model identifier
func ValidateModel(model string) error {
	if model == "" {
		return fmt.Errorf("model cannot be empty")
	}

	// Check against known models - support both short and full names
	validModels := map[string]bool{
		"veo-3.1":                       true,
		"veo-3.1-generate-preview":      true,
		"veo-3.1-fast-generate-preview": true,
		"veo-3-generate-preview":        true,
		"veo-3-fast-generate-preview":   true,
		"veo-3.0":                       true,
		"veo-2.0-generate-001":          true,
	}

	if !validModels[model] {
		return fmt.Errorf("unsupported model: %s", model)
	}

	return nil
}

// ValidateModelConstraints validates model-specific constraints
func ValidateModelConstraints(model string, params map[string]interface{}) error {
	if model == "" {
		return fmt.Errorf("model cannot be empty")
	}

	// Check if reference images are used
	refImages, hasRefImages := params["reference_images"]
	if hasRefImages {
		if refImageList, ok := refImages.([]string); ok && len(refImageList) > 0 {
			// Reference images require specific constraints

			// Check duration
			if duration, hasDuration := params["duration"]; hasDuration {
				if d, ok := duration.(int); ok && d != 8 {
					return fmt.Errorf("reference images require 8 seconds duration")
				}
			}

			// Check aspect ratio
			if aspectRatio, hasAspectRatio := params["aspect_ratio"]; hasAspectRatio {
				if ar, ok := aspectRatio.(string); ok && ar != "16:9" {
					return fmt.Errorf("reference images require 16:9 aspect ratio")
				}
			}
		}
	}

	return nil
}
