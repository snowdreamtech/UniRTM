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

func TestJavaProvider_Name(t *testing.T) {
	p := NewJavaProvider()
	assert.Equal(t, "java", p.Name())
}

func TestJavaProvider_GenerateShims(t *testing.T) {
	p := NewJavaProvider()
	installPath := "/fake/path"
	version := "17.0.0"

	shims, err := p.GenerateShims(installPath, version)
	assert.NoError(t, err)

	expectedExecutables := []string{"java", "javac", "jar", "javadoc"}
	for _, exe := range expectedExecutables {
		shim, ok := shims[exe]
		assert.True(t, ok, "missing shim for %s", exe)
		
		if runtime.GOOS == "windows" {
			assert.Contains(t, shim, "set \"JAVA_HOME=/fake/path\"")
			assert.Contains(t, shim, filepath.Join("/fake/path", "bin", exe+".exe"))
		} else {
			assert.Contains(t, shim, "export JAVA_HOME=\"/fake/path\"")
			assert.Contains(t, shim, filepath.Join("/fake/path", "bin", exe))
		}
	}
}

func TestJavaProvider_ListExecutables(t *testing.T) {
	p := NewJavaProvider()
	execs, err := p.ListExecutables("/fake/path", "17.0.0")
	assert.NoError(t, err)

	if runtime.GOOS == "windows" {
		assert.Contains(t, execs, "java.exe")
		assert.Contains(t, execs, "javac.exe")
	} else {
		assert.Contains(t, execs, "java")
		assert.Contains(t, execs, "javac")
	}
}

func TestJavaProvider_DetectVersionError(t *testing.T) {
	p := NewJavaProvider()
	ctx := context.Background()

	_, err := p.DetectVersion(ctx, "/fake/nonexistent/path")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to detect version"))
}
