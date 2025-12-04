package operations

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	client := &veo3.Client{}
	manager := NewManager(client)

	assert.NotNil(t, manager)
	assert.Equal(t, client, manager.client)
	assert.NotNil(t, manager.operations)
	assert.Equal(t, 0, len(manager.operations))
}

func TestManager_AddOperation(t *testing.T) {
	manager := NewManager(nil)
	op := &veo3.Operation{
		ID:        "op-123",
		Status:    veo3.StatusPending,
		StartTime: time.Now(),
	}

	manager.AddOperation(op)

	assert.Equal(t, 1, len(manager.operations))
	stored, err := manager.GetOperation("op-123")
	require.NoError(t, err)
	assert.Equal(t, op, stored)
}

func TestManager_GetOperation(t *testing.T) {
	manager := NewManager(nil)

	tests := []struct {
		name    string
		opID    string
		setup   func()
		wantErr bool
		wantOp  *veo3.Operation
	}{
		{
			name: "existing operation",
			opID: "op-exists",
			setup: func() {
				op := &veo3.Operation{
					ID:        "op-exists",
					Status:    veo3.StatusDone,
					StartTime: time.Now(),
				}
				manager.AddOperation(op)
			},
			wantErr: false,
		},
		{
			name:    "non-existent operation",
			opID:    "op-not-found",
			setup:   func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager = NewManager(nil) // Reset manager
			tt.setup()

			op, err := manager.GetOperation(tt.opID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			} else {
				require.NoError(t, err)
				assert.NotNil(t, op)
				assert.Equal(t, tt.opID, op.ID)
			}
		})
	}
}

func TestManager_UpdateOperation(t *testing.T) {
	manager := NewManager(nil)

	// Add initial operation
	op := &veo3.Operation{
		ID:        "op-123",
		Status:    veo3.StatusPending,
		StartTime: time.Now(),
	}
	manager.AddOperation(op)

	// Update it
	updatedOp := &veo3.Operation{
		ID:        "op-123",
		Status:    veo3.StatusRunning,
		Progress:  0.5,
		StartTime: op.StartTime,
	}
	err := manager.UpdateOperation(updatedOp)
	require.NoError(t, err)

	// Verify update
	stored, err := manager.GetOperation("op-123")
	require.NoError(t, err)
	assert.Equal(t, veo3.StatusRunning, stored.Status)
	assert.Equal(t, 0.5, stored.Progress)
}

