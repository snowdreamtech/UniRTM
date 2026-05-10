// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"testing"
)

func TestDenoProvider_Name(t *testing.T) {
	p := NewDenoProvider()
	if p.Name() != "deno" {
		t.Errorf("expected name 'deno', got %s", p.Name())
	}
}

func TestDenoProvider_DetectVersion(t *testing.T) {
	p := NewDenoProvider()
	ctx := context.Background()
	version, err := p.DetectVersion(ctx, "/tmp/unirtm/deno/1.37.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if version != "1.37.0" {
		t.Errorf("expected version '1.37.0', got %s", version)
	}
}
