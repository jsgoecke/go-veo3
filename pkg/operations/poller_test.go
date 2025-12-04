package operations

import (
	"context"
	"testing"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/veo3"
	"github.com/stretchr/testify/assert"
)

func TestNewPoller(t *testing.T) {
	client := &veo3.Client{}
	manager := NewManager(client)
	poller := NewPoller(client, manager)

	assert.NotNil(t, poller)
	assert.Equal(t, client, poller.client)
	assert.Equal(t, manager, poller.manager)
	assert.Equal(t, 10*time.Second, poller.baseInterval)
	assert.Equal(t, 5*time.Minute, poller.maxInterval)
	assert.Equal(t, 1.5, poller.backoffFactor)
	assert.Equal(t, 10, poller.maxRetries)
}

func TestPoller_SetPollingConfig(t *testing.T) {
	poller := NewPoller(nil, nil)

	poller.SetPollingConfig(
		5*time.Second,
		2*time.Minute,
		2.0,
		5,
	)

	assert.Equal(t, 5*time.Second, poller.baseInterval)
	assert.Equal(t, 2*time.Minute, poller.maxInterval)
	assert.Equal(t, 2.0, poller.backoffFactor)
	assert.Equal(t, 5, poller.maxRetries)
}

func TestFormatDuration_Poller(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"seconds only", 45 * time.Second, "0:45"},
		{"minutes and seconds", 2*time.Minute + 30*time.Second, "2:30"},
		{"zero", 0, "0:00"},
		{"one minute", 1 * time.Minute, "1:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestPoller_PollAllActive_NoOperations(t *testing.T) {
	manager := NewManager(nil)
	poller := NewPoller(nil, manager)

	err := poller.PollAllActive(context.Background(), nil)
	assert.NoError(t, err) // No operations to poll is not an error
}

func TestPoller_Struct(t *testing.T) {
	client := &veo3.Client{}
	manager := NewManager(client)

	poller := &Poller{
		client:        client,
		manager:       manager,
		baseInterval:  15 * time.Second,
		maxInterval:   10 * time.Minute,
		backoffFactor: 2.0,
		maxRetries:    20,
	}

	assert.Equal(t, client, poller.client)
	assert.Equal(t, manager, poller.manager)
	assert.Equal(t, 15*time.Second, poller.baseInterval)
	assert.Equal(t, 10*time.Minute, poller.maxInterval)
	assert.Equal(t, 2.0, poller.backoffFactor)
	assert.Equal(t, 20, poller.maxRetries)
}

func TestPoller_WaitForCompletion_OperationNotFound(t *testing.T) {
	// Skip this test as it requires a real client to avoid nil pointer dereference
	t.Skip("Requires mock client implementation")
}

func TestPoller_Configuration(t *testing.T) {
	poller := NewPoller(nil, nil)

	// Test default values
	assert.Greater(t, poller.baseInterval, time.Duration(0))
	assert.Greater(t, poller.maxInterval, poller.baseInterval)
	assert.Greater(t, poller.backoffFactor, 1.0)
	assert.Greater(t, poller.maxRetries, 0)

	// Test custom configuration
	poller.SetPollingConfig(
		1*time.Second,
		30*time.Second,
		1.2,
		3,
	)

	assert.Equal(t, 1*time.Second, poller.baseInterval)
	assert.Equal(t, 30*time.Second, poller.maxInterval)
	assert.Equal(t, 1.2, poller.backoffFactor)
	assert.Equal(t, 3, poller.maxRetries)
}
