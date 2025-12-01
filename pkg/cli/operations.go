package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jasongoecke/go-veo3/internal/format"
	"github.com/jasongoecke/go-veo3/pkg/config"
	"github.com/jasongoecke/go-veo3/pkg/operations"
	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newOperationsCmd creates the operations command group
func newOperationsCmd() *cobra.Command {
	// Create fresh command instance to avoid flag redefinition in tests
	operationsCmd := &cobra.Command{
		Use:   "operations",
		Short: "Manage long-running video generation operations",
		Long: `Manage long-running video generation operations.

This command group provides tools to list, check status, download, and cancel
video generation operations. This is useful for managing multiple concurrent
generations or recovering from interrupted sessions.`,
		Example: `  # List all operations
  veo3 operations list

  # Check status of specific operation
  veo3 operations status operations/abc123

  # Download completed video
  veo3 operations download operations/abc123

  # Cancel running operation
  veo3 operations cancel operations/abc123`,
	}

	// Add subcommands
	operationsCmd.AddCommand(newOperationsListCmd())
	operationsCmd.AddCommand(newOperationsStatusCmd())
	operationsCmd.AddCommand(newOperationsDownloadCmd())
	operationsCmd.AddCommand(newOperationsCancelCmd())

	return operationsCmd
}

// newOperationsListCmd creates the 'operations list' command
func newOperationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tracked operations",
		Long: `List all tracked video generation operations with their current status.

Operations are shown with their ID, status, progress (if available), start time,
and associated model. Completed operations show their completion time and
video download URI.`,
		Example: `  # List all operations
  veo3 operations list

  # List operations in JSON format
  veo3 operations list --json

  # List only running operations
  veo3 operations list --status running

  # List with detailed information
  veo3 operations list --detailed`,
		RunE: runOperationsList,
	}

	cmd.Flags().String("status", "", "Filter by status (pending, running, done, failed, cancelled)")
	cmd.Flags().Bool("detailed", false, "Show detailed information for each operation")
	cmd.Flags().Bool("pretty", false, "Pretty-print JSON output (with --json)")

	return cmd
}

// newOperationsStatusCmd creates the 'operations status' command
func newOperationsStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [operation-id]",
		Short: "Check the status of a specific operation",
		Long: `Check the detailed status of a specific video generation operation.

This command shows comprehensive information about an operation including
current status, progress, timing, metadata, and any error details if the
operation failed.`,
		Example: `  # Check operation status
  veo3 operations status operations/abc123

  # Get status in JSON format
  veo3 operations status operations/abc123 --json

  # Continuously watch operation status
  veo3 operations status operations/abc123 --watch`,
		Args: cobra.ExactArgs(1),
		RunE: runOperationsStatus,
	}

	cmd.Flags().Bool("watch", false, "Continuously watch operation status until completion")
	cmd.Flags().Bool("pretty", false, "Pretty-print JSON output (with --json)")

	return cmd
}

// newOperationsDownloadCmd creates the 'operations download' command
func newOperationsDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download [operation-id]",
		Short: "Download the video from a completed operation",
		Long: `Download the generated video from a completed operation.

This command is useful for downloading videos from operations that were
started with --no-download or from previous sessions. The operation must
be in DONE status with a valid video URI.`,
		Example: `  # Download video from completed operation
  veo3 operations download operations/abc123

  # Download to specific directory
  veo3 operations download operations/abc123 --output ./videos/

  # Download with custom filename
  veo3 operations download operations/abc123 --filename my-video.mp4`,
		Args: cobra.ExactArgs(1),
		RunE: runOperationsDownload,
	}

	cmd.Flags().String("output", "", "Output directory for downloaded video")
	cmd.Flags().String("filename", "", "Custom filename for output video")
	cmd.Flags().Bool("overwrite", false, "Overwrite existing file if it exists")

	return cmd
}

// newOperationsCancelCmd creates the 'operations cancel' command
func newOperationsCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel [operation-id]",
		Short: "Cancel a running operation",
		Long: `Cancel a pending or running video generation operation.

This command attempts to cancel an operation that is currently pending or
running. Once cancelled, the operation cannot be resumed and any partial
progress is lost.`,
		Example: `  # Cancel specific operation
  veo3 operations cancel operations/abc123

  # Cancel with confirmation prompt
  veo3 operations cancel operations/abc123 --confirm

  # Cancel all running operations
  veo3 operations cancel --all`,
		Args: cobra.MaximumNArgs(1),
		RunE: runOperationsCancel,
	}

	cmd.Flags().Bool("all", false, "Cancel all pending and running operations")
	cmd.Flags().Bool("confirm", false, "Prompt for confirmation before cancelling")

	return cmd
}

// Command implementations

func runOperationsList(cmd *cobra.Command, args []string) error {
	// Create operations manager (without client for listing stored operations)
	opsManager := operations.NewManager(nil)

	// Get filter status
	statusFilter, _ := cmd.Flags().GetString("status")
	detailed, _ := cmd.Flags().GetBool("detailed")
	jsonFormat := viper.GetBool("json")
	// pretty, _ := cmd.Flags().GetBool("pretty") // TODO: Use for formatted output

	// List operations
	var ops []*veo3.Operation
	if statusFilter != "" {
		// Parse status filter
		var status veo3.OperationStatus
		switch strings.ToLower(statusFilter) {
		case "pending":
			status = veo3.StatusPending
		case "running":
			status = veo3.StatusRunning
		case "done", "completed":
			status = veo3.StatusDone
		case "failed":
			status = veo3.StatusFailed
		case "cancelled":
			status = veo3.StatusCancelled
		default:
			return fmt.Errorf("invalid status filter: %s (valid: pending, running, done, failed, cancelled)", statusFilter)
		}
		ops = opsManager.FilterOperations(status)
	} else {
		ops = opsManager.ListOperations()
	}

	// Output operations
	if jsonFormat {
		jsonOutput, err := format.FormatOperationListJSON(ops)
		if err != nil {
			return err
		}
		fmt.Println(jsonOutput)
	} else {
		if detailed {
			for i, op := range ops {
				if i > 0 {
					fmt.Println() // Blank line between operations
				}
				output := format.FormatOperation(op)
				fmt.Print(output)
			}
		} else {
			output := format.FormatOperationList(ops)
			fmt.Print(output)
		}
	}

	return nil
}

