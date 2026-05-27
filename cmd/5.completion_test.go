package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCompletion(t *testing.T) {
	cmd := completionCmd
	
	// Set output buffer
	var out bytes.Buffer
	cmd.SetOut(&out)

	// Test missing args
	err := runCompletion(cmd, []string{})
	assert.Error(t, err)

	// Test unknown shell
	err = runCompletion(cmd, []string{"unknown-shell"})
	assert.Error(t, err)

	// Test valid shells
	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range shells {
		out.Reset()
		err = runCompletion(cmd, []string{shell})
		assert.NoError(t, err, "Shell %s should not error", shell)
	}
}
