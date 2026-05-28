// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func createTestZip(t *testing.T, dir string) string {
	zipPath := filepath.Join(dir, "test.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	f1, err := w.Create("hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f1.Write([]byte("hello world"))
	w.Close()

	return zipPath
}

func createTestTar(t *testing.T, dir string) string {
	tarPath := filepath.Join(dir, "test.tar")
	f, err := os.Create(tarPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tw := tar.NewWriter(f)

	hdr := &tar.Header{
		Name: "hello.txt",
		Mode: 0600,
		Size: int64(len("hello tar")),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("hello tar")); err != nil {
		t.Fatal(err)
	}

	tw.Close()

	return tarPath
}

func TestGenericProvider_Extract(t *testing.T) {
	p := NewGenericProvider()

	tmpDir := t.TempDir()

	zipPath := createTestZip(t, tmpDir)
	zipOut := filepath.Join(tmpDir, "zipout")
	err := p.extractZip(zipPath, zipOut)
	if err != nil {
		t.Errorf("extractZip failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(zipOut, "hello.txt")); err != nil {
		t.Errorf("expected hello.txt in zip extraction")
	}

	tarPath := createTestTar(t, tmpDir)
	tarOut := filepath.Join(tmpDir, "tarout")

	tf, err := os.Open(tarPath)
	if err != nil {
		t.Fatal(err)
	}
	err = p.extractTar(tf, tarOut)
	tf.Close()
	if err != nil {
		t.Errorf("extractTar failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tarOut, "hello.txt")); err != nil {
		t.Errorf("expected hello.txt in tar extraction")
	}
}

func TestGenericProvider_FlattenDirectory(t *testing.T) {
	p := NewGenericProvider()

	ctx := context.Background()
	tmpDir := t.TempDir()

	// create nested dir structure
	subDir := filepath.Join(tmpDir, "wrapper")
	targetDir := filepath.Join(subDir, "real_tool")
	os.MkdirAll(targetDir, 0755)

	f, _ := os.Create(filepath.Join(targetDir, "exec.bin"))
	f.Close()

	err := p.flattenDirectory(ctx, tmpDir)
	if err != nil {
		t.Errorf("flattenDirectory failed: %v", err)
	}

	// should be flattened to tmpDir/real_tool/exec.bin or tmpDir/exec.bin depending on logic
	// the logic in generic.go looks for a single dir inside the root
	if _, err := os.Stat(filepath.Join(tmpDir, "exec.bin")); err != nil {
		t.Errorf("flattenDirectory did not flatten properly: %v", err)
	}
}

func TestGenericProvider_CopyFile(t *testing.T) {
	p := NewGenericProvider()
	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "src.txt")
	dst := filepath.Join(tmpDir, "dst.txt")

	os.WriteFile(src, []byte("content"), 0644)

	err := p.copyFile(src, dst)
	if err != nil {
		t.Errorf("copyFile failed: %v", err)
	}

	b, _ := os.ReadFile(dst)
	if !bytes.Equal(b, []byte("content")) {
		t.Errorf("copyFile content mismatch")
	}
}

func TestGenericProvider_FindExecutables(t *testing.T) {
	p := NewGenericProvider()
	tmpDir := t.TempDir()

	exeName := "mytool"
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}

	f, _ := os.Create(filepath.Join(tmpDir, exeName))
	f.Close()
	if runtime.GOOS != "windows" {
		os.Chmod(filepath.Join(tmpDir, exeName), 0755)
	}

	execs, err := p.findExecutables(tmpDir)
	if err != nil {
		t.Errorf("findExecutables failed: %v", err)
	}
	if len(execs) == 0 {
		t.Errorf("findExecutables did not find anything")
	}
}

func TestGenericProvider_RelativizeAllSymlinks(t *testing.T) {
	p := NewGenericProvider()
	tmpDir := t.TempDir()

	// create a target
	target := filepath.Join(tmpDir, "target.txt")
	os.WriteFile(target, []byte(""), 0644)

	// create absolute symlink
	link := filepath.Join(tmpDir, "link.txt")
	os.Symlink(target, link)

	err := p.relativizeAllSymlinks(tmpDir)
	if err != nil {
		t.Errorf("relativizeAllSymlinks failed: %v", err)
	}
}

func TestGenericProvider_ValidateInstallDir(t *testing.T) {
	p := NewGenericProvider()

	err := p.validateInstallDir("nonexistent")
	if err == nil {
		t.Errorf("expected error for nonexistent dir")
	}

	tmpDir := t.TempDir()
	err = p.validateInstallDir(tmpDir)
	if err != nil {
		t.Errorf("expected no error for empty dir")
	}

	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte(""), 0644)
	err = p.validateInstallDir(tmpDir)
	if err != nil {
		t.Errorf("expected no error for valid dir")
	}
}

func TestGenericProvider_GenerateWindowsShim(t *testing.T) {
	p := NewGenericProvider()
	shimContent := p.generateWindowsShim("tool", "bin\\tool.exe")
	if len(shimContent) == 0 {
		t.Errorf("expected shim content")
	}
}

func TestGenericProvider_IsExecutableExtension(t *testing.T) {
	if !isExecutableExtension(".exe") {
		t.Errorf("expected .exe to be executable")
	}
	if isExecutableExtension(".longext") {
		t.Errorf("expected .longext to not be executable (too long)")
	}
	if isExecutableExtension(".123") {
		t.Errorf("expected .123 to not be executable (does not start with letter)")
	}
}

func TestGenericProvider_ExtractArtifact(t *testing.T) {
	p := NewGenericProvider()
	ctx := context.Background()
	tmpDir := t.TempDir()

	zipPath := createTestZip(t, tmpDir)
	outDir := filepath.Join(tmpDir, "out")
	err := p.extractArtifact(ctx, zipPath, outDir)
	if err != nil {
		t.Errorf("extractArtifact failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "hello.txt")); err != nil {
		t.Errorf("expected hello.txt in extracted artifact")
	}
}

func TestGenericProvider_PickBestExecutables(t *testing.T) {
	p := NewGenericProvider()
	execs := []string{"tool", "tool.md", "other"}
	best := p.pickBestExecutables(execs, "tool")
	if len(best) != 2 || best[0] != "tool" {
		t.Errorf("expected [tool, other], got %v", best)
	}
}

func TestGenericProvider_Install(t *testing.T) {
	p := NewGenericProvider()
	ctx := context.Background()

	tmpDir := t.TempDir()
	installPath := filepath.Join(tmpDir, "install")

	// Test failure cases for Install
	err := p.Install(ctx, "tool", installPath, "nonexistent", "1.0.0")
	if err == nil {
		t.Errorf("expected error for nonexistent artifact")
	}

	zipPath := createTestZip(t, tmpDir)

	err = p.Install(ctx, "tool", installPath, zipPath, "1.0.0")
	if err != nil {
		t.Errorf("expected no error for valid zip artifact install, got: %v", err)
	}

	// Also test non-archive file copy
	exePath := filepath.Join(tmpDir, "mytool.exe")
	os.WriteFile(exePath, []byte("binary"), 0755)

	installPath2 := filepath.Join(tmpDir, "install2")
	err = p.Install(ctx, "tool", installPath2, exePath, "1.0.0")
	if err != nil {
		t.Errorf("expected no error for single file install, got: %v", err)
	}
}
