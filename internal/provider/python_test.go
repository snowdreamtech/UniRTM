// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/stretchr/testify/assert"
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
	expectedLen := 2
	if env.RuntimeGOOS == "windows" {
		expectedLen = 3
	}
	if len(paths) != expectedLen {
		t.Errorf("expected %d paths, got %d", expectedLen, len(paths))
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

func TestPythonProvider_GetRealPythonPath(t *testing.T) {
	p := NewPythonProvider()
	tempDir := t.TempDir()

	origGOOS := env.RuntimeGOOS
	defer func() { env.RuntimeGOOS = origGOOS }()

	// 1. Windows: root python.exe
	env.RuntimeGOOS = "windows"
	rootPy := filepath.Join(tempDir, "python.exe")
	_ = os.WriteFile(rootPy, []byte("echo python"), 0755)
	assert.Equal(t, rootPy, p.getRealPythonPath(tempDir))
	_ = os.Remove(rootPy)

	// 2. Windows: install/python.exe
	installDir := filepath.Join(tempDir, "install")
	_ = os.MkdirAll(installDir, 0755)
	installPy := filepath.Join(installDir, "python.exe")
	_ = os.WriteFile(installPy, []byte("echo install python"), 0755)
	assert.Equal(t, installPy, p.getRealPythonPath(tempDir))
	_ = os.Remove(installPy)

	// 3. Windows: python_d.exe
	rootPyD := filepath.Join(tempDir, "python_d.exe")
	_ = os.WriteFile(rootPyD, []byte("echo debug python"), 0755)
	assert.Equal(t, rootPyD, p.getRealPythonPath(tempDir))
	_ = os.Remove(rootPyD)

	// 4. Unix: bin/python3
	env.RuntimeGOOS = "linux"
	binDir := filepath.Join(tempDir, "bin")
	_ = os.MkdirAll(binDir, 0755)
	binPy3 := filepath.Join(binDir, "python3")
	_ = os.WriteFile(binPy3, []byte("#!/bin/sh"), 0755)
	assert.Equal(t, binPy3, p.getRealPythonPath(tempDir))
	_ = os.Remove(binPy3)

	// 5. Linux: bin/python
	binPy := filepath.Join(binDir, "python")
	_ = os.WriteFile(binPy, []byte("#!/bin/sh"), 0755)
	assert.Equal(t, binPy, p.getRealPythonPath(tempDir))
	_ = os.Remove(binPy)

	// 6. Linux: versioned directory and binary (3.12.0 -> bin/python3.12)
	verDirLinux := filepath.Join(tempDir, "3.12.0")
	binDirLinux := filepath.Join(verDirLinux, "bin")
	_ = os.MkdirAll(binDirLinux, 0755)
	binPyVer := filepath.Join(binDirLinux, "python3.12")
	_ = os.WriteFile(binPyVer, []byte("#!/bin/sh"), 0755)
	assert.Equal(t, binPyVer, p.getRealPythonPath(verDirLinux))

	// 7. macOS: bin/python3
	env.RuntimeGOOS = "darwin"
	binPy3Darwin := filepath.Join(binDir, "python3")
	_ = os.WriteFile(binPy3Darwin, []byte("#!/bin/sh"), 0755)
	assert.Equal(t, binPy3Darwin, p.getRealPythonPath(tempDir))
	_ = os.Remove(binPy3Darwin)

	// 8. macOS: versioned directory and binary (3.14.7 -> bin/python3.14)
	verDirMac := filepath.Join(tempDir, "3.14.7")
	binDirMac := filepath.Join(verDirMac, "bin")
	_ = os.MkdirAll(binDirMac, 0755)
	binPy314 := filepath.Join(binDirMac, "python3.14")
	_ = os.WriteFile(binPy314, []byte("#!/bin/sh"), 0755)
	assert.Equal(t, binPy314, p.getRealPythonPath(verDirMac))
}

