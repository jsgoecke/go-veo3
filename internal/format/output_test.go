package format

import (
	"errors"
	"testing"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/operations"
	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/stretchr/testify/assert"
)

func TestFormatOperation(t *testing.T) {
	now := time.Now()
	endTime := now.Add(5 * time.Minute)

	tests := []struct {
		name  string
		op    *veo3.Operation
		check func(*testing.T, string)
	}{
		{
			name: "completed operation",
			op: &veo3.Operation{
				ID:        "operations/test-123",
				Status:    veo3.StatusDone,
				Progress:  1.0,
				StartTime: now,
				EndTime:   &endTime,
				VideoURI:  "gs://bucket/video.mp4",
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "operations/test-123")
				assert.Contains(t, output, "✅ DONE")
				assert.Contains(t, output, "100.0%")
				assert.Contains(t, output, "gs://bucket/video.mp4")
			},
		},
		{
			name: "running operation",
			op: &veo3.Operation{
				ID:        "operations/test-456",
				Status:    veo3.StatusRunning,
				Progress:  0.5,
				StartTime: now,
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "operations/test-456")
				assert.Contains(t, output, "🔄 RUNNING")
				assert.Contains(t, output, "50.0%")
			},
		},
		{
			name: "failed operation with error",
			op: &veo3.Operation{
				ID:        "operations/test-789",
				Status:    veo3.StatusFailed,
				StartTime: now,
				EndTime:   &endTime,
				Error: &veo3.OperationError{
					Code:       "GENERATION_FAILED",
					Message:    "Video generation failed",
					Suggestion: "Try again with a different prompt",
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "operations/test-789")
				assert.Contains(t, output, "❌ FAILED")
				assert.Contains(t, output, "Video generation failed")
				assert.Contains(t, output, "Try again with a different prompt")
			},
		},
		{
			name: "operation with metadata",
			op: &veo3.Operation{
				ID:        "operations/test-meta",
				Status:    veo3.StatusDone,
				StartTime: now,
				EndTime:   &endTime,
				Metadata: map[string]interface{}{
					"model":      "test-model",
					"resolution": "720p",
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "Metadata:")
				assert.Contains(t, output, "model")
				assert.Contains(t, output, "resolution")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatOperation(tt.op)
			tt.check(t, result)
		})
	}
}

func TestFormatOperationList(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		ops   []*veo3.Operation
		check func(*testing.T, string)
	}{
		{
			name: "empty list",
			ops:  []*veo3.Operation{},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "No operations found")
			},
		},
		{
			name: "multiple operations",
			ops: []*veo3.Operation{
				{
					ID:        "operations/op1",
					Status:    veo3.StatusDone,
					Progress:  1.0,
					StartTime: now,
					Metadata:  map[string]interface{}{"model": "test-model-1"},
				},
				{
					ID:        "operations/op2",
					Status:    veo3.StatusRunning,
					Progress:  0.5,
					StartTime: now,
					Metadata:  map[string]interface{}{"model": "test-model-2"},
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "OPERATION ID")
				assert.Contains(t, output, "STATUS")
				assert.Contains(t, output, "op1")
				assert.Contains(t, output, "op2")
				assert.Contains(t, output, "✅ DONE")
				assert.Contains(t, output, "🔄 RUNNING")
			},
		},
		{
			name: "long operation ID",
			ops: []*veo3.Operation{
				{
					ID:        "operations/very-long-operation-id-that-should-be-truncated-for-display",
					Status:    veo3.StatusDone,
					StartTime: now,
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "...")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatOperationList(tt.ops)
			tt.check(t, result)
		})
	}
}

func TestFormatOperationStats(t *testing.T) {
	stats := operations.OperationStats{
		Total:     15,
		Pending:   2,
		Running:   3,
		Completed: 8,
		Failed:    1,
		Cancelled: 1,
	}

	result := FormatOperationStats(stats)
	assert.Contains(t, result, "Operation Statistics")
	assert.Contains(t, result, "Total:     15")
	assert.Contains(t, result, "Pending:   2")
	assert.Contains(t, result, "Running:   3")
	assert.Contains(t, result, "Completed: 8")
	assert.Contains(t, result, "Failed:    1")
	assert.Contains(t, result, "Cancelled: 1")
}

func TestFormatModel(t *testing.T) {
	model := veo3.Model{
		Name:    "test-model",
		ID:      "model-123",
		Version: "1.0",
		Tier:    "preview",
		Capabilities: veo3.ModelCapabilities{
			Audio:           true,
			Extension:       false,
			ReferenceImages: true,
			Resolutions:     []string{"720p", "1080p"},
			Durations:       []int{4, 6, 8},
		},
		Constraints: veo3.ModelConstraints{
			MaxReferenceImages:  3,
			RequiredAspectRatio: "16:9",
			RequiredDuration:    8,
		},
	}

	result := FormatModel(model)
	assert.Contains(t, result, "Model: test-model")
	assert.Contains(t, result, "ID: model-123")
	assert.Contains(t, result, "Version: 1.0")
	assert.Contains(t, result, "Tier: preview")
	assert.Contains(t, result, "Capabilities:")
	assert.Contains(t, result, "Audio: Yes")
	assert.Contains(t, result, "Extension: No")
	assert.Contains(t, result, "Reference Images: Yes")
	assert.Contains(t, result, "720p, 1080p")
	assert.Contains(t, result, "4s, 6s, 8s")
	assert.Contains(t, result, "Constraints:")
	assert.Contains(t, result, "Max Reference Images: 3")
	assert.Contains(t, result, "Required Aspect Ratio: 16:9")
	assert.Contains(t, result, "Required Duration: 8s")
}

