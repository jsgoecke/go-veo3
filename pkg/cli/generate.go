package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jasongoecke/go-veo3/internal/format"
	"github.com/jasongoecke/go-veo3/pkg/config"
	"github.com/jasongoecke/go-veo3/pkg/operations"
	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newGenerateCmd creates the generate command with all subcommands
func newGenerateCmd() *cobra.Command {
	// Create fresh command instances each time to avoid flag redefinition in tests
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate videos from text prompts",
		Long: `Generate videos from text descriptions using Google Veo 3.1.

This command creates a video generation request and polls until completion.
The generated video is automatically downloaded to the current directory.

Generation typically takes 2-5 minutes depending on parameters.`,
		Example: `  # Generate 720p video with default settings
  veo3 generate --prompt "A serene sunset over ocean waves"

  # Generate 1080p 8-second video (1080p requires 8s duration)
  veo3 generate -p "City skyline at night" -r 1080p -d 8

  # Generate vertical video (9:16 aspect ratio)
  veo3 generate -p "Portrait of a cat" -a 9:16

  # Save to specific directory
  veo3 generate -p "Mountain landscape" --output ./videos/

  # Get machine-readable JSON output
  veo3 generate -p "Abstract art" --format=json

  # Start generation but don't wait (async mode)
  veo3 generate -p "Time-lapse clouds" --no-wait`,
		RunE: runGenerate,
	}

	generateTextCmd := &cobra.Command{
		Use:   "text",
		Short: "Generate video from text prompt",
		Long: `Generate a video from a text description using Google Veo 3.1.

This command creates a video generation request and polls until completion.
The generated video is automatically downloaded to the current directory.

Generation typically takes 2-5 minutes depending on parameters.`,
		Example: `  # Generate 720p video with default settings
  veo3 generate text --prompt "A serene sunset over ocean waves"

  # Generate 1080p 8-second video (1080p requires 8s duration)
  veo3 generate text -p "City skyline at night" -r 1080p -d 8

  # Generate vertical video (9:16 aspect ratio)
  veo3 generate text -p "Portrait of a cat" -a 9:16

  # Save to specific directory
  veo3 generate text -p "Mountain landscape" --output ./videos/

  # Get machine-readable JSON output
  veo3 generate text -p "Abstract art" --format=json

  # Start generation but don't wait (async mode)
  veo3 generate text -p "Time-lapse clouds" --no-wait`,
		RunE: runGenerateText,
	}

	// Add text subcommand to generate
	generateCmd.AddCommand(generateTextCmd)

	// Flags for the main generate command (backwards compatibility)
	generateCmd.Flags().StringP("prompt", "p", "", "Text prompt (required)")
	generateCmd.Flags().StringP("resolution", "r", "", "Resolution (720p or 1080p)")
	generateCmd.Flags().IntP("duration", "d", 0, "Duration in seconds (4, 6, or 8)")
	generateCmd.Flags().StringP("aspect-ratio", "a", "", "Aspect ratio (16:9 or 9:16)")
	generateCmd.Flags().StringP("model", "m", "", "Model to use")
	generateCmd.Flags().String("negative-prompt", "", "Negative prompt (elements to exclude)")
	generateCmd.Flags().StringSlice("reference", []string{}, "Reference image paths (max 3, requires 8s duration and 16:9 aspect ratio)")
	generateCmd.Flags().String("output", "", "Output directory for downloaded video")
	generateCmd.Flags().String("filename", "", "Custom filename for output video")
	generateCmd.Flags().Bool("no-wait", false, "Start generation and return immediately")
	generateCmd.Flags().Bool("no-download", false, "Skip automatic video download")
	generateCmd.Flags().Bool("pretty", false, "Pretty-print JSON output (with --json)")
	generateCmd.MarkFlagRequired("prompt")

	// Flags for the text subcommand
	generateTextCmd.Flags().StringP("prompt", "p", "", "Text prompt (required)")
	generateTextCmd.Flags().StringP("resolution", "r", "", "Resolution (720p or 1080p)")
	generateTextCmd.Flags().IntP("duration", "d", 0, "Duration in seconds (4, 6, or 8)")
	generateTextCmd.Flags().StringP("aspect-ratio", "a", "", "Aspect ratio (16:9 or 9:16)")
	generateTextCmd.Flags().StringP("model", "m", "", "Model to use")
	generateTextCmd.Flags().String("negative-prompt", "", "Negative prompt (elements to exclude)")
	generateTextCmd.Flags().StringSlice("reference", []string{}, "Reference image paths (max 3, requires 8s duration and 16:9 aspect ratio)")
	generateTextCmd.Flags().String("output", "", "Output directory for downloaded video")
	generateTextCmd.Flags().String("filename", "", "Custom filename for output video")
	generateTextCmd.Flags().Bool("no-wait", false, "Start generation and return immediately")
	generateTextCmd.Flags().Bool("no-download", false, "Skip automatic video download")
	generateTextCmd.Flags().Bool("pretty", false, "Pretty-print JSON output (with --json)")
	generateTextCmd.MarkFlagRequired("prompt")

	// Bind flags to viper for config integration
	viper.BindPFlag("model", generateCmd.Flags().Lookup("model"))
	viper.BindPFlag("resolution", generateCmd.Flags().Lookup("resolution"))
	viper.BindPFlag("duration", generateCmd.Flags().Lookup("duration"))
	viper.BindPFlag("aspect-ratio", generateCmd.Flags().Lookup("aspect-ratio"))
	viper.BindPFlag("output", generateCmd.Flags().Lookup("output"))

	viper.BindPFlag("model", generateTextCmd.Flags().Lookup("model"))
	viper.BindPFlag("resolution", generateTextCmd.Flags().Lookup("resolution"))
	viper.BindPFlag("duration", generateTextCmd.Flags().Lookup("duration"))
	viper.BindPFlag("aspect-ratio", generateTextCmd.Flags().Lookup("aspect-ratio"))
	viper.BindPFlag("output", generateTextCmd.Flags().Lookup("output"))

	return generateCmd
}

