// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"testing"
)

func TestSwiftProvider_Name(t *testing.T) {
	p := NewSwiftProvider()
	if p.Name() != "swift" {
		t.Errorf("expected name 'swift', got %s", p.Name())
	}
}

func TestSwiftProvider_DetectVersion(t *testing.T) {
	p := NewSwiftProvider()
	ctx := context.Background()
	version, err := p.DetectVersion(ctx, "swift", "/tmp/unirtm/swift/5.9.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if version != "5.9.0" {
		t.Errorf("expected version '5.9.0', got %s", version)
	}
}
