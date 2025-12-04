package batch_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/batch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockJobExecutor is a mock for testing
type MockJobExecutor struct {
	mock.Mock
}

func (m *MockJobExecutor) Execute(ctx context.Context, job batch.BatchJob) (*batch.JobResult, error) {
	args := m.Called(ctx, job)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*batch.JobResult), args.Error(1)
}

func TestNewProcessor(t *testing.T) {
	executor := &MockJobExecutor{}
	processor := batch.NewProcessor(executor, 3)

	assert.NotNil(t, processor)
	assert.Equal(t, 3, processor.Concurrency)
}

func TestProcessManifest_Success(t *testing.T) {
	executor := &MockJobExecutor{}
	processor := batch.NewProcessor(executor, 2)

	manifest := &batch.BatchManifest{
		Jobs: []batch.BatchJob{
			{ID: "job1", Type: "generate", Options: map[string]interface{}{"prompt": "A"}, Output: "a.mp4"},
			{ID: "job2", Type: "generate", Options: map[string]interface{}{"prompt": "B"}, Output: "b.mp4"},
		},
		Concurrency:     2,
		ContinueOnError: true,
	}

	// Mock successful execution
	executor.On("Execute", mock.Anything, mock.Anything).Return(&batch.JobResult{
		JobID:   "job1",
		Success: true,
		Output:  "a.mp4",
	}, nil).Once()

	executor.On("Execute", mock.Anything, mock.Anything).Return(&batch.JobResult{
		JobID:   "job2",
		Success: true,
		Output:  "b.mp4",
	}, nil).Once()

	ctx := context.Background()
	results, err := processor.ProcessManifest(ctx, manifest)

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.True(t, results[1].Success)
	executor.AssertExpectations(t)
}

func TestProcessManifest_WithFailures(t *testing.T) {
	executor := &MockJobExecutor{}
	processor := batch.NewProcessor(executor, 2)

	manifest := &batch.BatchManifest{
		Jobs: []batch.BatchJob{
			{ID: "job1", Type: "generate", Options: map[string]interface{}{"prompt": "A"}, Output: "a.mp4"},
			{ID: "job2", Type: "generate", Options: map[string]interface{}{"prompt": "B"}, Output: "b.mp4"},
			{ID: "job3", Type: "generate", Options: map[string]interface{}{"prompt": "C"}, Output: "c.mp4"},
		},
		Concurrency:     2,
		ContinueOnError: true,
	}

	// Job 1 succeeds
	executor.On("Execute", mock.Anything, mock.MatchedBy(func(j batch.BatchJob) bool {
		return j.ID == "job1"
	})).Return(&batch.JobResult{
		JobID:   "job1",
		Success: true,
	}, nil)

	// Job 2 fails
	executor.On("Execute", mock.Anything, mock.MatchedBy(func(j batch.BatchJob) bool {
		return j.ID == "job2"
	})).Return(&batch.JobResult{
		JobID:   "job2",
		Success: false,
		Error:   "safety filter triggered",
	}, nil)

	// Job 3 succeeds
	executor.On("Execute", mock.Anything, mock.MatchedBy(func(j batch.BatchJob) bool {
		return j.ID == "job3"
	})).Return(&batch.JobResult{
		JobID:   "job3",
		Success: true,
	}, nil)

	ctx := context.Background()
	results, err := processor.ProcessManifest(ctx, manifest)

	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Check individual results
	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		} else {
			failCount++
		}
	}
	assert.Equal(t, 2, successCount)
	assert.Equal(t, 1, failCount)
}

