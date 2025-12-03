package batch

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// JobExecutor is an interface for executing batch jobs
type JobExecutor interface {
	Execute(ctx context.Context, job BatchJob) (*JobResult, error)
}

// JobResult represents the result of a batch job execution
type JobResult struct {
	JobID     string        `json:"job_id"`
	Success   bool          `json:"success"`
	Output    string        `json:"output,omitempty"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
}

// Processor handles concurrent execution of batch jobs
type Processor struct {
	executor    JobExecutor
	Concurrency int
}

// BatchSummary provides statistics about batch execution
type BatchSummary struct {
	TotalJobs      int           `json:"total_jobs"`
	SuccessfulJobs int           `json:"successful_jobs"`
	FailedJobs     int           `json:"failed_jobs"`
	TotalDuration  time.Duration `json:"total_duration"`
	Results        []JobResult   `json:"results"`
}

// NewProcessor creates a new batch processor
func NewProcessor(executor JobExecutor, concurrency int) *Processor {
	if concurrency < 1 {
		concurrency = 3 // Default
	}

	return &Processor{
		executor:    executor,
		Concurrency: concurrency,
	}
}

// ProcessManifest processes all jobs in a manifest with concurrency control
func (p *Processor) ProcessManifest(ctx context.Context, manifest *BatchManifest) ([]JobResult, error) {
	// Use manifest's concurrency if set
	concurrency := manifest.Concurrency
	if concurrency < 1 {
		concurrency = p.Concurrency
	}

	// Create cancellable context for stopping on error
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create channels for job distribution and result collection
	jobChan := make(chan BatchJob, len(manifest.Jobs))
	resultChan := make(chan JobResult, len(manifest.Jobs))
	errorChan := make(chan error, 1)

	// Fill job channel
	for _, job := range manifest.Jobs {
		jobChan <- job
	}
	close(jobChan)

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.worker(workerCtx, jobChan, resultChan, errorChan, manifest.ContinueOnError)
		}()
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results
	var results []JobResult
	var firstError error

	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				// All results collected
				if firstError != nil && !manifest.ContinueOnError {
					return results, firstError
				}
				return results, nil
			}
			results = append(results, result)

		case err, ok := <-errorChan:
			if ok && err != nil && firstError == nil {
				firstError = err
				if !manifest.ContinueOnError {
					// Cancel context to stop other workers
					cancel()
					// Continue collecting results from workers that already started
				}
			}

		case <-ctx.Done():
			return results, fmt.Errorf("batch processing cancelled: %w", ctx.Err())
		}
	}
}

// worker processes jobs from the job channel
func (p *Processor) worker(ctx context.Context, jobs <-chan BatchJob, results chan<- JobResult, errors chan<- error, continueOnError bool) {
	for job := range jobs {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return
		default:
		}

		startTime := time.Now()
		result, err := p.executor.Execute(ctx, job)
		duration := time.Since(startTime)

		if err != nil {
			// Send error to error channel
			if !continueOnError {
				select {
				case errors <- err:
				default:
				}
				return
			}

			// Create failed result
			result = &JobResult{
				JobID:     job.ID,
				Success:   false,
				Error:     err.Error(),
				StartTime: startTime,
				EndTime:   time.Now(),
				Duration:  duration,
			}
		} else if result != nil {
			// Update timing information
			result.StartTime = startTime
			result.EndTime = time.Now()
			result.Duration = duration
		}

		if result != nil {
			results <- *result
		}
	}
}

// GenerateSummary creates a summary of batch execution results
func GenerateSummary(results []JobResult) BatchSummary {
	summary := BatchSummary{
		TotalJobs: len(results),
		Results:   results,
	}

	var totalDuration time.Duration
	for _, result := range results {
		if result.Success {
			summary.SuccessfulJobs++
		} else {
			summary.FailedJobs++
		}
		totalDuration += result.Duration
	}

	summary.TotalDuration = totalDuration

	return summary
}

// FormatSummary formats a batch summary for display
func (s BatchSummary) FormatSummary() string {
	successRate := 0.0
	if s.TotalJobs > 0 {
		successRate = float64(s.SuccessfulJobs) / float64(s.TotalJobs) * 100
	}

	return fmt.Sprintf(
		"Batch Processing Summary:\n"+
			"  Total Jobs: %d\n"+
			"  Successful: %d\n"+
			"  Failed: %d\n"+
			"  Success Rate: %.1f%%\n"+
			"  Total Duration: %s",
		s.TotalJobs,
		s.SuccessfulJobs,
		s.FailedJobs,
		successRate,
		s.TotalDuration.Round(time.Second),
	)
}
