package format

import (
	"encoding/json"
	"fmt"

	"github.com/jasongoecke/go-veo3/pkg/operations"
	"github.com/jasongoecke/go-veo3/pkg/veo3"
)

// JSONOutput represents a standardized JSON response format
type JSONOutput struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *JSONError  `json:"error,omitempty"`
}

// JSONError represents error information in JSON format
type JSONError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// FormatOperationJSON formats an operation as JSON
func FormatOperationJSON(op *veo3.Operation) (string, error) {
	output := JSONOutput{
		Success: true,
		Data:    op,
	}

	return marshalJSON(output)
}

// FormatOperationListJSON formats a list of operations as JSON
func FormatOperationListJSON(operations []*veo3.Operation) (string, error) {
	output := JSONOutput{
		Success: true,
		Data: map[string]interface{}{
			"operations": operations,
			"count":      len(operations),
		},
	}

	return marshalJSON(output)
}

// FormatOperationStatsJSON formats operation statistics as JSON
func FormatOperationStatsJSON(stats operations.OperationStats) (string, error) {
	output := JSONOutput{
		Success: true,
		Data:    stats,
	}

	return marshalJSON(output)
}

// FormatModelJSON formats a model as JSON
func FormatModelJSON(model veo3.Model) (string, error) {
	output := JSONOutput{
		Success: true,
		Data:    model,
	}

	return marshalJSON(output)
}

// FormatModelListJSON formats a list of models as JSON
func FormatModelListJSON(models []veo3.Model) (string, error) {
	output := JSONOutput{
		Success: true,
		Data: map[string]interface{}{
			"models": models,
			"count":  len(models),
		},
	}

	return marshalJSON(output)
}

// FormatGeneratedVideoJSON formats generated video metadata as JSON
func FormatGeneratedVideoJSON(video *veo3.GeneratedVideo) (string, error) {
	output := JSONOutput{
		Success: true,
		Data:    video,
	}

	return marshalJSON(output)
}

// FormatConfigJSON formats configuration as JSON (excluding sensitive data)
func FormatConfigJSON(config interface{}) (string, error) {
	output := JSONOutput{
		Success: true,
		Data:    config,
	}

	return marshalJSON(output)
}

// FormatSuccessJSON formats a simple success message as JSON
func FormatSuccessJSON(message string, data interface{}) (string, error) {
	output := JSONOutput{
		Success: true,
		Data: map[string]interface{}{
			"message": message,
			"result":  data,
		},
	}

	return marshalJSON(output)
}

// FormatErrorJSON formats an error as JSON
func FormatErrorJSON(code, message string, details interface{}) (string, error) {
	var detailsStr string
	if details != nil {
		if str, ok := details.(string); ok {
			detailsStr = str
		} else {
			// Convert complex details to JSON string
			if detailsBytes, err := json.Marshal(details); err == nil {
				detailsStr = string(detailsBytes)
			}
		}
	}

	output := JSONOutput{
		Success: false,
		Error: &JSONError{
			Code:    code,
			Message: message,
			Details: detailsStr,
		},
	}

	return marshalJSON(output)
}

// FormatOperationErrorJSON formats a Veo operation error as JSON
func FormatOperationErrorJSON(opErr *veo3.OperationError) (string, error) {
	var detailsStr string
	if opErr.Details != nil {
		if detailsBytes, err := json.Marshal(opErr.Details); err == nil {
			detailsStr = string(detailsBytes)
		}
	}

	return FormatErrorJSON(opErr.Code, opErr.Message, detailsStr)
}

// FormatValidationErrorJSON formats validation errors as JSON
func FormatValidationErrorJSON(field, message string) (string, error) {
	return FormatErrorJSON("VALIDATION_ERROR", message, map[string]string{
		"field": field,
	})
}

// FormatBatchResultJSON formats batch processing results as JSON
func FormatBatchResultJSON(results interface{}, summary interface{}) (string, error) {
	output := JSONOutput{
		Success: true,
		Data: map[string]interface{}{
			"results": results,
			"summary": summary,
		},
	}

	return marshalJSON(output)
}

// FormatProgressJSON formats progress updates as JSON
func FormatProgressJSON(operationID string, status veo3.OperationStatus, progress float64, message string) (string, error) {
	output := JSONOutput{
		Success: true,
		Data: map[string]interface{}{
			"operation_id": operationID,
			"status":       status,
			"progress":     progress,
			"message":      message,
		},
	}

	return marshalJSON(output)
}

// FormatGenericJSON formats any data structure as JSON with success wrapper
func FormatGenericJSON(data interface{}) (string, error) {
	output := JSONOutput{
		Success: true,
		Data:    data,
	}

	return marshalJSON(output)
}

// marshalJSON is a helper function to marshal JSON with consistent formatting
func marshalJSON(v interface{}) (string, error) {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(bytes), nil
}

// marshalJSONCompact marshals JSON without indentation for compact output
func marshalJSONCompact(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(bytes), nil
}

// FormatCompactJSON formats any data structure as compact JSON
func FormatCompactJSON(data interface{}) (string, error) {
	output := JSONOutput{
		Success: true,
		Data:    data,
	}

	return marshalJSONCompact(output)
}

// ParseJSONError parses a JSON error response and returns the error details
func ParseJSONError(jsonStr string) (*JSONError, error) {
	var output JSONOutput
	if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if output.Error == nil {
		return nil, fmt.Errorf("no error found in JSON response")
	}

	return output.Error, nil
}

// IsJSONOutput checks if the given string is a valid JSON output format
func IsJSONOutput(str string) bool {
	var output JSONOutput
	return json.Unmarshal([]byte(str), &output) == nil
}
