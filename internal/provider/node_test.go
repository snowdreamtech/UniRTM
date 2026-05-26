// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"path/filepath"
	"testing"
)

func TestNodeProvider_Name(t *testing.T) {
	p := NewNodeProvider()
	if p.Name() != "node" {
		t.Errorf("expected node, got %s", p.Name())
	}
}

func TestNodeProvider_ListExecutables(t *testing.T) {
	p := NewNodeProvider()
	execs, err := p.ListExecutables("node", "/tmp", "18.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(execs) != 3 {
		t.Errorf("expected 3 executables, got %d", len(execs))
	}
}

func TestNodeProvider_GetBinPaths(t *testing.T) {
	p := NewNodeProvider()
	paths, err := p.GetBinPaths("node", "/tmp/node", "18.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(paths))
	}
}

func TestNodeProvider_GetEnvVars(t *testing.T) {
	p := NewNodeProvider()
	vars, err := p.GetEnvVars("node", "/tmp/node", "18.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	npmGlobal := filepath.Join("/tmp/node", "npm-global")
	if val, ok := vars["NPM_CONFIG_PREFIX"]; !ok || val != npmGlobal {
		t.Errorf("expected NPM_CONFIG_PREFIX=%s, got %s", npmGlobal, val)
	}
}

func TestNodeProvider_GenerateShims(t *testing.T) {
	p := NewNodeProvider()
	shims, err := p.GenerateShims("node", "/tmp/node", "18.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(shims) != 3 {
		t.Errorf("expected 3 shims, got %d", len(shims))
	}
}
