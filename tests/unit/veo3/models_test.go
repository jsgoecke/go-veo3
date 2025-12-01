package veo3_test

import (
	"testing"

	"github.com/jasongoecke/go-veo3/pkg/veo3"
)

func TestGetModel(t *testing.T) {
	tests := []struct {
		name      string
		modelID   string
		wantFound bool
		wantName  string
	}{
		{
			name:      "veo-3.1-generate-preview exists",
			modelID:   "veo-3.1-generate-preview",
			wantFound: true,
			wantName:  "Veo 3.1 Preview",
		},
		{
			name:      "veo-3.1-fast-generate-preview exists",
			modelID:   "veo-3.1-fast-generate-preview",
			wantFound: true,
			wantName:  "Veo 3.1 Fast Preview",
		},
		{
			name:      "veo-3-generate-preview exists",
			modelID:   "veo-3-generate-preview",
			wantFound: true,
			wantName:  "Veo 3.0 Preview",
		},
		{
			name:      "veo-3-fast-generate-preview exists",
			modelID:   "veo-3-fast-generate-preview",
			wantFound: true,
			wantName:  "Veo 3.0 Fast Preview",
		},
		{
			name:      "veo-2.0-generate-001 exists",
			modelID:   "veo-2.0-generate-001",
			wantFound: true,
			wantName:  "Veo 2.0",
		},
		{
			name:      "unknown model returns false",
			modelID:   "veo-unknown",
			wantFound: false,
		},
		{
			name:      "empty model ID returns false",
			modelID:   "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, found := veo3.GetModel(tt.modelID)
			if found != tt.wantFound {
				t.Errorf("GetModel() found = %v, want %v", found, tt.wantFound)
			}
			if tt.wantFound && model.Name != tt.wantName {
				t.Errorf("GetModel() name = %v, want %v", model.Name, tt.wantName)
			}
		})
	}
}

func TestModelCapabilities_Audio(t *testing.T) {
	tests := []struct {
		name      string
		modelID   string
		wantAudio bool
	}{
		{
			name:      "veo-3.1-generate-preview supports audio",
			modelID:   "veo-3.1-generate-preview",
			wantAudio: true,
		},
		{
			name:      "veo-3.1-fast-generate-preview supports audio",
			modelID:   "veo-3.1-fast-generate-preview",
			wantAudio: true,
		},
		{
			name:      "veo-3-generate-preview supports audio",
			modelID:   "veo-3-generate-preview",
			wantAudio: true,
		},
		{
			name:      "veo-3-fast-generate-preview supports audio",
			modelID:   "veo-3-fast-generate-preview",
			wantAudio: true,
		},
		{
			name:      "veo-2.0-generate-001 does not support audio",
			modelID:   "veo-2.0-generate-001",
			wantAudio: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, found := veo3.GetModel(tt.modelID)
			if !found {
				t.Fatalf("Model %s not found", tt.modelID)
			}
			if model.Capabilities.Audio != tt.wantAudio {
				t.Errorf("Audio capability = %v, want %v", model.Capabilities.Audio, tt.wantAudio)
			}
		})
	}
}

func TestModelCapabilities_Extension(t *testing.T) {
	tests := []struct {
		name          string
		modelID       string
		wantExtension bool
	}{
		{
			name:          "veo-3.1-generate-preview supports extension",
			modelID:       "veo-3.1-generate-preview",
			wantExtension: true,
		},
		{
			name:          "veo-3.1-fast-generate-preview supports extension",
			modelID:       "veo-3.1-fast-generate-preview",
			wantExtension: true,
		},
		{
			name:          "veo-3-generate-preview does not support extension",
			modelID:       "veo-3-generate-preview",
			wantExtension: false,
		},
		{
			name:          "veo-3-fast-generate-preview does not support extension",
			modelID:       "veo-3-fast-generate-preview",
			wantExtension: false,
		},
		{
			name:          "veo-2.0-generate-001 does not support extension",
			modelID:       "veo-2.0-generate-001",
			wantExtension: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, found := veo3.GetModel(tt.modelID)
			if !found {
				t.Fatalf("Model %s not found", tt.modelID)
			}
			if model.Capabilities.Extension != tt.wantExtension {
				t.Errorf("Extension capability = %v, want %v", model.Capabilities.Extension, tt.wantExtension)
			}
		})
	}
}

