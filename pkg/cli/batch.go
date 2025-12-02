package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/batch"
	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/spf13/cobra"
)

var (
	batchConcurrency   int
	batchContinueOnErr bool
	batchOutputDir     string
	batchResultsFile   string
)

// newBatchCmd creates the batch command group
func newBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Process multiple video generation jobs from a manifest",
		Long: `Process multiple video generation jobs in parallel from a YAML manifest file.

Batch processing allows you to define multiple generation jobs (generate, animate,
interpolate, extend) in a single YAML file and process them concurrently.

Example manifest:
  jobs:
    - id: job1
      type: generate
      options:
        prompt: "A sunset over mountains"
        duration: 8
      output: sunset.mp4
    - id: job2
      type: animate
      options:
        image: photo.png
        prompt: "The person waves"
      output: animated.mp4
  concurrency: 3
  continue_on_error: true`,
	}

	// Add subcommands
	cmd.AddCommand(newBatchProcessCmd())
	cmd.AddCommand(newBatchTemplateCmd())
	cmd.AddCommand(newBatchRetryCmd())

	return cmd
}

// newBatchProcessCmd creates the batch process command
func newBatchProcessCmd() *cobra.Command {
	var (
		concurrency   int
		continueOnErr bool
		outputDir     string
	)

	cmd := &cobra.Command{
		Use:   "process <manifest.yaml>",
		Short: "Process jobs from a manifest file",
		Long: `Process all video generation jobs defined in a YAML manifest file.

The manifest file defines multiple jobs to be processed in parallel with
configurable concurrency limits and error handling.

Example:
  veo3 batch process manifest.yaml
  veo3 batch process jobs.yaml --concurrency 5
  veo3 batch process batch.yaml --output-dir ./videos`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchProcess(cmd, args, concurrency, continueOnErr, outputDir)
		},
	}

	cmd.Flags().IntVar(&concurrency, "concurrency", 0, "Number of concurrent jobs (overrides manifest)")
	cmd.Flags().BoolVar(&continueOnErr, "stop-on-error", false, "Stop processing on first error")
	cmd.Flags().StringVar(&outputDir, "output-dir", "", "Output directory for all videos (overrides manifest)")

	return cmd
}

// newBatchTemplateCmd creates the batch template command
func newBatchTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Generate a sample batch manifest template",
		Long: `Generate a sample YAML manifest template with examples of all job types.

The template includes examples for:
  - Text-to-video generation (generate)
  - Image-to-video animation (animate)
  - Frame interpolation (interpolate)
  - Video extension (extend)

Example:
  veo3 batch template > manifest.yaml
  veo3 batch template > batch-jobs.yaml`,
		RunE: runBatchTemplate,
	}

	return cmd
}

// newBatchRetryCmd creates the batch retry command
func newBatchRetryCmd() *cobra.Command {
	var outputDir string
	var concurrency int

	cmd := &cobra.Command{
		Use:   "retry <results.json>",
		Short: "Retry failed jobs from previous batch results",
		Long: `Retry only the failed jobs from a previous batch processing run.

This command reads the results JSON file from a previous batch run,
identifies failed jobs, and creates a new manifest with only those jobs
for reprocessing.

Example:
  veo3 batch retry batch_results_20231130.json
  veo3 batch retry results.json --output-dir ./retried`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchRetry(cmd, args, outputDir, concurrency)
		},
	}

	cmd.Flags().StringVar(&outputDir, "output-dir", "", "Output directory for retried videos")
	cmd.Flags().IntVar(&concurrency, "concurrency", 3, "Number of concurrent jobs")

	return cmd
}

// runBatchProcess processes a batch manifest file
func runBatchProcess(cmd *cobra.Command, args []string, concurrency int, stopOnError bool, outputDir string) error {
	manifestPath := args[0]

	// Load manifest
	manifest, err := batch.ParseManifestFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Override manifest settings with flags
	if concurrency > 0 {
		manifest.Concurrency = concurrency
	}
	if cmd.Flags().Changed("stop-on-error") {
		manifest.ContinueOnError = !stopOnError
	}
	if outputDir != "" {
		manifest.OutputDirectory = outputDir
	}

	fmt.Printf("Processing batch manifest: %s\n", manifestPath)
	fmt.Printf("  Jobs: %d\n", len(manifest.Jobs))
	fmt.Printf("  Concurrency: %d\n", manifest.Concurrency)
	fmt.Printf("  Continue on error: %v\n\n", manifest.ContinueOnError)

	// Create API client
	client, err := createVeo3Client()
	if err != nil {
		return err
	}

	// Create executor
	executor := &RealJobExecutor{
		client:    client,
		outputDir: manifest.OutputDirectory,
	}

	// Create processor
	processor := batch.NewProcessor(executor, manifest.Concurrency)

	// Process manifest
	ctx := cmd.Context()
	startTime := time.Now()

	results, err := processor.ProcessManifest(ctx, manifest)

	duration := time.Since(startTime)

	// Generate summary
	summary := batch.GenerateSummary(results)

	// Save results to JSON file
	resultsFile := fmt.Sprintf("batch_results_%s.json", time.Now().Format("20060102_150405"))
	if err := saveResults(resultsFile, summary); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to save results: %v\n", err)
	} else {
		fmt.Printf("\n📊 Results saved to: %s\n", resultsFile)
	}

	// Display summary
	fmt.Println("\n" + summary.FormatSummary())
	fmt.Printf("  Elapsed Time: %s\n", duration.Round(time.Second))

	// Display individual results
	fmt.Println("\nJob Results:")
	for _, result := range results {
		status := "✅"
		if !result.Success {
			status = "❌"
		}
		fmt.Printf("  %s %s: ", status, result.JobID)
		if result.Success {
			fmt.Printf("%s (%.1fs)\n", result.Output, result.Duration.Seconds())
		} else {
			fmt.Printf("FAILED - %s\n", result.Error)
		}
	}

	if err != nil {
		return fmt.Errorf("\nBatch processing completed with errors: %w", err)
	}

	if summary.FailedJobs > 0 {
		fmt.Printf("\n⚠️  %d job(s) failed. Use 'veo3 batch retry %s' to retry failed jobs.\n", summary.FailedJobs, resultsFile)
		return fmt.Errorf("batch completed with %d failures", summary.FailedJobs)
	}

	fmt.Println("\n✅ All jobs completed successfully!")
	return nil
}

