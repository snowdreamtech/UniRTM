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

func TestEnvCommandStructure(t *testing.T) {
	assert.Contains(t, envCmd.Use, "env")
	assert.NotEmpty(t, envCmd.Short)
	assert.True(t, envCmd.Run != nil || envCmd.RunE != nil)
}

func TestEnvFlags(t *testing.T) {
	assert.NotNil(t, envCmd.Flags().Lookup("shell"))
	assert.NotNil(t, envCmd.Flags().Lookup("info"))
}

func TestResolveShell(t *testing.T) {
	tests := []struct{ flag, want string }{
		{"bash", "bash"},
		{"fish", "fish"},
		{"nu", "nu"},
		{"FISH", "fish"},
		{"", "bash"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, resolveShell(tt.flag), "resolveShell(%q)", tt.flag)
	}
}

func TestEmitShellEnv_Bash(t *testing.T) {
	out := captureStdoutFunc(t, func() {
		_ = emitShellEnv("bash",
			[]string{"/usr/local/share/unirtm/shims"},
			[]envVarEntry{{Name: "UNIRTM_NODE_VERSION", Value: "22.14.0"}},
		)
	})
	assert.Contains(t, out, "export PATH=")
	assert.Contains(t, out, "UNIRTM_NODE_VERSION")
}

func TestEmitShellEnv_Fish(t *testing.T) {
	out := captureStdoutFunc(t, func() {
		_ = emitShellEnv("fish",
			[]string{"/usr/local/share/unirtm/shims"},
			[]envVarEntry{{Name: "UNIRTM_NODE_VERSION", Value: "22.14.0"}},
		)
	})
	assert.Contains(t, out, "set -gx PATH")
}

func TestEmitShellEnv_Nu(t *testing.T) {
	out := captureStdoutFunc(t, func() {
		_ = emitShellEnv("nu",
			[]string{"/usr/local/share/unirtm/shims"},
			[]envVarEntry{},
		)
	})
	assert.Contains(t, out, "$env.PATH")
}

func TestEmitShellEnv_NoPathDirs(t *testing.T) {
	out := captureStdoutFunc(t, func() {
		_ = emitShellEnv("bash", []string{}, []envVarEntry{})
	})
	assert.NotContains(t, out, "export PATH=")
}

func TestRunEnvInfo(t *testing.T) {
	out := captureStdoutFunc(t, func() {
		_ = runEnvInfo()
	})
	assert.True(t, strings.Contains(out, "ProjectName=") || strings.Contains(out, "GOOS="))
}

// captureStdoutFunc redirects os.Stdout for the duration of f and returns output.
func captureStdoutFunc(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = orig
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	r.Close()
	return buf.String()
}
