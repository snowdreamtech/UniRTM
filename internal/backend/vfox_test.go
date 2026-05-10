// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"
)

func TestVfoxBackend_Name(t *testing.T) {
	b := NewVfoxBackend()
	if b.Name() != "vfox" {
		t.Errorf("expected name 'vfox', got %s", b.Name())
	}
}

func TestVfoxBackend_ResolveVersion(t *testing.T) {
	b := NewVfoxBackend()
	
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}
	
	info, err := b.ResolveVersion(ctx, "java", "21.0.1", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "21.0.1" {
		t.Errorf("expected version '21.0.1', got %s", info.Version)
	}
}
