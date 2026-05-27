// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
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

	shims, err := p.GenerateShims("java", installPath, version)
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
	execs, err := p.ListExecutables("java", "/fake/path", "17.0.0")
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

	_, err := p.DetectVersion(ctx, "java", "/fake/nonexistent/path")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to detect version"))
}

func TestJavaProvider_getJavaHome(t *testing.T) {
	p := NewJavaProvider()
	tmpDir := t.TempDir()

	// Default case
	home1 := p.getJavaHome(tmpDir)
	assert.Equal(t, tmpDir, home1)

	// macOS case
	macOSPath := filepath.Join(tmpDir, "Contents", "Home", "bin")
	err := os.MkdirAll(macOSPath, 0755)
	assert.NoError(t, err)

	home2 := p.getJavaHome(tmpDir)
	assert.Equal(t, filepath.Join(tmpDir, "Contents", "Home"), home2)
}

func TestJavaProvider_DetectVersionSuccess(t *testing.T) {
	p := NewJavaProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	err := os.MkdirAll(binDir, 0755)
	assert.NoError(t, err)

	javaPath := filepath.Join(binDir, "java")
	if runtime.GOOS == "windows" {
		javaPath += ".exe"
	}

	// Mock java output to stderr
	mockJava := `#!/bin/sh
echo 'openjdk version "17.0.8" 2023-07-18 LTS' >&2
exit 0
`
	if runtime.GOOS == "windows" {
		mockJava = `@echo off
echo openjdk version "17.0.8" 2023-07-18 LTS 1>&2
exit 0
`
	}
	err = os.WriteFile(javaPath, []byte(mockJava), 0755)
	assert.NoError(t, err)

	ctx := context.Background()
	version, err := p.DetectVersion(ctx, "java", tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "17.0.8", version)
}

