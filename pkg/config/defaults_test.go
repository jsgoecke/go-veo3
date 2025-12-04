package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	// Test that default values are as expected
	assert.Equal(t, "veo-3.1-generate-preview", DefaultModel)
	assert.Equal(t, "720p", DefaultResolution)
	assert.Equal(t, "16:9", DefaultAspectRatio)
	assert.Equal(t, 8, DefaultDuration)
	assert.Equal(t, 10, DefaultPollInterval)
	assert.Equal(t, 3, DefaultConcurrency)
	assert.Equal(t, "1.0", DefaultConfigVersion)
}

func TestConstraintConstants(t *testing.T) {
	// Test constraint values
	assert.Equal(t, 20*1024*1024, MaxImageSize)
	assert.Equal(t, 141, MaxVideoLength)
	assert.Equal(t, 1024, MaxPromptLength)
	assert.Equal(t, 3, MaxReferenceImages)
}

func TestDefaultValuesAreReasonable(t *testing.T) {
	// Sanity checks for default values
	assert.Greater(t, DefaultPollInterval, 0, "Poll interval should be positive")
	assert.Greater(t, DefaultConcurrency, 0, "Concurrency should be positive")
	assert.Greater(t, DefaultDuration, 0, "Duration should be positive")
	assert.Greater(t, MaxImageSize, 0, "Max image size should be positive")
	assert.Greater(t, MaxVideoLength, 0, "Max video length should be positive")
	assert.Greater(t, MaxPromptLength, 0, "Max prompt length should be positive")
	assert.Greater(t, MaxReferenceImages, 0, "Max reference images should be positive")
}

func TestMaxReferenceImagesMatchesConstraint(t *testing.T) {
	// Ensure MaxReferenceImages constant matches DefaultConcurrency is reasonable
	assert.LessOrEqual(t, DefaultConcurrency, 10, "Default concurrency should be reasonable")
	assert.LessOrEqual(t, MaxReferenceImages, 5, "Max reference images should be reasonable")
}
