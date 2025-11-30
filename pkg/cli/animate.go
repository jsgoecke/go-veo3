package cli

import (
	"context"
	"fmt"

	"github.com/jasongoecke/go-veo3/pkg/config"
	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// animateCmd represents the animate command
var animateCmd = &cobra.Command{
	Use:   "animate [image-path]",
	Short: "Generate video from image",
	Long: `Generate a video from an input image using Google Veo 3.1.

This command takes a static image and animates it into a video.
The image is used as the first frame of the generated video.

Supported image formats: JPEG, PNG, WebP
Maximum image size: 20MB`,
	Example: `  # Animate image with default settings
  veo3 animate image.jpg

  # Animate with custom prompt
  veo3 animate image.png --prompt "Add subtle motion to the water"

  # Generate 1080p video (requires 8s duration)
  veo3 animate image.jpg --resolution 1080p --duration 8

  # Save to specific directory
  veo3 animate image.jpg --output ./videos/`,
	Args: cobra.ExactArgs(1),
	RunE: runAnimate,
}

// newAnimateCmd creates the animate command
func newAnimateCmd() *cobra.Command {
	// Add flags
	animateCmd.Flags().StringP("prompt", "p", "", "Optional prompt to enhance the animation")
	animateCmd.Flags().StringP("resolution", "r", "", "Resolution (720p or 1080p)")
	animateCmd.Flags().IntP("duration", "d", 0, "Duration in seconds (4, 6, or 8)")
	animateCmd.Flags().StringP("aspect-ratio", "a", "", "Aspect ratio (16:9 or 9:16)")
	animateCmd.Flags().StringP("model", "m", "", "Model to use")
	animateCmd.Flags().String("negative-prompt", "", "Negative prompt (elements to exclude)")
	animateCmd.Flags().String("output", "", "Output directory for downloaded video")
	animateCmd.Flags().String("filename", "", "Custom filename for output video")
	animateCmd.Flags().Bool("no-wait", false, "Start generation and return immediately")
	animateCmd.Flags().Bool("no-download", false, "Skip automatic video download")
	animateCmd.Flags().Bool("pretty", false, "Pretty-print JSON output (with --json)")

	// Bind flags to viper for config integration
	viper.BindPFlag("model", animateCmd.Flags().Lookup("model"))
	viper.BindPFlag("resolution", animateCmd.Flags().Lookup("resolution"))
	viper.BindPFlag("duration", animateCmd.Flags().Lookup("duration"))
	viper.BindPFlag("aspect-ratio", animateCmd.Flags().Lookup("aspect-ratio"))
	viper.BindPFlag("output", animateCmd.Flags().Lookup("output"))

	return animateCmd
}

// runAnimate handles image-to-video generation
func runAnimate(cmd *cobra.Command, args []string) error {
	imagePath := args[0]

	// Load configuration using manager
	manager := config.NewManager("")
	cfg, err := manager.Load()
	if err != nil {
		// Use defaults if config load fails
		cfg = &config.Configuration{
			DefaultModel:        config.DefaultModel,
			DefaultResolution:   config.DefaultResolution,
			DefaultAspectRatio:  config.DefaultAspectRatio,
			DefaultDuration:     config.DefaultDuration,
			OutputDirectory:     ".",
			PollIntervalSeconds: config.DefaultPollInterval,
		}
	}

	// Get flag values with config fallbacks
	prompt, _ := cmd.Flags().GetString("prompt")
	resolution := getStringWithDefault(cmd, "resolution", cfg.DefaultResolution)
	duration := getIntWithDefault(cmd, "duration", cfg.DefaultDuration)
	aspectRatio := getStringWithDefault(cmd, "aspect-ratio", cfg.DefaultAspectRatio)
	model := getStringWithDefault(cmd, "model", cfg.DefaultModel)
	negativePrompt, _ := cmd.Flags().GetString("negative-prompt")
	outputDir := getStringWithDefault(cmd, "output", cfg.OutputDirectory)
	filename, _ := cmd.Flags().GetString("filename")
	noWait, _ := cmd.Flags().GetBool("no-wait")
	noDownload, _ := cmd.Flags().GetBool("no-download")
	jsonFormat := viper.GetBool("json")
	pretty, _ := cmd.Flags().GetBool("pretty")

	// Create image request
	request := &veo3.ImageRequest{
		GenerationRequest: veo3.GenerationRequest{
			Prompt:           prompt,
			NegativePrompt:   negativePrompt,
			Model:            model,
			AspectRatio:      aspectRatio,
			Resolution:       resolution,
			DurationSeconds:  duration,
			PersonGeneration: "", // Use default
		},
		ImagePath: imagePath,
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

	client, err := veo3.NewClient(apiKey)
	if err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Show upload progress for image
	if !jsonFormat {
		fmt.Printf("⬆ Uploading image: %s\n", imagePath)

		// Get image info for display
		imageInfo, err := veo3.GetImageInfo(imagePath)
		if err == nil {
			sizeMB := float64(imageInfo.SizeBytes) / (1024 * 1024)
			fmt.Printf("📸 %dx%d %s (%.1f MB)\n",
				imageInfo.Width, imageInfo.Height, imageInfo.Format, sizeMB)
		}
	}

	// Submit animation request
	ctx := context.Background()
	operation, err := client.AnimateImage(ctx, request)
	if err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Handle async mode
	if noWait {
		return outputOperation(operation, jsonFormat, pretty)
	}

	// Poll for completion with progress display
	if !jsonFormat {
		fmt.Printf("🎬 Animating image... (Operation: %s)\n", operation.ID)
	}

	operation, err = pollOperation(ctx, client, operation.ID, jsonFormat)
	if err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Download video if completed and not disabled
	if operation.Status == veo3.StatusDone && !noDownload {
		if !jsonFormat {
			fmt.Println("✓ Animation completed!")
		}

		err = downloadVideo(ctx, operation, outputDir, filename, jsonFormat)
		if err != nil {
			return handleError(err, jsonFormat, pretty)
		}
	}

	return outputOperation(operation, jsonFormat, pretty)
}
