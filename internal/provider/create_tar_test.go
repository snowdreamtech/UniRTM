// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTarCoverage(t *testing.T) {
	p := NewGenericProvider()
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar.gz")
	destDir := filepath.Join(tmpDir, "dest")

	// Create a tar.gz file
	f, err := os.Create(tarPath)
	if err != nil {
		t.Fatal(err)
	}
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	// Add a file
	content := []byte("hello tar")
	hdr := &tar.Header{
		Name: "testtar.txt",
		Mode: 0600,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}

	tw.Close()
	gw.Close()
	f.Close()

	err = p.extractArtifact(context.Background(), tarPath, destDir)
	if err != nil {
		t.Errorf("Unexpected error extracting tar: %v", err)
	}

	if _, err := os.Stat(filepath.Join(destDir, "testtar.txt")); os.IsNotExist(err) {
		t.Errorf("File was not extracted")
	}
}
