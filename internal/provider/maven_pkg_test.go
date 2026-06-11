// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMavenPkgProvider_Name(t *testing.T) {
	p := NewMavenPkgProvider()
	if p.Name() != "maven-pkg" {
		t.Errorf("Expected name 'maven-pkg', got '%s'", p.Name())
	}
}

// In a real scenario, we might mock DownloadFile or skip actual downloads in unit tests.
// For now, we will verify that parsing failure paths work correctly.
func TestMavenPkgProvider_Install_InvalidTool(t *testing.T) {
	p := NewMavenPkgProvider()
	ctx := context.Background()
	installPath := t.TempDir()

	err := p.Install(ctx, "invalid-tool", installPath, "", "1.0.0")
	if err == nil {
		t.Errorf("Expected error for invalid tool name format")
	} else if !strings.Contains(err.Error(), "must be in the format") {
		t.Errorf("Expected format error, got: %v", err)
	}
}

// We test GenerateShims functionality. Since Install() writes the actual wrapper scripts,
// we'll manually create a dummy wrapper to test GenerateShims.
func TestMavenPkgProvider_GenerateShims(t *testing.T) {
	p := NewMavenPkgProvider()
	installPath := t.TempDir()
	binDir := filepath.Join(installPath, "bin")
	os.MkdirAll(binDir, 0755)

	// Create dummy shims
	unixShim := filepath.Join(binDir, "mytool")
	winShim := filepath.Join(binDir, "mytool2.cmd")

	os.WriteFile(unixShim, []byte("#!/bin/sh\n"), 0755)
	os.WriteFile(winShim, []byte("@echo off\n"), 0755)

	shims, err := p.GenerateShims("org.example/mytool", installPath, "1.0.0")
	if err != nil {
		t.Fatalf("GenerateShims failed: %v", err)
	}

	if len(shims) != 2 {
		t.Errorf("Expected 2 shims, got %d", len(shims))
	}

	if val, ok := shims["mytool"]; !ok || val != unixShim {
		t.Errorf("Missing or incorrect unix shim mapping: %v", val)
	}

	if _, ok := shims["mytool2.cmd"]; ok {
		t.Errorf("Expected .cmd suffix to be stripped from shim name")
	}
	if val, ok := shims["mytool2"]; !ok || val != winShim {
		t.Errorf("Missing or incorrect windows shim mapping: %v", val)
	}
}
