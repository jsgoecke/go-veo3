package operations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDownloader(t *testing.T) {
	tests := []struct {
		name         string
		showProgress bool
	}{
		{"with progress", true},
		{"without progress", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			downloader := NewDownloader(tt.showProgress)
			assert.NotNil(t, downloader)
			assert.NotNil(t, downloader.client)
			assert.Equal(t, tt.showProgress, downloader.showProgress)
		})
	}
}

func TestDownloader_DownloadVideo(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", "1024")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(make([]byte, 1024)) // Send 1KB of data
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test-video.mp4")

	downloader := NewDownloader(false)
	op := &veo3.Operation{
		ID:        "op-123",
		Status:    veo3.StatusDone,
		VideoURI:  server.URL,
		StartTime: time.Now(),
		EndTime:   &time.Time{},
		Metadata: map[string]interface{}{
			"model":            "test-model",
			"prompt":           "test prompt",
			"duration_seconds": 8,
			"resolution":       "720p",
			"aspect_ratio":     "16:9",
		},
	}
	*op.EndTime = op.StartTime.Add(2 * time.Minute)

	video, err := downloader.DownloadVideo(context.Background(), op, outputPath)
	require.NoError(t, err)
	assert.NotNil(t, video)

	// Verify file was created
	info, err := os.Stat(outputPath)
	require.NoError(t, err)
	assert.Equal(t, int64(1024), info.Size())

	// Verify metadata
	assert.Equal(t, outputPath, video.FilePath)
	assert.Equal(t, "op-123", video.OperationID)
	assert.Equal(t, int64(1024), video.FileSizeBytes)
	assert.Equal(t, "test-model", video.Model)
	assert.Equal(t, "test prompt", video.Prompt)
	assert.Equal(t, 8, video.DurationSeconds)
	assert.Equal(t, "720p", video.Resolution)
	assert.Equal(t, "16:9", video.AspectRatio)
	assert.Equal(t, 120, video.GenerationTimeSeconds)
}

func TestDownloader_DownloadVideo_NoVideoURI(t *testing.T) {
	downloader := NewDownloader(false)
	op := &veo3.Operation{
		ID:        "op-123",
		Status:    veo3.StatusDone,
		VideoURI:  "",
		StartTime: time.Now(),
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test-video.mp4")

	_, err := downloader.DownloadVideo(context.Background(), op, outputPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no video URI")
}

func TestDownloader_DownloadVideo_NotComplete(t *testing.T) {
	downloader := NewDownloader(false)
	op := &veo3.Operation{
		ID:        "op-123",
		Status:    veo3.StatusRunning,
		VideoURI:  "http://example.com/video.mp4",
		StartTime: time.Now(),
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test-video.mp4")

	_, err := downloader.DownloadVideo(context.Background(), op, outputPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not complete")
}

func TestDownloader_DownloadVideo_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	downloader := NewDownloader(false)
	op := &veo3.Operation{
		ID:        "op-123",
		Status:    veo3.StatusDone,
		VideoURI:  server.URL,
		StartTime: time.Now(),
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test-video.mp4")

	_, err := downloader.DownloadVideo(context.Background(), op, outputPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "download failed with status")
}

func TestDownloader_DownloadVideoWithRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "video/mp4")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(make([]byte, 512))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test-video.mp4")

	downloader := NewDownloader(false)
	op := &veo3.Operation{
		ID:        "op-123",
		Status:    veo3.StatusDone,
		VideoURI:  server.URL,
		StartTime: time.Now(),
	}

	video, err := downloader.DownloadVideoWithRetry(context.Background(), op, outputPath, 3)
	require.NoError(t, err)
	assert.NotNil(t, video)
	assert.Equal(t, 2, attempts)
}

func TestDownloader_DownloadVideoWithRetry_AllFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test-video.mp4")

	downloader := NewDownloader(false)
	op := &veo3.Operation{
		ID:        "op-123",
		Status:    veo3.StatusDone,
		VideoURI:  server.URL,
		StartTime: time.Now(),
	}

	_, err := downloader.DownloadVideoWithRetry(context.Background(), op, outputPath, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "download failed after")
}

func TestDownloader_CheckVideoAvailability(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"available", http.StatusOK, false},
		{"not found", http.StatusNotFound, true},
		{"server error", http.StatusInternalServerError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			downloader := NewDownloader(false)
			err := downloader.CheckVideoAvailability(context.Background(), server.URL)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDownloader_CheckVideoAvailability_EmptyURI(t *testing.T) {
	downloader := NewDownloader(false)
	err := downloader.CheckVideoAvailability(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "video URI is empty")
}

func TestDownloader_GetVideoInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", "2048576")
		w.Header().Set("Last-Modified", "Mon, 02 Jan 2023 15:04:05 GMT")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	downloader := NewDownloader(false)
	info, err := downloader.GetVideoInfo(context.Background(), server.URL)
	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, int64(2048576), info.ContentLength)
	assert.Equal(t, "video/mp4", info.ContentType)
	assert.Equal(t, "Mon, 02 Jan 2023 15:04:05 GMT", info.LastModified)
}

func TestDownloader_GetVideoInfo_EmptyURI(t *testing.T) {
	downloader := NewDownloader(false)
	_, err := downloader.GetVideoInfo(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "video URI is empty")
}

func TestDownloader_GetVideoInfo_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	downloader := NewDownloader(false)
	_, err := downloader.GetVideoInfo(context.Background(), server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "video info request failed")
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"bytes", 512, "512 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"megabytes", 1048576, "1.0 MB"},
		{"gigabytes", 1073741824, "1.0 GB"},
		{"terabytes", 1099511627776, "1.0 TB"},
		{"mixed", 1536, "1.5 KB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileSize(tt.bytes)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestVideoInfo_Struct(t *testing.T) {
	info := VideoInfo{
		ContentLength: 1024000,
		ContentType:   "video/mp4",
		LastModified:  "Mon, 02 Jan 2023 15:04:05 GMT",
	}

	assert.Equal(t, int64(1024000), info.ContentLength)
	assert.Equal(t, "video/mp4", info.ContentType)
	assert.Equal(t, "Mon, 02 Jan 2023 15:04:05 GMT", info.LastModified)
}

func TestDownloader_DownloadVideo_CreatesDirectory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(make([]byte, 100))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	// Nested directory that doesn't exist
	outputPath := filepath.Join(tmpDir, "nested", "dir", "video.mp4")

	downloader := NewDownloader(false)
	op := &veo3.Operation{
		ID:        "op-123",
		Status:    veo3.StatusDone,
		VideoURI:  server.URL,
		StartTime: time.Now(),
	}

	_, err := downloader.DownloadVideo(context.Background(), op, outputPath)
	require.NoError(t, err)

	// Verify directory was created
	dir := filepath.Dir(outputPath)
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
