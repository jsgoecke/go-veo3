package operations

import (
	"context"
	"fmt"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/veo3"
)

// Poller handles status polling with exponential backoff
type Poller struct {
	client        *veo3.Client
	manager       *Manager
	baseInterval  time.Duration
	maxInterval   time.Duration
	backoffFactor float64
	maxRetries    int
}

// NewPoller creates a new operation poller
func NewPoller(client *veo3.Client, manager *Manager) *Poller {
	return &Poller{
		client:        client,
		manager:       manager,
		baseInterval:  10 * time.Second, // Start with 10 seconds
		maxInterval:   5 * time.Minute,  // Max 5 minutes between polls
		backoffFactor: 1.5,              // Increase by 50% each time
		maxRetries:    10,               // Max 10 consecutive failures
	}
}

// PollOperation polls a single operation until completion
func (p *Poller) PollOperation(ctx context.Context, operationID string, progressCallback func(*veo3.Operation)) error {
	interval := p.baseInterval
	retries := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Get current operation status from API
		op, err := p.client.GetOperation(ctx, operationID)
		if err != nil {
			retries++
			if retries > p.maxRetries {
				return fmt.Errorf("max retries exceeded polling operation %s: %w", operationID, err)
			}

			// Wait before retry with exponential backoff
			time.Sleep(interval)
			interval = time.Duration(float64(interval) * p.backoffFactor)
			if interval > p.maxInterval {
				interval = p.maxInterval
			}
			continue
		}

		// Reset retry counter on successful request
		retries = 0
		interval = p.baseInterval

		// Update operation in manager
		p.manager.UpdateOperation(op)

		// Call progress callback if provided
		if progressCallback != nil {
			progressCallback(op)
		}

		// Check if operation is complete
		switch op.Status {
		case veo3.StatusDone, veo3.StatusFailed, veo3.StatusCancelled:
			return nil
		case veo3.StatusPending, veo3.StatusRunning:
			// Continue polling
		default:
			return fmt.Errorf("unknown operation status: %s", op.Status)
		}

		// Wait before next poll
		time.Sleep(interval)

		// Gradually increase interval for long-running operations
		interval = time.Duration(float64(interval) * p.backoffFactor)
		if interval > p.maxInterval {
			interval = p.maxInterval
		}
	}
}

// PollAllActive polls all active operations concurrently
func (p *Poller) PollAllActive(ctx context.Context, progressCallback func(*veo3.Operation)) error {
	activeOps := p.manager.ListActiveOperations()
	if len(activeOps) == 0 {
		return nil
	}

	// Create a context that can be cancelled
	pollCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Channel to collect errors
	errChan := make(chan error, len(activeOps))

	// Start polling each active operation
	for _, op := range activeOps {
		go func(operationID string) {
			err := p.PollOperation(pollCtx, operationID, progressCallback)
			errChan <- err
		}(op.ID)
	}

	// Wait for all polls to complete or fail
	var lastErr error
	for i := 0; i < len(activeOps); i++ {
		if err := <-errChan; err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// StartContinuousPolling starts continuous polling of all active operations
func (p *Poller) StartContinuousPolling(ctx context.Context, progressCallback func(*veo3.Operation)) {
	ticker := time.NewTicker(p.baseInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			activeOps := p.manager.ListActiveOperations()
			if len(activeOps) == 0 {
				continue
			}

			// Poll each active operation once
			for _, op := range activeOps {
				go func(operationID string) {
					updated, err := p.client.GetOperation(ctx, operationID)
					if err != nil {
						return // Ignore polling errors in continuous mode
					}

					p.manager.UpdateOperation(updated)
					if progressCallback != nil {
						progressCallback(updated)
					}
				}(op.ID)
			}
		}
	}
}

// WaitForCompletion waits for an operation to complete with progress updates
func (p *Poller) WaitForCompletion(ctx context.Context, operationID string, showProgress bool) (*veo3.Operation, error) {
	var progressCallback func(*veo3.Operation)

	if showProgress {
		progressCallback = func(op *veo3.Operation) {
			elapsed := time.Since(op.StartTime)
			status := string(op.Status)

			if op.Progress > 0 {
				fmt.Printf("\r⏳ %s... (%.0f%%, elapsed: %s)", status, op.Progress*100, formatDuration(elapsed))
			} else {
				fmt.Printf("\r⏳ %s... (elapsed: %s)", status, formatDuration(elapsed))
			}
		}
	}

	err := p.PollOperation(ctx, operationID, progressCallback)
	if err != nil {
		return nil, err
	}

	// Get final operation state
	op, exists := p.manager.GetOperation(operationID)
	if !exists {
		return nil, fmt.Errorf("operation not found after polling: %s", operationID)
	}

	if showProgress {
		fmt.Println() // New line after progress updates
	}

	return op, nil
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60

	if minutes > 0 {
		return fmt.Sprintf("%d:%02d", minutes, seconds)
	}
	return fmt.Sprintf("0:%02d", seconds)
}

// SetPollingConfig allows customization of polling behavior
func (p *Poller) SetPollingConfig(baseInterval, maxInterval time.Duration, backoffFactor float64, maxRetries int) {
	p.baseInterval = baseInterval
	p.maxInterval = maxInterval
	p.backoffFactor = backoffFactor
	p.maxRetries = maxRetries
}
