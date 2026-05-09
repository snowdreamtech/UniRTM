// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"testing"
)

func TestNpmProvider_Name(t *testing.T) {
	p := NewNpmProvider()
	if p.Name() != "npm" {
		t.Errorf("expected 'npm', got '%s'", p.Name())
	}
}

func TestNpmProvider_Interface(t *testing.T) {
	var _ Provider = (*NpmProvider)(nil)
}