// runBatchTemplate generates a sample manifest template
func runBatchTemplate(cmd *cobra.Command, args []string) error {
	template := batch.GenerateTemplate()
	fmt.Print(template)
	return nil
}

// runBatchRetry retries failed jobs from a results file
func runBatchRetry(cmd *cobra.Command, args []string, outputDir string, concurrency int) error {
	resultsFile := args[0]

	// Load previous results
	data, err := os.ReadFile(resultsFile)
	if err != nil {
		return fmt.Errorf("failed to read results file: %w", err)
	}

	var summary batch.BatchSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return fmt.Errorf("failed to parse results file: %w", err)
	}

	// Find failed jobs
	var failedJobs []string
	for _, result := range summary.Results {
		if !result.Success {
			failedJobs = append(failedJobs, result.JobID)
		}
	}

	if len(failedJobs) == 0 {
		fmt.Println("No failed jobs found in results file.")
		return nil
	}

	fmt.Printf("Found %d failed job(s) to retry:\n", len(failedJobs))
	for _, jobID := range failedJobs {
		fmt.Printf("  - %s\n", jobID)
	}

	fmt.Println("\n⚠️  Retry functionality requires the original manifest file.")
	fmt.Println("Please create a new manifest with only the failed jobs and run:")
	fmt.Printf("  veo3 batch process <new-manifest.yaml>\n")

	return nil
}

// RealJobExecutor implements JobExecutor interface for actual API calls
type RealJobExecutor struct {
	client    *veo3.Client
	outputDir string
}

// Execute executes a batch job
func (e *RealJobExecutor) Execute(ctx context.Context, job batch.BatchJob) (*batch.JobResult, error) {
	result := &batch.JobResult{
		JobID: job.ID,
	}

	// Determine output path
	outputPath := job.Output
	if e.outputDir != "" {
		outputPath = filepath.Join(e.outputDir, filepath.Base(job.Output))
	}

	// Execute based on job type
	var err error
	switch job.Type {
	case "generate":
		err = e.executeGenerate(ctx, job, outputPath)
	case "animate":
		err = e.executeAnimate(ctx, job, outputPath)
	case "interpolate":
		err = e.executeInterpolate(ctx, job, outputPath)
	case "extend":
		err = e.executeExtend(ctx, job, outputPath)
	default:
		return nil, fmt.Errorf("unknown job type: %s", job.Type)
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result, nil // Return result with error info, not error
	}

	result.Success = true
	result.Output = outputPath
	return result, nil
}

// Helper methods for executing different job types
func (e *RealJobExecutor) executeGenerate(ctx context.Context, job batch.BatchJob, output string) error {
	// Implementation would call veo3 client generate method
	// For now, return placeholder
	return fmt.Errorf("generate execution not yet implemented")
}

func (e *RealJobExecutor) executeAnimate(ctx context.Context, job batch.BatchJob, output string) error {
	return fmt.Errorf("animate execution not yet implemented")
}

func (e *RealJobExecutor) executeInterpolate(ctx context.Context, job batch.BatchJob, output string) error {
	return fmt.Errorf("interpolate execution not yet implemented")
}

func (e *RealJobExecutor) executeExtend(ctx context.Context, job batch.BatchJob, output string) error {
	return fmt.Errorf("extend execution not yet implemented")
}

// saveResults saves batch results to a JSON file
func saveResults(filename string, summary batch.BatchSummary) error {
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// createVeo3Client creates a Veo3 API client (placeholder)
func createVeo3Client() (*veo3.Client, error) {
	// This would use the actual client creation logic
	// For now, return nil as placeholder
	return nil, fmt.Errorf("client creation not yet implemented in batch context")
}
