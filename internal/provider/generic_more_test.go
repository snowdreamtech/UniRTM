// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGenericProvider_More(t *testing.T) {
	p := NewGenericProvider()

	dir := t.TempDir()

	dummyFile := filepath.Join(dir, "dummy.txt")
	os.WriteFile(dummyFile, []byte("test"), 0644)

	info, _ := os.Stat(dummyFile)
	p.isExecutable(info)

	dest := filepath.Join(dir, "dest.txt")
	p.copyFile(dummyFile, dest)

	p.generateShimScript(filepath.Join(dir, "shim"), dest)

	p.validateInstallDir(dir)

	ctx := context.Background()
	p.extractArtifact(ctx, dummyFile, filepath.Join(dir, "out"))

	dummyZip := filepath.Join(dir, "dummy.zip")
	os.WriteFile(dummyZip, []byte("dummy zip"), 0644)
	p.extractArtifact(ctx, dummyZip, filepath.Join(dir, "outzip"))

	dummyTar := filepath.Join(dir, "dummy.tar.gz")
	os.WriteFile(dummyTar, []byte("dummy tar"), 0644)
	p.extractArtifact(ctx, dummyTar, filepath.Join(dir, "outtar"))

	isExecutableExtension(".exe")
}
