// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"testing"
)

func TestVfoxProvider_Name(t *testing.T) {
	p := NewVfoxProvider()
	if p.Name() != "vfox" {
		t.Errorf("expected name 'vfox', got %s", p.Name())
	}
}

func TestVfoxProvider_DetectVersion(t *testing.T) {
	p := NewVfoxProvider()

	ctx := context.Background()
	version, err := p.DetectVersion(ctx, "/fake/path/tool/1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %s", version)
	}
}
