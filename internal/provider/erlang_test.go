// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"testing"
)

func TestErlangProvider_Name(t *testing.T) {
	p := NewErlangProvider()
	if p.Name() != "erlang" {
		t.Errorf("expected name 'erlang', got %s", p.Name())
	}
}

func TestErlangProvider_DetectVersion(t *testing.T) {
	p := NewErlangProvider()
	ctx := context.Background()
	version, err := p.DetectVersion(ctx, "/tmp/unirtm/erlang/26.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if version != "26.0" {
		t.Errorf("expected version '26.0', got %s", version)
	}
}
