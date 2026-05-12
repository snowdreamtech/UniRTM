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

func TestRubyProvider_Name(t *testing.T) {
	p := NewRubyProvider(NewNativeProvider())
	assert.Equal(t, "ruby", p.Name())
}

func TestRubyProvider_GenerateShims(t *testing.T) {
	p := NewRubyProvider(NewNativeProvider())
	installPath := "/fake/path"
	version := "3.2.2"

	shims, err := p.GenerateShims(installPath, version)
	assert.NoError(t, err)

	expectedExecutables := []string{"ruby", "gem", "irb", "bundle"}
	for _, exe := range expectedExecutables {
		// on windows it might append .exe depending on logic, let's just test exact base
		if runtime.GOOS == "windows" {
			exe += ".exe"
		}
		shim, ok := shims[exe]
		assert.True(t, ok, "missing shim for %s", exe)
		
		if runtime.GOOS == "windows" {
			assert.Contains(t, shim, "set \"GEM_HOME=/fake/path/gem-global\"")
		} else {
			assert.Contains(t, shim, "export GEM_HOME=\"/fake/path/gem-global\"")
			assert.Contains(t, shim, filepath.Join("/fake/path", "gem-global", "bin", exe))
		}
	}
}

func TestRubyProvider_ListExecutables(t *testing.T) {
	p := NewRubyProvider(NewNativeProvider())
	execs, err := p.ListExecutables("/fake/path", "3.2.2")
	assert.NoError(t, err)

	if runtime.GOOS == "windows" {
		assert.Contains(t, execs, "ruby.exe")
		assert.Contains(t, execs, "gem.exe")
	} else {
		assert.Contains(t, execs, "ruby")
		assert.Contains(t, execs, "gem")
	}
}

func TestRubyProvider_DetectVersionError(t *testing.T) {
	p := NewRubyProvider(NewNativeProvider())
	ctx := context.Background()

	_, err := p.DetectVersion(ctx, "/fake/nonexistent/path")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to detect version"))
}
