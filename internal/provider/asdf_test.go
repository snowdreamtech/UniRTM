// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"testing"
)

func TestAsdfProvider_Name(t *testing.T) {
	p := NewAsdfProvider()
	if p.Name() != "asdf" {
		t.Errorf("expected 'asdf', got '%s'", p.Name())
	}
}

func TestAsdfProvider_Interface(t *testing.T) {
	var _ Provider = (*AsdfProvider)(nil)
}
