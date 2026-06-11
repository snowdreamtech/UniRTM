// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/stretchr/testify/require"
)

func TestNpmProvider_checkAndWarnLifecycleScripts_POSIX(t *testing.T) {
	if env.RuntimeGOOS == "windows" {
		t.Skip("Skipping POSIX test on windows")
	}

	p := NewNpmProvider()
	tmpDir := t.TempDir()

	// Mock package.json path for POSIX: lib/node_modules/test_pkg/package.json
	pkgDir := filepath.Join(tmpDir, "lib", "node_modules", "test_pkg")
	require.NoError(t, os.MkdirAll(pkgDir, 0755))
	
	pkgJsonPath := filepath.Join(pkgDir, "package.json")
	
	// Test 1: No lifecycle scripts
	err := os.WriteFile(pkgJsonPath, []byte(`{"name":"test_pkg","version":"1.0.0"}`), 0644)
	require.NoError(t, err)
	// Should not panic or error
	p.checkAndWarnLifecycleScripts("test_pkg", tmpDir)

	// Test 2: Has postinstall script
	err = os.WriteFile(pkgJsonPath, []byte(`{"name":"test_pkg","scripts":{"postinstall":"echo malicious"}}`), 0644)
	require.NoError(t, err)
	p.checkAndWarnLifecycleScripts("test_pkg", tmpDir)
}

func TestNpmProvider_checkAndWarnLifecycleScripts_Windows(t *testing.T) {
	if env.RuntimeGOOS != "windows" {
		t.Skip("Skipping Windows test on POSIX")
	}

	p := NewNpmProvider()
	tmpDir := t.TempDir()

	// Mock package.json path for Windows: node_modules/test_pkg/package.json
	pkgDir := filepath.Join(tmpDir, "node_modules", "test_pkg")
	require.NoError(t, os.MkdirAll(pkgDir, 0755))
	
	pkgJsonPath := filepath.Join(pkgDir, "package.json")
	
	// Test 1: No lifecycle scripts
	err := os.WriteFile(pkgJsonPath, []byte(`{"name":"test_pkg","version":"1.0.0"}`), 0644)
	require.NoError(t, err)
	p.checkAndWarnLifecycleScripts("test_pkg", tmpDir)

	// Test 2: Has preinstall script
	err = os.WriteFile(pkgJsonPath, []byte(`{"name":"test_pkg","scripts":{"preinstall":"echo malicious"}}`), 0644)
	require.NoError(t, err)
	p.checkAndWarnLifecycleScripts("test_pkg", tmpDir)
}
