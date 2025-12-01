package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/spf13/cobra"
)

// newModelsCmd creates the models command group
func newModelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "List and inspect available Veo models",
		Long: `List available Veo models and view detailed information about their capabilities.

Use this to discover which models are available, what features they support,
and their constraints.`,
	}

	cmd.AddCommand(newModelsListCmd())
	cmd.AddCommand(newModelsInfoCmd())

	return cmd
}

// newModelsListCmd creates the models list subcommand
func newModelsListCmd() *cobra.Command {
	var jsonFormat bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available Veo models",
		Long: `List all available Veo models with their key capabilities.

Shows model ID, name, version, and supported features like audio,
video extension, and reference images.`,
		Example: `  # List all models
  veo3 models list

  # List models in JSON format
  veo3 models list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModelsList(cmd, jsonFormat)
		},
	}

	cmd.Flags().BoolVar(&jsonFormat, "json", false, "Output in JSON format")

	return cmd
}

// newModelsInfoCmd creates the models info subcommand
func newModelsInfoCmd() *cobra.Command {
	var jsonFormat bool

	cmd := &cobra.Command{
		Use:   "info <model-id>",
		Short: "Show detailed information about a specific model",
		Long: `Display detailed specifications for a specific Veo model.

Shows complete information about model capabilities, constraints,
supported resolutions, durations, and other technical details.`,
		Example: `  # Get info about a specific model
  veo3 models info veo-3.1-generate-preview

  # Get model info in JSON format
  veo3 models info veo-3.1-generate-preview --json`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("model ID is required\n\nUse 'veo3 models list' to see available models")
			}
			if len(args) > 1 {
				return fmt.Errorf("only one model ID is accepted")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModelsInfo(cmd, args[0], jsonFormat)
		},
	}

	cmd.Flags().BoolVar(&jsonFormat, "json", false, "Output in JSON format")

	return cmd
}

// runModelsList lists all available models
func runModelsList(cmd *cobra.Command, jsonFormat bool) error {
	models := veo3.ListModels()

	if jsonFormat {
		output := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"models": models,
				"count":  len(models),
			},
		}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		cmd.Println(string(data))
		return nil
	}

	// Human-readable table output
	cmd.Println("Available Veo Models:")
	cmd.Println()
	cmd.Printf("%-35s %-15s %-8s %s\n", "MODEL ID", "NAME", "VERSION", "CAPABILITIES")
	cmd.Println(strings.Repeat("-", 100))

	for _, model := range models {
		capabilities := []string{}
		if model.Capabilities.Audio {
			capabilities = append(capabilities, "Audio")
		}
		if model.Capabilities.Extension {
			capabilities = append(capabilities, "Extension")
		}
		if model.Capabilities.ReferenceImages {
			capabilities = append(capabilities, "Reference Images")
		}
		if len(capabilities) == 0 {
			capabilities = append(capabilities, "Basic")
		}

		cmd.Printf("%-35s %-15s %-8s %s\n",
			model.ID,
			model.Name,
			model.Version,
			strings.Join(capabilities, ", "))
	}

	cmd.Println()
	cmd.Printf("Total: %d models\n", len(models))
	cmd.Println()
	cmd.Println("Use 'veo3 models info <model-id>' for detailed information about a specific model")

	return nil
}

// runModelsInfo displays detailed information about a specific model
func runModelsInfo(cmd *cobra.Command, modelID string, jsonFormat bool) error {
	model, found := veo3.GetModel(modelID)
	if !found {
		return fmt.Errorf("model not found: %s\n\nUse 'veo3 models list' to see available models", modelID)
	}

	if jsonFormat {
		output := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"model": model,
			},
		}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		cmd.Println(string(data))
		return nil
	}

	// Human-readable detailed output
	cmd.Println()
	cmd.Printf("Model: %s\n", model.Name)
	cmd.Printf("ID: %s\n", model.ID)
	cmd.Printf("Version: %s\n", model.Version)
	cmd.Printf("Tier: %s\n", model.Tier)
	cmd.Println()

	// Capabilities
	cmd.Println("Capabilities:")
	cmd.Printf("  Audio Generation:        %s\n", formatBool(model.Capabilities.Audio))
	cmd.Printf("  Video Extension:         %s\n", formatBool(model.Capabilities.Extension))
	cmd.Printf("  Reference Images:        %s\n", formatBool(model.Capabilities.ReferenceImages))
	cmd.Println()

	// Supported configurations
	cmd.Println("Supported Configurations:")
	cmd.Printf("  Resolutions:             %s\n", strings.Join(model.Capabilities.Resolutions, ", "))
	cmd.Printf("  Durations:               %s\n", formatDurations(model.Capabilities.Durations))
	cmd.Println()

	// Constraints
	if model.Capabilities.ReferenceImages {
		cmd.Println("Constraints:")
		cmd.Printf("  Max Reference Images:    %d\n", model.Constraints.MaxReferenceImages)
		if model.Constraints.RequiredAspectRatio != "" {
			cmd.Printf("  Required Aspect Ratio:   %s\n", model.Constraints.RequiredAspectRatio)
		}
		if model.Constraints.RequiredDuration > 0 {
			cmd.Printf("  Required Duration:       %ds\n", model.Constraints.RequiredDuration)
		}
		cmd.Println()
	}

	// Usage notes
	cmd.Println("Usage Notes:")
	if model.Capabilities.Audio {
		cmd.Println("  • Supports native audio generation with sound effects and ambience")
	}
	if model.Capabilities.Extension {
		cmd.Println("  • Can extend previously generated videos up to 7 seconds")
	}
	if model.Capabilities.ReferenceImages {
		cmd.Printf("  • Supports up to %d reference images for style/content guidance\n", model.Constraints.MaxReferenceImages)
		cmd.Println("  • Reference images require 8s duration and 16:9 aspect ratio")
	}
	if !model.Capabilities.Audio && !model.Capabilities.Extension && !model.Capabilities.ReferenceImages {
		cmd.Println("  • Basic text-to-video generation")
	}
	cmd.Println()

	return nil
}

// formatBool formats a boolean as a checkmark or X
func formatBool(b bool) string {
	if b {
		return "✓ Yes"
	}
	return "✗ No"
}

// formatDurations formats a list of durations with "seconds" suffix
func formatDurations(durations []int) string {
	parts := make([]string, len(durations))
	for i, d := range durations {
		parts[i] = fmt.Sprintf("%ds", d)
	}
	return strings.Join(parts, ", ")
}
