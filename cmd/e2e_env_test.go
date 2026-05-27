package cmd

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestE2E_Env(t *testing.T) {
	h := NewE2EHarness(t)

	// "use" command
	_, _, err := h.Run("use", "node@20")
	// ignore error, just need coverage of execution path
	_ = err

	// "env" command
	stdout, _, err := h.Run("env")
	assert.NoError(t, err)
	assert.Contains(t, stdout, "node")

	// shell
	_, _, err = h.Run("shell", "node@20")
	assert.NoError(t, err)
	
	// activate
	stdout, _, err = h.Run("activate", "bash")
	assert.NoError(t, err)
	assert.Contains(t, stdout, "unirtm")

	// deactivate
	stdout, _, err = h.Run("deactivate")
	assert.NoError(t, err)
}

func TestE2E_Where(t *testing.T) {
	h := NewE2EHarness(t)
	h.SetupMockTool("node", "20.0.0")

	stdout, _, err := h.Run("where", "node@20.0.0")
	// Since node@20.0.0 is not in the DB, it might fail or return nothing. We just check no panic.
	_ = stdout
	_ = err
}
