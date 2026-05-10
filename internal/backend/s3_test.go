// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"
)

func TestS3Backend_Name(t *testing.T) {
	b := NewS3Backend()
	if b.Name() != "s3" {
		t.Errorf("expected name 's3', got %s", b.Name())
	}
}

func TestS3Backend_ResolveVersion(t *testing.T) {
	b := NewS3Backend()
	
	ctx := context.Background()
	p := Platform{OS: "linux", Arch: "amd64"}
	
	info, err := b.ResolveVersion(ctx, "mytool", "1.0.0", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", info.Version)
	}
}
