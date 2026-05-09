// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"
)

func TestAsdfBackend_Name(t *testing.T) {
	b := NewAsdfBackend()
	if b.Name() != "asdf" {
		t.Errorf("expected 'asdf', got '%s'", b.Name())
	}
}

func TestAsdfBackend_Interface(t *testing.T) {
	var _ Backend = (*AsdfBackend)(nil)
}

func TestAsdfBackend_ResolveVersion(t *testing.T) {
	b := NewAsdfBackend()
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	info, err := b.ResolveVersion(ctx, "nodejs", "20.0.0", platform)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if info.Version != "20.0.0" {
		t.Errorf("expected version 20.0.0, got %s", info.Version)
	}
}
