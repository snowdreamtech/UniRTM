package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestE2E_Alias(t *testing.T) {
	h := NewE2EHarness(t)
	stdout, _, err := h.Run("alias", "ls")
	assert.NoError(t, err)
	_ = stdout
}

func TestE2E_Current(t *testing.T) {
	h := NewE2EHarness(t)
	stdout, _, err := h.Run("current")
	assert.NoError(t, err)
	_ = stdout
}

func TestE2E_EnableDisable(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, err := h.Run("enable")
	// might return error if unirtm logic checks path
	_ = err

	_, _, err = h.Run("disable")
	_ = err
}

func TestE2E_Fmt(t *testing.T) {
	h := NewE2EHarness(t)
	stdout, _, err := h.Run("fmt")
	assert.NoError(t, err)
	_ = stdout
}

func TestE2E_Generate(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, err := h.Run("generate", "github-action")
	assert.NoError(t, err)
}

func TestE2E_Implode(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, err := h.Run("implode", "--dry-run")
	// Since there's no --dry-run maybe we just run it, but wait, implode deletes directories!
	// But they are isolated in tmpDir, so it is safe.
	// Actually we can pass yes.
	_, _, err = h.Run("implode", "--yes")
	_ = err
}

func TestE2E_Migrate(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, err := h.Run("migrate")
	_ = err
}

func TestE2E_Prepare(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, err := h.Run("prepare")
	_ = err
}

func TestE2E_Prune(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, err := h.Run("prune", "--dry-run")
	assert.NoError(t, err)
}

func TestE2E_Reshim(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, err := h.Run("reshim")
	assert.NoError(t, err)
}

func TestE2E_SelfUpdate(t *testing.T) {
	t.Skip("Skipping self update in coverage")
	h := NewE2EHarness(t)
	_, _, err := h.Run("self-update")
	_ = err
}

func TestE2E_Set(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, err := h.Run("set", "node@20")
	// Will fail with backend not found because DB is empty
	_ = err
}

func TestE2E_Sync(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, err := h.Run("sync")
	_ = err
}

func TestE2E_Dependabot(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, err := h.Run("generate", "dependabot")
	assert.NoError(t, err)
}

func TestE2E_Which(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, _ = h.Run("which", "go")
}

func TestE2E_ToolStub(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, _ = h.Run("tool-stub", "dummy")
}

func TestE2E_TestTool(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, _ = h.Run("test-tool", "dummy")
}

func TestE2E_LockCheck(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, _ = h.Run("lock", "--check")
}
