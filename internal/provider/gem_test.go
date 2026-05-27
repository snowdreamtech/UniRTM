// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGemProvider_Interface(t *testing.T) {
	var _ Provider = (*GemProvider)(nil)
}

func TestGemProvider_FindGem(t *testing.T) {
	p := NewGemProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)

	scriptPath := filepath.Join(binDir, "gem")
	os.WriteFile(scriptPath, []byte("#!/bin/sh\necho gem"), 0755)

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	gem, err := p.findGem()
	if err != nil {
		t.Fatalf("expected to find gem, but got error: %v", err)
	}
	if gem != scriptPath {
		t.Errorf("expected %s, got %s", scriptPath, gem)
	}
}

func TestGemProvider_Install(t *testing.T) {
	p := NewGemProvider()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)
	scriptPath := filepath.Join(binDir, "gem")
	os.WriteFile(scriptPath, []byte("#!/bin/sh\necho installing..."), 0755)

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	installPath := filepath.Join(tmpDir, "install")

	ctx := context.Background()

	err := p.Install(ctx, "tool", installPath, "", "1.0.0")
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
}
