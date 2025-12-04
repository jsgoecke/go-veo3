package format

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/operations"
	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatOperationJSON(t *testing.T) {
	now := time.Now()
	op := &veo3.Operation{
		ID:        "operations/test-123",
		Status:    veo3.StatusDone,
		StartTime: now,
		EndTime:   &now,
		VideoURI:  "gs://bucket/video.mp4",
	}

	result, err := FormatOperationJSON(op)
	require.NoError(t, err)
	assert.Contains(t, result, `"success": true`)
	assert.Contains(t, result, "operations/test-123")
}

func TestFormatOperationListJSON(t *testing.T) {
	now := time.Now()
	ops := []*veo3.Operation{
		{ID: "op1", Status: veo3.StatusDone, StartTime: now},
		{ID: "op2", Status: veo3.StatusRunning, StartTime: now},
	}

	result, err := FormatOperationListJSON(ops)
	require.NoError(t, err)
	assert.Contains(t, result, `"success": true`)
	assert.Contains(t, result, `"count": 2`)
	assert.Contains(t, result, "op1")
	assert.Contains(t, result, "op2")
}

func TestFormatOperationStatsJSON(t *testing.T) {
	stats := operations.OperationStats{
		Total:     10,
		Pending:   2,
		Running:   3,
		Completed: 4,
		Failed:    1,
		Cancelled: 0,
	}

	result, err := FormatOperationStatsJSON(stats)
	require.NoError(t, err)
	assert.Contains(t, result, `"success": true`)
	assert.Contains(t, result, `"Total": 10`)
	assert.Contains(t, result, `"Pending": 2`)
}

func TestFormatModelJSON(t *testing.T) {
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
			Durations:       []int{4, 8},
		},
	}

	result, err := FormatModelJSON(model)
	require.NoError(t, err)
	assert.Contains(t, result, `"success": true`)
	assert.Contains(t, result, "test-model")
}

func TestFormatModelListJSON(t *testing.T) {
	models := []veo3.Model{
		{Name: "model1", ID: "id1", Version: "1.0"},
		{Name: "model2", ID: "id2", Version: "2.0"},
	}

	result, err := FormatModelListJSON(models)
	require.NoError(t, err)
	assert.Contains(t, result, `"success": true`)
	assert.Contains(t, result, `"count": 2`)
	assert.Contains(t, result, "model1")
	assert.Contains(t, result, "model2")
}

func TestFormatGeneratedVideoJSON(t *testing.T) {
	video := &veo3.GeneratedVideo{
		FilePath:              "/path/to/video.mp4",
		OperationID:           "op-123",
		FileSizeBytes:         1024000,
		DurationSeconds:       8,
		Resolution:            "720p",
		AspectRatio:           "16:9",
		Model:                 "test-model",
		Prompt:                "test prompt",
		GenerationTimeSeconds: 120,
		CreatedAt:             time.Now(),
	}

	result, err := FormatGeneratedVideoJSON(video)
	require.NoError(t, err)
	assert.Contains(t, result, `"success": true`)
	assert.Contains(t, result, "/path/to/video.mp4")
	assert.Contains(t, result, "op-123")
}

func TestFormatConfigJSON(t *testing.T) {
	config := map[string]interface{}{
		"default_model":      "test-model",
		"default_resolution": "720p",
	}

	result, err := FormatConfigJSON(config)
	require.NoError(t, err)
	assert.Contains(t, result, `"success": true`)
	assert.Contains(t, result, "test-model")
	assert.Contains(t, result, "720p")
}

func TestFormatSuccessJSON(t *testing.T) {
	result, err := FormatSuccessJSON("Operation completed", map[string]string{"video": "path.mp4"})
	require.NoError(t, err)
	assert.Contains(t, result, `"success": true`)
	assert.Contains(t, result, "Operation completed")
	assert.Contains(t, result, "path.mp4")
}

func TestFormatErrorJSON(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		message string
		details interface{}
		check   func(*testing.T, string)
	}{
		{
			name:    "simple error",
			code:    "ERROR",
			message: "Something went wrong",
			details: nil,
			check: func(t *testing.T, result string) {
				assert.Contains(t, result, `"success": false`)
				assert.Contains(t, result, "ERROR")
				assert.Contains(t, result, "Something went wrong")
			},
		},
		{
			name:    "error with string details",
			code:    "VALIDATION_ERROR",
			message: "Invalid input",
			details: "Field 'name' is required",
			check: func(t *testing.T, result string) {
				assert.Contains(t, result, `"success": false`)
				assert.Contains(t, result, "VALIDATION_ERROR")
				assert.Contains(t, result, "Invalid input")
				assert.Contains(t, result, "Field 'name' is required")
			},
		},
		{
			name:    "error with complex details",
			code:    "API_ERROR",
			message: "Request failed",
			details: map[string]interface{}{"status": 500, "retry": true},
			check: func(t *testing.T, result string) {
				assert.Contains(t, result, `"success": false`)
				assert.Contains(t, result, "API_ERROR")
				assert.Contains(t, result, "Request failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatErrorJSON(tt.code, tt.message, tt.details)
			require.NoError(t, err)
			tt.check(t, result)
		})
	}
}