func runOperationsStatus(cmd *cobra.Command, args []string) error {
	operationID := args[0]

	// Create operations manager
	opsManager := operations.NewManager(nil)

	watch, _ := cmd.Flags().GetBool("watch")
	jsonFormat := viper.GetBool("json")
	pretty, _ := cmd.Flags().GetBool("pretty")

	if watch {
		// TODO: Implement watch mode with live updates
		return fmt.Errorf("watch mode not implemented yet")
	}

	// Get operation
	op, err := opsManager.GetOperation(operationID)
	if err != nil {
		return handleError(err, jsonFormat, pretty)
	}

	// Output operation status
	return outputOperation(op, jsonFormat, pretty)
}

func runOperationsDownload(cmd *cobra.Command, args []string) error {
	operationID := args[0]

	// Load configuration
	manager := config.NewManager("")
	cfg, _ := manager.Load()
	if cfg == nil {
		cfg = &config.Configuration{OutputDirectory: "."}
	}

	outputDir, _ := cmd.Flags().GetString("output")
	if outputDir == "" {
		outputDir = cfg.OutputDirectory
	}

	filename, _ := cmd.Flags().GetString("filename")
	overwrite, _ := cmd.Flags().GetBool("overwrite")
	jsonFormat := viper.GetBool("json")

	// Create operations manager
	opsManager := operations.NewManager(nil)

	// Get operation
	op, err := opsManager.GetOperation(operationID)
	if err != nil {
		return handleError(err, jsonFormat, false)
	}

	// Check operation status
	if op.Status != veo3.StatusDone {
		return handleError(fmt.Errorf("operation %s is not completed (status: %s)", operationID, op.Status), jsonFormat, false)
	}

	if op.VideoURI == "" {
		return handleError(fmt.Errorf("operation %s has no video URI", operationID), jsonFormat, false)
	}

	// Download video
	if !jsonFormat {
		fmt.Printf("⬇ Downloading video from operation %s...\n", operationID)
	}

	// Create downloader
	downloader := operations.NewDownloader(!jsonFormat)

	// Generate output path
	if filename == "" {
		shortID := operationID[strings.LastIndex(operationID, "/")+1:]
		filename = fmt.Sprintf("%s.mp4", shortID)
	}

	outputPath := fmt.Sprintf("%s/%s", outputDir, filename)

	// Check if file exists and handle overwrite
	if !overwrite {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("file already exists: %s (use --overwrite to replace)", outputPath)
		}
	}

	ctx := context.Background()
	_, err = downloader.DownloadVideo(ctx, op, outputPath)
	if err != nil {
		return handleError(err, jsonFormat, false)
	}

	if jsonFormat {
		result := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"operation_id": operationID,
				"output_path":  outputPath,
				"status":       "downloaded",
			},
		}
		jsonOutput, _ := format.FormatGenericJSON(result)
		fmt.Println(jsonOutput)
	}

	return nil
}

func runOperationsCancel(cmd *cobra.Command, args []string) error {
	cancelAll, _ := cmd.Flags().GetBool("all")
	confirm, _ := cmd.Flags().GetBool("confirm")
	jsonFormat := viper.GetBool("json")

	if !cancelAll && len(args) == 0 {
		return fmt.Errorf("operation ID required, or use --all to cancel all operations")
	}

	// Create API client
	apiKey := viper.GetString("api-key")
	client, err := veo3.NewClient(context.Background(), apiKey)
	if err != nil {
		return handleError(err, jsonFormat, false)
	}

	// Create operations manager
	opsManager := operations.NewManager(client)

	ctx := context.Background()

	if cancelAll {
		// Cancel all active operations
		activeOps := opsManager.ListActiveOperations()
		if len(activeOps) == 0 {
			if !jsonFormat {
				fmt.Println("No active operations to cancel")
			}
			return nil
		}

		if confirm && !jsonFormat {
			fmt.Printf("Cancel %d active operations? (y/N): ", len(activeOps))
			var response string
			_, _ = fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Cancelled")
				return nil
			}
		}

		cancelled := 0
		for _, op := range activeOps {
			if err := opsManager.CancelOperation(ctx, op.ID); err != nil {
				if !jsonFormat {
					fmt.Printf("Failed to cancel %s: %v\n", op.ID, err)
				}
			} else {
				cancelled++
			}
		}

		if jsonFormat {
			result := map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"cancelled_count": cancelled,
					"total_count":     len(activeOps),
				},
			}
			jsonOutput, _ := format.FormatGenericJSON(result)
			fmt.Println(jsonOutput)
		} else {
			fmt.Printf("Cancelled %d out of %d operations\n", cancelled, len(activeOps))
		}
	} else {
		// Cancel specific operation
		operationID := args[0]

		err := opsManager.CancelOperation(ctx, operationID)
		if err != nil {
			return handleError(err, jsonFormat, false)
		}

		if jsonFormat {
			result := map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"operation_id": operationID,
					"status":       "cancelled",
				},
			}
			jsonOutput, _ := format.FormatGenericJSON(result)
			fmt.Println(jsonOutput)
		} else {
			fmt.Printf("✓ Operation %s cancelled successfully\n", operationID)
		}
	}

	return nil
}
