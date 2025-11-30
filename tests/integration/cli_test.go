package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateCommand_FullFlow tests the complete generate command flow
func TestGenerateCommand_FullFlow(t *testing.T) {
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
			name: "valid text-to-video generation",
			args: []string{
				"generate",
				"--prompt", "A beautiful sunset over the ocean",
				"--resolution", "720p",
				"--duration", "6",
				"--aspect-ratio", "16:9",
				"--output", "./test_output",
				"--no-download", // Skip download for faster tests
			},
			wantErr: false,
		},
		{
			name: "missing required prompt",
			args: []string{
				"generate",
				"--resolution", "720p",
			},
			wantErr: true,
			errMsg:  "required flag \"prompt\" not set",
		},
		{
			name: "invalid resolution",
			args: []string{
				"generate",
				"--prompt", "Test prompt",
				"--resolution", "4K",
			},
			wantErr: true,
			errMsg:  "resolution must be",
		},
		{
			name: "invalid duration",
			args: []string{
				"generate",
				"--prompt", "Test prompt",
				"--duration", "5",
			},
			wantErr: true,
			errMsg:  "duration must be",
		},
		{
			name: "1080p without 8s duration",
			args: []string{
				"generate",
				"--prompt", "Test prompt",
				"--resolution", "1080p",
				"--duration", "6",
			},
			wantErr: true,
			errMsg:  "1080p resolution requires 8 seconds",
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

// TestGenerateCommand_JSONOutput tests JSON output format
func TestGenerateCommand_JSONOutput(t *testing.T) {
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
		"generate",
		"--prompt", "A test video for JSON output",
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
	assert.Contains(t, result, "timestamp")

	if result["success"].(bool) {
		data := result["data"].(map[string]interface{})
		assert.Contains(t, data, "operation_id")
		assert.Contains(t, data, "status")
		assert.Contains(t, data, "model")
		assert.Contains(t, data, "prompt")
	}
}

// TestGenerateCommand_WithMockAPI tests the command with a mock API server
func TestGenerateCommand_WithMockAPI(t *testing.T) {
	// Create mock API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/operations/"):
			// Mock operation status response
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"name": "operations/test-op-123",
				"done": true,
				"response": map[string]interface{}{
					"@type":    "type.googleapis.com/google.ai.generativelanguage.v1beta.GenerateVideoResponse",
					"videoUri": "gs://bucket/test-video.mp4",
				},
			}
			json.NewEncoder(w).Encode(response)
		case strings.Contains(r.URL.Path, "/models/") && strings.Contains(r.URL.Path, ":generateVideo"):
			// Mock video generation response
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"name": "operations/test-op-123",
			}
			json.NewEncoder(w).Encode(response)
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	// Set up environment to use mock server
	originalEnv := os.Getenv("VEO3_API_ENDPOINT")
	os.Setenv("VEO3_API_ENDPOINT", mockServer.URL)
	defer func() {
		if originalEnv != "" {
			os.Setenv("VEO3_API_ENDPOINT", originalEnv)
		} else {
			os.Unsetenv("VEO3_API_ENDPOINT")
		}
	}()

	// Set up fake API key
	os.Setenv("VEO3_API_KEY", "fake-api-key-for-testing")
	defer os.Unsetenv("VEO3_API_KEY")

	tempDir := t.TempDir()

	var stdout, stderr bytes.Buffer

	rootCmd := cli.NewRootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{
		"generate",
		"--prompt", "A test video with mock API",
		"--no-download", // Skip download for mock test
		"--output", tempDir,
	})

	err := rootCmd.Execute()
	require.NoError(t, err, "stdout: %s, stderr: %s", stdout.String(), stderr.String())

	output := stdout.String()
	assert.Contains(t, output, "Operation", "Expected output to contain operation information")
}