func TestFormatOperationErrorJSON(t *testing.T) {
	opErr := &veo3.OperationError{
		Code:    "GENERATION_FAILED",
		Message: "Video generation failed",
		Details: map[string]interface{}{"reason": "timeout"},
	}

	result, err := FormatOperationErrorJSON(opErr)
	require.NoError(t, err)
	assert.Contains(t, result, `"success": false`)
	assert.Contains(t, result, "GENERATION_FAILED")
	assert.Contains(t, result, "Video generation failed")
}

func TestFormatValidationErrorJSON(t *testing.T) {
	result, err := FormatValidationErrorJSON("prompt", "Prompt is required")
	require.NoError(t, err)
	assert.Contains(t, result, `"success": false`)
	assert.Contains(t, result, "VALIDATION_ERROR")
	assert.Contains(t, result, "Prompt is required")
	assert.Contains(t, result, "prompt")
}

func TestFormatBatchResultJSON(t *testing.T) {
	results := []map[string]string{
		{"id": "job1", "status": "completed"},
		{"id": "job2", "status": "failed"},
	}
	summary := map[string]int{
		"total":     2,
		"completed": 1,
		"failed":    1,
	}

	result, err := FormatBatchResultJSON(results, summary)
	require.NoError(t, err)
	assert.Contains(t, result, `"success": true`)
	assert.Contains(t, result, "job1")
	assert.Contains(t, result, "job2")
	assert.Contains(t, result, `"total": 2`)
}

func TestFormatProgressJSON(t *testing.T) {
	result, err := FormatProgressJSON("op-123", veo3.StatusRunning, 0.5, "Generating video")
	require.NoError(t, err)
	assert.Contains(t, result, `"success": true`)
	assert.Contains(t, result, "op-123")
	assert.Contains(t, result, "RUNNING")
	assert.Contains(t, result, "0.5")
	assert.Contains(t, result, "Generating video")
}

func TestFormatGenericJSON(t *testing.T) {
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	result, err := FormatGenericJSON(data)
	require.NoError(t, err)
	assert.Contains(t, result, `"success": true`)
	assert.Contains(t, result, "value1")
	assert.Contains(t, result, "42")
}

func TestFormatCompactJSON(t *testing.T) {
	data := map[string]interface{}{
		"key": "value",
	}

	result, err := FormatCompactJSON(data)
	require.NoError(t, err)

	// Compact JSON should not have extra whitespace
	assert.NotContains(t, result, "\n  ")
	assert.Contains(t, result, `"success":true`)
	assert.Contains(t, result, "value")
}

func TestParseJSONError(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
		check   func(*testing.T, *JSONError)
	}{
		{
			name:    "valid error JSON",
			jsonStr: `{"success":false,"error":{"code":"ERROR","message":"Failed"}}`,
			wantErr: false,
			check: func(t *testing.T, err *JSONError) {
				assert.Equal(t, "ERROR", err.Code)
				assert.Equal(t, "Failed", err.Message)
			},
		},
		{
			name:    "no error in JSON",
			jsonStr: `{"success":true,"data":{}}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			jsonStr: `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseJSONError(tt.jsonStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, result)
				}
			}
		})
	}
}

func TestIsJSONOutput(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{
			name: "valid JSON output",
			str:  `{"success":true,"data":{}}`,
			want: true,
		},
		{
			name: "valid error JSON",
			str:  `{"success":false,"error":{"code":"ERROR","message":"Failed"}}`,
			want: true,
		},
		{
			name: "invalid JSON",
			str:  `{invalid}`,
			want: false,
		},
		{
			name: "empty string",
			str:  ``,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsJSONOutput(tt.str)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	output := JSONOutput{
		Success: true,
		Data:    map[string]string{"key": "value"},
	}

	result, err := marshalJSON(output)
	require.NoError(t, err)

	// Should be pretty-printed with indentation
	assert.Contains(t, result, "\n")
	assert.Contains(t, result, `"success": true`)

	// Should be valid JSON
	var parsed JSONOutput
	err = json.Unmarshal([]byte(result), &parsed)
	require.NoError(t, err)
	assert.True(t, parsed.Success)
}

func TestMarshalJSONCompact(t *testing.T) {
	output := JSONOutput{
		Success: true,
		Data:    map[string]string{"key": "value"},
	}

	result, err := marshalJSONCompact(output)
	require.NoError(t, err)

	// Should be compact without extra whitespace
	assert.NotContains(t, result, "\n  ")
	assert.Contains(t, result, `"success":true`)
}
