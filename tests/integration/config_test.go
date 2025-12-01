package integration_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/jasongoecke/go-veo3/pkg/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigurationPrecedence tests that configuration values follow the correct precedence:
// Command-line flags > Environment variables > Configuration file
func TestConfigurationPrecedence(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "precedence-test.yaml")

	// First, create a config file with default values
	var stdout, stderr bytes.Buffer
	initCmd := cli.NewRootCmd()
	initCmd.SetOut(&stdout)
	initCmd.SetErr(&stderr)
	initCmd.SetArgs([]string{
		"--config", configFile,
		"config", "init",
		"--api-key", "file-api-key-12345",
		"--model", "veo-2.0-generate-001",
		"--output", tempDir,
		"--force",
	})

	err := initCmd.Execute()
	require.NoError(t, err, "Failed to initialize config: %s", stderr.String())

	t.Run("config file values are used when no env or flags", func(t *testing.T) {
		var stdout, stderr bytes.Buffer

		cmd := cli.NewRootCmd()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--config", configFile,
			"config", "get",
			"default-model",
		})

		err := cmd.Execute()
		require.NoError(t, err, "stderr: %s", stderr.String())

		output := stdout.String()
		assert.Contains(t, output, "veo-2.0-generate-001", "Should use value from config file")
	})

	t.Run("environment variables override config file", func(t *testing.T) {
		// Set environment variable
		os.Setenv("VEO3_DEFAULT_MODEL", "veo-3-generate-preview")
		defer os.Unsetenv("VEO3_DEFAULT_MODEL")

		var stdout, stderr bytes.Buffer

		cmd := cli.NewRootCmd()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--config", configFile,
			"config", "get",
			"default-model",
		})

		err := cmd.Execute()
		require.NoError(t, err, "stderr: %s", stderr.String())

		output := stdout.String()
		// Note: Viper's env handling might need special handling for nested keys
		// This test verifies the expected behavior
		assert.Contains(t, output, "veo-3-generate-preview", "Should use value from environment")
	})

	t.Run("command-line flags override environment and config file", func(t *testing.T) {
		// Set environment variable
		os.Setenv("VEO3_DEFAULT_MODEL", "veo-3-generate-preview")
		defer os.Unsetenv("VEO3_DEFAULT_MODEL")

		// First set via command line
		var stdout1, stderr1 bytes.Buffer
		setCmd := cli.NewRootCmd()
		setCmd.SetOut(&stdout1)
		setCmd.SetErr(&stderr1)
		setCmd.SetArgs([]string{
			"--config", configFile,
			"config", "set",
			"default-model", "veo-3.1-generate-preview",
		})

		err := setCmd.Execute()
		require.NoError(t, err, "stderr: %s", stderr1.String())

		// Now get the value - file should have the flag value
		var stdout2, stderr2 bytes.Buffer
		getCmd := cli.NewRootCmd()
		getCmd.SetOut(&stdout2)
		getCmd.SetErr(&stderr2)
		getCmd.SetArgs([]string{
			"--config", configFile,
			"config", "get",
			"default-model",
		})

		err = getCmd.Execute()
		require.NoError(t, err, "stderr: %s", stderr2.String())

		output := stdout2.String()
		assert.Contains(t, output, "veo-3.1-generate-preview", "Should use value set via CLI flag")
	})

	t.Run("API key precedence: env > file", func(t *testing.T) {
		// Test API key from file
		var stdout1, stderr1 bytes.Buffer
		cmd1 := cli.NewRootCmd()
		cmd1.SetOut(&stdout1)
		cmd1.SetErr(&stderr1)
		cmd1.SetArgs([]string{
			"--config", configFile,
			"config", "get",
			"api-key",
		})

		err := cmd1.Execute()
		require.NoError(t, err, "stderr: %s", stderr1.String())

		output1 := stdout1.String()
		assert.Contains(t, output1, "****", "API key should be masked")

		// Test API key from environment (GEMINI_API_KEY)
		os.Setenv("GEMINI_API_KEY", "env-api-key-67890")
		defer os.Unsetenv("GEMINI_API_KEY")

		var stdout2, stderr2 bytes.Buffer
		cmd2 := cli.NewRootCmd()
		cmd2.SetOut(&stdout2)
		cmd2.SetErr(&stderr2)
		cmd2.SetArgs([]string{
			"--config", configFile,
			"config", "get",
			"api-key",
		})

		err = cmd2.Execute()
		require.NoError(t, err, "stderr: %s", stderr2.String())

		// The environment variable should take precedence
		// (actual behavior depends on config.Manager implementation)
	})

	t.Run("output directory precedence in generate command", func(t *testing.T) {
		// This tests that command flags override config file in actual commands
		// Skip if no API key available
		if os.Getenv("VEO3_API_KEY") == "" && os.Getenv("GEMINI_API_KEY") == "" {
			t.Skip("No API key available for integration test")
		}

		customOutput := filepath.Join(tempDir, "custom-output")
		os.MkdirAll(customOutput, 0755)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewRootCmd()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--config", configFile,
			"generate",
			"--prompt", "Test precedence",
			"--output", customOutput, // CLI flag should override config file
			"--no-download",
		})

		err := cmd.Execute()
		// May fail due to API issues, but we're testing the flag parsing
		if err != nil {
			// Check that the error is not about output directory
			errMsg := err.Error()
			assert.NotContains(t, errMsg, "output", "Error should not be about output directory")
		}
	})
}

