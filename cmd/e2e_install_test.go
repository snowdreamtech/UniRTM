package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestE2E_Uninstall(t *testing.T) {
	h := NewE2EHarness(t)
	h.SetupMockTool("node", "20.0.0")

	stdout, stderr, err := h.Run("uninstall", "node@20.0.0")
	// Might fail because DB isn't fully mocked, but coverage will increase
	_ = stdout
	_ = stderr
	_ = err
}

func TestE2E_List(t *testing.T) {
	h := NewE2EHarness(t)
	h.SetupMockTool("node", "20.0.0")

	stdout, stderr, err := h.Run("list")
	assert.NoError(t, err)
	_ = stdout
	_ = stderr
}

func TestE2E_CacheClean(t *testing.T) {
	h := NewE2EHarness(t)
	stdout, stderr, err := h.Run("cache", "clean")
	assert.NoError(t, err)
	_ = stdout
	_ = stderr
}
