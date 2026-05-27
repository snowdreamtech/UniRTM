package cmd

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGenerateDeactivationScript(t *testing.T) {
	script := generateDeactivationScript("bash")
	assert.Contains(t, script, "_unirtm_clean_path")

	script = generateDeactivationScript("zsh")
	assert.Contains(t, script, "_unirtm_clean_path")

	script = generateDeactivationScript("fish")
	assert.Contains(t, script, "set -gx PATH $new_path")

	script = generateDeactivationScript("powershell")
	assert.Contains(t, script, "$env:PATH = ($env:PATH -split")
}
