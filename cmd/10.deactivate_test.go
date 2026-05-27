// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeactivateCommandStructure(t *testing.T) {
	assert.Equal(t, "deactivate", deactivateCmd.Use)
	assert.NotEmpty(t, deactivateCmd.Short)
	assert.NotNil(t, deactivateCmd.RunE)
}

func TestRunDeactivate(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := deactivateCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Test default execution (usually bash/zsh depending on env)
	err := runDeactivate(cmd, []string{})
	assert.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout

	out, _ := io.ReadAll(r)
	script := string(out)

	assert.Contains(t, script, "unset UNIRTM_PATH")
	assert.Contains(t, script, "unset UNIRTM_ACTIVATION_SCOPE")
}

func TestRunDeactivate_ShellSpecific(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cmd := deactivateCmd
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			deactivateShell = shell
			err := runDeactivate(cmd, []string{})
			deactivateShell = "" // Reset
			
			assert.NoError(t, err)

			w.Close()
			os.Stdout = oldStdout

			out, _ := io.ReadAll(r)
			script := string(out)

			if shell == "fish" {
				assert.Contains(t, script, "set -e UNIRTM_PATH")
			} else if shell == "powershell" {
				assert.Contains(t, script, "Remove-Item Env:\\UNIRTM_PATH")
			} else {
				assert.Contains(t, script, "unset UNIRTM_PATH")
			}
		})
	}
}

func TestRunDeactivate_ShellArgument(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := deactivateCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runDeactivate(cmd, []string{"fish"})
	assert.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout

	out, _ := io.ReadAll(r)
	script := string(out)

	assert.Contains(t, script, "set -e UNIRTM_PATH")
}

func TestGenerateDeactivationScript(t *testing.T) {
	bashScript := generateDeactivationScript("bash")
	assert.True(t, strings.HasPrefix(bashScript, "# UniRTM deactivation script"))
	assert.Contains(t, bashScript, "unset UNIRTM_PATH")

	fishScript := generateDeactivationScript("fish")
	assert.True(t, strings.HasPrefix(fishScript, "# UniRTM deactivation script (fish)"))
	assert.Contains(t, fishScript, "set -e UNIRTM_PATH")

	psScript := generateDeactivationScript("powershell")
	assert.True(t, strings.HasPrefix(psScript, "# UniRTM deactivation script (PowerShell)"))
	assert.Contains(t, psScript, "Remove-Item Env:\\UNIRTM_PATH")
}
