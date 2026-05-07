package cmd

import (
	"bytes"
	"context"
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
	testInstallPath := filepath.Join(tmpDir, "tools", "node", "20.0.0")
	err = os.MkdirAll(testInstallPath, 0755)
	require.NoError(t, err)

	installation := &repository.Installation{
		Tool:        "node",
		Version:     "20.0.0",
		Backend:     "github",
		InstallPath: testInstallPath,
		Checksum:    "abc123",
	}
	err = installRepo.Create(ctx, installation)
	require.NoError(t, err)

	// Verify installation exists
	found, err := installRepo.FindByToolAndVersion(ctx, "node", "20.0.0")
	require.NoError(t, err)
	require.NotNil(t, found)

	// Test uninstall command structure
	t.Run("command structure", func(t *testing.T) {
		assert.Equal(t, "uninstall <tool> <version>", uninstallCmd.Use)
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
			name:        "valid arguments",
			args:        []string{"node", "20.0.0"},
			expectError: false,
		},
		{
			name:        "missing version",
			args:        []string{"node"},
			expectError: true,
		},
		{
			name:        "missing tool and version",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "too many arguments",
			args:        []string{"node", "20.0.0", "extra"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary root command for testing
			testRootCmd := &cobra.Command{Use: "unirtm"}
			testUninstallCmd := &cobra.Command{
				Use:  "uninstall <tool> <version>",
				Args: cobra.ExactArgs(2),
				RunE: func(cmd *cobra.Command, args []string) error {
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
