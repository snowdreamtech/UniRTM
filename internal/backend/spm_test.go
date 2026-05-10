// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"
)

func TestSpmBackend_Name(t *testing.T) {
	b := NewSpmBackend()
	if b.Name() != "spm" {
		t.Errorf("expected name 'spm', got %s", b.Name())
	}
}

func TestSpmBackend_ResolveVersion(t *testing.T) {
	b := NewSpmBackend()
	
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}
	
	info, err := b.ResolveVersion(ctx, "apple/swift-format", "509.0.0", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "509.0.0" {
		t.Errorf("expected version '509.0.0', got %s", info.Version)
	}
}
