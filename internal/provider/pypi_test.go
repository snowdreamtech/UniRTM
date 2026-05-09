// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"testing"
)

func TestPypiProvider_Name(t *testing.T) {
	p := NewPypiProvider()
	if p.Name() != "pypi" {
		t.Errorf("expected 'pypi', got '%s'", p.Name())
	}
}

func TestPypiProvider_Interface(t *testing.T) {
	var _ Provider = (*PypiProvider)(nil)
}
