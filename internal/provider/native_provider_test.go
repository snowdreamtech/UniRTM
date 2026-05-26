// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"testing"
)

func TestNativeProvider_Name(t *testing.T) {
	p := NewNativeProvider()
	if p.Name() != "native" {
		t.Errorf("expected native, got %s", p.Name())
	}
}

func TestNativeProvider_ListVersions_NoRecipe(t *testing.T) {
	p := NewNativeProvider()
	_, err := p.ListVersions(context.Background(), "nonexistent_tool")
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}
}

func TestNativeProvider_Install_EmptyArtifact(t *testing.T) {
	p := NewNativeProvider()
	err := p.Install(context.Background(), "tool", "/tmp", "", "1.0")
	if err == nil {
		t.Error("expected error for empty artifact path")
	}
}

func TestNativeProvider_Verify(t *testing.T) {
	p := NewNativeProvider()
	err := p.Verify(context.Background(), "tool", "1.0", "/tmp")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestNativeProvider_IsExecutable(t *testing.T) {
	tmpDir := t.TempDir()
	f, err := os.CreateTemp(tmpDir, "test")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	info, err := os.Stat(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	// We're just testing the standalone isExecutable function in native_provider.go
	if isExecutable(info) {
		t.Errorf("expected false for non-executable file")
	}
}
