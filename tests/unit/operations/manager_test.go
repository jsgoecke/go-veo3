package operations_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/operations"
	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_ListOperations(t *testing.T) {
	manager := operations.NewManager()

	// Initially empty
	ops := manager.ListOperations()
	assert.Empty(t, ops)

	// Add some operations
	op1 := &veo3.Operation{
		ID:        "operations/test-op-1",
		Status:    veo3.StatusRunning,
		StartTime: time.Now(),
	}
	op2 := &veo3.Operation{
		ID:        "operations/test-op-2",
		Status:    veo3.StatusDone,
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   timePtr(time.Now().Add(-30 * time.Minute)),
		VideoURI:  "gs://bucket/video.mp4",
	}

	manager.AddOperation(op1)
	manager.AddOperation(op2)

	// List should return both operations
	ops = manager.ListOperations()
	assert.Len(t, ops, 2)

	// Should be ordered by start time (newest first)
	assert.Equal(t, "operations/test-op-1", ops[0].ID)
	assert.Equal(t, "operations/test-op-2", ops[1].ID)
}

func TestManager_GetOperation(t *testing.T) {
	manager := operations.NewManager()

	// Get non-existent operation
	op, err := manager.GetOperation("operations/nonexistent")
	assert.Error(t, err)
	assert.Nil(t, op)
	assert.Contains(t, err.Error(), "not found")

	// Add an operation
	testOp := &veo3.Operation{
		ID:        "operations/test-op",
		Status:    veo3.StatusRunning,
		StartTime: time.Now(),
	}
	manager.AddOperation(testOp)

	// Get existing operation
	op, err = manager.GetOperation("operations/test-op")
	assert.NoError(t, err)
	require.NotNil(t, op)
	assert.Equal(t, "operations/test-op", op.ID)
	assert.Equal(t, veo3.StatusRunning, op.Status)
}

func TestManager_UpdateOperation(t *testing.T) {
	manager := operations.NewManager()

	// Add initial operation
	testOp := &veo3.Operation{
		ID:        "operations/test-op",
		Status:    veo3.StatusPending,
		StartTime: time.Now(),
	}
	manager.AddOperation(testOp)

	// Update operation status
	updatedOp := &veo3.Operation{
		ID:        "operations/test-op",
		Status:    veo3.StatusDone,
		StartTime: testOp.StartTime,
		EndTime:   timePtr(time.Now()),
		VideoURI:  "gs://bucket/completed.mp4",
	}

	err := manager.UpdateOperation(updatedOp)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := manager.GetOperation("operations/test-op")
	assert.NoError(t, err)
	assert.Equal(t, veo3.StatusDone, retrieved.Status)
	assert.Equal(t, "gs://bucket/completed.mp4", retrieved.VideoURI)
	assert.NotNil(t, retrieved.EndTime)
}

func TestManager_RemoveOperation(t *testing.T) {
	manager := operations.NewManager()

	// Add operation
	testOp := &veo3.Operation{
		ID:        "operations/test-op",
		Status:    veo3.StatusDone,
		StartTime: time.Now(),
	}
	manager.AddOperation(testOp)

	// Verify it exists
	ops := manager.ListOperations()
	assert.Len(t, ops, 1)

	// Remove operation
	err := manager.RemoveOperation("operations/test-op")
	assert.NoError(t, err)

	// Verify it's gone
	ops = manager.ListOperations()
	assert.Empty(t, ops)

	// Try to remove non-existent operation
	err = manager.RemoveOperation("operations/nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_GetOperationStats(t *testing.T) {
	manager := operations.NewManager()

	// Initially empty stats
	stats := manager.GetOperationStats()
	assert.Equal(t, 0, stats.Total)
	assert.Equal(t, 0, stats.Pending)
	assert.Equal(t, 0, stats.Running)
	assert.Equal(t, 0, stats.Completed)
	assert.Equal(t, 0, stats.Failed)
	assert.Equal(t, 0, stats.Cancelled)

	// Add operations with different statuses
	operations := []*veo3.Operation{
		{ID: "op1", Status: veo3.StatusPending, StartTime: time.Now()},
		{ID: "op2", Status: veo3.StatusRunning, StartTime: time.Now()},
		{ID: "op3", Status: veo3.StatusDone, StartTime: time.Now()},
		{ID: "op4", Status: veo3.StatusFailed, StartTime: time.Now()},
		{ID: "op5", Status: veo3.StatusCancelled, StartTime: time.Now()},
		{ID: "op6", Status: veo3.StatusDone, StartTime: time.Now()},
	}

	for _, op := range operations {
		manager.AddOperation(op)
	}

	// Check stats
	stats = manager.GetOperationStats()
	assert.Equal(t, 6, stats.Total)
	assert.Equal(t, 1, stats.Pending)
	assert.Equal(t, 1, stats.Running)
	assert.Equal(t, 2, stats.Completed)
	assert.Equal(t, 1, stats.Failed)
	assert.Equal(t, 1, stats.Cancelled)
}

func TestManager_FilterOperations(t *testing.T) {
	manager := operations.NewManager()

	// Add operations with different statuses
	operations := []*veo3.Operation{
		{ID: "op1", Status: veo3.StatusPending, StartTime: time.Now()},
		{ID: "op2", Status: veo3.StatusRunning, StartTime: time.Now()},
		{ID: "op3", Status: veo3.StatusDone, StartTime: time.Now()},
		{ID: "op4", Status: veo3.StatusFailed, StartTime: time.Now()},
	}

	for _, op := range operations {
		manager.AddOperation(op)
	}

	// Filter by status
	runningOps := manager.FilterOperations(veo3.StatusRunning)
	assert.Len(t, runningOps, 1)
	assert.Equal(t, "op2", runningOps[0].ID)

	doneOps := manager.FilterOperations(veo3.StatusDone)
	assert.Len(t, doneOps, 1)
	assert.Equal(t, "op3", doneOps[0].ID)

	// Filter by non-existent status
	cancelledOps := manager.FilterOperations(veo3.StatusCancelled)
	assert.Empty(t, cancelledOps)
}

func TestManager_ConcurrentAccess(t *testing.T) {
	manager := operations.NewManager()

	// Test concurrent add/get operations
	done := make(chan bool, 2)

	// Goroutine 1: Add operations
	go func() {
		for i := 0; i < 10; i++ {
			op := &veo3.Operation{
				ID:        fmt.Sprintf("operations/concurrent-%d", i),
				Status:    veo3.StatusRunning,
				StartTime: time.Now(),
			}
			manager.AddOperation(op)
		}
		done <- true
	}()

	// Goroutine 2: Read operations
	go func() {
		for i := 0; i < 10; i++ {
			manager.ListOperations()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Verify final state
	ops := manager.ListOperations()
	assert.Len(t, ops, 10)
}

// Helper function
func timePtr(t time.Time) *time.Time {
	return &t
}
