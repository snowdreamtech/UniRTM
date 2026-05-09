// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"
)

func TestCargoBackend_Name(t *testing.T) {
	b := NewCargoBackend()
	if b.Name() != "cargo" {
		t.Errorf("expected 'cargo', got '%s'", b.Name())
	}
}

func TestCargoBackend_Interface(t *testing.T) {
	var _ Backend = (*CargoBackend)(nil)
}

func TestCargoBackend_ResolveVersion(t *testing.T) {
	b := NewCargoBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.ResolveVersion(ctx, "ripgrep", "13.0.0", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if info.Version != "13.0.0" {
		t.Errorf("expected version 13.0.0, got %s", info.Version)
	}
}
