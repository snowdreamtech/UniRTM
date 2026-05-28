// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"path/filepath"
	"testing"
)

func TestPythonProvider_Name(t *testing.T) {
	p := NewPythonProvider()
	if p.Name() != "python" {
		t.Errorf("expected python, got %s", p.Name())
	}
}

func TestPythonProvider_ListExecutables(t *testing.T) {
	p := NewPythonProvider()
	execs, err := p.ListExecutables("python", "/tmp", "3.10.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(execs) != 4 {
		t.Errorf("expected 4 executables, got %d", len(execs))
	}
}

func TestPythonProvider_GetBinPaths(t *testing.T) {
	p := NewPythonProvider()
	paths, err := p.GetBinPaths("python", "/tmp/py", "3.10.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(paths))
	}
	expected := filepath.Join("/tmp/py", "bin")
	if paths[0] != expected {
		t.Errorf("expected %s, got %s", expected, paths[0])
	}
}

func TestPythonProvider_GetEnvVars(t *testing.T) {
	p := NewPythonProvider()
	vars, err := p.GetEnvVars("python", "/tmp/py", "3.10.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	venv := filepath.Join("/tmp/py", "venv")
	if val, ok := vars["VIRTUAL_ENV"]; !ok || val != venv {
		t.Errorf("expected VIRTUAL_ENV=%s, got %s", venv, val)
	}
}

func TestPythonProvider_GenerateShims(t *testing.T) {
	p := NewPythonProvider()
	shims, err := p.GenerateShims("python", "/tmp/py", "3.10.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(shims) != 4 {
		t.Errorf("expected 4 shims, got %d", len(shims))
	}
}
