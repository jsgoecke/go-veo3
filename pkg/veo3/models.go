package veo3

import "fmt"

// Model represents a Veo model
type Model struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Capabilities ModelCapabilities `json:"capabilities"`
	Constraints  ModelConstraints  `json:"constraints"`
	Tier         string            `json:"tier"`
	Version      string            `json:"version"`
}

type ModelCapabilities struct {
	Audio           bool     `json:"audio"`
	Extension       bool     `json:"extension"`
	ReferenceImages bool     `json:"reference_images"`
	Resolutions     []string `json:"resolutions"`
	Durations       []int    `json:"durations"`
}

type ModelConstraints struct {
	MaxReferenceImages  int    `json:"max_reference_images"`
	RequiredAspectRatio string `json:"required_aspect_ratio,omitempty"`
	RequiredDuration    int    `json:"required_duration,omitempty"`
}

// ModelRegistry holds available models
var ModelRegistry = []Model{
	{
		ID:   "veo-3.1-generate-preview",
		Name: "Veo 3.1 Preview",
		Capabilities: ModelCapabilities{
			Audio:           true,
			Extension:       true,
			ReferenceImages: true,
			Resolutions:     []string{"720p", "1080p"},
			Durations:       []int{4, 6, 8},
		},
		Constraints: ModelConstraints{
			MaxReferenceImages: 3,
		},
		Tier:    "standard",
		Version: "3.1",
	},
	{
		ID:   "veo-3.1-fast-generate-preview",
		Name: "Veo 3.1 Fast Preview",
		Capabilities: ModelCapabilities{
			Audio:           true,
			Extension:       true,
			ReferenceImages: true,
			Resolutions:     []string{"720p", "1080p"},
			Durations:       []int{4, 6, 8},
		},
		Constraints: ModelConstraints{
			MaxReferenceImages: 3,
		},
		Tier:    "standard",
		Version: "3.1",
	},
	{
		ID:   "veo-3-generate-preview",
		Name: "Veo 3.0 Preview",
		Capabilities: ModelCapabilities{
			Audio:           true,
			Extension:       false,
			ReferenceImages: false,
			Resolutions:     []string{"720p", "1080p"},
			Durations:       []int{4, 6, 8},
		},
		Constraints: ModelConstraints{
			MaxReferenceImages: 0,
		},
		Tier:    "standard",
		Version: "3.0",
	},
	{
		ID:   "veo-3-fast-generate-preview",
		Name: "Veo 3.0 Fast Preview",
		Capabilities: ModelCapabilities{
			Audio:           true,
			Extension:       false,
			ReferenceImages: false,
			Resolutions:     []string{"720p", "1080p"},
			Durations:       []int{4, 6, 8},
		},
		Constraints: ModelConstraints{
			MaxReferenceImages: 0,
		},
		Tier:    "standard",
		Version: "3.0",
	},
	{
		ID:   "veo-2.0-generate-001",
		Name: "Veo 2.0",
		Capabilities: ModelCapabilities{
			Audio:           false,
			Extension:       false,
			ReferenceImages: false,
			Resolutions:     []string{"720p"},
			Durations:       []int{5, 6, 8},
		},
		Constraints: ModelConstraints{
			MaxReferenceImages: 0,
		},
		Tier:    "legacy",
		Version: "2.0",
	},
}

// GetModel retrieves a model by ID
func GetModel(id string) (Model, bool) {
	for _, m := range ModelRegistry {
		if m.ID == id {
			return m, true
		}
	}
	return Model{}, false
}

// ValidateModelForReferenceImages checks if model supports reference images
func ValidateModelForReferenceImages(modelID string, count int) error {
	model, exists := GetModel(modelID)
	if !exists {
		return fmt.Errorf("unknown model: %s", modelID)
	}
	if !model.Capabilities.ReferenceImages {
		return fmt.Errorf("model %s does not support reference images", modelID)
	}
	if count > model.Constraints.MaxReferenceImages {
		return fmt.Errorf("too many reference images: %d (max %d)", count, model.Constraints.MaxReferenceImages)
	}
	return nil
}

// ValidateModelForExtension checks if model supports extension
func ValidateModelForExtension(modelID string) error {
	model, exists := GetModel(modelID)
	if !exists {
		return fmt.Errorf("unknown model: %s", modelID)
	}
	if !model.Capabilities.Extension {
		return fmt.Errorf("model %s does not support video extension", modelID)
	}
	return nil
}

// ValidateModelForResolution checks if model supports the specified resolution
func ValidateModelForResolution(modelID string, resolution string) error {
	model, exists := GetModel(modelID)
	if !exists {
		return fmt.Errorf("unknown model: %s", modelID)
	}
	for _, res := range model.Capabilities.Resolutions {
		if res == resolution {
			return nil
		}
	}
	return fmt.Errorf("model %s does not support resolution %s (supported: %v)", modelID, resolution, model.Capabilities.Resolutions)
}

// ValidateModelForDuration checks if model supports the specified duration
func ValidateModelForDuration(modelID string, duration int) error {
	model, exists := GetModel(modelID)
	if !exists {
		return fmt.Errorf("unknown model: %s", modelID)
	}
	for _, dur := range model.Capabilities.Durations {
		if dur == duration {
			return nil
		}
	}
	return fmt.Errorf("model %s does not support duration %ds (supported: %v)", modelID, duration, model.Capabilities.Durations)
}

// ListModels returns all available models
func ListModels() []Model {
	return ModelRegistry
}

// ListModelsByTier returns models filtered by tier
func ListModelsByTier(tier string) []Model {
	var models []Model
	for _, m := range ModelRegistry {
		if m.Tier == tier {
			models = append(models, m)
		}
	}
	return models
}

// ListModelsByCapability returns models that have a specific capability
func ListModelsByCapability(capability string) []Model {
	var models []Model
	for _, m := range ModelRegistry {
		hasCapability := false
		switch capability {
		case "audio":
			hasCapability = m.Capabilities.Audio
		case "extension":
			hasCapability = m.Capabilities.Extension
		case "reference_images":
			hasCapability = m.Capabilities.ReferenceImages
		}
		if hasCapability {
			models = append(models, m)
		}
	}
	return models
}