func TestModelCapabilities_ReferenceImages(t *testing.T) {
	tests := []struct {
		name                string
		modelID             string
		wantReferenceImages bool
		wantMaxImages       int
	}{
		{
			name:                "veo-3.1-generate-preview supports reference images",
			modelID:             "veo-3.1-generate-preview",
			wantReferenceImages: true,
			wantMaxImages:       3,
		},
		{
			name:                "veo-3.1-fast-generate-preview supports reference images",
			modelID:             "veo-3.1-fast-generate-preview",
			wantReferenceImages: true,
			wantMaxImages:       3,
		},
		{
			name:                "veo-3-generate-preview does not support reference images",
			modelID:             "veo-3-generate-preview",
			wantReferenceImages: false,
			wantMaxImages:       0,
		},
		{
			name:                "veo-3-fast-generate-preview does not support reference images",
			modelID:             "veo-3-fast-generate-preview",
			wantReferenceImages: false,
			wantMaxImages:       0,
		},
		{
			name:                "veo-2.0-generate-001 does not support reference images",
			modelID:             "veo-2.0-generate-001",
			wantReferenceImages: false,
			wantMaxImages:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, found := veo3.GetModel(tt.modelID)
			if !found {
				t.Fatalf("Model %s not found", tt.modelID)
			}
			if model.Capabilities.ReferenceImages != tt.wantReferenceImages {
				t.Errorf("ReferenceImages capability = %v, want %v", model.Capabilities.ReferenceImages, tt.wantReferenceImages)
			}
			if model.Constraints.MaxReferenceImages != tt.wantMaxImages {
				t.Errorf("MaxReferenceImages = %v, want %v", model.Constraints.MaxReferenceImages, tt.wantMaxImages)
			}
		})
	}
}

func TestModelCapabilities_Resolutions(t *testing.T) {
	tests := []struct {
		name            string
		modelID         string
		wantResolutions []string
	}{
		{
			name:            "veo-3.1-generate-preview supports 720p and 1080p",
			modelID:         "veo-3.1-generate-preview",
			wantResolutions: []string{"720p", "1080p"},
		},
		{
			name:            "veo-3.1-fast-generate-preview supports 720p and 1080p",
			modelID:         "veo-3.1-fast-generate-preview",
			wantResolutions: []string{"720p", "1080p"},
		},
		{
			name:            "veo-3-generate-preview supports 720p and 1080p",
			modelID:         "veo-3-generate-preview",
			wantResolutions: []string{"720p", "1080p"},
		},
		{
			name:            "veo-3-fast-generate-preview supports 720p and 1080p",
			modelID:         "veo-3-fast-generate-preview",
			wantResolutions: []string{"720p", "1080p"},
		},
		{
			name:            "veo-2.0-generate-001 supports only 720p",
			modelID:         "veo-2.0-generate-001",
			wantResolutions: []string{"720p"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, found := veo3.GetModel(tt.modelID)
			if !found {
				t.Fatalf("Model %s not found", tt.modelID)
			}
			if len(model.Capabilities.Resolutions) != len(tt.wantResolutions) {
				t.Errorf("Resolutions count = %v, want %v", len(model.Capabilities.Resolutions), len(tt.wantResolutions))
			}
			for i, res := range tt.wantResolutions {
				if i >= len(model.Capabilities.Resolutions) || model.Capabilities.Resolutions[i] != res {
					t.Errorf("Resolution[%d] = %v, want %v", i, model.Capabilities.Resolutions[i], res)
				}
			}
		})
	}
}

