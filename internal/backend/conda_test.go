// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"
)

func TestCondaBackend_Name(t *testing.T) {
	b := NewCondaBackend()
	if b.Name() != "conda" {
		t.Errorf("expected name 'conda', got %s", b.Name())
	}
}

func TestCondaBackend_ResolveVersion(t *testing.T) {
	b := NewCondaBackend()
	
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}
	
	info, err := b.ResolveVersion(ctx, "numpy", "1.24.0", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "1.24.0" {
		t.Errorf("expected version '1.24.0', got %s", info.Version)
	}
}
