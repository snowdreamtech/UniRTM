// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelfUpdateStructure(t *testing.T) {
	assert.Contains(t, selfUpdateCmd.Use, "self-update", "selfUpdateCmd command use should contain 'self-update'")
	assert.NotEmpty(t, selfUpdateCmd.Short, "selfUpdateCmd command short description should not be empty")
	assert.True(t, selfUpdateCmd.Run != nil || selfUpdateCmd.RunE != nil, "Run or RunE function should be set for selfUpdateCmd")
}

func TestRunSelfUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	cmd := selfUpdateCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Set selfUpdateYes to true to skip confirmation prompt
	prevYes := selfUpdateYes
	selfUpdateYes = true
	t.Cleanup(func() { selfUpdateYes = prevYes })

	// Mock fetchGitHubRelease to avoid network calls
	prevFetch := fetchGitHubRelease
	fetchGitHubRelease = func(version string) (*githubRelease, error) {
		return &githubRelease{TagName: "v1.0.0", Name: "v1.0.0", Body: "Dummy"}, nil
	}
	t.Cleanup(func() { fetchGitHubRelease = prevFetch })

	prevExecCommand := execCommand
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command(os.Args[0], "-h")
	}
	t.Cleanup(func() { execCommand = prevExecCommand })

	err := runSelfUpdate(cmd, []string{})
	assert.NoError(t, err)
}