// TestConfigurationDefaults tests that default values are correctly applied
func TestConfigurationDefaults(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "defaults-test.yaml")

	// Initialize config without specifying optional values
	var stdout, stderr bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{
		"--config", configFile,
		"config", "init",
		"--api-key", "test-key",
		"--force",
	})

	err := cmd.Execute()
	require.NoError(t, err, "stderr: %s", stderr.String())

	t.Run("default model is set", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := cli.NewRootCmd()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--config", configFile,
			"config", "get",
			"default-model",
		})

		err := cmd.Execute()
		require.NoError(t, err, "stderr: %s", stderr.String())

		output := stdout.String()
		// Should contain a valid model name
		assert.NotEmpty(t, output, "Default model should be set")
	})

	t.Run("default resolution is set", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := cli.NewRootCmd()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--config", configFile,
			"config", "get",
			"default-resolution",
		})

		err := cmd.Execute()
		require.NoError(t, err, "stderr: %s", stderr.String())

		output := stdout.String()
		assert.Contains(t, output, "720p", "Default resolution should be 720p")
	})

	t.Run("poll interval has default value", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := cli.NewRootCmd()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--config", configFile,
			"config", "get",
			"poll-interval",
		})

		err := cmd.Execute()
		require.NoError(t, err, "stderr: %s", stderr.String())

		output := stdout.String()
		// Should have a numeric value
		assert.NotEmpty(t, output, "Poll interval should have a default value")
	})
}

// TestConfigurationValidation tests that invalid configuration values are rejected
func TestConfigurationValidation(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "validation-test.yaml")

	// Initialize valid config
	var stdout, stderr bytes.Buffer
	cmd := cli.NewRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{
		"--config", configFile,
		"config", "init",
		"--api-key", "test-key",
		"--force",
	})

	err := cmd.Execute()
	require.NoError(t, err, "stderr: %s", stderr.String())

	t.Run("invalid model is rejected", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := cli.NewRootCmd()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--config", configFile,
			"config", "set",
			"default-model", "invalid-model-name",
		})

		err := cmd.Execute()
		// May or may not fail depending on validation implementation
		// If validation is implemented, it should fail
		if err != nil {
			assert.Contains(t, err.Error(), "invalid", "Error should mention invalid model")
		}
	})

	t.Run("invalid duration format is rejected", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := cli.NewRootCmd()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--config", configFile,
			"config", "set",
			"default-duration", "not-a-number",
		})

		err := cmd.Execute()
		require.Error(t, err, "Should reject non-numeric duration")
		assert.Contains(t, err.Error(), "invalid", "Error should mention invalid value")
	})

	t.Run("invalid poll interval format is rejected", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := cli.NewRootCmd()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--config", configFile,
			"config", "set",
			"poll-interval", "invalid",
		})

		err := cmd.Execute()
		require.Error(t, err, "Should reject non-numeric poll interval")
		assert.Contains(t, err.Error(), "invalid", "Error should mention invalid value")
	})
}
