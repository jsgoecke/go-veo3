package operations

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/veo3"
)

// Manager handles operation lifecycle management
type Manager struct {
	client     *veo3.Client
	operations map[string]*veo3.Operation
	mutex      sync.RWMutex
}

// NewManager creates a new operation manager with optional client
func NewManager(client ...*veo3.Client) *Manager {
	var c *veo3.Client
	if len(client) > 0 {
		c = client[0]
	}
	return &Manager{
		client:     c,
		operations: make(map[string]*veo3.Operation),
	}
}

// SubmitOperation submits a new operation and starts tracking it
func (m *Manager) SubmitOperation(ctx context.Context, req *veo3.GenerationRequest) (*veo3.Operation, error) {
	// Submit the operation to the API
	op, err := m.client.GenerateVideo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit operation: %w", err)
	}

	// Store the operation for tracking
	m.mutex.Lock()
	m.operations[op.ID] = op
	m.mutex.Unlock()

	return op, nil
}

// GetOperation retrieves an operation by ID
func (m *Manager) GetOperation(operationID string) (*veo3.Operation, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	op, exists := m.operations[operationID]
	if !exists {
		return nil, fmt.Errorf("operation not found: %s", operationID)
	}
	return op, nil
}

// GetOperationExists retrieves an operation by ID (legacy interface)
func (m *Manager) GetOperationExists(operationID string) (*veo3.Operation, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	op, exists := m.operations[operationID]
	return op, exists
}

// UpdateOperation updates the status of an operation
func (m *Manager) UpdateOperation(op *veo3.Operation) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.operations[op.ID] = op
	return nil
}

// AddOperation adds a new operation to tracking
func (m *Manager) AddOperation(op *veo3.Operation) error {
	if op == nil {
		return fmt.Errorf("operation cannot be nil")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.operations[op.ID] = op
	return nil
}

// RemoveOperation removes an operation from tracking
func (m *Manager) RemoveOperation(operationID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.operations[operationID]; !exists {
		return fmt.Errorf("operation not found: %s", operationID)
	}

	delete(m.operations, operationID)
	return nil
}

// FilterOperations returns operations filtered by status
func (m *Manager) FilterOperations(status veo3.OperationStatus) []*veo3.Operation {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var filtered []*veo3.Operation
	for _, op := range m.operations {
		if op.Status == status {
			filtered = append(filtered, op)
		}
	}
	return filtered
}

// ListOperations returns all tracked operations
func (m *Manager) ListOperations() []*veo3.Operation {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	ops := make([]*veo3.Operation, 0, len(m.operations))
	for _, op := range m.operations {
		ops = append(ops, op)
	}
	return ops
}

// ListActiveOperations returns operations that are still running
func (m *Manager) ListActiveOperations() []*veo3.Operation {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var activeOps []*veo3.Operation
	for _, op := range m.operations {
		if op.Status == veo3.StatusPending || op.Status == veo3.StatusRunning {
			activeOps = append(activeOps, op)
		}
	}
	return activeOps
}

// CancelOperation cancels a running operation
func (m *Manager) CancelOperation(ctx context.Context, operationID string) error {
	// Check if operation exists and is cancellable
	m.mutex.RLock()
	op, exists := m.operations[operationID]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("operation not found: %s", operationID)
	}

	if op.Status != veo3.StatusPending && op.Status != veo3.StatusRunning {
		return fmt.Errorf("operation %s is not cancellable (status: %s)", operationID, op.Status)
	}

	// Cancel via API
	if err := m.client.CancelOperation(ctx, operationID); err != nil {
		return fmt.Errorf("failed to cancel operation: %w", err)
	}

	// Update local status
	now := time.Now()
	op.Status = veo3.StatusCancelled
	op.EndTime = &now
	m.UpdateOperation(op)

	return nil
}

// CleanupCompletedOperations removes old completed operations from memory
func (m *Manager) CleanupCompletedOperations(maxAge time.Duration) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	deleted := 0

	for id, op := range m.operations {
		if op.EndTime != nil && op.EndTime.Before(cutoff) {
			delete(m.operations, id)
			deleted++
		}
	}

	return deleted
}

// GetOperationStats returns statistics about operations
func (m *Manager) GetOperationStats() OperationStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := OperationStats{}

	for _, op := range m.operations {
		stats.Total++
		switch op.Status {
		case veo3.StatusPending:
			stats.Pending++
		case veo3.StatusRunning:
			stats.Running++
		case veo3.StatusDone:
			stats.Completed++
		case veo3.StatusFailed:
			stats.Failed++
		case veo3.StatusCancelled:
			stats.Cancelled++
		}
	}

	return stats
}

// OperationStats holds statistics about operations
type OperationStats struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Running   int `json:"running"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
	Cancelled int `json:"cancelled"`
}