func TestFormatModelList(t *testing.T) {
	tests := []struct {
		name   string
		models []veo3.Model
		check  func(*testing.T, string)
	}{
		{
			name:   "empty list",
			models: []veo3.Model{},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "No models available")
			},
		},
		{
			name: "multiple models",
			models: []veo3.Model{
				{
					Name:    "model-1",
					Version: "1.0",
					Tier:    "preview",
					Capabilities: veo3.ModelCapabilities{
						Audio:           true,
						Extension:       false,
						ReferenceImages: true,
					},
				},
				{
					Name:    "model-2",
					Version: "2.0",
					Tier:    "stable",
					Capabilities: veo3.ModelCapabilities{
						Audio:           false,
						Extension:       true,
						ReferenceImages: false,
					},
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "MODEL NAME")
				assert.Contains(t, output, "model-1")
				assert.Contains(t, output, "model-2")
				assert.Contains(t, output, "1.0")
				assert.Contains(t, output, "2.0")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatModelList(tt.models)
			tt.check(t, result)
		})
	}
}

func TestFormatGeneratedVideo(t *testing.T) {
	video := &veo3.GeneratedVideo{
		FilePath:              "/path/to/video.mp4",
		OperationID:           "op-123",
		FileSizeBytes:         10485760, // 10 MB
		DurationSeconds:       8,
		Resolution:            "720p",
		AspectRatio:           "16:9",
		Model:                 "test-model",
		Prompt:                "A beautiful sunset",
		GenerationTimeSeconds: 120,
		CreatedAt:             time.Now(),
	}

	result := FormatGeneratedVideo(video)
	assert.Contains(t, result, "✅ Video generated successfully!")
	assert.Contains(t, result, "📁 File: /path/to/video.mp4")
	assert.Contains(t, result, "Duration: 8s")
	assert.Contains(t, result, "Resolution: 720p")
	assert.Contains(t, result, "Generation time:")
	assert.Contains(t, result, "🤖 Model: test-model")
	assert.Contains(t, result, "💭 Prompt: A beautiful sunset")
}

func TestFormatError(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		check func(*testing.T, string)
	}{
		{
			name: "nil error",
			err:  nil,
			check: func(t *testing.T, output string) {
				assert.Empty(t, output)
			},
		},
		{
			name: "operation error",
			err: &veo3.OperationError{
				Code:       "GENERATION_FAILED",
				Message:    "Video generation failed",
				Suggestion: "Try again later",
				Details: map[string]interface{}{
					"reason": "timeout",
				},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "❌ GENERATION_FAILED")
				assert.Contains(t, output, "Video generation failed")
				assert.Contains(t, output, "💡 Try again later")
				assert.Contains(t, output, "Details:")
				assert.Contains(t, output, "reason")
			},
		},
		{
			name: "generic error",
			err:  errors.New("something went wrong"),
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "❌ Error:")
				assert.Contains(t, output, "something went wrong")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatError(tt.err)
			tt.check(t, result)
		})
	}
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		status veo3.OperationStatus
		want   string
	}{
		{veo3.StatusPending, "⏳ PENDING"},
		{veo3.StatusRunning, "🔄 RUNNING"},
		{veo3.StatusDone, "✅ DONE"},
		{veo3.StatusFailed, "❌ FAILED"},
		{veo3.StatusCancelled, "🚫 CANCELLED"},
		{veo3.OperationStatus("UNKNOWN"), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := formatStatus(tt.status)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"seconds only", 45 * time.Second, "0:45"},
		{"minutes and seconds", 2*time.Minute + 30*time.Second, "2:30"},
		{"hours", 1*time.Hour + 15*time.Minute + 30*time.Second, "1:15:30"},
		{"zero", 0, "0:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFormatBool(t *testing.T) {
	assert.Equal(t, "Yes", formatBool(true))
	assert.Equal(t, "No", formatBool(false))
}

func TestFormatIntSlice(t *testing.T) {
	tests := []struct {
		name string
		ints []int
		want string
	}{
		{"empty", []int{}, ""},
		{"single", []int{4}, "4s"},
		{"multiple", []int{4, 6, 8}, "4s, 6s, 8s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatIntSlice(tt.ints)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"bytes", 512, "512 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"megabytes", 1048576, "1.0 MB"},
		{"gigabytes", 1073741824, "1.0 GB"},
		{"mixed", 1536, "1.5 KB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileSize(tt.bytes)
			assert.Equal(t, tt.want, result)
		})
	}
}
