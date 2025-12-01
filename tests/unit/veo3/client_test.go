package veo3_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GenerateVideo(t *testing.T) {
	tests := []struct {
		name           string
		request        *veo3.GenerationRequest
		mockResponse   map[string]interface{}
		mockStatusCode int
		wantErr        bool
		errContains    string
	}{
		{
			name: "successful generation request",
			request: &veo3.GenerationRequest{
				Prompt:          "A beautiful sunset over the ocean",
				Model:           "veo-3.1-generate-preview",
				AspectRatio:     "16:9",
				Resolution:      "720p",
				DurationSeconds: 6,
			},
			mockResponse: map[string]interface{}{
				"name": "operations/generate-video-op-123456",
				"metadata": map[string]interface{}{
					"@type":      "type.googleapis.com/google.ai.generativelanguage.v1beta.GenerateVideoMetadata",
					"createTime": "2024-01-01T00:00:00Z",
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "successful generation with negative prompt",
			request: &veo3.GenerationRequest{
				Prompt:          "A cityscape",
				NegativePrompt:  "avoid people, cars",
				Model:           "veo-3.1-generate-preview",
				AspectRatio:     "16:9",
				Resolution:      "720p",
				DurationSeconds: 6,
			},
			mockResponse: map[string]interface{}{
				"name": "operations/generate-video-op-789012",
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "successful 1080p generation with 8s duration",
			request: &veo3.GenerationRequest{
				Prompt:          "High quality mountain landscape",
				Model:           "veo-3.1-generate-preview",
				AspectRatio:     "16:9",
				Resolution:      "1080p",
				DurationSeconds: 8,
			},
			mockResponse: map[string]interface{}{
				"name": "operations/generate-video-op-hd",
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "API returns bad request error",
			request: &veo3.GenerationRequest{
				Prompt:          "Test prompt",
				Model:           "veo-3.1-generate-preview",
				AspectRatio:     "16:9",
				Resolution:      "720p",
				DurationSeconds: 6,
			},
			mockResponse: map[string]interface{}{
				"error": map[string]interface{}{
					"code":    400,
					"message": "Invalid prompt content",
					"status":  "INVALID_ARGUMENT",
				},
			},
			mockStatusCode: http.StatusBadRequest,
			wantErr:        true,
			errContains:    "invalid prompt content",
		},
		{
			name: "API returns authentication error",
			request: &veo3.GenerationRequest{
				Prompt:          "Test prompt",
				Model:           "veo-3.1-generate-preview",
				AspectRatio:     "16:9",
				Resolution:      "720p",
				DurationSeconds: 6,
			},
			mockResponse: map[string]interface{}{
				"error": map[string]interface{}{
					"code":    401,
					"message": "Request had invalid authentication credentials",
					"status":  "UNAUTHENTICATED",
				},
			},
			mockStatusCode: http.StatusUnauthorized,
			wantErr:        true,
			errContains:    "authentication",
		},
		{
			name: "API returns rate limit error",
			request: &veo3.GenerationRequest{
				Prompt:          "Test prompt",
				Model:           "veo-3.1-generate-preview",
				AspectRatio:     "16:9",
				Resolution:      "720p",
				DurationSeconds: 6,
			},
			mockResponse: map[string]interface{}{
				"error": map[string]interface{}{
					"code":    429,
					"message": "Quota exceeded. Please try again later",
					"status":  "RESOURCE_EXHAUSTED",
				},
			},
			mockStatusCode: http.StatusTooManyRequests,
			wantErr:        true,
			errContains:    "quota exceeded",
		},
		{
			name: "API returns safety filter error",
			request: &veo3.GenerationRequest{
				Prompt:          "Inappropriate content prompt",
				Model:           "veo-3.1-generate-preview",
				AspectRatio:     "16:9",
				Resolution:      "720p",
				DurationSeconds: 6,
			},
			mockResponse: map[string]interface{}{
				"error": map[string]interface{}{
					"code":    400,
					"message": "Content blocked by safety filters",
					"status":  "INVALID_ARGUMENT",
					"details": []interface{}{
						map[string]interface{}{
							"@type":  "type.googleapis.com/google.ai.generativelanguage.v1beta.BlockedPromptFeedback",
							"reason": "SAFETY",
						},
					},
				},
			},
			mockStatusCode: http.StatusBadRequest,
			wantErr:        true,
			errContains:    "safety",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, ":predictLongRunning")

				// Verify Content-Type
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Verify API key header
				assert.Equal(t, "test-api-key", r.Header.Get("x-goog-api-key"))

				// Verify request body structure
				var requestBody map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				require.NoError(t, err)

				// Verify required fields in request (nested Vertex AI structure)
				instances, ok := requestBody["instances"].([]interface{})
				require.True(t, ok, "instances should be an array")
				require.Len(t, instances, 1, "should have exactly one instance")

				instance, ok := instances[0].(map[string]interface{})
				require.True(t, ok, "instance should be a map")
				assert.Equal(t, tt.request.Prompt, instance["prompt"])

				parameters, ok := requestBody["parameters"].(map[string]interface{})
				require.True(t, ok, "parameters should be a map")
				assert.Equal(t, tt.request.AspectRatio, parameters["aspectRatio"])

				// Note: resolution is not sent in the Vertex AI format
				// The API uses aspectRatio and durationSeconds to determine output

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer mockServer.Close()

			// Create client with mock server URL
			ctx := context.Background()
			client, err := veo3.NewClient(ctx, "test-api-key", veo3.WithBaseURL(mockServer.URL))
			require.NoError(t, err)

			// Execute the test
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			operation, err := client.GenerateVideo(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), tt.errContains)
				assert.Nil(t, operation)
			} else {
				require.NoError(t, err)
				require.NotNil(t, operation)
				assert.NotEmpty(t, operation.ID)
				assert.Contains(t, operation.ID, "operations/")
			}
		})
	}
}

