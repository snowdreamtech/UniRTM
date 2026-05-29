// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvCommand_InjectsProviderEnvVars(t *testing.T) {
	// Create a temp directory for our mock project
	tempDir, err := os.MkdirTemp("", "unirtm-env-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Write a mock .unirtm.toml
	tomlContent := `
[tools]
go = "1.26.3"
`
	err = os.WriteFile(filepath.Join(tempDir, ".unirtm.toml"), []byte(tomlContent), 0644)
	assert.NoError(t, err)

	// Mock the installs directory so GetEnvVars doesn't panic or fail
	installsDir := filepath.Join(tempDir, "installs")
	t.Setenv("UNIRTM_INSTALLS_DIR", installsDir)

	// Change working directory to the temp dir so config.LoadFull picks it up
	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	// Temporarily set the database path to a dummy memory DB or temp file to avoid nil pointers
	t.Setenv("UNIRTM_DATABASE_PATH", filepath.Join(tempDir, "unirtm.db"))

	out := captureStdoutFunc(t, func() {
		// Run the env command in shell "bash" mode
		rootCmd.SetArgs([]string{"env", "--shell", "bash"})
		_ = rootCmd.Execute()
	})

	// Assert that GOROOT is present in the output
	assert.Contains(t, out, "GOROOT")
	assert.Contains(t, out, "1.26.3")
}
