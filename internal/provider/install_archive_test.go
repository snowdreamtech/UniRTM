// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenericProvider_InstallArchiveCoverage(t *testing.T) {
	p := NewGenericProvider()
	ctx := context.Background()

	tmpDir := t.TempDir()
	installPath := filepath.Join(tmpDir, "install_dir")
	artifactPath := filepath.Join(tmpDir, "test.zip")

	// Create a valid zip file
	f, err := os.Create(artifactPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)

	// Add a directory
	dirHeader := &zip.FileHeader{
		Name:   "subdir/",
		Method: zip.Store,
	}
	dirHeader.SetMode(0755 | os.ModeDir)
	zw.CreateHeader(dirHeader)

	// Add an executable file inside
	header := &zip.FileHeader{
		Name:   "subdir/mytool",
		Method: zip.Store,
	}
	header.SetMode(0755) // Executable mode

	w, err := zw.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("binary data"))

	// Add a Zip Slip file
	slipHeader := &zip.FileHeader{
		Name:   "../slip.txt",
		Method: zip.Store,
	}
	wSlip, _ := zw.CreateHeader(slipHeader)
	wSlip.Write([]byte("slip"))

	zw.Close()
	f.Close()

	err = p.Install(ctx, "mytool", installPath, artifactPath, "1.0.0")
	if err != nil {
		t.Errorf("Unexpected error for archive install: %v", err)
	}

	// Check if bin/mytool was created (since flattenDirectory will lift subdir/mytool)
	dstPath := filepath.Join(installPath, "bin", "mytool")
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Errorf("Expected executable to be linked to %s", dstPath)
	}
}

func TestGenericProvider_ExtractArtifactMoreCoverage(t *testing.T) {
	p := NewGenericProvider()
	ctx := context.Background()

	tmpDir := t.TempDir()

	// Test unsupported archive type
	unsupportedPath := filepath.Join(tmpDir, "test.unsupported")
	os.WriteFile(unsupportedPath, []byte("dummy"), 0644)
	err := p.extractArtifact(ctx, unsupportedPath, tmpDir)
	if err == nil {
		t.Errorf("Expected error for unsupported archive type")
	}

	// Test extracting a single compressed file (.gz)
	gzPath := filepath.Join(tmpDir, "single.gz")
	fGz, _ := os.Create(gzPath)
	gw := gzip.NewWriter(fGz)
	gw.Write([]byte("single decompressed data"))
	gw.Close()
	fGz.Close()

	err = p.extractArtifact(ctx, gzPath, tmpDir)
	if err != nil {
		t.Errorf("Unexpected error for .gz extraction: %v", err)
	}
	extractedSinglePath := filepath.Join(tmpDir, "single")
	if _, err := os.Stat(extractedSinglePath); os.IsNotExist(err) {
		t.Errorf("Expected single file to be extracted to %s", extractedSinglePath)
	}

	// Test extracting a tar file
	tarPath := filepath.Join(tmpDir, "test.tar")
	fTar, _ := os.Create(tarPath)
	tw := tar.NewWriter(fTar)
	tw.WriteHeader(&tar.Header{
		Name:     "tardir/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	})
	tw.WriteHeader(&tar.Header{
		Name:     "tarfile.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len("tar data")),
	})
	tw.Write([]byte("tar data"))
	tw.WriteHeader(&tar.Header{
		Name:     "tarsymlink",
		Typeflag: tar.TypeSymlink,
		Linkname: "tarfile.txt",
	})
	tw.WriteHeader(&tar.Header{
		Name:     "../tarslip.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     4,
	})
	tw.Write([]byte("slip"))
	tw.Close()
	fTar.Close()

	err = p.extractArtifact(ctx, tarPath, tmpDir)
	if err != nil {
		t.Errorf("Unexpected error for .tar extraction: %v", err)
	}

	// Test extracting a tar.gz file
	targzPath := filepath.Join(tmpDir, "test.tar.gz")
	fTargz, _ := os.Create(targzPath)
	gw2 := gzip.NewWriter(fTargz)
	tw2 := tar.NewWriter(gw2)
	tw2.WriteHeader(&tar.Header{
		Name:     "targzdir/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	})
	tw2.WriteHeader(&tar.Header{
		Name:     "targzfile.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len("targz data")),
	})
	tw2.Write([]byte("targz data"))
	tw2.WriteHeader(&tar.Header{
		Name:     "targzsymlink",
		Typeflag: tar.TypeSymlink,
		Linkname: "targzfile.txt",
	})
	tw2.Close()
	gw2.Close()
	fTargz.Close()

	err = p.extractArtifact(ctx, targzPath, tmpDir)
	if err != nil {
		t.Errorf("Unexpected error for .tar.gz extraction: %v", err)
	}

	// Test extracting an unsupported archive format
	unsupportedPath2 := filepath.Join(tmpDir, "test.unsupported2")
	os.WriteFile(unsupportedPath2, []byte("data"), 0644)
	err = p.extractArtifact(ctx, unsupportedPath2, tmpDir)
	if err == nil || !strings.Contains(err.Error(), "unsupported archive type") {
		t.Errorf("Expected unsupported archive type error, got %v", err)
	}

	// Test extracting into a read-only directory to trigger os.MkdirAll / os.OpenFile errors
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	os.MkdirAll(readOnlyDir, 0555) // Read and execute only, no write

	// Create a dummy zip file
	zipPath := filepath.Join(tmpDir, "test_readonly.zip")
	fZip, _ := os.Create(zipPath)
	zw := zip.NewWriter(fZip)
	fWriter, _ := zw.Create("file.txt")
	fWriter.Write([]byte("data"))
	zw.Close()
	fZip.Close()

	// zip extraction error
	err = p.extractArtifact(ctx, zipPath, readOnlyDir)
	if err == nil {
		t.Error("Expected error when extracting zip to read-only directory, got nil")
	}

	// tar extraction error
	err = p.extractArtifact(ctx, tarPath, readOnlyDir)
	if err == nil {
		t.Error("Expected error when extracting tar to read-only directory, got nil")
	}

	// Reset permissions so tmpDir cleanup doesn't fail
	os.Chmod(readOnlyDir, 0755)
}
