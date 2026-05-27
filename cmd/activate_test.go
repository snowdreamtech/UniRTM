package cmd

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestResolveShellType(t *testing.T) {
	// Valid shells
	shell, err := resolveShellType("bash")
	assert.NoError(t, err)
	assert.Equal(t, "bash", shell)

	shell, err = resolveShellType("ZSH")
	assert.NoError(t, err)
	assert.Equal(t, "zsh", shell)

	shell, err = resolveShellType("fish")
	assert.NoError(t, err)
	assert.Equal(t, "fish", shell)

	shell, err = resolveShellType("powershell")
	assert.NoError(t, err)
	assert.Equal(t, "powershell", shell)

	shell, err = resolveShellType("pwsh")
	assert.NoError(t, err)
	assert.Equal(t, "powershell", shell)

	// Invalid shell
	_, err = resolveShellType("cmd")
	assert.Error(t, err)
}