func TestProcessManifest_StopOnError(t *testing.T) {
	executor := &MockJobExecutor{}
	processor := batch.NewProcessor(executor, 2)

	manifest := &batch.BatchManifest{
		Jobs: []batch.BatchJob{
			{ID: "job1", Type: "generate", Options: map[string]interface{}{"prompt": "A"}, Output: "a.mp4"},
			{ID: "job2", Type: "generate", Options: map[string]interface{}{"prompt": "B"}, Output: "b.mp4"},
			{ID: "job3", Type: "generate", Options: map[string]interface{}{"prompt": "C"}, Output: "c.mp4"},
		},
		Concurrency:     1, // Sequential for predictable order
		ContinueOnError: false,
	}

	// Job 1 succeeds
	executor.On("Execute", mock.Anything, mock.MatchedBy(func(j batch.BatchJob) bool {
		return j.ID == "job1"
	})).Return(&batch.JobResult{
		JobID:   "job1",
		Success: true,
	}, nil)

	// Job 2 fails - should stop processing
	executor.On("Execute", mock.Anything, mock.MatchedBy(func(j batch.BatchJob) bool {
		return j.ID == "job2"
	})).Return(nil, errors.New("critical error"))

	ctx := context.Background()
	results, err := processor.ProcessManifest(ctx, manifest)

	// Should return error and partial results
	assert.Error(t, err)
	assert.NotNil(t, results)
	// Job 3 should not have been attempted
	assert.LessOrEqual(t, len(results), 2)
}

func TestProcessManifest_Concurrency(t *testing.T) {
	executor := &MockJobExecutor{}
	processor := batch.NewProcessor(executor, 5)

	// Create 10 jobs
	jobs := make([]batch.BatchJob, 10)
	for i := 0; i < 10; i++ {
		jobs[i] = batch.BatchJob{
			ID:      string(rune('A' + i)),
			Type:    "generate",
			Options: map[string]interface{}{"prompt": "Test"},
			Output:  "out.mp4",
		}
	}

	manifest := &batch.BatchManifest{
		Jobs:            jobs,
		Concurrency:     5,
		ContinueOnError: true,
	}

	// Track concurrent executions with proper synchronization
	executing := make(chan bool, 10)
	var mu sync.Mutex
	maxConcurrent := 0
	currentConcurrent := 0

	executor.On("Execute", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		executing <- true

		mu.Lock()
		currentConcurrent++
		if currentConcurrent > maxConcurrent {
			maxConcurrent = currentConcurrent
		}
		mu.Unlock()

		time.Sleep(10 * time.Millisecond) // Simulate work

		<-executing
		mu.Lock()
		currentConcurrent--
		mu.Unlock()
	}).Return(&batch.JobResult{Success: true}, nil)

	ctx := context.Background()
	results, err := processor.ProcessManifest(ctx, manifest)

	require.NoError(t, err)
	assert.Len(t, results, 10)

	mu.Lock()
	finalMaxConcurrent := maxConcurrent
	mu.Unlock()

	assert.LessOrEqual(t, finalMaxConcurrent, 5, "Should respect concurrency limit")
}

func TestProcessManifest_ContextCancellation(t *testing.T) {
	executor := &MockJobExecutor{}
	processor := batch.NewProcessor(executor, 2)

	manifest := &batch.BatchManifest{
		Jobs: []batch.BatchJob{
			{ID: "job1", Type: "generate", Options: map[string]interface{}{"prompt": "A"}, Output: "a.mp4"},
			{ID: "job2", Type: "generate", Options: map[string]interface{}{"prompt": "B"}, Output: "b.mp4"},
		},
		Concurrency:     2,
		ContinueOnError: true,
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after first job starts
	executor.On("Execute", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		cancel()
		time.Sleep(50 * time.Millisecond)
	}).Return(&batch.JobResult{Success: true}, nil)

	results, err := processor.ProcessManifest(ctx, manifest)

	// Should handle cancellation gracefully
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	// Results may be nil or partial when context is cancelled
	_ = results
}

func TestGenerateSummary(t *testing.T) {
	results := []batch.JobResult{
		{JobID: "job1", Success: true, Duration: 45 * time.Second},
		{JobID: "job2", Success: true, Duration: 50 * time.Second},
		{JobID: "job3", Success: false, Error: "failed", Duration: 10 * time.Second},
		{JobID: "job4", Success: true, Duration: 55 * time.Second},
	}

	summary := batch.GenerateSummary(results)

	assert.Equal(t, 4, summary.TotalJobs)
	assert.Equal(t, 3, summary.SuccessfulJobs)
	assert.Equal(t, 1, summary.FailedJobs)
	assert.Greater(t, summary.TotalDuration, 150*time.Second)
}
