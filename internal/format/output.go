package format

import (
	"fmt"
	"strings"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/operations"
	"github.com/jasongoecke/go-veo3/pkg/veo3"
)

// FormatOperation formats an operation for human-readable display
func FormatOperation(op *veo3.Operation) string {
	var output strings.Builder

	// Basic operation info
	output.WriteString(fmt.Sprintf("Operation: %s\n", op.ID))
	output.WriteString(fmt.Sprintf("Status: %s\n", formatStatus(op.Status)))

	// Progress if available
	if op.Progress > 0 {
		output.WriteString(fmt.Sprintf("Progress: %.1f%%\n", op.Progress*100))
	}

	// Timing information
	output.WriteString(fmt.Sprintf("Started: %s\n", op.StartTime.Format(time.RFC3339)))
	if op.EndTime != nil {
		output.WriteString(fmt.Sprintf("Completed: %s\n", op.EndTime.Format(time.RFC3339)))
		duration := op.EndTime.Sub(op.StartTime)
		output.WriteString(fmt.Sprintf("Duration: %s\n", formatDuration(duration)))
	} else {
		elapsed := time.Since(op.StartTime)
		output.WriteString(fmt.Sprintf("Elapsed: %s\n", formatDuration(elapsed)))
	}

	// Video URI if available
	if op.VideoURI != "" {
		output.WriteString(fmt.Sprintf("Video URI: %s\n", op.VideoURI))
	}

	// Error details if failed
	if op.Error != nil {
		output.WriteString(fmt.Sprintf("Error: %s\n", op.Error.Message))
		if op.Error.Suggestion != "" {
			output.WriteString(fmt.Sprintf("Suggestion: %s\n", op.Error.Suggestion))
		}
	}

	// Metadata if available
	if len(op.Metadata) > 0 {
		output.WriteString("Metadata:\n")
		for key, value := range op.Metadata {
			output.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	return output.String()
}

// FormatOperationList formats a list of operations in table format
func FormatOperationList(operations []*veo3.Operation) string {
	if len(operations) == 0 {
		return "No operations found.\n"
	}

	var output strings.Builder

	// Header
	output.WriteString("OPERATION ID                    STATUS      PROGRESS  STARTED              MODEL\n")
	output.WriteString("──────────────────────────────  ──────────  ────────  ───────────────────  ─────────────────────\n")

	for _, op := range operations {
		// Extract short ID (last 8 characters)
		shortID := op.ID
		if len(op.ID) > 30 {
			shortID = "..." + op.ID[len(op.ID)-27:]
		}

		// Format progress
		progress := ""
		if op.Progress > 0 {
			progress = fmt.Sprintf("%.1f%%", op.Progress*100)
		}

		// Format started time
		started := op.StartTime.Format("2006-01-02 15:04:05")

		// Extract model from metadata
		model := ""
		if op.Metadata != nil {
			if m, ok := op.Metadata["model"].(string); ok {
				model = m
			}
		}
		if len(model) > 20 {
			model = model[:17] + "..."
		}

		output.WriteString(fmt.Sprintf("%-30s  %-10s  %-8s  %-19s  %-20s\n",
			shortID, formatStatus(op.Status), progress, started, model))
	}

	return output.String()
}

// FormatOperationStats formats operation statistics
func FormatOperationStats(stats operations.OperationStats) string {
	var output strings.Builder

	output.WriteString("Operation Statistics:\n")
	output.WriteString(fmt.Sprintf("  Total:     %d\n", stats.Total))
	output.WriteString(fmt.Sprintf("  Pending:   %d\n", stats.Pending))
	output.WriteString(fmt.Sprintf("  Running:   %d\n", stats.Running))
	output.WriteString(fmt.Sprintf("  Completed: %d\n", stats.Completed))
	output.WriteString(fmt.Sprintf("  Failed:    %d\n", stats.Failed))
	output.WriteString(fmt.Sprintf("  Cancelled: %d\n", stats.Cancelled))

	return output.String()
}

// FormatModel formats a model for human-readable display
func FormatModel(model veo3.Model) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Model: %s\n", model.Name))
	output.WriteString(fmt.Sprintf("ID: %s\n", model.ID))
	output.WriteString(fmt.Sprintf("Version: %s\n", model.Version))
	output.WriteString(fmt.Sprintf("Tier: %s\n", model.Tier))

	output.WriteString("\nCapabilities:\n")
	output.WriteString(fmt.Sprintf("  Audio: %s\n", formatBool(model.Capabilities.Audio)))
	output.WriteString(fmt.Sprintf("  Extension: %s\n", formatBool(model.Capabilities.Extension)))
	output.WriteString(fmt.Sprintf("  Reference Images: %s\n", formatBool(model.Capabilities.ReferenceImages)))
	output.WriteString(fmt.Sprintf("  Resolutions: %s\n", strings.Join(model.Capabilities.Resolutions, ", ")))
	output.WriteString(fmt.Sprintf("  Durations: %s\n", formatIntSlice(model.Capabilities.Durations)))

	if model.Constraints.MaxReferenceImages > 0 {
		output.WriteString("\nConstraints:\n")
		output.WriteString(fmt.Sprintf("  Max Reference Images: %d\n", model.Constraints.MaxReferenceImages))
		if model.Constraints.RequiredAspectRatio != "" {
			output.WriteString(fmt.Sprintf("  Required Aspect Ratio: %s\n", model.Constraints.RequiredAspectRatio))
		}
		if model.Constraints.RequiredDuration > 0 {
			output.WriteString(fmt.Sprintf("  Required Duration: %ds\n", model.Constraints.RequiredDuration))
		}
	}

	return output.String()
}

// FormatModelList formats a list of models in table format
func FormatModelList(models []veo3.Model) string {
	if len(models) == 0 {
		return "No models available.\n"
	}

	var output strings.Builder

	// Header
	output.WriteString("MODEL NAME                      VERSION  TIER      AUDIO  EXTENSION  REFERENCE\n")
	output.WriteString("──────────────────────────────  ───────  ────────  ─────  ─────────  ─────────\n")

	for _, model := range models {
		name := model.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		output.WriteString(fmt.Sprintf("%-30s  %-7s  %-8s  %-5s  %-9s  %-9s\n",
			name,
			model.Version,
			model.Tier,
			formatBool(model.Capabilities.Audio),
			formatBool(model.Capabilities.Extension),
			formatBool(model.Capabilities.ReferenceImages),
		))
	}

	return output.String()
}

// FormatGeneratedVideo formats generated video metadata
func FormatGeneratedVideo(video *veo3.GeneratedVideo) string {
	var output strings.Builder

	output.WriteString("✅ Video generated successfully!\n")
	output.WriteString(fmt.Sprintf("📁 File: %s\n", video.FilePath))
	output.WriteString(fmt.Sprintf("📊 Duration: %ds | Resolution: %s | Size: %s\n",
		video.DurationSeconds,
		video.Resolution,
		formatFileSize(video.FileSizeBytes)))

	if video.GenerationTimeSeconds > 0 {
		output.WriteString(fmt.Sprintf("⏱️  Generation time: %s\n",
			formatDuration(time.Duration(video.GenerationTimeSeconds)*time.Second)))
	}

	if video.Model != "" {
		output.WriteString(fmt.Sprintf("🤖 Model: %s\n", video.Model))
	}

	if video.Prompt != "" {
		output.WriteString(fmt.Sprintf("💭 Prompt: %s\n", video.Prompt))
	}

	return output.String()
}

// FormatError formats an error for human-readable display
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	var output strings.Builder

	// Check if it's a veo3.OperationError for enhanced formatting
	if opErr, ok := err.(*veo3.OperationError); ok {
		output.WriteString(fmt.Sprintf("❌ %s: %s\n", opErr.Code, opErr.Message))

		if opErr.Suggestion != "" {
			output.WriteString(fmt.Sprintf("\n💡 %s\n", opErr.Suggestion))
		}

		if len(opErr.Details) > 0 {
			output.WriteString("\nDetails:\n")
			for key, value := range opErr.Details {
				output.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
			}
		}
	} else {
		// Generic error formatting
		output.WriteString(fmt.Sprintf("❌ Error: %s\n", err.Error()))
	}

	return output.String()
}

// Helper functions

func formatStatus(status veo3.OperationStatus) string {
	switch status {
	case veo3.StatusPending:
		return "⏳ PENDING"
	case veo3.StatusRunning:
		return "🔄 RUNNING"
	case veo3.StatusDone:
		return "✅ DONE"
	case veo3.StatusFailed:
		return "❌ FAILED"
	case veo3.StatusCancelled:
		return "🚫 CANCELLED"
	default:
		return string(status)
	}
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%d:%02d", minutes, seconds)
	}
	return fmt.Sprintf("0:%02d", seconds)
}

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func formatIntSlice(ints []int) string {
	strs := make([]string, len(ints))
	for i, v := range ints {
		strs[i] = fmt.Sprintf("%ds", v)
	}
	return strings.Join(strs, ", ")
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