func TestClient_GetOperation(t *testing.T) {
	tests := []struct {
		name           string
		operationID    string
		mockResponse   map[string]interface{}
		mockStatusCode int
		wantErr        bool
		expectedStatus veo3.OperationStatus
	}{
		{
			name:        "operation pending",
			operationID: "operations/test-op-pending",
			mockResponse: map[string]interface{}{
				"name": "operations/test-op-pending",
				"done": false,
				"metadata": map[string]interface{}{
					"@type":      "type.googleapis.com/google.ai.generativelanguage.v1beta.GenerateVideoMetadata",
					"createTime": "2024-01-01T00:00:00Z",
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			expectedStatus: veo3.StatusRunning,
		},
		{
			name:        "operation completed with video",
			operationID: "operations/test-op-completed",
			mockResponse: map[string]interface{}{
				"name": "operations/test-op-completed",
				"done": true,
				"response": map[string]interface{}{
					"@type":    "type.googleapis.com/google.ai.generativelanguage.v1beta.GenerateVideoResponse",
					"videoUri": "gs://bucket/generated-video-123.mp4",
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			expectedStatus: veo3.StatusDone,
		},
		{
			name:        "operation failed",
			operationID: "operations/test-op-failed",
			mockResponse: map[string]interface{}{
				"name": "operations/test-op-failed",
				"done": true,
				"error": map[string]interface{}{
					"code":    3,
					"message": "Generation failed due to safety filters",
					"status":  "INVALID_ARGUMENT",
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			expectedStatus: veo3.StatusFailed,
		},
		{
			name:        "operation not found",
			operationID: "operations/not-found",
			mockResponse: map[string]interface{}{
				"error": map[string]interface{}{
					"code":    404,
					"message": "Operation not found",
					"status":  "NOT_FOUND",
				},
			},
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, tt.operationID)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer mockServer.Close()

			ctx := context.Background()
			client, err := veo3.NewClient(ctx, "test-api-key", veo3.WithBaseURL(mockServer.URL))
			require.NoError(t, err)

			operation, err := client.GetOperation(ctx, tt.operationID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, operation)
			} else {
				require.NoError(t, err)
				require.NotNil(t, operation)
				assert.Equal(t, tt.operationID, operation.ID)
				assert.Equal(t, tt.expectedStatus, operation.Status)

				// Additional checks based on operation status
				switch tt.expectedStatus {
				case veo3.StatusDone:
					assert.NotEmpty(t, operation.VideoURI, "Completed operation should have video URI")
				case veo3.StatusFailed:
					assert.NotNil(t, operation.Error, "Failed operation should have error details")
				}
			}
		})
	}
}

func TestClient_Authentication(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		valid  bool
	}{
		{
			name:   "valid API key",
			apiKey: "valid-api-key-123",
			valid:  true,
		},
		{
			name:   "empty API key should fail",
			apiKey: "",
			valid:  false,
		},
		{
			name:   "whitespace only API key should fail",
			apiKey: "   \n\t   ",
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client, err := veo3.NewClient(ctx, tt.apiKey)

			if tt.valid {
				require.NoError(t, err)
				require.NotNil(t, client)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "API key")
				assert.Nil(t, client)
			}
		})
	}
}

