// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpmProvider_Name(t *testing.T) {
	p := NewSpmProvider()
	if p.Name() != "spm" {
		t.Errorf("expected name 'spm', got %s", p.Name())
	}
}

func TestSpmProvider_DetectVersion(t *testing.T) {
	p := NewSpmProvider()

	ctx := context.Background()
	version, err := p.DetectVersion(ctx, "spm", "/fake/path/tool/1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %s", version)
	}
}

func TestSpmProvider_findSwift(t *testing.T) {
	tmpData := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpData)

	p := NewSpmProvider()

	// Create fake swift installation
	swiftDir := filepath.Join(tmpData, "installs", "swift", "5.8", "bin")
	os.MkdirAll(swiftDir, 0755)
	swiftPath := filepath.Join(swiftDir, "swift")
	os.WriteFile(swiftPath, []byte("fake binary"), 0755)

	found, err := p.findSwift()
	require.NoError(t, err)
	require.Equal(t, swiftPath, found)
}
