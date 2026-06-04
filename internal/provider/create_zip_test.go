// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractZipCoverage(t *testing.T) {
	p := NewGenericProvider()
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")
	destDir := filepath.Join(tmpDir, "dest")

	// Create a zip file
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)

	// Add a file
	w, err := zw.Create("testfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("hello"))

	// Add a directory
	_, err = zw.Create("testdir/")
	if err != nil {
		t.Fatal(err)
	}

	zw.Close()
	f.Close()

	err = p.extractArtifact(context.Background(), zipPath, destDir)
	if err != nil {
		t.Errorf("Unexpected error extracting zip: %v", err)
	}

	if _, err := os.Stat(filepath.Join(destDir, "testfile.txt")); os.IsNotExist(err) {
		t.Errorf("File was not extracted")
	}
}
