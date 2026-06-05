// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGenericProvider_InstallCoverage(t *testing.T) {
	p := NewGenericProvider()
	ctx := context.Background()

	// Test 1: extractArtifact fails (non-existent artifact)
	err := p.Install(ctx, "testtool", "installPath", "artifactPath", "1.0.0")
	if err == nil {
		t.Errorf("Expected error for non-existent artifact")
	}

	// Test 2: non-archive file (copy directly to binDir)
	tmpDir := t.TempDir()
	installPath := filepath.Join(tmpDir, "install_dir")
	artifactPath := filepath.Join(tmpDir, "mytool.exe")
	os.WriteFile(artifactPath, []byte("binary data"), 0755)

	err = p.Install(ctx, "mytool", installPath, artifactPath, "1.0.0")
	if err != nil {
		t.Errorf("Unexpected error for non-archive install: %v", err)
	}

	// Check if bin/mytool.exe was created
	dstPath := filepath.Join(installPath, "bin", "mytool.exe")
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Errorf("Expected executable to be copied to %s", dstPath)
	}

	// Test 3: isExecutable test
	tmpFile := filepath.Join(tmpDir, "exec.exe")
	os.WriteFile(tmpFile, []byte("data"), 0755)
	fi, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if !p.isExecutable(fi) {
		t.Errorf("Expected file to be executable")
	}
}
