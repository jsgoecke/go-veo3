package operations

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/veo3"
)

// Manager manages operation lifecycle and state
type Manager struct {
	client     *veo3.Client
	operations map[string]*veo3.Operation
	mu         sync.RWMutex
}

// OperationStats contains statistics about operations
type OperationStats struct {
	Total     int
	Pending   int
	Running   int
	Completed int
	Failed    int
	Cancelled int
}

// NewManager creates a new operation manager
func NewManager(client *veo3.Client) *Manager {
	return &Manager{
		client:     client,
		operations: make(map[string]*veo3.Operation),
	}
}

// AddOperation adds an operation to the manager
func (m *Manager) AddOperation(op *veo3.Operation) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.operations[op.ID] = op
}

// GetOperation retrieves an operation by ID
func (m *Manager) GetOperation(operationID string) (*veo3.Operation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	op, exists := m.operations[operationID]
	if !exists {
		return nil, fmt.Errorf("operation %s not found", operationID)
	}

	return op, nil
}

// UpdateOperation updates an existing operation
func (m *Manager) UpdateOperation(op *veo3.Operation) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.operations[op.ID]; !exists {
		return fmt.Errorf("operation %s not found", op.ID)
	}

	m.operations[op.ID] = op
	return nil
}

// RemoveOperation removes an operation from the manager
func (m *Manager) RemoveOperation(operationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.operations[operationID]; !exists {
		return fmt.Errorf("operation %s not found", operationID)
	}

	delete(m.operations, operationID)
	return nil
}

// ListOperations returns all operations, ordered by start time (newest first)
func (m *Manager) ListOperations() []*veo3.Operation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ops := make([]*veo3.Operation, 0, len(m.operations))
	for _, op := range m.operations {
		ops = append(ops, op)
	}

	// Sort by start time, newest first
	for i := 0; i < len(ops)-1; i++ {
		for j := i + 1; j < len(ops); j++ {
			if ops[i].StartTime.Before(ops[j].StartTime) {
				ops[i], ops[j] = ops[j], ops[i]
			}
		}
	}

	return ops
}

// ListActiveOperations returns operations that are pending or running
func (m *Manager) ListActiveOperations() []*veo3.Operation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active []*veo3.Operation
	for _, op := range m.operations {
		if op.Status == veo3.StatusPending || op.Status == veo3.StatusRunning {
			active = append(active, op)
		}
	}

	return active
}

// FilterOperations returns operations matching the given status
func (m *Manager) FilterOperations(status veo3.OperationStatus) []*veo3.Operation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []*veo3.Operation
	for _, op := range m.operations {
		if op.Status == status {
			filtered = append(filtered, op)
		}
	}

	return filtered
}

// GetOperationStats returns statistics about all operations
func (m *Manager) GetOperationStats() OperationStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := OperationStats{
		Total: len(m.operations),
	}

	for _, op := range m.operations {
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

// CancelOperation cancels a running operation via the API
func (m *Manager) CancelOperation(ctx context.Context, operationID string) error {
	if m.client == nil {
		return fmt.Errorf("no client available")
	}

	// Cancel via API
	err := m.client.CancelOperation(ctx, operationID)
	if err != nil {
		return fmt.Errorf("failed to cancel operation: %w", err)
	}

	// Update local state
	m.mu.Lock()
	defer m.mu.Unlock()

	if op, exists := m.operations[operationID]; exists {
		op.Status = veo3.StatusCancelled
		now := time.Now()
		op.EndTime = &now
	}

	return nil
}

// List retrieves operations from the API
func (m *Manager) List(ctx context.Context, filter veo3.OperationStatus) ([]*veo3.Operation, error) {
	if m.client == nil {
		return nil, fmt.Errorf("no client available")
	}

	return m.client.ListOperations(ctx, filter)
}

// GetStatus retrieves the current status of an operation from the API
func (m *Manager) GetStatus(ctx context.Context, operationID string) (*veo3.Operation, error) {
	if m.client == nil {
		return nil, fmt.Errorf("no client available")
	}

	return m.client.GetOperationStatus(ctx, operationID)
}

// Cancel cancels an operation via the API (alias for CancelOperation)
func (m *Manager) Cancel(ctx context.Context, operationID string) error {
	return m.CancelOperation(ctx, operationID)
}
