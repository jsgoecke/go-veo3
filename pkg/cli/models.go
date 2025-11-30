package cli

import (
	"fmt"
	"strings"

	"github.com/jasongoecke/go-veo3/internal/format"
	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// modelsCmd represents the models command group
var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available models and view capabilities",
	Long: `List available Veo models and view their capabilities and constraints.

This command group helps you discover available models, understand their
capabilities (audio, extension, reference images), and check compatibility
constraints for different generation types.`,
	Example: `  # List all available models
  veo3 models list

  # Get detailed information about a specific model
  veo3 models info veo-3.1

  # List models that support specific features
  veo3 models list --feature extension
  veo3 models list --feature reference-images`,
}

// newModelsCmd creates the models command group
func newModelsCmd() *cobra.Command {
	// Add subcommands
	modelsCmd.AddCommand(newModelsListCmd())
	modelsCmd.AddCommand(newModelsInfoCmd())

	return modelsCmd
}

// newModelsListCmd creates the 'models list' command
func newModelsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available models",
		Long: `List all available Veo models with their basic information.

Models are shown with their name, version, tier (standard/fast), and key
capabilities like audio support, extension support, and reference image support.`,
		Example: `  # List all models
  veo3 models list

  # List models in JSON format
  veo3 models list --json

  # List only models that support audio
  veo3 models list --feature audio

  # List only Veo 3.1 models
  veo3 models list --version 3.1`,
		RunE: runModelsList,
	}

	cmd.Flags().String("feature", "", "Filter by feature (audio, extension, reference-images)")
	cmd.Flags().String("version", "", "Filter by version (3.1, 3.0, 2.0)")
	cmd.Flags().String("tier", "", "Filter by tier (standard, fast)")
	cmd.Flags().Bool("pretty", false, "Pretty-print JSON output (with --json)")

	return cmd
}

// newModelsInfoCmd creates the 'models info' command
func newModelsInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [model-id]",
		Short: "Get detailed information about a specific model",
		Long: `Get comprehensive information about a specific Veo model.

This shows all capabilities, constraints, supported resolutions, durations,
and any special requirements for the model.`,
		Example: `  # Get model information
  veo3 models info veo-3.1

  # Get information in JSON format
  veo3 models info veo-3.1-generate-preview --json

  # Get capabilities summary
  veo3 models info veo-3.1 --capabilities-only`,
		Args: cobra.ExactArgs(1),
		RunE: runModelsInfo,
	}

	cmd.Flags().Bool("capabilities-only", false, "Show only capabilities, not constraints")
	cmd.Flags().Bool("pretty", false, "Pretty-print JSON output (with --json)")

	return cmd
}

// Command implementations

func runModelsList(cmd *cobra.Command, args []string) error {
	featureFilter, _ := cmd.Flags().GetString("feature")
	versionFilter, _ := cmd.Flags().GetString("version")
	tierFilter, _ := cmd.Flags().GetString("tier")
	jsonFormat := viper.GetBool("json")
	// pretty, _ := cmd.Flags().GetBool("pretty") // TODO: Use for formatted output

	// Get all models
	allModels := veo3.ListModels()

	// Apply filters
	var filteredModels []veo3.Model
	for _, model := range allModels {
		// Feature filter
		if featureFilter != "" {
			switch strings.ToLower(featureFilter) {
			case "audio":
				if !model.Capabilities.Audio {
					continue
				}
			case "extension":
				if !model.Capabilities.Extension {
					continue
				}
			case "reference-images":
				if !model.Capabilities.ReferenceImages {
					continue
				}
			default:
				return fmt.Errorf("invalid feature filter: %s (valid: audio, extension, reference-images)", featureFilter)
			}
		}

		// Version filter
		if versionFilter != "" {
			if model.Version != versionFilter {
				continue
			}
		}

		// Tier filter
		if tierFilter != "" {
			if model.Tier != tierFilter {
				continue
			}
		}

		filteredModels = append(filteredModels, model)
	}

	// Output models
	if jsonFormat {
		jsonOutput, err := format.FormatModelListJSON(filteredModels)
		if err != nil {
			return err
		}
		fmt.Println(jsonOutput)
	} else {
		output := format.FormatModelList(filteredModels)
		fmt.Print(output)
	}

	return nil
}

func runModelsInfo(cmd *cobra.Command, args []string) error {
	modelID := args[0]

	capabilitiesOnly, _ := cmd.Flags().GetBool("capabilities-only")
	jsonFormat := viper.GetBool("json")
	// pretty, _ := cmd.Flags().GetBool("pretty") // TODO: Use for formatted output

	// Get model
	model, exists := veo3.GetModel(modelID)
	if !exists {
		return fmt.Errorf("model not found: %s", modelID)
	}

	// Output model information
	if jsonFormat {
		if capabilitiesOnly {
			jsonOutput, err := format.FormatGenericJSON(model.Capabilities)
			if err != nil {
				return err
			}
			fmt.Println(jsonOutput)
		} else {
			jsonOutput, err := format.FormatModelJSON(model)
			if err != nil {
				return err
			}
			fmt.Println(jsonOutput)
		}
	} else {
		if capabilitiesOnly {
			// Show only capabilities
			fmt.Printf("Model: %s\n", model.Name)
			fmt.Printf("Capabilities:\n")
			fmt.Printf("  Audio: %s\n", formatBool(model.Capabilities.Audio))
			fmt.Printf("  Extension: %s\n", formatBool(model.Capabilities.Extension))
			fmt.Printf("  Reference Images: %s\n", formatBool(model.Capabilities.ReferenceImages))
			fmt.Printf("  Resolutions: %s\n", strings.Join(model.Capabilities.Resolutions, ", "))
			fmt.Printf("  Durations: %s\n", formatIntSlice(model.Capabilities.Durations))
		} else {
			// Show full model information
			output := format.FormatModel(model)
			fmt.Print(output)
		}
	}

	return nil
}

// Helper functions

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
