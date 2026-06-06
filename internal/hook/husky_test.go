package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestHuskyRunner_Detect(t *testing.T) {
	tmpDir := t.TempDir()

	h := HuskyRunner{}

	// No .husky dir — should not detect
	if h.Detect(tmpDir) {
		t.Errorf("Detect() should return false when .husky does not exist")
	}

	// Create .husky as a regular file — should not detect (must be dir)
	huskyFile := filepath.Join(tmpDir, ".husky")
	if err := os.WriteFile(huskyFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if h.Detect(tmpDir) {
		t.Errorf("Detect() should return false when .husky is a file, not a directory")
	}

	// Remove file, create as directory — should detect
	os.Remove(huskyFile)
	if err := os.Mkdir(huskyFile, 0755); err != nil {
		t.Fatal(err)
	}
	if !h.Detect(tmpDir) {
		t.Errorf("Detect() should return true when .husky is a directory")
	}
}

func TestHuskyRunner_ValidateHookName_BlocksPathTraversal(t *testing.T) {
	// Path traversal must be blocked at CLI entry before reaching HuskyRunner.
	// This test confirms that ValidateHookName catches dangerous names that
	// would otherwise be used in filepath.Join(".husky", hookName).
	dangerous := []string{
		"../etc/passwd",
		"../../root/.ssh/authorized_keys",
		"/etc/passwd",
		"pre-commit; rm -rf /",
	}
	for _, name := range dangerous {
		t.Run(name, func(t *testing.T) {
			if err := ValidateHookName(name); err == nil {
				t.Errorf("ValidateHookName(%q) must reject path traversal attempt", name)
			}
		})
	}
}

func TestHuskyRunner_Run_SilentOnMissingScript(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .husky dir but no pre-commit script inside
	huskyDir := filepath.Join(tmpDir, ".husky")
	if err := os.Mkdir(huskyDir, 0755); err != nil {
		t.Fatal(err)
	}

	h := HuskyRunner{}
	// Must succeed silently (no hook script defined)
	if err := h.Run(context.Background(), "pre-commit", nil); err != nil {
		t.Errorf("Run() should return nil when hook script does not exist, got: %v", err)
	}
}