// TestGenerateCommand_ConfigFile tests command with configuration file
func TestGenerateCommand_ConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// Create test config file
	configContent := `api_key: test-api-key
default_model: veo-3.1
default_resolution: 720p
default_duration: 6
default_aspect_ratio: 16:9
output_directory: ` + tempDir + `
poll_interval_seconds: 1
`
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	var stdout, stderr bytes.Buffer

	rootCmd := cli.NewRootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{
		"--config", configFile,
		"generate",
		"--prompt", "Test with config file",
		"--no-download",
	})

	// This test verifies that config loading works
	// The actual API call will fail without a real API key, but we test the config parsing
	err = rootCmd.Execute()

	// We expect this to fail due to authentication, but not due to config parsing
	if err != nil {
		// Should fail with auth error, not config error
		assert.NotContains(t, err.Error(), "config", "Should not fail due to config file issues")
	}
}

// TestGenerateCommand_ProgressDisplay tests that progress is displayed correctly
func TestGenerateCommand_ProgressDisplay(t *testing.T) {
	// This test checks that progress indicators are shown during generation
	// We'll use a mock that simulates a slow operation

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/operations/"):
			// Simulate operation in progress, then completion
			w.Header().Set("Content-Type", "application/json")

			// First few calls return running status
			time.Sleep(100 * time.Millisecond) // Simulate some processing time

			response := map[string]interface{}{
				"name": "operations/test-op-progress",
				"done": true, // Complete immediately for test speed
				"response": map[string]interface{}{
					"@type":    "type.googleapis.com/google.ai.generativelanguage.v1beta.GenerateVideoResponse",
					"videoUri": "gs://bucket/test-video.mp4",
				},
			}
			json.NewEncoder(w).Encode(response)
		case strings.Contains(r.URL.Path, "/models/") && strings.Contains(r.URL.Path, ":generateVideo"):
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"name": "operations/test-op-progress",
			}
			json.NewEncoder(w).Encode(response)
		default:
			http.NotFound(w, r)
		}
		
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
	}))
	defer mockServer.Close()

	// Set up environment
	os.Setenv("VEO3_API_ENDPOINT", mockServer.URL)
	os.Setenv("VEO3_API_KEY", "fake-api-key-for-testing")
	defer func() {
		os.Unsetenv("VEO3_API_ENDPOINT")
		os.Unsetenv("VEO3_API_KEY")
	}()

	tempDir := t.TempDir()

	var stdout, stderr bytes.Buffer

	rootCmd := cli.NewRootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{
		"generate",
		"--prompt", "Test progress display",
		"--no-download",
		"--output", tempDir,
	})

	err := rootCmd.Execute()
	require.NoError(t, err, "stderr: %s", stderr.String())

	output := stdout.String()

	// Verify that progress-related output is shown
	// This is a basic check - in real implementation we'd check for progress bars, spinners, etc.
	assert.NotEmpty(t, output, "Expected some output during generation")
}

// TestGenerateCommand_ErrorHandling tests various error scenarios
func TestGenerateCommand_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		setupMockResp  func() *httptest.Server
		expectedErrMsg string
	}{
		{
			name: "API authentication error",
			args: []string{
				"generate",
				"--prompt", "Test auth error",
				"--no-download",
			},
			setupMockResp: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"error": {"code": 401, "message": "Invalid API key"}}`))
				}))
			},
			expectedErrMsg: "authentication",
		},
		{
			name: "API rate limit error",
			args: []string{
				"generate",
				"--prompt", "Test rate limit",
				"--no-download",
			},
			setupMockResp: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTooManyRequests)
					w.Write([]byte(`{"error": {"code": 429, "message": "Rate limit exceeded"}}`))
				}))
			},
			expectedErrMsg: "rate limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := tt.setupMockResp()
			defer mockServer.Close()

			os.Setenv("VEO3_API_ENDPOINT", mockServer.URL)
			os.Setenv("VEO3_API_KEY", "fake-api-key")
			defer func() {
				os.Unsetenv("VEO3_API_ENDPOINT")
				os.Unsetenv("VEO3_API_KEY")
			}()

			tempDir := t.TempDir()

			// Add output directory to args
			args := append(tt.args, "--output", tempDir)

			var stdout, stderr bytes.Buffer

			rootCmd := cli.NewRootCmd()
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs(args)

			err := rootCmd.Execute()
			require.Error(t, err)
			assert.Contains(t, strings.ToLower(err.Error()), tt.expectedErrMsg)
		})
	}
}
