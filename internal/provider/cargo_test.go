// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"testing"
)

func TestCargoProvider_Name(t *testing.T) {
	p := NewCargoProvider()
	if p.Name() != "cargo" {
		t.Errorf("expected 'cargo', got '%s'", p.Name())
	}
}

func TestCargoProvider_Interface(t *testing.T) {
	var _ Provider = (*CargoProvider)(nil)
}
