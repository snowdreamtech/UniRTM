// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRustProvider_Name(t *testing.T) {
	p := NewRustProvider()
	assert.Equal(t, "rust", p.Name())
}

func TestRustProvider_GenerateShims(t *testing.T) {
	p := NewRustProvider()
	installPath := "/fake/path"
	version := "1.70.0"

	shims, err := p.GenerateShims("rust", installPath, version)
	assert.NoError(t, err)

	expectedExecutables := []string{"rustc", "cargo", "rustdoc", "rustfmt", "rustup"}
	for _, exe := range expectedExecutables {
		if runtime.GOOS == "windows" {
			exe += ".exe"
		}
		shim, ok := shims[exe]
		assert.True(t, ok, "missing shim for %s", exe)

		if runtime.GOOS == "windows" {
			assert.Contains(t, shim, "set \"CARGO_HOME=/fake/path/cargo\"")
			assert.Contains(t, shim, "set \"RUSTUP_HOME=/fake/path/rustup\"")
		} else {
			assert.Contains(t, shim, "export CARGO_HOME=\"/fake/path/cargo\"")
			assert.Contains(t, shim, "export RUSTUP_HOME=\"/fake/path/rustup\"")
			assert.Contains(t, shim, filepath.Join("/fake/path", "cargo", "bin", exe))
		}
	}
}

func TestRustProvider_ListExecutables(t *testing.T) {
	p := NewRustProvider()
	execs, err := p.ListExecutables("rust", "/fake/path", "1.70.0")
	assert.NoError(t, err)

	if runtime.GOOS == "windows" {
		assert.Contains(t, execs, "rustc.exe")
		assert.Contains(t, execs, "cargo.exe")
	} else {
		assert.Contains(t, execs, "rustc")
		assert.Contains(t, execs, "cargo")
	}
}

func TestRustProvider_DetectVersionError(t *testing.T) {
	p := NewRustProvider()
	ctx := context.Background()

	_, err := p.DetectVersion(ctx, "rust", "/fake/nonexistent/path")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to detect version"))
}
