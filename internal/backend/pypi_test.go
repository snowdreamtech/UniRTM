// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"
)

func TestPypiBackend_Name(t *testing.T) {
	b := NewPypiBackend()
	if b.Name() != "pypi" {
		t.Errorf("expected 'pypi', got '%s'", b.Name())
	}
}

func TestPypiBackend_Interface(t *testing.T) {
	var _ Backend = (*PypiBackend)(nil)
}

func TestPypiBackend_ResolveVersion(t *testing.T) {
	b := NewPypiBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.ResolveVersion(ctx, "black", "23.3.0", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if info.Version != "23.3.0" {
		t.Errorf("expected version 23.3.0, got %s", info.Version)
	}
}
