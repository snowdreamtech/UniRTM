// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"fmt"
	"os"
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
	installPath := t.TempDir()
	version := "3.2.2"

	os.MkdirAll(filepath.Join(installPath, "bin"), 0755)
	os.MkdirAll(filepath.Join(installPath, "gem-global", "bin"), 0755)

	if runtime.GOOS == "windows" {
		os.WriteFile(filepath.Join(installPath, "bin", "ruby.exe"), []byte(""), 0755)
		os.WriteFile(filepath.Join(installPath, "gem-global", "bin", "gem.exe"), []byte(""), 0755)
		os.WriteFile(filepath.Join(installPath, "gem-global", "bin", "irb.exe"), []byte(""), 0755)
		os.WriteFile(filepath.Join(installPath, "gem-global", "bin", "bundle.exe"), []byte(""), 0755)
	} else {
		os.WriteFile(filepath.Join(installPath, "bin", "ruby"), []byte(""), 0755)
		os.WriteFile(filepath.Join(installPath, "gem-global", "bin", "gem"), []byte(""), 0755)
		os.WriteFile(filepath.Join(installPath, "gem-global", "bin", "irb"), []byte(""), 0755)
		os.WriteFile(filepath.Join(installPath, "gem-global", "bin", "bundle"), []byte(""), 0755)
	}

	shims, err := p.GenerateShims("ruby", installPath, version)
	t.Logf("shims: %v, err: %v", shims, err)
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
			assert.Contains(t, shim, fmt.Sprintf("set \"GEM_HOME=%s\"", filepath.Join(installPath, "gem-global")))
		} else {
			assert.Contains(t, shim, fmt.Sprintf("export GEM_HOME=\"%s\"", filepath.Join(installPath, "gem-global")))
			if exe == "ruby" {
				assert.Contains(t, shim, filepath.Join(installPath, "bin", exe))
			} else {
				assert.Contains(t, shim, filepath.Join(installPath, "gem-global", "bin", exe))
			}
		}
	}
}

func TestRubyProvider_ListExecutables(t *testing.T) {
	p := NewRubyProvider(NewNativeProvider())
	installPath := t.TempDir()
	os.MkdirAll(filepath.Join(installPath, "bin"), 0755)
	os.MkdirAll(filepath.Join(installPath, "gem-global", "bin"), 0755)

	if runtime.GOOS == "windows" {
		os.WriteFile(filepath.Join(installPath, "bin", "ruby.exe"), []byte(""), 0755)
		os.WriteFile(filepath.Join(installPath, "gem-global", "bin", "gem.exe"), []byte(""), 0755)
	} else {
		os.WriteFile(filepath.Join(installPath, "bin", "ruby"), []byte(""), 0755)
		os.WriteFile(filepath.Join(installPath, "gem-global", "bin", "gem"), []byte(""), 0755)
	}

	execs, err := p.ListExecutables("ruby", installPath, "3.2.2")
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

	_, err := p.DetectVersion(ctx, "ruby", "/fake/nonexistent/path")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to detect version"))
}
