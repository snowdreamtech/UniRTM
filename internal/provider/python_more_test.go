// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPythonProvider_More(t *testing.T) {
	p := NewPythonProvider()

	dir := t.TempDir()

	// Create dummy python executable
	binDir := filepath.Join(dir, "bin")
	os.MkdirAll(binDir, 0755)

	pyExe := filepath.Join(binDir, "python")
	os.WriteFile(pyExe, []byte("#!/bin/sh\nexit 0\n"), 0755)

	p.getRealPythonPath(pyExe)
	p.PostInstall(context.Background(), "tool", dir, "1.0.0")
	p.DetectVersion(context.Background(), "tool", dir)
	p.ListExecutables("tool", dir, "1.0.0")
	p.Uninstall(context.Background(), "tool", dir, "1.0.0")
	p.generatePythonShim("python", pyExe, dir, "1.0.0")
}
