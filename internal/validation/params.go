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

// ValidateResolution checks if the resolution is supported
func ValidateResolution(resolution string) error {
	switch resolution {
	case "720p", "1080p":
		return nil
	default:
		return fmt.Errorf("unsupported resolution: %s (supported: 720p, 1080p)", resolution)
	}
}

// ValidateAspectRatio checks if the aspect ratio is supported
func ValidateAspectRatio(ratio string) error {
	switch ratio {
	case "16:9", "9:16":
		return nil
	default:
		return fmt.Errorf("unsupported aspect ratio: %s (supported: 16:9, 9:16)", ratio)
	}
}

// ValidateDuration checks if the duration is supported
func ValidateDuration(seconds int) error {
	switch seconds {
	case 4, 6, 8:
		return nil
	default:
		return fmt.Errorf("unsupported duration: %d seconds (supported: 4, 6, 8)", seconds)
	}
}
