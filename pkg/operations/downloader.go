package operations

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/schollz/progressbar/v3"
)

// Downloader handles video download with streaming and progress
type Downloader struct {
	client       *http.Client
	showProgress bool
}

// NewDownloader creates a new video downloader
func NewDownloader(showProgress bool) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 10 * time.Minute, // Allow up to 10 minutes for large video downloads
		},
		showProgress: showProgress,
	}
}

// DownloadVideo downloads a video from the given URI to the specified path
func (d *Downloader) DownloadVideo(ctx context.Context, op *veo3.Operation, outputPath string) (*veo3.GeneratedVideo, error) {
	if op.VideoURI == "" {
		return nil, fmt.Errorf("operation %s has no video URI", op.ID)
	}

	if op.Status != veo3.StatusDone {
		return nil, fmt.Errorf("operation %s is not complete (status: %s)", op.ID, op.Status)
	}

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0750); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", op.VideoURI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	// Make the request
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Create the output file
	file, err := os.Create(outputPath) // #nosec G304 -- User-specified output path is validated
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Setup progress bar if needed
	var reader io.Reader = resp.Body
	var bar *progressbar.ProgressBar

	if d.showProgress && resp.ContentLength > 0 {
		bar = progressbar.DefaultBytes(
			resp.ContentLength,
			"Downloading video",
		)
		progressReader := progressbar.NewReader(resp.Body, bar)
		reader = &progressReader
	}

	// Stream the download
	start := time.Now()
	written, err := io.Copy(file, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to save video: %w", err)
	}

	downloadTime := time.Since(start)

	if bar != nil {
		_ = bar.Finish()
		fmt.Println() // New line after progress bar
	}

	// Create GeneratedVideo metadata
	generatedVideo := &veo3.GeneratedVideo{
		FilePath:      outputPath,
		OperationID:   op.ID,
		FileSizeBytes: written,
		CreatedAt:     time.Now(),
	}

	// Extract metadata from operation if available
	if op.Metadata != nil {
		if model, ok := op.Metadata["model"].(string); ok {
			generatedVideo.Model = model
		}
		if prompt, ok := op.Metadata["prompt"].(string); ok {
			generatedVideo.Prompt = prompt
		}
		if duration, ok := op.Metadata["duration_seconds"].(int); ok {
			generatedVideo.DurationSeconds = duration
		}
		if resolution, ok := op.Metadata["resolution"].(string); ok {
			generatedVideo.Resolution = resolution
		}
		if aspectRatio, ok := op.Metadata["aspect_ratio"].(string); ok {
			generatedVideo.AspectRatio = aspectRatio
		}
	}

	// Calculate generation time if available
	if op.StartTime != (time.Time{}) && op.EndTime != nil {
		generationTime := op.EndTime.Sub(op.StartTime)
		generatedVideo.GenerationTimeSeconds = int(generationTime.Seconds())
	}

	if d.showProgress {
		fmt.Printf("✅ Video downloaded successfully!\n")
		fmt.Printf("📁 Saved to: %s\n", outputPath)
		fmt.Printf("📊 Size: %s | Download time: %s\n",
			formatFileSize(written),
			downloadTime.Round(time.Second))
	}

	return generatedVideo, nil
}

// DownloadVideoWithRetry downloads with automatic retry on failure
func (d *Downloader) DownloadVideoWithRetry(ctx context.Context, op *veo3.Operation, outputPath string, maxRetries int) (*veo3.GeneratedVideo, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		video, err := d.DownloadVideo(ctx, op, outputPath)
		if err == nil {
			return video, nil
		}

		lastErr = err
		if attempt < maxRetries {
			if d.showProgress {
				fmt.Printf("Download attempt %d failed, retrying in 5 seconds...\n", attempt)
			}
			time.Sleep(5 * time.Second)
		}
	}

	return nil, fmt.Errorf("download failed after %d attempts: %w", maxRetries, lastErr)
}

// CheckVideoAvailability checks if a video is ready for download
func (d *Downloader) CheckVideoAvailability(ctx context.Context, videoURI string) error {
	if videoURI == "" {
		return fmt.Errorf("video URI is empty")
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", videoURI, nil)
	if err != nil {
		return fmt.Errorf("failed to create availability check request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check video availability: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("video not available (status %d)", resp.StatusCode)
	}

	return nil
}

// GetVideoInfo returns information about the video without downloading it
func (d *Downloader) GetVideoInfo(ctx context.Context, videoURI string) (*VideoInfo, error) {
	if videoURI == "" {
		return nil, fmt.Errorf("video URI is empty")
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", videoURI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create info request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("video info request failed (status %d)", resp.StatusCode)
	}

	info := &VideoInfo{
		ContentLength: resp.ContentLength,
		ContentType:   resp.Header.Get("Content-Type"),
		LastModified:  resp.Header.Get("Last-Modified"),
	}

	return info, nil
}

// VideoInfo holds metadata about a video
type VideoInfo struct {
	ContentLength int64  `json:"content_length"`
	ContentType   string `json:"content_type"`
	LastModified  string `json:"last_modified"`
}

// formatFileSize formats bytes into human readable format
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
	return fmt.Sprintf("%.1f %cB",
		float64(bytes)/float64(div), "KMGTPE"[exp])
}
