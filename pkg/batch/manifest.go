package batch

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// BatchManifest represents a collection of batch generation jobs
type BatchManifest struct {
	Jobs            []BatchJob `yaml:"jobs" json:"jobs"`
	Concurrency     int        `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
	ContinueOnError bool       `yaml:"continue_on_error,omitempty" json:"continue_on_error,omitempty"`
	OutputDirectory string     `yaml:"output_directory,omitempty" json:"output_directory,omitempty"`
}

// BatchJob represents a single video generation job in a batch
type BatchJob struct {
	ID      string                 `yaml:"id" json:"id"`
	Type    string                 `yaml:"type" json:"type"` // "generate", "animate", "interpolate", "extend"
	Options map[string]interface{} `yaml:"options" json:"options"`
	Output  string                 `yaml:"output" json:"output"`
}

// ParseManifest parses a YAML manifest from bytes
func ParseManifest(data []byte) (*BatchManifest, error) {
	var manifest BatchManifest

	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Apply defaults
	ApplyDefaults(&manifest)

	// Validate
	if err := ValidateManifest(&manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	return &manifest, nil
}

// ParseManifestFile parses a YAML manifest from a file
func ParseManifestFile(path string) (*BatchManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	return ParseManifest(data)
}

// ValidateManifest validates a batch manifest
func ValidateManifest(manifest *BatchManifest) error {
	if len(manifest.Jobs) == 0 {
		return fmt.Errorf("manifest must contain at least one job")
	}

	if manifest.Concurrency < 1 {
		return fmt.Errorf("concurrency must be at least 1")
	}

	// Check for duplicate job IDs
	seen := make(map[string]bool)
	for i, job := range manifest.Jobs {
		if job.ID == "" {
			return fmt.Errorf("job at index %d missing required field: id", i)
		}

		if seen[job.ID] {
			return fmt.Errorf("duplicate job ID: %s", job.ID)
		}
		seen[job.ID] = true

		if job.Output == "" {
			return fmt.Errorf("job %s missing required field: output", job.ID)
		}

		// Validate job type
		validTypes := map[string]bool{
			"generate":    true,
			"animate":     true,
			"interpolate": true,
			"extend":      true,
		}
		if !validTypes[job.Type] {
			return fmt.Errorf("job %s has invalid type: %s (must be generate, animate, interpolate, or extend)", job.ID, job.Type)
		}

		// Validate job-specific options
		if err := validateJobOptions(job); err != nil {
			return fmt.Errorf("job %s: %w", job.ID, err)
		}
	}

	return nil
}

// validateJobOptions validates job-specific required options
func validateJobOptions(job BatchJob) error {
	switch job.Type {
	case "generate":
		if _, ok := job.Options["prompt"]; !ok {
			return fmt.Errorf("generate job requires 'prompt' option")
		}

	case "animate":
		if _, ok := job.Options["image"]; !ok {
			return fmt.Errorf("animate job requires 'image' option")
		}
		if _, ok := job.Options["prompt"]; !ok {
			return fmt.Errorf("animate job requires 'prompt' option")
		}

	case "interpolate":
		if _, ok := job.Options["first_frame"]; !ok {
			return fmt.Errorf("interpolate job requires 'first_frame' option")
		}
		if _, ok := job.Options["last_frame"]; !ok {
			return fmt.Errorf("interpolate job requires 'last_frame' option")
		}

	case "extend":
		if _, ok := job.Options["video"]; !ok {
			return fmt.Errorf("extend job requires 'video' option")
		}
	}

	return nil
}

// ApplyDefaults applies default values to a manifest
func ApplyDefaults(manifest *BatchManifest) {
	if manifest.Concurrency == 0 {
		manifest.Concurrency = 3 // Default concurrency
	}

	if !manifest.ContinueOnError {
		manifest.ContinueOnError = true // Default to continuing on error
	}
}

// GenerateTemplate generates a sample manifest template
func GenerateTemplate() string {
	return `# Veo3 CLI Batch Processing Manifest
# This file defines multiple video generation jobs to be processed in parallel

# Global settings
concurrency: 3              # Number of jobs to run in parallel (default: 3)
continue_on_error: true     # Continue processing if a job fails (default: true)
output_directory: ./videos  # Optional: Override output directory for all jobs

# List of jobs to process
jobs:
  # Text-to-video generation
  - id: job1
    type: generate
    options:
      prompt: "A cinematic shot of a sunset over mountains"
      model: "veo-3.1-generate-preview"
      resolution: "720p"
      duration: 8
      aspect_ratio: "16:9"
    output: sunset.mp4

  # Image-to-video animation
  - id: job2
    type: animate
    options:
      image: path/to/image.png
      prompt: "The person waves and smiles"
      duration: 8
    output: animated.mp4

  # Frame interpolation
  - id: job3
    type: interpolate
    options:
      first_frame: path/to/start.png
      last_frame: path/to/end.png
    output: interpolated.mp4

  # Video extension
  - id: job4
    type: extend
    options:
      video: path/to/existing.mp4
      prompt: "The action continues"
    output: extended.mp4
`
}
