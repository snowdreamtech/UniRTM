// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"testing"
)

func TestElixirProvider_Name(t *testing.T) {
	p := NewElixirProvider()
	if p.Name() != "elixir" {
		t.Errorf("expected name 'elixir', got %s", p.Name())
	}
}

func TestElixirProvider_DetectVersion(t *testing.T) {
	p := NewElixirProvider()
	ctx := context.Background()
	version, err := p.DetectVersion(ctx, "/tmp/unirtm/elixir/1.15.4")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if version != "1.15.4" {
		t.Errorf("expected version '1.15.4', got %s", version)
	}
}