func TestManager_UpdateOperation_NotFound(t *testing.T) {
	manager := NewManager(nil)

	op := &veo3.Operation{
		ID:        "op-not-exists",
		Status:    veo3.StatusRunning,
		StartTime: time.Now(),
	}

	err := manager.UpdateOperation(op)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_RemoveOperation(t *testing.T) {
	manager := NewManager(nil)

	op := &veo3.Operation{
		ID:        "op-123",
		Status:    veo3.StatusDone,
		StartTime: time.Now(),
	}
	manager.AddOperation(op)

	// Remove it
	err := manager.RemoveOperation("op-123")
	require.NoError(t, err)

	// Verify removal
	_, err = manager.GetOperation("op-123")
	assert.Error(t, err)
}

func TestManager_RemoveOperation_NotFound(t *testing.T) {
	manager := NewManager(nil)

	err := manager.RemoveOperation("op-not-exists")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_ListOperations(t *testing.T) {
	manager := NewManager(nil)

	now := time.Now()
	ops := []*veo3.Operation{
		{ID: "op-1", Status: veo3.StatusDone, StartTime: now.Add(-2 * time.Hour)},
		{ID: "op-2", Status: veo3.StatusRunning, StartTime: now.Add(-1 * time.Hour)},
		{ID: "op-3", Status: veo3.StatusPending, StartTime: now},
	}

	for _, op := range ops {
		manager.AddOperation(op)
	}

	list := manager.ListOperations()
	assert.Equal(t, 3, len(list))

	// Should be sorted by start time, newest first
	assert.Equal(t, "op-3", list[0].ID)
	assert.Equal(t, "op-2", list[1].ID)
	assert.Equal(t, "op-1", list[2].ID)
}

func TestManager_ListOperations_Empty(t *testing.T) {
	manager := NewManager(nil)

	list := manager.ListOperations()
	assert.NotNil(t, list)
	assert.Equal(t, 0, len(list))
}

func TestManager_ListActiveOperations(t *testing.T) {
	manager := NewManager(nil)

	ops := []*veo3.Operation{
		{ID: "op-done", Status: veo3.StatusDone, StartTime: time.Now()},
		{ID: "op-pending", Status: veo3.StatusPending, StartTime: time.Now()},
		{ID: "op-running", Status: veo3.StatusRunning, StartTime: time.Now()},
		{ID: "op-failed", Status: veo3.StatusFailed, StartTime: time.Now()},
	}

	for _, op := range ops {
		manager.AddOperation(op)
	}

	active := manager.ListActiveOperations()
	assert.Equal(t, 2, len(active))

	// Should only include pending and running
	ids := make([]string, len(active))
	for i, op := range active {
		ids[i] = op.ID
	}
	assert.Contains(t, ids, "op-pending")
	assert.Contains(t, ids, "op-running")
}

func TestManager_FilterOperations(t *testing.T) {
	manager := NewManager(nil)

	ops := []*veo3.Operation{
		{ID: "op-done-1", Status: veo3.StatusDone, StartTime: time.Now()},
		{ID: "op-done-2", Status: veo3.StatusDone, StartTime: time.Now()},
		{ID: "op-running", Status: veo3.StatusRunning, StartTime: time.Now()},
		{ID: "op-failed", Status: veo3.StatusFailed, StartTime: time.Now()},
	}

	for _, op := range ops {
		manager.AddOperation(op)
	}

	tests := []struct {
		name   string
		status veo3.OperationStatus
		want   int
	}{
		{"filter done", veo3.StatusDone, 2},
		{"filter running", veo3.StatusRunning, 1},
		{"filter failed", veo3.StatusFailed, 1},
		{"filter cancelled", veo3.StatusCancelled, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := manager.FilterOperations(tt.status)
			assert.Equal(t, tt.want, len(filtered))

			for _, op := range filtered {
				assert.Equal(t, tt.status, op.Status)
			}
		})
	}
}

func TestManager_GetOperationStats(t *testing.T) {
	manager := NewManager(nil)

	ops := []*veo3.Operation{
		{ID: "op-pending-1", Status: veo3.StatusPending, StartTime: time.Now()},
		{ID: "op-pending-2", Status: veo3.StatusPending, StartTime: time.Now()},
		{ID: "op-running-1", Status: veo3.StatusRunning, StartTime: time.Now()},
		{ID: "op-running-2", Status: veo3.StatusRunning, StartTime: time.Now()},
		{ID: "op-running-3", Status: veo3.StatusRunning, StartTime: time.Now()},
		{ID: "op-done-1", Status: veo3.StatusDone, StartTime: time.Now()},
		{ID: "op-done-2", Status: veo3.StatusDone, StartTime: time.Now()},
		{ID: "op-done-3", Status: veo3.StatusDone, StartTime: time.Now()},
		{ID: "op-done-4", Status: veo3.StatusDone, StartTime: time.Now()},
		{ID: "op-failed", Status: veo3.StatusFailed, StartTime: time.Now()},
		{ID: "op-cancelled", Status: veo3.StatusCancelled, StartTime: time.Now()},
	}

	for _, op := range ops {
		manager.AddOperation(op)
	}

	stats := manager.GetOperationStats()
	assert.Equal(t, 11, stats.Total)
	assert.Equal(t, 2, stats.Pending)
	assert.Equal(t, 3, stats.Running)
	assert.Equal(t, 4, stats.Completed)
	assert.Equal(t, 1, stats.Failed)
	assert.Equal(t, 1, stats.Cancelled)
}

func TestManager_GetOperationStats_Empty(t *testing.T) {
	manager := NewManager(nil)

	stats := manager.GetOperationStats()
	assert.Equal(t, 0, stats.Total)
	assert.Equal(t, 0, stats.Pending)
	assert.Equal(t, 0, stats.Running)
	assert.Equal(t, 0, stats.Completed)
	assert.Equal(t, 0, stats.Failed)
	assert.Equal(t, 0, stats.Cancelled)
}

func TestManager_CancelOperation_NoClient(t *testing.T) {
	manager := NewManager(nil)

	err := manager.CancelOperation(context.Background(), "op-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no client available")
}

func TestManager_List_NoClient(t *testing.T) {
	manager := NewManager(nil)

	_, err := manager.List(context.Background(), veo3.StatusDone)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no client available")
}

func TestManager_GetStatus_NoClient(t *testing.T) {
	manager := NewManager(nil)

	_, err := manager.GetStatus(context.Background(), "op-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no client available")
}

func TestManager_Cancel_Alias(t *testing.T) {
	manager := NewManager(nil)

	// Cancel is an alias for CancelOperation
	err := manager.Cancel(context.Background(), "op-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no client available")
}

func TestManager_ThreadSafety(t *testing.T) {
	manager := NewManager(nil)

	// Test concurrent access
	done := make(chan bool)

	// Add operations concurrently
	for i := 0; i < 10; i++ {
		go func(id int) {
			op := &veo3.Operation{
				ID:        fmt.Sprintf("op-%d", id),
				Status:    veo3.StatusPending,
				StartTime: time.Now(),
			}
			manager.AddOperation(op)
			done <- true
		}(i)
	}

	// Wait for all additions
	for i := 0; i < 10; i++ {
		<-done
	}

	// List and verify
	list := manager.ListOperations()
	assert.Equal(t, 10, len(list))
}

func TestOperationStats_Struct(t *testing.T) {
	stats := OperationStats{
		Total:     100,
		Pending:   10,
		Running:   20,
		Completed: 60,
		Failed:    8,
		Cancelled: 2,
	}

	assert.Equal(t, 100, stats.Total)
	assert.Equal(t, 10, stats.Pending)
	assert.Equal(t, 20, stats.Running)
	assert.Equal(t, 60, stats.Completed)
	assert.Equal(t, 8, stats.Failed)
	assert.Equal(t, 2, stats.Cancelled)
}
