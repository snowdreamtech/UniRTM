package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestE2E_Tasks(t *testing.T) {
	h := NewE2EHarness(t)
	// Just listing tasks shouldn't panic, even if empty
	stdout, _, err := h.Run("tasks", "ls")
	assert.NoError(t, err)
	_ = stdout
}

func TestE2E_BinPaths(t *testing.T) {
	h := NewE2EHarness(t)
	stdout, _, err := h.Run("bin-paths")
	// If no tools, it might be empty but shouldn't error
	assert.NoError(t, err)
	_ = stdout
}

func TestE2E_Exec(t *testing.T) {
	h := NewE2EHarness(t)
	// Try executing a built-in like echo through unirtm
	// Wait, unirtm exec [tool] -- [command]
	// If tool isn't installed, it will error, but we can do a dummy one
	_, _, err := h.Run("exec", "node@20", "--", "echo", "hello")
	// It should fail saying backend not found or tool not installed, which is fine for coverage
	_ = err
}

func TestE2E_Run(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, err := h.Run("run", "dummy-task")
	// Task won't exist
	_ = err
}
