// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInstallCommand_ArgumentParsing tests that the install command correctly parses arguments.
func TestInstallCommand_ArgumentParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid arguments",
			args:    []string{"node", "20.0.0"},
			wantErr: false,
		},
		{
			name:        "missing version",
			args:        []string{"node"},
			wantErr:     true,
			errContains: "accepts 2 arg(s), received 1",
		},
		{
			name:        "missing tool and version",
			args:        []string{},
			wantErr:     true,
			errContains: "accepts 2 arg(s), received 0",
		},
		{
			name:        "too many arguments",
			args:        []string{"node", "20.0.0", "extra"},
			wantErr:     true,
			errContains: "accepts 2 arg(s), received 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command instance for each test
			cmd := &cobra.Command{
				Use:  "install <tool> <version>",
				Args: cobra.ExactArgs(2),
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}

			// Set arguments
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()

			// Check error
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestInstallCommand_Validation tests input validation.
func TestInstallCommand_Validation(t *testing.T) {
	tests := []struct {
		name        string
		tool        string
		version     string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty tool name",
			tool:        "",
			version:     "20.0.0",
			wantErr:     true,
			errContains: "tool name is required",
		},
		{
			name:        "empty version",
			tool:        "node",
			version:     "",
			wantErr:     true,
			errContains: "version is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			cmd := &cobra.Command{}
			args := []string{tt.tool, tt.version}

			// Set quiet mode to suppress output during tests
			originalQuiet := quiet
			quiet = true
			defer func() { quiet = originalQuiet }()

			// Redirect output
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = originalStdout }()

			// Run the install command logic (validation only)
			err := runInstall(cmd, args)

			// Close writer and read output
			w.Close()
			_, _ = buf.ReadFrom(r)

			// All validation cases expect errors
			require.Error(t, err)
			if tt.errContains != "" {
				assert.Contains(t, err.Error(), tt.errContains)
			}
		})
	}
}

// TestInstallCommand_BackendFlag tests the --backend flag.
func TestInstallCommand_BackendFlag(t *testing.T) {
	tests := []struct {
		name            string
		backendFlag     string
		expectedBackend string
	}{
		{
			name:            "default backend",
			backendFlag:     "",
			expectedBackend: "github",
		},
		{
			name:            "custom backend",
			backendFlag:     "aqua",
			expectedBackend: "aqua",
		},
		{
			name:            "http backend",
			backendFlag:     "http",
			expectedBackend: "http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the backend flag
			installBackend = tt.backendFlag

			// Get the backend name
			backendName := getBackendName()

			// Check the result
			assert.Equal(t, tt.expectedBackend, backendName)
		})
	}
}

// TestInstallCommand_OutputFormat tests output format selection.
func TestInstallCommand_OutputFormat(t *testing.T) {
	tests := []struct {
		name           string
		jsonFlag       bool
		expectedFormat string
	}{
		{
			name:           "human format",
			jsonFlag:       false,
			expectedFormat: "human",
		},
		{
			name:           "json format",
			jsonFlag:       true,
			expectedFormat: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the json flag
			jsonOutput = tt.jsonFlag

			// Get the output format
			format := getOutputFormat()

			// Check the result
			assert.Equal(t, tt.expectedFormat, string(format))
		})
	}
}

// TestGetDefaultDatabasePath tests the default database path generation.
func TestGetDefaultDatabasePath(t *testing.T) {
	// Save original environment
	originalXDGDataHome := os.Getenv("XDG_DATA_HOME")
	defer func() {
		if originalXDGDataHome != "" {
			os.Setenv("XDG_DATA_HOME", originalXDGDataHome)
		} else {
			os.Unsetenv("XDG_DATA_HOME")
		}
	}()

	t.Run("with XDG_DATA_HOME set", func(t *testing.T) {
		testDir := "/tmp/test-xdg-data"
		os.Setenv("XDG_DATA_HOME", testDir)

		path := getDefaultDatabasePath()

		assert.Contains(t, path, testDir)
		assert.Contains(t, path, "unirtm")
		assert.Contains(t, path, "unirtm.db")
	})

	t.Run("without XDG_DATA_HOME", func(t *testing.T) {
		os.Unsetenv("XDG_DATA_HOME")

		path := getDefaultDatabasePath()

		// Should contain .local/share/unirtm or fallback to ./unirtm.db
		assert.True(t,
			path == "./unirtm.db" || (len(path) >= 9 && path[len(path)-9:] == "unirtm.db"),
			"path should end with unirtm.db: %s", path)
	})
}

// TestInstallCommand_Integration tests the full install command integration.
// This is a placeholder for integration tests that would require a real database and backend.
func TestInstallCommand_Integration(t *testing.T) {
	t.Skip("Integration test requires database and backend setup")

	// TODO: Implement integration tests with:
	// - Mock database
	// - Mock backend
	// - Mock provider
	// - Mock download manager
	// - Verify full workflow: parse → validate → download → install → record
}
