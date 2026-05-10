// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"testing"
)

func TestBunProvider_Name(t *testing.T) {
	p := NewBunProvider()
	if p.Name() != "bun" {
		t.Errorf("expected name 'bun', got %s", p.Name())
	}
}

func TestBunProvider_DetectVersion(t *testing.T) {
	p := NewBunProvider()
	ctx := context.Background()
	version, err := p.DetectVersion(ctx, "/tmp/unirtm/bun/1.0.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", version)
	}
}
