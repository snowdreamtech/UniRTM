// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestContainerProvider_Name(t *testing.T) {
	p := NewContainerProvider("podman")
	if p.Name() != "podman" {
		t.Errorf("expected podman, got %s", p.Name())
	}
}

func TestContainerProvider_Install(t *testing.T) {
	// Create a temporary directory for our mock environment
	tempDir := t.TempDir()

	// Create a fake docker executable to bypass engine detection and pull
	fakeDockerDir := filepath.Join(tempDir, "fake-bin")
	if err := os.MkdirAll(fakeDockerDir, 0755); err != nil {
		t.Fatalf("failed to create fake bin dir: %v", err)
	}

	fakeDockerPath := filepath.Join(fakeDockerDir, "docker")
	if runtime.GOOS == "windows" {
		fakeDockerPath += ".bat"
		os.WriteFile(fakeDockerPath, []byte("@echo off\nexit 0"), 0755)
	} else {
		os.WriteFile(fakeDockerPath, []byte("#!/bin/sh\nexit 0"), 0755)
	}

	// Prepend fake bin dir to PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)
	os.Setenv("PATH", fakeDockerDir+string(os.PathListSeparator)+originalPath)

	p := NewContainerProvider("docker")
	ctx := context.Background()
	installPath := filepath.Join(tempDir, "install")

	// Test Installation
	err := p.Install(ctx, "ghcr.io/aquasec/trivy", installPath, "", "0.48.0")
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify wrapper script was created
	basename := "trivy"
	wrapperName := basename
	if runtime.GOOS == "windows" {
		wrapperName += ".cmd"
	}
	wrapperPath := filepath.Join(installPath, "bin", wrapperName)

	content, err := os.ReadFile(wrapperPath)
	if err != nil {
		t.Fatalf("Failed to read wrapper script: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "CONTAINER_ENGINE=") {
		t.Errorf("Wrapper script missing CONTAINER_ENGINE variable")
	}
	if !strings.Contains(contentStr, "IMAGE=") {
		t.Errorf("Wrapper script missing IMAGE variable")
	}
	if runtime.GOOS == "windows" {
		if !strings.Contains(contentStr, "ghcr.io/aquasec/trivy:0.48.0") {
			t.Errorf("Wrapper script missing correct image tag")
		}
	} else {
		if !strings.Contains(contentStr, "ghcr.io/aquasec/trivy:0.48.0") {
			t.Errorf("Wrapper script missing correct image tag")
		}
		if !strings.Contains(contentStr, "exec \"$CONTAINER_ENGINE\" run") {
			t.Errorf("Wrapper script missing exec command")
		}
	}
}

func TestContainerProvider_Install_Digest(t *testing.T) {
	tempDir := t.TempDir()

	fakeDockerDir := filepath.Join(tempDir, "fake-bin")
	os.MkdirAll(fakeDockerDir, 0755)

	fakeDockerPath := filepath.Join(fakeDockerDir, "docker")
	if runtime.GOOS == "windows" {
		fakeDockerPath += ".bat"
		os.WriteFile(fakeDockerPath, []byte("@echo off\nexit 0"), 0755)
	} else {
		os.WriteFile(fakeDockerPath, []byte("#!/bin/sh\nexit 0"), 0755)
	}

	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)
	os.Setenv("PATH", fakeDockerDir+string(os.PathListSeparator)+originalPath)

	p := NewContainerProvider("docker")
	ctx := context.Background()
	installPath := filepath.Join(tempDir, "install")

	digest := "sha256:123456"
	err := p.Install(ctx, "alpine", installPath, "", digest)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	wrapperName := "alpine"
	if runtime.GOOS == "windows" {
		wrapperName += ".cmd"
	}
	wrapperPath := filepath.Join(installPath, "bin", wrapperName)

	content, err := os.ReadFile(wrapperPath)
	if err != nil {
		t.Fatalf("Failed to read wrapper script: %v", err)
	}

	if !strings.Contains(string(content), "alpine@sha256:123456") {
		t.Errorf("Wrapper script missing correct digest image reference")
	}
}

func TestContainerProvider_Methods(t *testing.T) {
	p := NewContainerProvider("container")
	ctx := context.Background()

	// PostInstall
	if err := p.PostInstall(ctx, "alpine", "/tmp", "latest"); err != nil {
		t.Errorf("PostInstall failed: %v", err)
	}

	// GenerateShims
	shims, err := p.GenerateShims("alpine", "/tmp", "latest")
	if err != nil {
		t.Errorf("GenerateShims failed: %v", err)
	}
	if len(shims) != 0 {
		t.Errorf("GenerateShims should return empty map, got %d items", len(shims))
	}

	// DetectVersion
	ver, err := p.DetectVersion(ctx, "alpine", "/tmp")
	if err != nil {
		t.Errorf("DetectVersion failed: %v", err)
	}
	if ver != "unknown" {
		t.Errorf("expected unknown, got %s", ver)
	}

	// ListExecutables
	execs, err := p.ListExecutables("ghcr.io/aquasec/trivy", "/tmp", "latest")
	if err != nil {
		t.Errorf("ListExecutables failed: %v", err)
	}
	if len(execs) != 1 {
		t.Fatalf("expected 1 executable, got %d", len(execs))
	}

	expectedExec := "trivy"
	if runtime.GOOS == "windows" {
		expectedExec = "trivy.cmd"
	}
	if execs[0] != expectedExec {
		t.Errorf("expected %s, got %s", expectedExec, execs[0])
	}

	// GetBinPaths
	bins, err := p.GetBinPaths("alpine", "/tmp/install", "latest")
	if err != nil {
		t.Errorf("GetBinPaths failed: %v", err)
	}
	if len(bins) != 1 || filepath.Base(bins[0]) != "bin" {
		t.Errorf("expected bin path ending with 'bin', got %v", bins)
	}

	// GetEnvVars
	envs, err := p.GetEnvVars("alpine", "/tmp", "latest")
	if err != nil {
		t.Errorf("GetEnvVars failed: %v", err)
	}
	if len(envs) != 0 {
		t.Errorf("expected 0 env vars, got %d", len(envs))
	}

	// Uninstall
	if err := p.Uninstall(ctx, "alpine", "/tmp", "latest"); err != nil {
		t.Errorf("Uninstall failed: %v", err)
	}
}
