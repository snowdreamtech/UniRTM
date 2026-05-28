package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestE2E_CommandsHelp(t *testing.T) {
	commands := []string{
		"activate", "deactivate", "search", "update", "cache", "config",
		"use", "shell", "which", "plugin", "trust", "untrust", "watch",
		"settings", "lock", "outdated", "latest", "tool", "backends",
		"link", "unuse", "install-into", "edit", "token", "mcp", "index",
	}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			h := NewE2EHarness(t)
			stdout, _, err := h.Run(cmd, "--help")
			assert.NoError(t, err)
			assert.Contains(t, stdout, "Usage:")
		})
	}
}
