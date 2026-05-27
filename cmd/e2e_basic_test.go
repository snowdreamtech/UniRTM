package cmd

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestE2E_Version(t *testing.T) {
	h := NewE2EHarness(t)
	stdout, _, err := h.Run("version")
	assert.NoError(t, err)
	assert.Contains(t, stdout, "unirtm version")
}

func TestE2E_Help(t *testing.T) {
	h := NewE2EHarness(t)
	stdout, _, err := h.Run("help")
	assert.NoError(t, err)
	assert.Contains(t, stdout, "Usage:")
	assert.Contains(t, stdout, "Available Commands:")
}

func TestE2E_Doctor(t *testing.T) {
	h := NewE2EHarness(t)
	_, _, err := h.Run("doctor")
	// ignore error, just run it
	_ = err
}
