// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"testing"
)

func TestZigProvider_Name(t *testing.T) {
	p := NewZigProvider()
	if p.Name() != "zig" {
		t.Errorf("expected name 'zig', got %s", p.Name())
	}
}

func TestZigProvider_DetectVersion(t *testing.T) {
	p := NewZigProvider()
	ctx := context.Background()
	version, err := p.DetectVersion(ctx, "zig", "/tmp/unirtm/zig/0.11.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if version != "0.11.0" {
		t.Errorf("expected version '0.11.0', got %s", version)
	}
}
