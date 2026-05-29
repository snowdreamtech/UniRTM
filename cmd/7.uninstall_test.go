// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUninstallCommand(t *testing.T) {
	// Create a temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize database
	ctx := context.Background()
	db, err := database.Open(ctx, database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	require.NoError(t, err)
	defer db.Close()

	// Create installation repository
	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)

	// Create a test installation
	testInstallPath := filepath.Join(tmpDir, "tools", "dummy-tool", "20.0.0")
	err = os.MkdirAll(testInstallPath, 0755)
	require.NoError(t, err)

	installation := &repository.Installation{
		Tool:        "dummy-tool",
		Version:     "20.0.0",
		Backend:     "native",
		InstallPath: testInstallPath,
		Checksum:    "abc123",
	}
	err = installRepo.Create(ctx, installation)
	require.NoError(t, err)

	// Verify installation exists
	found, err := installRepo.FindByToolAndVersion(ctx, "dummy-tool", "20.0.0")
	require.NoError(t, err)
	require.NotNil(t, found)

	// Test uninstall command structure
	t.Run("command structure", func(t *testing.T) {
		assert.Equal(t, "uninstall [tool[@version]...]", uninstallCmd.Use)
		assert.NotEmpty(t, uninstallCmd.Short)
		assert.NotEmpty(t, uninstallCmd.Long)
		assert.NotNil(t, uninstallCmd.RunE)
	})

	t.Run("command flags", func(t *testing.T) {
		forceFlag := uninstallCmd.Flags().Lookup("force")
		require.NotNil(t, forceFlag)
		assert.Equal(t, "f", forceFlag.Shorthand)
		assert.Equal(t, "false", forceFlag.DefValue)
	})
}

func TestPromptConfirmation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "yes confirmation",
			input:    "yes\n",
			expected: true,
		},
		{
			name:     "y confirmation",
			input:    "y\n",
			expected: true,
		},
		{
			name:     "Y confirmation (uppercase)",
			input:    "Y\n",
			expected: true,
		},
		{
			name:     "YES confirmation (uppercase)",
			input:    "YES\n",
			expected: true,
		},
		{
			name:     "no confirmation",
			input:    "no\n",
			expected: false,
		},
		{
			name:     "n confirmation",
			input:    "n\n",
			expected: false,
		},
		{
			name:     "empty input",
			input:    "\n",
			expected: false,
		},
		{
			name:     "random input",
			input:    "maybe\n",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stdin
			oldStdin := os.Stdin
			r, w, err := os.Pipe()
			require.NoError(t, err)
			os.Stdin = r

			// Write test input
			go func() {
				defer w.Close()
				w.Write([]byte(tt.input))
			}()

			// Redirect stdout to capture prompt
			oldStdout := os.Stdout
			_, captureW, err := os.Pipe()
			require.NoError(t, err)
			os.Stdout = captureW

			// Test confirmation
			result, err := promptConfirmation("Test prompt")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)

			// Restore stdin and stdout
			os.Stdin = oldStdin
			os.Stdout = oldStdout
			captureW.Close()
		})
	}
}

func TestUninstallCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "valid arguments with version",
			args:        []string{"dummy-tool", "20.0.0"},
			expectError: false,
		},
		{
			name:        "valid arguments with at syntax",
			args:        []string{"dummy-tool@20.0.0"},
			expectError: false,
		},
		{
			name:        "missing version is valid for args validation",
			args:        []string{"dummy-tool"},
			expectError: false,
		},
		{
			name:        "missing tool is an error (handled in RunE)",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "multiple arguments are valid",
			args:        []string{"dummy-tool@20.0.0", "go@1.22.1", "python@3.12.0"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary root command for testing
			testRootCmd := &cobra.Command{Use: "unirtm"}
			testUninstallCmd := &cobra.Command{
				Use:  "uninstall [tool[@version]...]",
				Args: cobra.ArbitraryArgs,
				RunE: func(cmd *cobra.Command, args []string) error {
					if len(args) == 0 {
						return fmt.Errorf("tool specification is required")
					}
					return nil
				},
			}
			testRootCmd.AddCommand(testUninstallCmd)

			// Capture output
			buf := new(bytes.Buffer)
			testRootCmd.SetOut(buf)
			testRootCmd.SetErr(buf)

			// Set arguments
			testRootCmd.SetArgs(append([]string{"uninstall"}, tt.args...))

			// Execute command
			err := testRootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunUninstallExecution(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	dbPath := filepath.Join(tmpDir, "unirtm.db")
	db, err := database.Open(context.Background(), database.Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	repo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)

	testInstallPath := filepath.Join(tmpDir, "tools", "dummy-tool", "20.0.0")
	err = os.MkdirAll(testInstallPath, 0755)
	require.NoError(t, err)

	err = repo.Create(context.Background(), &repository.Installation{Tool: "dummy-tool", Version: "20.0.0", Backend: "native", InstallPath: testInstallPath})
	require.NoError(t, err)

	uninstallForce = true

	cmd := uninstallCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runUninstall(cmd, []string{"dummy-tool", "20.0.0"})
	assert.NoError(t, err)

	_, err = repo.FindByToolAndVersion(context.Background(), "dummy-tool", "20.0.0")
	assert.Error(t, err) // Should be uninstalled
}

func TestRunUninstallExecution_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	dbPath := filepath.Join(tmpDir, "unirtm.db")
	db, err := database.Open(context.Background(), database.Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	repo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)

	testInstallPath := filepath.Join(tmpDir, "tools", "dummy-tool", "20.0.0")
	err = os.MkdirAll(testInstallPath, 0755)
	require.NoError(t, err)

	err = repo.Create(context.Background(), &repository.Installation{Tool: "dummy-tool", Version: "20.0.0", Backend: "native", InstallPath: testInstallPath})
	require.NoError(t, err)

	uninstallForce = true
	dryRun = true
	defer func() { dryRun = false }()

	cmd := uninstallCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runUninstall(cmd, []string{"dummy-tool", "20.0.0"})
	assert.NoError(t, err)

	found, err := repo.FindByToolAndVersion(context.Background(), "dummy-tool", "20.0.0")
	assert.NoError(t, err)
	assert.NotNil(t, found) // Still installed
}

func TestRunUninstallExecution_Multi(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	dbPath := filepath.Join(tmpDir, "unirtm.db")
	db, err := database.Open(context.Background(), database.Config{Path: dbPath})
	require.NoError(t, err)
	defer db.Close()

	repo, err := sqlite.NewInstallationRepository(db.Conn())
	require.NoError(t, err)

	testInstallPath := filepath.Join(tmpDir, "tools", "dummy-tool", "20.0.0")
	err = os.MkdirAll(testInstallPath, 0755)
	require.NoError(t, err)

	err = repo.Create(context.Background(), &repository.Installation{Tool: "dummy-tool", Version: "20.0.0", Backend: "native", InstallPath: testInstallPath})
	require.NoError(t, err)

	testInstallPath2 := filepath.Join(tmpDir, "tools", "dummy-tool2", "1.0.0")
	err = os.MkdirAll(testInstallPath2, 0755)
	require.NoError(t, err)

	err = repo.Create(context.Background(), &repository.Installation{Tool: "dummy-tool2", Version: "1.0.0", Backend: "native", InstallPath: testInstallPath2})
	require.NoError(t, err)

	uninstallForce = true

	cmd := uninstallCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = runUninstall(cmd, []string{"dummy-tool@20.0.0", "dummy-tool2@1.0.0"})
	assert.NoError(t, err)

	_, err = repo.FindByToolAndVersion(context.Background(), "dummy-tool", "20.0.0")
	assert.Error(t, err)

	_, err = repo.FindByToolAndVersion(context.Background(), "dummy-tool2", "1.0.0")
	assert.Error(t, err)
}
