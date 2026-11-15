// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildTerminalEnv_Basic(t *testing.T) {
	env := BuildTerminalEnv()
	assert.NotEmpty(t, env)

	envMap := make(map[string]string)
	for _, kv := range env {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	assert.Equal(t, "xterm-256color", envMap["TERM"])
	assert.Equal(t, "truecolor", envMap["COLORTERM"])
	assert.Equal(t, "1", envMap["UNIRTM_TERMINAL"])
	assert.Contains(t, envMap["PATH"], ".unirtm")
}

func TestBuildTerminalEnv_CustomShell(t *testing.T) {
	testCases := []struct {
		name      string
		shellPath string
		expected  string
	}{
		{
			name:      "POSIX sh override",
			shellPath: "/bin/sh",
			expected:  "/bin/sh",
		},
		{
			name:      "Bash override",
			shellPath: "/bin/bash",
			expected:  "/bin/bash",
		},
		{
			name:      "Zsh override",
			shellPath: "/usr/bin/zsh",
			expected:  "/usr/bin/zsh",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := BuildTerminalEnv(tc.shellPath)
			found := false
			for _, item := range env {
				if item == "SHELL="+tc.expected {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected SHELL to be set to %s", tc.expected)
		})
	}
}
