// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestUbiProvider_Name(t *testing.T) {
	provider := NewUbiProvider()
	if provider.Name() != "ubi" {
		t.Errorf("Expected name 'ubi', got '%s'", provider.Name())
	}
}

func TestUbiProvider_DetectVersion(t *testing.T) {
	provider := NewUbiProvider()
	ctx := context.Background()
	installPath := "/fake/path/to/installs/houseabsolute/precious/0.0.12"
	
	version, err := provider.DetectVersion(ctx, installPath)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if version != "0.0.12" {
		t.Errorf("Expected version '0.0.12', got '%s'", version)
	}
}

func TestUbiProvider_GenerateShims(t *testing.T) {
	provider := NewUbiProvider()
	
	// Create a temporary bin directory with a mock executable
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin dir: %v", err)
	}
	
	exePath := filepath.Join(binDir, "precious")
	if err := os.WriteFile(exePath, []byte("#!/bin/sh\necho 'precious'"), 0755); err != nil {
		t.Fatalf("Failed to create mock executable: %v", err)
	}
	
	shims, err := provider.GenerateShims(tmpDir, "0.0.12")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(shims) != 1 {
		t.Errorf("Expected 1 shim, got %d", len(shims))
	}
	
	if path, ok := shims["precious"]; !ok || path != exePath {
		t.Errorf("Expected precious shim to point to %s, got %s", exePath, path)
	}
}
