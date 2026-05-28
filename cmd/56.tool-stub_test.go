package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunToolStub_FailsReadingFile(t *testing.T) {
	err := runToolStub(toolStubCmd, []string{"/path/does/not/exist"})
	assert.Error(t, err)
}

func TestRunToolStub_FailsParsingTOML(t *testing.T) {
	tmpDir := t.TempDir()
	invalidTomlPath := filepath.Join(tmpDir, "invalid")

	content := `#!/bin/sh
# UNIRTM_TOOL_STUB:
# invalid_toml ==== "mytool"
# :UNIRTM_TOOL_STUB
`
	err := os.WriteFile(invalidTomlPath, []byte(content), 0755)
	assert.NoError(t, err)

	err = runToolStub(toolStubCmd, []string{invalidTomlPath})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse stub TOML")
}
