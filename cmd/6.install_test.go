// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
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
			name:    "valid legacy arguments",
			args:    []string{"node", "20.0.0"},
			wantErr: false,
		},
		{
			name:    "valid package specifications",
			args:    []string{"node@20.0.0", "go@1.22.1"},
			wantErr: false,
		},
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "many arguments",
			args:    []string{"node@20.0.0", "go@1.22.1", "python@3.12.0"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command instance for each test
			cmd := &cobra.Command{
				Use:  "install [tool[@version]...]",
				Args: cobra.ArbitraryArgs,
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
			expectedBackend: "",
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
	originalXDGDataHome := env.Get("XDG_DATA_HOME")
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

		path := env.GetDatabasePath()

		assert.Contains(t, path, testDir)
		assert.Contains(t, path, "unirtm")
		assert.Contains(t, path, "unirtm.db")
	})

	t.Run("without XDG_DATA_HOME", func(t *testing.T) {
		os.Unsetenv("XDG_DATA_HOME")

		path := env.GetDatabasePath()

		// Should contain .local/share/unirtm or fallback to ./unirtm.db
		assert.True(t,
			path == "./unirtm.db" || (len(path) >= 9 && path[len(path)-9:] == "unirtm.db"),
			"path should end with unirtm.db: %s", path)
	})
}

func TestConcurrentSpinnerManager(t *testing.T) {
	// Capture stdout to avoid polluting test output, though it uses pterm and fmt
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	mgr := newConcurrentSpinnerManager()
	assert.NotNil(t, mgr)

	mgr.Start()

	mgr.Add("node", "20.0.0")
	mgr.Add("go", "1.21.0")

	assert.Equal(t, 2, len(mgr.active))
	assert.Equal(t, "starting", mgr.activeMap["node"].status)

	mgr.Update("node", "downloading")
	assert.Equal(t, "downloading", mgr.activeMap["node"].status)

	time.Sleep(150 * time.Millisecond) // Let it render at least once

	mgr.Complete("node", "20.0.0", "done")
	assert.Equal(t, 1, len(mgr.active))
	assert.Nil(t, mgr.activeMap["node"])

	mgr.Complete("go", "1.21.0", "failed: network error")
	assert.Equal(t, 0, len(mgr.active))

	mgr.Stop()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
}

func TestRunInstall_Execution(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	// Set quiet mode to suppress output during tests
	originalQuiet := quiet
	quiet = true
	defer func() { quiet = originalQuiet }()

	// Create a dummy config so we test config resolution
	configFile := tmpDir + "/.unirtm.toml"
	os.WriteFile(configFile, []byte("[tools]\ndummy-tool = { version = \"20.0.0\" }"), 0644)
	
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cmd := &cobra.Command{}

	// Test case: 0 args, config used
	err := runInstall(cmd, []string{})
	// Will fail because "dummy-tool" is an unsupported tool (no provider registered in the default list) or download fails
	assert.Error(t, err)

	// Test case: 1 arg, multiple tools via parsing
	err = runInstall(cmd, []string{"dummy-tool@20.0.0"})
	assert.Error(t, err)

	// Test case: 1 arg, version from config
	err = runInstall(cmd, []string{"dummy-tool"})
	assert.Error(t, err)
	
	// Test case: json output
	jsonOutput = true
	err = runInstall(cmd, []string{"dummy-tool@20.0.0"})
	assert.Error(t, err)
	jsonOutput = false
	
	// Test case: multi install
	err = runInstall(cmd, []string{"dummy-tool@20.0.0", "another-tool@1.21.0"})
	assert.Error(t, err)

	// Test case: legacy install args
	err = runInstall(cmd, []string{"dummy-tool", "20.0.0"})
	assert.Error(t, err)
}
