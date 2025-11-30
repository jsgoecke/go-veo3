package cli

import (
	"context"
	"fmt"

	"github.com/jasongoecke/go-veo3/pkg/config"
	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// interpolateCmd represents the interpolate command
var interpolateCmd = &cobra.Command{
	Use:   "interpolate [first-frame] [last-frame]",
	Short: "Generate video by interpolating between two frames",
	Long: `Generate a video by smoothly transitioning between two input images.

This command takes two images as the first and last frames and creates
a smooth interpolation between them. The interpolation is constrained
to 8 seconds duration and 16:9 aspect ratio as per API requirements.

Supported image formats: JPEG, PNG, WebP
Maximum image size: 20MB each
Images must have identical dimensions.`,
	Example: `  # Interpolate between two frames
  veo3 interpolate start.jpg end.jpg

  # Add a prompt to guide the transition
  veo3 interpolate frame1.png frame2.png --prompt "Smooth morphing transition"

  # Generate 1080p interpolation
  veo3 interpolate start.jpg end.jpg --resolution 1080p

  # Save to specific directory
  veo3 interpolate frame1.jpg frame2.jpg --output ./videos/`,
	Args: cobra.ExactArgs(2),
	RunE: runInterpolate,
}

// newInterpolateCmd creates the interpolate command
func newInterpolateCmd() *cobra.Command {
	// Add flags (note: duration and aspect-ratio are fixed for interpolation)
	interpolateCmd.Flags().StringP("prompt", "p", "", "Optional prompt to guide the transition")
	interpolateCmd.Flags().StringP("resolution", "r", "", "Resolution (720p or 1080p)")
	interpolateCmd.Flags().StringP("model", "m", "", "Model to use (must support interpolation)")
	interpolateCmd.Flags().String("negative-prompt", "", "Negative prompt (elements to exclude)")
	interpolateCmd.Flags().String("output", "", "Output directory for downloaded video")
	interpolateCmd.Flags().String("filename", "", "Custom filename for output video")
	interpolateCmd.Flags().Bool("no-wait", false, "Start generation and return immediately")
	interpolateCmd.Flags().Bool("no-download", false, "Skip automatic video download")
	interpolateCmd.Flags().Bool("pretty", false, "Pretty-print JSON output (with --json)")

	// Note: duration and aspect-ratio are NOT configurable for interpolation
	// They are fixed at 8s and 16:9 respectively

	// Bind flags to viper for config integration
	viper.BindPFlag("model", interpolateCmd.Flags().Lookup("model"))
	viper.BindPFlag("resolution", interpolateCmd.Flags().Lookup("resolution"))
	viper.BindPFlag("output", interpolateCmd.Flags().Lookup("output"))

	return interpolateCmd
}

// runInterpolate handles frame interpolation
func runInterpolate(cmd *cobra.Command, args []string) error {
	firstFramePath := args[0]
	lastFramePath := args[1]

	// Load configuration using manager
	manager := config.NewManager("")
	cfg, err := manager.Load()
	if err != nil {
		// Use defaults if config load fails
		cfg = &config.Configuration{
			DefaultModel:        config.DefaultModel,
			DefaultResolution:   config.DefaultResolution,
			DefaultAspectRatio:  "16:9", // Fixed for interpolation
			DefaultDuration:     8,      // Fixed for interpolation
			OutputDirectory:     ".",
			PollIntervalSeconds: config.DefaultPollInterval,
		}
	}

	// Get flag values with config fallbacks
	prompt, _ := cmd.Flags().GetString("prompt")
	resolution := getStringWithDefault(cmd, "resolution", cfg.DefaultResolution)
	model := getStringWithDefault(cmd, "model", cfg.DefaultModel)
	negativePrompt, _ := cmd.Flags().GetString("negative-prompt")
	outputDir := getStringWithDefault(cmd, "output", cfg.OutputDirectory)
	filename, _ := cmd.Flags().GetString("filename")
	noWait, _ := cmd.Flags().GetBool("no-wait")
	noDownload, _ := cmd.Flags().GetBool("no-download")
	jsonFormat := viper.GetBool("json")
	pretty, _ := cmd.Flags().GetBool("pretty")

	// Create interpolation request with fixed constraints
	request := &veo3.InterpolationRequest{
		GenerationRequest: veo3.GenerationRequest{
			Prompt:           prompt,
			NegativePrompt:   negativePrompt,
			Model:            model,
			AspectRatio:      "16:9", // Fixed for interpolation
			Resolution:       resolution,
			DurationSeconds:  8,  // Fixed for interpolation
			PersonGeneration: "", // Use default
		},
		FirstFramePath: firstFramePath,
		LastFramePath:  lastFramePath,
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Create API client
	apiKey := viper.GetString("api-key")
	if apiKey == "" {
		apiKey = cfg.APIKey
	}

	client, err := veo3.NewClient(context.Background(), apiKey)
	if err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Show upload progress for images
	if !jsonFormat {
		fmt.Printf("⬆ Uploading first frame: %s\n", firstFramePath)

		// Get image info for display
		firstInfo, err := veo3.GetImageInfo(firstFramePath)
		if err == nil {
			sizeMB := float64(firstInfo.SizeBytes) / (1024 * 1024)
			fmt.Printf("📸 %dx%d %s (%.1f MB)\n",
				firstInfo.Width, firstInfo.Height, firstInfo.Format, sizeMB)
		}

		fmt.Printf("⬆ Uploading last frame: %s\n", lastFramePath)

		// Get image info for display
		lastInfo, err := veo3.GetImageInfo(lastFramePath)
		if err == nil {
			sizeMB := float64(lastInfo.SizeBytes) / (1024 * 1024)
			fmt.Printf("📸 %dx%d %s (%.1f MB)\n",
				lastInfo.Width, lastInfo.Height, lastInfo.Format, sizeMB)
		}
	}

	// Submit interpolation request
	ctx := context.Background()
	operation, err := client.InterpolateFrames(ctx, request)
	if err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Handle async mode
	if noWait {
		return outputOperation(operation, jsonFormat, pretty)
	}

	// Poll for completion with progress display
	if !jsonFormat {
		fmt.Printf("🔄 Interpolating frames... (Operation: %s)\n", operation.ID)
	}

	operation, err = pollOperation(ctx, client, operation.ID, jsonFormat)
	if err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Download video if completed and not disabled
	if operation.Status == veo3.StatusDone && !noDownload {
		if !jsonFormat {
			fmt.Println("✓ Interpolation completed!")
		}

		err = downloadVideo(ctx, operation, outputDir, filename, jsonFormat)
		if err != nil {
			return handleError(err, jsonFormat, pretty)
		}
	}

	return outputOperation(operation, jsonFormat, pretty)
}
