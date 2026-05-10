// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"testing"
)

func TestPhpProvider_Name(t *testing.T) {
	p := NewPhpProvider()
	if p.Name() != "php" {
		t.Errorf("expected name 'php', got %s", p.Name())
	}
}

func TestPhpProvider_DetectVersion(t *testing.T) {
	p := NewPhpProvider()
	ctx := context.Background()
	version, err := p.DetectVersion(ctx, "/tmp/unirtm/php/8.2.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if version != "8.2.0" {
		t.Errorf("expected version '8.2.0', got %s", version)
	}
}
