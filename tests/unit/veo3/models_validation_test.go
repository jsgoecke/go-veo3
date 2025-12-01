package veo3_test

import (
	"strings"
	"testing"

	"github.com/jasongoecke/go-veo3/pkg/veo3"
)

func TestValidateModelForReferenceImages(t *testing.T) {
	tests := []struct {
		name      string
		modelID   string
		count     int
		wantError bool
		errSubstr string
	}{
		{
			name:      "veo-3.1 with 1 reference image succeeds",
			modelID:   "veo-3.1-generate-preview",
			count:     1,
			wantError: false,
		},
		{
			name:      "veo-3.1 with 3 reference images succeeds",
			modelID:   "veo-3.1-generate-preview",
			count:     3,
			wantError: false,
		},
		{
			name:      "veo-3.1-fast with 2 reference images succeeds",
			modelID:   "veo-3.1-fast-generate-preview",
			count:     2,
			wantError: false,
		},
		{
			name:      "veo-3.1 with 4 reference images fails",
			modelID:   "veo-3.1-generate-preview",
			count:     4,
			wantError: true,
			errSubstr: "too many reference images",
		},
		{
			name:      "veo-3.1 with 0 reference images succeeds",
			modelID:   "veo-3.1-generate-preview",
			count:     0,
			wantError: false,
		},
		{
			name:      "veo-3.0 does not support reference images",
			modelID:   "veo-3-generate-preview",
			count:     1,
			wantError: true,
			errSubstr: "does not support reference images",
		},
		{
			name:      "veo-3.0-fast does not support reference images",
			modelID:   "veo-3-fast-generate-preview",
			count:     1,
			wantError: true,
			errSubstr: "does not support reference images",
		},
		{
			name:      "veo-2.0 does not support reference images",
			modelID:   "veo-2.0-generate-001",
			count:     1,
			wantError: true,
			errSubstr: "does not support reference images",
		},
		{
			name:      "unknown model returns error",
			modelID:   "veo-unknown",
			count:     1,
			wantError: true,
			errSubstr: "unknown model",
		},
		{
			name:      "empty model ID returns error",
			modelID:   "",
			count:     1,
			wantError: true,
			errSubstr: "unknown model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := veo3.ValidateModelForReferenceImages(tt.modelID, tt.count)
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateModelForReferenceImages() expected error containing %q, got nil", tt.errSubstr)
				} else if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("ValidateModelForReferenceImages() error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateModelForReferenceImages() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateModelForExtension(t *testing.T) {
	tests := []struct {
		name      string
		modelID   string
		wantError bool
		errSubstr string
	}{
		{
			name:      "veo-3.1 supports extension",
			modelID:   "veo-3.1-generate-preview",
			wantError: false,
		},
		{
			name:      "veo-3.1-fast supports extension",
			modelID:   "veo-3.1-fast-generate-preview",
			wantError: false,
		},
		{
			name:      "veo-3.0 does not support extension",
			modelID:   "veo-3-generate-preview",
			wantError: true,
			errSubstr: "does not support video extension",
		},
		{
			name:      "veo-3.0-fast does not support extension",
			modelID:   "veo-3-fast-generate-preview",
			wantError: true,
			errSubstr: "does not support video extension",
		},
		{
			name:      "veo-2.0 does not support extension",
			modelID:   "veo-2.0-generate-001",
			wantError: true,
			errSubstr: "does not support video extension",
		},
		{
			name:      "unknown model returns error",
			modelID:   "veo-unknown",
			wantError: true,
			errSubstr: "unknown model",
		},
		{
			name:      "empty model ID returns error",
			modelID:   "",
			wantError: true,
			errSubstr: "unknown model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := veo3.ValidateModelForExtension(tt.modelID)
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateModelForExtension() expected error containing %q, got nil", tt.errSubstr)
				} else if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("ValidateModelForExtension() error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateModelForExtension() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateModelForResolution(t *testing.T) {
	tests := []struct {
		name       string
		modelID    string
		resolution string
		wantError  bool
		errSubstr  string
	}{
		{
			name:       "veo-3.1 supports 720p",
			modelID:    "veo-3.1-generate-preview",
			resolution: "720p",
			wantError:  false,
		},
		{
			name:       "veo-3.1 supports 1080p",
			modelID:    "veo-3.1-generate-preview",
			resolution: "1080p",
			wantError:  false,
		},
		{
			name:       "veo-3.1-fast supports 720p",
			modelID:    "veo-3.1-fast-generate-preview",
			resolution: "720p",
			wantError:  false,
		},
		{
			name:       "veo-3.1-fast supports 1080p",
			modelID:    "veo-3.1-fast-generate-preview",
			resolution: "1080p",
			wantError:  false,
		},
		{
			name:       "veo-3.0 supports 720p",
			modelID:    "veo-3-generate-preview",
			resolution: "720p",
			wantError:  false,
		},
		{
			name:       "veo-3.0 supports 1080p",
			modelID:    "veo-3-generate-preview",
			resolution: "1080p",
			wantError:  false,
		},
		{
			name:       "veo-2.0 supports 720p",
			modelID:    "veo-2.0-generate-001",
			resolution: "720p",
			wantError:  false,
		},
		{
			name:       "veo-2.0 does not support 1080p",
			modelID:    "veo-2.0-generate-001",
			resolution: "1080p",
			wantError:  true,
			errSubstr:  "does not support resolution",
		},
		{
			name:       "invalid resolution fails",
			modelID:    "veo-3.1-generate-preview",
			resolution: "4K",
			wantError:  true,
			errSubstr:  "does not support resolution",
		},
		{
			name:       "unknown model returns error",
			modelID:    "veo-unknown",
			resolution: "720p",
			wantError:  true,
			errSubstr:  "unknown model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := veo3.ValidateModelForResolution(tt.modelID, tt.resolution)
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateModelForResolution() expected error containing %q, got nil", tt.errSubstr)
				} else if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("ValidateModelForResolution() error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateModelForResolution() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateModelForDuration(t *testing.T) {
	tests := []struct {
		name      string
		modelID   string
		duration  int
		wantError bool
		errSubstr string
	}{
		{
			name:      "veo-3.1 supports 4 seconds",
			modelID:   "veo-3.1-generate-preview",
			duration:  4,
			wantError: false,
		},
		{
			name:      "veo-3.1 supports 6 seconds",
			modelID:   "veo-3.1-generate-preview",
			duration:  6,
			wantError: false,
		},
		{
			name:      "veo-3.1 supports 8 seconds",
			modelID:   "veo-3.1-generate-preview",
			duration:  8,
			wantError: false,
		},
		{
			name:      "veo-2.0 supports 5 seconds",
			modelID:   "veo-2.0-generate-001",
			duration:  5,
			wantError: false,
		},
		{
			name:      "veo-2.0 supports 6 seconds",
			modelID:   "veo-2.0-generate-001",
			duration:  6,
			wantError: false,
		},
		{
			name:      "veo-2.0 supports 8 seconds",
			modelID:   "veo-2.0-generate-001",
			duration:  8,
			wantError: false,
		},
		{
			name:      "veo-2.0 does not support 4 seconds",
			modelID:   "veo-2.0-generate-001",
			duration:  4,
			wantError: true,
			errSubstr: "does not support duration",
		},
		{
			name:      "veo-3.1 does not support 5 seconds",
			modelID:   "veo-3.1-generate-preview",
			duration:  5,
			wantError: true,
			errSubstr: "does not support duration",
		},
		{
			name:      "invalid duration fails",
			modelID:   "veo-3.1-generate-preview",
			duration:  10,
			wantError: true,
			errSubstr: "does not support duration",
		},
		{
			name:      "unknown model returns error",
			modelID:   "veo-unknown",
			duration:  8,
			wantError: true,
			errSubstr: "unknown model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := veo3.ValidateModelForDuration(tt.modelID, tt.duration)
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateModelForDuration() expected error containing %q, got nil", tt.errSubstr)
				} else if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("ValidateModelForDuration() error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateModelForDuration() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestListModels(t *testing.T) {
	models := veo3.ListModels()

	// Should return all models in registry
	if len(models) != len(veo3.ModelRegistry) {
		t.Errorf("ListModels() returned %d models, want %d", len(models), len(veo3.ModelRegistry))
	}

	// Verify models are returned
	for _, model := range models {
		if model.ID == "" {
			t.Error("ListModels() returned model with empty ID")
		}
		if model.Name == "" {
			t.Error("ListModels() returned model with empty Name")
		}
	}
}

func TestListModelsByTier(t *testing.T) {
	tests := []struct {
		name      string
		tier      string
		wantCount int
	}{
		{
			name:      "standard tier has 4 models",
			tier:      "standard",
			wantCount: 4,
		},
		{
			name:      "legacy tier has 1 model",
			tier:      "legacy",
			wantCount: 1,
		},
		{
			name:      "unknown tier has 0 models",
			tier:      "unknown",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			models := veo3.ListModelsByTier(tt.tier)
			if len(models) != tt.wantCount {
				t.Errorf("ListModelsByTier(%q) returned %d models, want %d", tt.tier, len(models), tt.wantCount)
			}
			for _, model := range models {
				if model.Tier != tt.tier {
					t.Errorf("ListModelsByTier(%q) returned model with tier %q", tt.tier, model.Tier)
				}
			}
		})
	}
}

func TestListModelsByCapability(t *testing.T) {
	tests := []struct {
		name       string
		capability string
		wantMin    int // minimum expected models with this capability
	}{
		{
			name:       "audio capability",
			capability: "audio",
			wantMin:    4, // veo 3.x models
		},
		{
			name:       "extension capability",
			capability: "extension",
			wantMin:    2, // veo 3.1 models
		},
		{
			name:       "reference_images capability",
			capability: "reference_images",
			wantMin:    2, // veo 3.1 models
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			models := veo3.ListModelsByCapability(tt.capability)
			if len(models) < tt.wantMin {
				t.Errorf("ListModelsByCapability(%q) returned %d models, want at least %d", tt.capability, len(models), tt.wantMin)
			}

			// Verify all returned models have the capability
			for _, model := range models {
				hasCapability := false
				switch tt.capability {
				case "audio":
					hasCapability = model.Capabilities.Audio
				case "extension":
					hasCapability = model.Capabilities.Extension
				case "reference_images":
					hasCapability = model.Capabilities.ReferenceImages
				}
				if !hasCapability {
					t.Errorf("ListModelsByCapability(%q) returned model %s without that capability", tt.capability, model.ID)
				}
			}
		})
	}
}