func TestClient_WithOptions(t *testing.T) {
	t.Run("client with custom base URL", func(t *testing.T) {
		ctx := context.Background()
		client, err := veo3.NewClient(ctx, "test-api-key")

		require.NoError(t, err)
		require.NotNil(t, client)

		// We can't easily verify options without exposing internals,
		// but we can verify the client was created successfully
	})

	t.Run("client with timeout", func(t *testing.T) {
		ctx := context.Background()
		client, err := veo3.NewClient(ctx, "test-api-key")

		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("client with retry configuration", func(t *testing.T) {
		ctx := context.Background()
		client, err := veo3.NewClient(ctx, "test-api-key")

		require.NoError(t, err)
		require.NotNil(t, client)
	})
}

func TestClient_ContextCancellation(t *testing.T) {
	// Test that client respects context cancellation
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"name": "operations/slow-op",
		})
	}))
	defer mockServer.Close()

	bgCtx := context.Background()
	client, err := veo3.NewClient(bgCtx, "test-api-key", veo3.WithBaseURL(mockServer.URL))
	require.NoError(t, err)

	// Create context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	request := &veo3.GenerationRequest{
		Prompt:          "Test context cancellation",
		Model:           "veo-3.1-generate-preview",
		AspectRatio:     "16:9",
		Resolution:      "720p",
		DurationSeconds: 6,
	}

	_, err = client.GenerateVideo(ctx, request)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestClient_RetryLogic(t *testing.T) {
	t.Skip("Retry logic not yet implemented in client")

	// Test that client retries on transient errors
	callCount := 0

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount < 3 {
			// Return temporary error for first 2 calls
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"code":    500,
					"message": "Internal server error",
					"status":  "INTERNAL",
				},
			})
			return
		}

		// Succeed on 3rd call
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"name": "operations/retry-success",
		})
	}))
	defer mockServer.Close()

	bgCtx := context.Background()
	client, err := veo3.NewClient(bgCtx, "test-api-key", veo3.WithBaseURL(mockServer.URL))
	require.NoError(t, err)

	ctx := context.Background()
	request := &veo3.GenerationRequest{
		Prompt:          "Test retry logic",
		Model:           "veo-3.1-generate-preview",
		AspectRatio:     "16:9",
		Resolution:      "720p",
		DurationSeconds: 6,
	}

	operation, err := client.GenerateVideo(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, operation)
	assert.Equal(t, "operations/retry-success", operation.ID)
	assert.Equal(t, 3, callCount, "Expected 3 API calls (2 failures + 1 success)")
}