// runGenerate handles the main generate command (backwards compatibility)
func runGenerate(cmd *cobra.Command, args []string) error {
	return runGenerateText(cmd, args)
}

// runGenerateText handles text-to-video generation
func runGenerateText(cmd *cobra.Command, args []string) error {
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
	referenceImages, _ := cmd.Flags().GetStringSlice("reference")
	outputDir := getStringWithDefault(cmd, "output", cfg.OutputDirectory)
	filename, _ := cmd.Flags().GetString("filename")
	noWait, _ := cmd.Flags().GetBool("no-wait")
	noDownload, _ := cmd.Flags().GetBool("no-download")
	jsonFormat := viper.GetBool("json")
	pretty, _ := cmd.Flags().GetBool("pretty")

	// Create context
	ctx := context.Background()

	// Create API client - check multiple sources in priority order
	apiKey := viper.GetString("api-key")
	if apiKey == "" {
		// Check environment variables directly (viper's flag binding can override env vars)
		if envKey := os.Getenv("VEO3_API_KEY"); envKey != "" {
			apiKey = envKey
		} else if envKey := os.Getenv("GEMINI_API_KEY"); envKey != "" {
			apiKey = envKey
		} else {
			apiKey = cfg.APIKey
		}
	}

	// Check for custom API endpoint (for testing)
	opts := []veo3.ClientOption{}
	if apiEndpoint := os.Getenv("VEO3_API_ENDPOINT"); apiEndpoint != "" {
		opts = append(opts, veo3.WithBaseURL(apiEndpoint))
	}

	client, err := veo3.NewClient(context.Background(), apiKey, opts...)
	if err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Check if reference images are provided
	if len(referenceImages) > 0 {
		return handleReferenceImageGeneration(ctx, client, prompt, negativePrompt, model,
			resolution, duration, aspectRatio, referenceImages, outputDir, filename,
			noWait, noDownload, jsonFormat, pretty)
	}

	// Regular text-to-video generation
	request := &veo3.GenerationRequest{
		Prompt:           prompt,
		NegativePrompt:   negativePrompt,
		Model:            model,
		AspectRatio:      aspectRatio,
		Resolution:       resolution,
		DurationSeconds:  duration,
		PersonGeneration: "", // Use default
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Submit generation request
	operation, err := client.GenerateVideo(ctx, request)
	if err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Handle async mode
	if noWait {
		return outputOperation(operation, jsonFormat, pretty)
	}

	// Poll for completion with progress display
	if !jsonFormat {
		fmt.Printf("⠋ Generating video... (Operation: %s)\n", operation.ID)
	}

	operation, err = pollOperation(ctx, client, operation.ID, jsonFormat)
	if err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Download video if completed and not disabled
	if operation.Status == veo3.StatusDone && !noDownload {
		if !jsonFormat {
			fmt.Println("✓ Generation completed!")
		}

		err = downloadVideo(ctx, operation, outputDir, filename, jsonFormat)
		if err != nil {
			return handleError(err, jsonFormat, pretty)
		}
	}

	return outputOperation(operation, jsonFormat, pretty)
}

// Helper functions

func getStringWithDefault(cmd *cobra.Command, flag string, defaultValue string) string {
	value, _ := cmd.Flags().GetString(flag)
	if value == "" {
		value = viper.GetString(flag)
	}
	if value == "" {
		value = defaultValue
	}
	return value
}

func getIntWithDefault(cmd *cobra.Command, flag string, defaultValue int) int {
	value, _ := cmd.Flags().GetInt(flag)
	if value == 0 {
		value = viper.GetInt(flag)
	}
	if value == 0 {
		value = defaultValue
	}
	return value
}

func handleError(err error, jsonFormat bool, pretty bool) error {
	if jsonFormat {
		jsonOutput, _ := format.FormatErrorJSON("ERROR", err.Error(), nil)
		fmt.Println(jsonOutput)
	} else {
		// Human-readable error formatting
		fmt.Fprintf(os.Stderr, "❌ Error: %s\n", err.Error())
	}
	return err // Return the error for proper error handling in tests
}

func outputOperation(operation *veo3.Operation, jsonFormat bool, pretty bool) error {
	if jsonFormat {
		jsonOutput, err := format.FormatOperationJSON(operation)
		if err != nil {
			return err
		}
		fmt.Println(jsonOutput)
		return nil
	}

	// Human-readable output
	output := format.FormatOperation(operation)
	fmt.Println(output)
	return nil
}

func pollOperation(ctx context.Context, client *veo3.Client, operationID string, jsonFormat bool) (*veo3.Operation, error) {
	var bar *progressbar.ProgressBar
	if !jsonFormat {
		bar = progressbar.NewOptions(-1,
			progressbar.OptionSetDescription("Generating video..."),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionShowCount(),
			progressbar.OptionShowElapsedTimeOnFinish(),
		)
	}

	ticker := time.NewTicker(5 * time.Second) // Poll every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			operation, err := client.GetOperation(ctx, operationID)
			if err != nil {
				return nil, err
			}

			if !jsonFormat && bar != nil {
				bar.Add(1)
			}

			switch operation.Status {
			case veo3.StatusDone, veo3.StatusFailed, veo3.StatusCancelled:
				if !jsonFormat && bar != nil {
					bar.Finish()
				}
				return operation, nil
			}
		}
	}
}

