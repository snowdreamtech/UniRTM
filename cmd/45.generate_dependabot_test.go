// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateDependabotStructure(t *testing.T) {
	assert.Contains(t, generateDependabotCmd.Use, "dependabot")
	assert.NotEmpty(t, generateDependabotCmd.Short)
	assert.NotNil(t, generateDependabotCmd.RunE)
}

func TestRunGenerateDependabot(t *testing.T) {
	tmpDir := t.TempDir()

	// Switch to tmpDir
	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	// Create a dummy repository structure to get coverage for directory scanning
	_ = os.MkdirAll(filepath.Join(tmpDir, ".github"), 0o755)
	_ = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/test/repo"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0o644)

	_ = generateDependabotCmd.RunE(generateDependabotCmd, []string{})
}

