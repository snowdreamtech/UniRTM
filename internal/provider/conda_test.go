// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"testing"
)

func TestCondaProvider_Name(t *testing.T) {
	p := NewCondaProvider()
	if p.Name() != "conda" {
		t.Errorf("expected name 'conda', got %s", p.Name())
	}
}

func TestCondaProvider_DetectVersion(t *testing.T) {
	p := NewCondaProvider()
	
	ctx := context.Background()
	version, err := p.DetectVersion(ctx, "/fake/path/tool/1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %s", version)
	}
}