func TestModelCapabilities_Durations(t *testing.T) {
	tests := []struct {
		name          string
		modelID       string
		wantDurations []int
	}{
		{
			name:          "veo-3.1-generate-preview supports 4, 6, 8 seconds",
			modelID:       "veo-3.1-generate-preview",
			wantDurations: []int{4, 6, 8},
		},
		{
			name:          "veo-3.1-fast-generate-preview supports 4, 6, 8 seconds",
			modelID:       "veo-3.1-fast-generate-preview",
			wantDurations: []int{4, 6, 8},
		},
		{
			name:          "veo-3-generate-preview supports 4, 6, 8 seconds",
			modelID:       "veo-3-generate-preview",
			wantDurations: []int{4, 6, 8},
		},
		{
			name:          "veo-3-fast-generate-preview supports 4, 6, 8 seconds",
			modelID:       "veo-3-fast-generate-preview",
			wantDurations: []int{4, 6, 8},
		},
		{
			name:          "veo-2.0-generate-001 supports 5, 6, 8 seconds",
			modelID:       "veo-2.0-generate-001",
			wantDurations: []int{5, 6, 8},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, found := veo3.GetModel(tt.modelID)
			if !found {
				t.Fatalf("Model %s not found", tt.modelID)
			}
			if len(model.Capabilities.Durations) != len(tt.wantDurations) {
				t.Errorf("Durations count = %v, want %v", len(model.Capabilities.Durations), len(tt.wantDurations))
			}
			for i, dur := range tt.wantDurations {
				if i >= len(model.Capabilities.Durations) || model.Capabilities.Durations[i] != dur {
					t.Errorf("Duration[%d] = %v, want %v", i, model.Capabilities.Durations[i], dur)
				}
			}
		})
	}
}

func TestModelRegistry_Count(t *testing.T) {
	// Should have exactly 5 models as per spec.md
	expectedCount := 5
	if len(veo3.ModelRegistry) != expectedCount {
		t.Errorf("ModelRegistry has %d models, want %d", len(veo3.ModelRegistry), expectedCount)
	}
}

func TestModelRegistry_Tiers(t *testing.T) {
	tests := []struct {
		name     string
		modelID  string
		wantTier string
	}{
		{
			name:     "veo-3.1-generate-preview is standard tier",
			modelID:  "veo-3.1-generate-preview",
			wantTier: "standard",
		},
		{
			name:     "veo-3.1-fast-generate-preview is standard tier",
			modelID:  "veo-3.1-fast-generate-preview",
			wantTier: "standard",
		},
		{
			name:     "veo-3-generate-preview is standard tier",
			modelID:  "veo-3-generate-preview",
			wantTier: "standard",
		},
		{
			name:     "veo-3-fast-generate-preview is standard tier",
			modelID:  "veo-3-fast-generate-preview",
			wantTier: "standard",
		},
		{
			name:     "veo-2.0-generate-001 is legacy tier",
			modelID:  "veo-2.0-generate-001",
			wantTier: "legacy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, found := veo3.GetModel(tt.modelID)
			if !found {
				t.Fatalf("Model %s not found", tt.modelID)
			}
			if model.Tier != tt.wantTier {
				t.Errorf("Tier = %v, want %v", model.Tier, tt.wantTier)
			}
		})
	}
}

func TestModelRegistry_Versions(t *testing.T) {
	tests := []struct {
		name        string
		modelID     string
		wantVersion string
	}{
		{
			name:        "veo-3.1-generate-preview version",
			modelID:     "veo-3.1-generate-preview",
			wantVersion: "3.1",
		},
		{
			name:        "veo-3.1-fast-generate-preview version",
			modelID:     "veo-3.1-fast-generate-preview",
			wantVersion: "3.1",
		},
		{
			name:        "veo-3-generate-preview version",
			modelID:     "veo-3-generate-preview",
			wantVersion: "3.0",
		},
		{
			name:        "veo-3-fast-generate-preview version",
			modelID:     "veo-3-fast-generate-preview",
			wantVersion: "3.0",
		},
		{
			name:        "veo-2.0-generate-001 version",
			modelID:     "veo-2.0-generate-001",
			wantVersion: "2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, found := veo3.GetModel(tt.modelID)
			if !found {
				t.Fatalf("Model %s not found", tt.modelID)
			}
			if model.Version != tt.wantVersion {
				t.Errorf("Version = %v, want %v", model.Version, tt.wantVersion)
			}
		})
	}
}
