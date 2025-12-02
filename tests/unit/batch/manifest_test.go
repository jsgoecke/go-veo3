package batch_test

import (
	"testing"

	"github.com/jasongoecke/go-veo3/pkg/batch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseManifest(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		validate    func(*testing.T, *batch.BatchManifest)
	}{
		{
			name: "valid manifest with generate jobs",
			yamlContent: `
jobs:
  - id: job1
    type: generate
    options:
      prompt: "A sunset over mountains"
      model: "veo-3.1-generate-preview"
      duration: 8
    output: sunset.mp4
  - id: job2
    type: generate
    options:
      prompt: "Ocean waves"
      resolution: "1080p"
    output: ocean.mp4
concurrency: 3
continue_on_error: true
output_directory: /tmp/videos
`,
			wantErr: false,
			validate: func(t *testing.T, m *batch.BatchManifest) {
				assert.Len(t, m.Jobs, 2)
				assert.Equal(t, 3, m.Concurrency)
				assert.True(t, m.ContinueOnError)
				assert.Equal(t, "/tmp/videos", m.OutputDirectory)
				assert.Equal(t, "job1", m.Jobs[0].ID)
				assert.Equal(t, "generate", m.Jobs[0].Type)
			},
		},
		{
			name: "manifest with mixed job types",
			yamlContent: `
jobs:
  - id: gen1
    type: generate
    options:
      prompt: "Test"
    output: out1.mp4
  - id: anim1
    type: animate
    options:
      image: test.png
      prompt: "Animate this"
    output: out2.mp4
  - id: interp1
    type: interpolate
    options:
      first_frame: start.png
      last_frame: end.png
    output: out3.mp4
`,
			wantErr: false,
			validate: func(t *testing.T, m *batch.BatchManifest) {
				assert.Len(t, m.Jobs, 3)
				assert.Equal(t, "generate", m.Jobs[0].Type)
				assert.Equal(t, "animate", m.Jobs[1].Type)
				assert.Equal(t, "interpolate", m.Jobs[2].Type)
			},
		},
		{
			name: "manifest with defaults",
			yamlContent: `
jobs:
  - id: job1
    type: generate
    options:
      prompt: "Test"
    output: out.mp4
`,
			wantErr: false,
			validate: func(t *testing.T, m *batch.BatchManifest) {
				// Should apply default values
				assert.Equal(t, 3, m.Concurrency) // Default
				assert.True(t, m.ContinueOnError) // Default true
			},
		},
		{
			name: "invalid yaml",
			yamlContent: `
jobs:
  - id: job1
    invalid yaml structure
`,
			wantErr: true,
		},
		{
			name: "empty manifest",
			yamlContent: `
jobs: []
`,
			wantErr: true, // Should error on empty jobs
		},
		{
			name: "missing required fields",
			yamlContent: `
jobs:
  - type: generate
    options:
      prompt: "Test"
`,
			wantErr: true, // Missing ID and output
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := batch.ParseManifest([]byte(tt.yamlContent))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, manifest)

			if tt.validate != nil {
				tt.validate(t, manifest)
			}
		})
	}
}

func TestValidateManifest(t *testing.T) {
	tests := []struct {
		name     string
		manifest *batch.BatchManifest
		wantErr  bool
	}{
		{
			name: "valid manifest",
			manifest: &batch.BatchManifest{
				Jobs: []batch.BatchJob{
					{
						ID:   "job1",
						Type: "generate",
						Options: map[string]interface{}{
							"prompt": "Test",
						},
						Output: "out.mp4",
					},
				},
				Concurrency: 3,
			},
			wantErr: false,
		},
		{
			name: "duplicate job IDs",
			manifest: &batch.BatchManifest{
				Jobs: []batch.BatchJob{
					{ID: "job1", Type: "generate", Options: map[string]interface{}{"prompt": "A"}, Output: "a.mp4"},
					{ID: "job1", Type: "generate", Options: map[string]interface{}{"prompt": "B"}, Output: "b.mp4"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid job type",
			manifest: &batch.BatchManifest{
				Jobs: []batch.BatchJob{
					{ID: "job1", Type: "invalid_type", Options: map[string]interface{}{}, Output: "out.mp4"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing prompt in generate job",
			manifest: &batch.BatchManifest{
				Jobs: []batch.BatchJob{
					{ID: "job1", Type: "generate", Options: map[string]interface{}{}, Output: "out.mp4"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid concurrency",
			manifest: &batch.BatchManifest{
				Jobs: []batch.BatchJob{
					{ID: "job1", Type: "generate", Options: map[string]interface{}{"prompt": "Test"}, Output: "out.mp4"},
				},
				Concurrency: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := batch.ValidateManifest(tt.manifest)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	manifest := &batch.BatchManifest{
		Jobs: []batch.BatchJob{
			{ID: "job1", Type: "generate", Options: map[string]interface{}{"prompt": "Test"}, Output: "out.mp4"},
		},
	}

	batch.ApplyDefaults(manifest)

	assert.Equal(t, 3, manifest.Concurrency, "Should apply default concurrency")
	assert.True(t, manifest.ContinueOnError, "Should apply default continue_on_error")
}