func downloadVideo(ctx context.Context, operation *veo3.Operation, outputDir string, filename string, jsonFormat bool) error {
	if operation.VideoURI == "" {
		return fmt.Errorf("no video URI in completed operation")
	}

	// Use default output directory if not specified
	if outputDir == "" {
		outputDir = "."
	}

	// Generate filename if not specified
	if filename == "" {
		operationIDParts := strings.Split(operation.ID, "/")
		shortID := operationIDParts[len(operationIDParts)-1]
		filename = fmt.Sprintf("%s.mp4", shortID)
	}

	outputPath := filepath.Join(outputDir, filename)

	// Create downloader with progress display (opposite of jsonFormat)
	downloader := operations.NewDownloader(!jsonFormat)

	if !jsonFormat {
		fmt.Printf("⬇ Downloading video to %s...\n", outputPath)
	}

	_, err := downloader.DownloadVideo(ctx, operation, outputPath)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	return nil
}

// handleReferenceImageGeneration handles generation with reference images
func handleReferenceImageGeneration(ctx context.Context, client *veo3.Client, prompt, negativePrompt, model,
	resolution string, duration int, aspectRatio string, referenceImages []string, outputDir, filename string,
	noWait, noDownload, jsonFormat, pretty bool) error {

	// Create reference image request
	request := &veo3.ReferenceImageRequest{
		GenerationRequest: veo3.GenerationRequest{
			Prompt:           prompt,
			NegativePrompt:   negativePrompt,
			Model:            model,
			AspectRatio:      aspectRatio,
			Resolution:       resolution,
			DurationSeconds:  duration,
			PersonGeneration: "", // Use default
		},
		ReferenceImagePaths: referenceImages,
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Show upload progress for reference images
	if !jsonFormat {
		fmt.Printf("⬆ Uploading %d reference image(s)...\n", len(referenceImages))

		for i, imagePath := range referenceImages {
			fmt.Printf("  %d. %s\n", i+1, imagePath)

			// Get image info for display
			imageInfo, err := veo3.GetImageInfo(imagePath)
			if err == nil {
				sizeMB := float64(imageInfo.SizeBytes) / (1024 * 1024)
				fmt.Printf("     📸 %dx%d %s (%.1f MB)\n",
					imageInfo.Width, imageInfo.Height, imageInfo.Format, sizeMB)
			}
		}
	}

	// Submit reference image generation request
	operation, err := client.GenerateWithReferenceImages(ctx, request)
	if err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Handle async mode
	if noWait {
		return outputOperation(operation, jsonFormat, pretty)
	}

	// Poll for completion with progress display
	if !jsonFormat {
		fmt.Printf("🎨 Generating with reference style... (Operation: %s)\n", operation.ID)
	}

	operation, err = pollOperation(ctx, client, operation.ID, jsonFormat)
	if err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Download video if completed and not disabled
	if operation.Status == veo3.StatusDone && !noDownload {
		if !jsonFormat {
			fmt.Println("✓ Reference-guided generation completed!")
		}

		err = downloadVideo(ctx, operation, outputDir, filename, jsonFormat)
		if err != nil {
			return handleError(err, jsonFormat, pretty)
		}
	}

	return outputOperation(operation, jsonFormat, pretty)
}
