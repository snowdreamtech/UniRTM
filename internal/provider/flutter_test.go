// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"testing"
)

func TestFlutterProvider_Name(t *testing.T) {
	p := NewFlutterProvider()
	if p.Name() != "flutter" {
		t.Errorf("expected name 'flutter', got %s", p.Name())
	}
}

func TestFlutterProvider_DetectVersion(t *testing.T) {
	p := NewFlutterProvider()
	ctx := context.Background()
	version, err := p.DetectVersion(ctx, "flutter", "/tmp/unirtm/flutter/3.13.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if version != "3.13.0" {
		t.Errorf("expected version '3.13.0', got %s", version)
	}
}
