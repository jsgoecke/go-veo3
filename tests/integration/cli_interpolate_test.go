package integration_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/jasongoecke/go-veo3/pkg/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInterpolateCommand_FullFlow tests the complete interpolate command flow
func TestInterpolateCommand_FullFlow(t *testing.T) {
	// Skip this test if no API key is available
	if os.Getenv("VEO3_API_KEY") == "" && os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("No API key available for integration test")
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid frame interpolation",
			args: []string{
				"interpolate",
				"testdata/frame1.jpg",
				"testdata/frame2.jpg",
				"--prompt", "Smooth transition between frames",
				"--resolution", "720p",
				"--output", "./test_output",
				"--no-download", // Skip download for faster tests
			},
			wantErr: false,
		},
		{
			name: "interpolation without prompt",
			args: []string{
				"interpolate",
				"testdata/frame1.jpg",
				"testdata/frame2.jpg",
				"--resolution", "720p",
				"--no-download",
			},
			wantErr: false,
		},
		{
			name: "missing first frame argument",
			args: []string{
				"interpolate",
				"--resolution", "720p",
			},
			wantErr: true,
			errMsg:  "accepts 2 arg(s), received",
		},
		{
			name: "same file for both frames",
			args: []string{
				"interpolate",
				"testdata/frame1.jpg",
				"testdata/frame1.jpg",
				"--no-download",
			},
			wantErr: true,
			errMsg:  "cannot be the same file",
		},
		{
			name: "invalid duration for interpolation",
			args: []string{
				"interpolate",
				"testdata/frame1.jpg",
				"testdata/frame2.jpg",
				"--duration", "6", // Should be fixed at 8s
				"--no-download",
			},
			wantErr: true,
			errMsg:  "interpolation requires 8 seconds",
		},
		{
			name: "invalid aspect ratio for interpolation",
			args: []string{
				"interpolate",
				"testdata/frame1.jpg",
				"testdata/frame2.jpg",
				"--aspect-ratio", "9:16", // Should be fixed at 16:9
				"--no-download",
			},
			wantErr: true,
			errMsg:  "interpolation requires 16:9 aspect ratio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for test outputs
			tempDir := t.TempDir()

			// Replace output directory in args
			for i, arg := range tt.args {
				if arg == "./test_output" {
					tt.args[i] = tempDir
				}
			}

			// Capture stdout and stderr
			var stdout, stderr bytes.Buffer

			// Create root command
			rootCmd := cli.NewRootCmd()
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs(tt.args)

			// Execute command
			err := rootCmd.Execute()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err, "stdout: %s, stderr: %s", stdout.String(), stderr.String())

				// Verify output contains operation information
				output := stdout.String()
				assert.Contains(t, output, "Operation", "Expected output to contain operation information")
			}
		})
	}
}

// TestInterpolateCommand_JSONOutput tests JSON output format for interpolation
func TestInterpolateCommand_JSONOutput(t *testing.T) {
	// Skip this test if no API key is available
	if os.Getenv("VEO3_API_KEY") == "" && os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("No API key available for integration test")
	}

	tempDir := t.TempDir()

	var stdout, stderr bytes.Buffer

	rootCmd := cli.NewRootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{
		"interpolate",
		"testdata/frame1.jpg",
		"testdata/frame2.jpg",
		"--prompt", "Smooth morphing transition",
		"--json",
		"--no-download",
		"--output", tempDir,
	})

	err := rootCmd.Execute()
	require.NoError(t, err, "stderr: %s", stderr.String())

	// Verify JSON output
	output := stdout.String()

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err, "Output should be valid JSON: %s", output)

	// Verify JSON structure
	assert.Contains(t, result, "success")
	assert.Contains(t, result, "data")

	if result["success"].(bool) {
		data := result["data"].(map[string]interface{})
		assert.Contains(t, data, "id")
		assert.Contains(t, data, "status")

		// Check interpolation-specific metadata
		if metadata, ok := data["metadata"].(map[string]interface{}); ok {
			assert.Contains(t, metadata, "first_frame_path")
			assert.Contains(t, metadata, "last_frame_path")
		}
	}
}
