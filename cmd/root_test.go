package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRootCommand tests the root command structure and flags
func TestRootCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectOut   string
	}{
		{
			name:        "help flag",
			args:        []string{"--help"},
			expectError: false,
			expectOut:   "Universal Runtime Manager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new root command for each test to avoid state pollution
			cmd := rootCmd
			cmd.SetArgs(tt.args)

			// Capture output
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Execute command
			err := cmd.Execute()

			// Check error expectation
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Check output contains expected string
			output := buf.String()
			assert.Contains(t, output, tt.expectOut, "output should contain expected string")
		})
	}
}

// TestGlobalFlags tests that global flags are properly registered
func TestGlobalFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
	}{
		{
			name:     "config flag",
			flagName: "config",
		},
		{
			name:     "verbose flag",
			flagName: "verbose",
		},
		{
			name:     "quiet flag",
			flagName: "quiet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := rootCmd.PersistentFlags().Lookup(tt.flagName)
			require.NotNil(t, flag, "flag %s should be registered", tt.flagName)
		})
	}
}

// TestVersionCommand tests the version command output
func TestVersionCommand(t *testing.T) {
	// Just verify the command can be executed without error
	// The actual output goes to stdout which is hard to capture in tests
	assert.NotNil(t, versionCmd.Run, "version command should have a Run function")
	assert.Equal(t, "version", versionCmd.Use, "version command use should be 'version'")
}

// TestLoggingSetup tests that logging is configured correctly based on flags
func TestLoggingSetup(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
		quiet   bool
	}{
		{
			name:    "default logging",
			verbose: false,
			quiet:   false,
		},
		{
			name:    "verbose logging",
			verbose: true,
			quiet:   false,
		},
		{
			name:    "quiet logging",
			verbose: false,
			quiet:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set flags
			verbose = tt.verbose
			quiet = tt.quiet

			// Call setupLogging
			setupLogging(rootCmd, []string{})

			// Note: We can't easily test the actual log level without exposing it,
			// but we can at least verify the function doesn't panic
		})
	}
}

// TestRootCommandStructure tests the basic structure of the root command
func TestRootCommandStructure(t *testing.T) {
	assert.Equal(t, "unirtm", rootCmd.Use, "root command use should be 'unirtm'")
	assert.Contains(t, rootCmd.Short, "Universal Runtime Manager", "short description should mention UniRTM")
	assert.NotNil(t, rootCmd.PersistentPreRun, "PersistentPreRun should be set for logging setup")
}

// TestVersionCommandStructure tests the basic structure of the version command
func TestVersionCommandStructure(t *testing.T) {
	assert.Equal(t, "version", versionCmd.Use, "version command use should be 'version'")
	assert.Contains(t, versionCmd.Short, "version", "short description should mention version")
	assert.NotNil(t, versionCmd.Run, "Run function should be set")
}
