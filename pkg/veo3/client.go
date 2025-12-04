package veo3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"
)

// Client handles interaction with the Veo API
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithBaseURL sets a custom base URL for the client (useful for testing)
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// NewClient creates a new Veo API client
func NewClient(ctx context.Context, apiKey string, opts ...ClientOption) (*Client, error) {
	// Trim whitespace and validate
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	client := &Client{
		APIKey:     apiKey,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// CancelOperation cancels a running operation
func (c *Client) CancelOperation(ctx context.Context, operationID string) error {
	// TODO: Implement actual API call to cancel operation
	// For now, this is a placeholder
	return fmt.Errorf("CancelOperation not yet implemented")
}

// ListOperations retrieves operations from the API
func (c *Client) ListOperations(ctx context.Context, filter OperationStatus) ([]*Operation, error) {
	// TODO: Implement actual API call to list operations
	// For now, this is a placeholder
	return nil, fmt.Errorf("ListOperations not yet implemented")
}

// GetOperation retrieves an operation's current status from the API
func (c *Client) GetOperation(ctx context.Context, operationID string) (*Operation, error) {
	if operationID == "" {
		return nil, fmt.Errorf("operation ID cannot be empty")
	}

	// Build URL - operationID should be like "operations/generate/abc123"
	url := fmt.Sprintf("%s/%s", c.BaseURL, operationID)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key header
	req.Header.Set("x-goog-api-key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Debug logging if VEO3_DEBUG environment variable is set
	if os.Getenv("VEO3_DEBUG") != "" {
		log.Printf("[DEBUG] GetOperation response body: %s", string(body))
	}

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp.StatusCode, body)
	}

	// Parse response with flexible structure to handle various API response formats
	var apiResp struct {
		Name     string `json:"name"`
		Done     bool   `json:"done"`
		Metadata struct {
			Type            string  `json:"@type"`
			State           string  `json:"state"`
			ProgressPercent float64 `json:"progressPercent"`
		} `json:"metadata"`
		Response *struct {
			Type     string `json:"@type"`
			VideoURI string `json:"videoUri"`  // Format: videoUri (camelCase)
			VideoUri string `json:"video_uri"` // Format: video_uri (snake_case)
			Videos   []struct {
				URI      string `json:"uri"`
				Uri      string `json:"Uri"` // Capitalized variant
				MimeType string `json:"mimeType"`
			} `json:"videos"` // Array format
			Video *struct {
				URI      string `json:"uri"`
				MimeType string `json:"mimeType"`
			} `json:"video"` // Single video object format
		} `json:"response,omitempty"`
		Error *struct {
			Code    int                      `json:"code"`
			Message string                   `json:"message"`
			Status  string                   `json:"status"`
			Details []map[string]interface{} `json:"details"`
		} `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Also parse as a generic map for fallback extraction
	var genericResp map[string]interface{}
	_ = json.Unmarshal(body, &genericResp)

	// Map API response to Operation
	op := &Operation{
		ID:        apiResp.Name,
		StartTime: time.Now(), // TODO: Parse from metadata if available
		Progress:  apiResp.Metadata.ProgressPercent,
		Metadata:  make(map[string]interface{}),
	}

	// Determine status based on done flag and presence of response/error
	if !apiResp.Done {
		// Operation is still in progress
		if apiResp.Metadata.State == "PENDING" {
			op.Status = StatusPending
		} else {
			op.Status = StatusRunning
		}
	} else {
		// Operation is complete
		if apiResp.Error != nil {
			// Failed operation
			op.Status = StatusFailed
			op.Error = &OperationError{
				Code:    fmt.Sprintf("%d", apiResp.Error.Code),
				Message: apiResp.Error.Message,
			}
			now := time.Now()
			op.EndTime = &now
		} else if apiResp.Response != nil {
			// Successful operation
			op.Status = StatusDone

			// Extract video URI - support multiple response formats
			videoURI := extractVideoURI(apiResp.Response, genericResp)

			if videoURI != "" {
				op.VideoURI = videoURI
			} else if os.Getenv("VEO3_DEBUG") != "" {
				// Log response structure for debugging when URI extraction fails
				log.Printf("[DEBUG] Failed to extract video URI from response. Response structure: %+v", apiResp.Response)
				if resp, ok := genericResp["response"].(map[string]interface{}); ok {
					log.Printf("[DEBUG] Generic response map: %+v", resp)
				}
			}

			now := time.Now()
			op.EndTime = &now
		} else {
			// Done but no response or error (shouldn't happen)
			op.Status = StatusDone
			now := time.Now()
			op.EndTime = &now
		}
	}

	return op, nil
}

// GetOperationStatus retrieves the current status of an operation (alias for GetOperation)
func (c *Client) GetOperationStatus(ctx context.Context, operationID string) (*Operation, error) {
	return c.GetOperation(ctx, operationID)
}

// GenerateVideo generates a video from a text prompt
func (c *Client) GenerateVideo(ctx context.Context, request *GenerationRequest) (*Operation, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Build API payload with correct Vertex AI structure
	// Based on: https://docs.cloud.google.com/vertex-ai/generative-ai/docs/model-reference/veo-video-generation
	parameters := map[string]interface{}{
		"aspectRatio":     request.AspectRatio,
		"durationSeconds": request.DurationSeconds, // INTEGER, not string!
	}

	// Add optional parameters
	if request.NegativePrompt != "" {
		parameters["negativePrompt"] = request.NegativePrompt
	}
	if request.Seed != nil {
		parameters["seed"] = *request.Seed
	}
	if request.PersonGeneration != "" {
		parameters["personGeneration"] = request.PersonGeneration
	}

	// Build full payload with instances and parameters
	payload := map[string]interface{}{
		"instances": []map[string]interface{}{
			{
				"prompt": request.Prompt,
			},
		},
		"parameters": parameters,
	}

	// Marshal payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Build URL
	url := fmt.Sprintf("%s/models/%s:predictLongRunning", c.BaseURL, request.Model)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("x-goog-api-key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp.StatusCode, body)
	}

	// Parse response
	var apiResp struct {
		Name     string `json:"name"`
		Metadata struct {
			Type  string `json:"@type"`
			State string `json:"state"`
		} `json:"metadata"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Create operation
	op := &Operation{
		ID:        apiResp.Name,
		Status:    StatusPending,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Map initial state
	if apiResp.Metadata.State == "RUNNING" {
		op.Status = StatusRunning
	}

	return op, nil
}

// parseErrorResponse parses API error responses
func parseErrorResponse(statusCode int, body []byte) error {
	var errResp struct {
		Error struct {
			Code    int                      `json:"code"`
			Message string                   `json:"message"`
			Status  string                   `json:"status"`
			Details []map[string]interface{} `json:"details"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("HTTP %d: %s", statusCode, string(body))
	}

	return &OperationError{
		Code:    errResp.Error.Status,
		Message: errResp.Error.Message,
		Details: map[string]interface{}{
			"http_code": statusCode,
		},
	}
}

// extractVideoURI attempts to extract video URI from various response formats
func extractVideoURI(response *struct {
	Type     string `json:"@type"`
	VideoURI string `json:"videoUri"`
	VideoUri string `json:"video_uri"`
	Videos   []struct {
		URI      string `json:"uri"`
		Uri      string `json:"Uri"`
		MimeType string `json:"mimeType"`
	} `json:"videos"`
	Video *struct {
		URI      string `json:"uri"`
		MimeType string `json:"mimeType"`
	} `json:"video"`
}, genericResp map[string]interface{}) string {
	// Try structured formats first
	if uri := extractFromStructured(response); uri != "" {
		return uri
	}

	// Fallback: Try extracting from generic map for unknown formats
	if genericResp != nil {
		if resp, ok := genericResp["response"].(map[string]interface{}); ok {
			return extractFromGenericMap(resp)
		}
	}

	return ""
}

// extractFromStructured extracts URI from structured response formats
func extractFromStructured(response *struct {
	Type     string `json:"@type"`
	VideoURI string `json:"videoUri"`
	VideoUri string `json:"video_uri"`
	Videos   []struct {
		URI      string `json:"uri"`
		Uri      string `json:"Uri"`
		MimeType string `json:"mimeType"`
	} `json:"videos"`
	Video *struct {
		URI      string `json:"uri"`
		MimeType string `json:"mimeType"`
	} `json:"video"`
}) string {
	// Format 1: Direct videoUri field (camelCase)
	if response.VideoURI != "" {
		return response.VideoURI
	}

	// Format 2: Direct video_uri field (snake_case)
	if response.VideoUri != "" {
		return response.VideoUri
	}

	// Format 3: videos array with uri field
	if len(response.Videos) > 0 {
		if response.Videos[0].URI != "" {
			return response.Videos[0].URI
		}
		if response.Videos[0].Uri != "" {
			return response.Videos[0].Uri
		}
	}

	// Format 4: Single video object
	if response.Video != nil && response.Video.URI != "" {
		return response.Video.URI
	}

	return ""
}

// extractFromGenericMap extracts URI from generic map for unknown formats
func extractFromGenericMap(resp map[string]interface{}) string {
	// Format 5: Nested generateVideoResponse format
	if uri := extractFromGenerateVideoResponse(resp); uri != "" {
		return uri
	}

	// Try direct field name patterns
	if uri := extractFromDirectFields(resp); uri != "" {
		return uri
	}

	// Try videos array
	if uri := extractFromVideosArray(resp); uri != "" {
		return uri
	}

	// Try single video object
	if uri := extractFromVideoObject(resp); uri != "" {
		return uri
	}

	return ""
}

// extractFromGenerateVideoResponse extracts URI from generateVideoResponse.generatedSamples[0].video.uri
func extractFromGenerateVideoResponse(resp map[string]interface{}) string {
	genVideoResp, ok := resp["generateVideoResponse"].(map[string]interface{})
	if !ok {
		return ""
	}

	samples, ok := genVideoResp["generatedSamples"].([]interface{})
	if !ok || len(samples) == 0 {
		return ""
	}

	sample, ok := samples[0].(map[string]interface{})
	if !ok {
		return ""
	}

	video, ok := sample["video"].(map[string]interface{})
	if !ok {
		return ""
	}

	if uri, ok := video["uri"].(string); ok && uri != "" {
		return uri
	}

	return ""
}

// extractFromDirectFields tries various direct field name patterns
func extractFromDirectFields(resp map[string]interface{}) string {
	fieldNames := []string{"videoUri", "video_uri", "VideoUri", "VideoURI", "uri"}
	for _, field := range fieldNames {
		if uri, ok := resp[field].(string); ok && uri != "" {
			return uri
		}
	}
	return ""
}

// extractFromVideosArray extracts URI from videos array
func extractFromVideosArray(resp map[string]interface{}) string {
	videos, ok := resp["videos"].([]interface{})
	if !ok || len(videos) == 0 {
		return ""
	}

	video, ok := videos[0].(map[string]interface{})
	if !ok {
		return ""
	}

	if uri, ok := video["uri"].(string); ok && uri != "" {
		return uri
	}
	if uri, ok := video["Uri"].(string); ok && uri != "" {
		return uri
	}

	return ""
}

// extractFromVideoObject extracts URI from single video object
func extractFromVideoObject(resp map[string]interface{}) string {
	video, ok := resp["video"].(map[string]interface{})
	if !ok {
		return ""
	}

	if uri, ok := video["uri"].(string); ok && uri != "" {
		return uri
	}

	return ""
}

// generateID generates a random ID for operations
func generateID() string {
	// Simple implementation for now, using nanoseconds
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
