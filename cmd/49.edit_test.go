package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunEdit_WithFileArg(t *testing.T) {
	tmpDir := t.TempDir()

	// Fake editor script that appends some text
	editorScript := filepath.Join(tmpDir, "fake_editor.sh")
	scriptContent := `#!/bin/sh
# $1 is the file to edit
echo "key = 'value'" >> "$1"
`
	err := os.WriteFile(editorScript, []byte(scriptContent), 0755)
	assert.NoError(t, err)

	os.Setenv("UNIRTM_EDITOR", editorScript)
	defer os.Unsetenv("UNIRTM_EDITOR")

	targetFile := filepath.Join(tmpDir, "test.toml")

	// Run edit command passing the specific file
	err = runEdit(editCmd, []string{targetFile})
	assert.NoError(t, err)

	// Check if file was modified correctly
	data, err := os.ReadFile(targetFile)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "key = 'value'")
}

func TestDiscoverConfigFiles(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create dummy files
	os.WriteFile(filepath.Join(tmpDir, ".unirtm.toml"), []byte(""), 0644)
	
	// Change working directory to tmpDir temporarily
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	candidates := discoverConfigFiles()
	
	found := false
	for _, c := range candidates {
		if c.Name == ".unirtm.toml" {
			found = true
			break
		}
	}
	
	assert.True(t, found, "Should discover .unirtm.toml in current directory")
}
