// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GolangProvider implements the Provider interface for Go SDK.
type GolangProvider struct {
	generic *GenericProvider
}

// NewGolangProvider creates a new Go provider.
func NewGolangProvider() *GolangProvider {
	return &GolangProvider{
		generic: NewGenericProvider(),
	}
}

// Name returns the provider identifier.
func (g *GolangProvider) Name() string {
	return "go"
}

// Install performs Go-specific installation.
func (g *GolangProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	return g.generic.Install(ctx, installPath, artifactPath, version)
}

// PostInstall sets up GOPATH.
func (g *GolangProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	gopath := filepath.Join(installPath, "gopath")
	if err := os.MkdirAll(filepath.Join(gopath, "bin"), 0755); err != nil {
		return NewProviderError("golang", "go", version, "failed to create GOPATH", err)
	}
	if err := os.MkdirAll(filepath.Join(gopath, "pkg"), 0755); err != nil {
		return NewProviderError("golang", "go", version, "failed to create GOPATH/pkg", err)
	}
	if err := os.MkdirAll(filepath.Join(gopath, "src"), 0755); err != nil {
		return NewProviderError("golang", "go", version, "failed to create GOPATH/src", err)
	}
	return nil
}

// GenerateShims generates shims for go and gofmt.
func (g *GolangProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
	shims := make(map[string]string)

	executables := []string{"go", "gofmt"}
	for _, exe := range executables {
		exePath := filepath.Join(installPath, "bin", exe)
		if runtime.GOOS == "windows" {
			exePath += ".exe"
		}

		shimContent := g.generateGoShim(exe, exePath, installPath, version)
		shims[exe] = shimContent
	}

	return shims, nil
}

// DetectVersion detects Go version.
func (g *GolangProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	goPath := filepath.Join(installPath, "bin", "go")
	if runtime.GOOS == "windows" {
		goPath += ".exe"
	}

	cmd := exec.CommandContext(ctx, goPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return "", NewProviderError("golang", "go", "", "failed to detect version", err)
	}

	version := strings.TrimSpace(string(output))
	parts := strings.Fields(version)
	if len(parts) >= 3 {
		version = strings.TrimPrefix(parts[2], "go")
		return version, nil
	}

	return "", NewProviderError("golang", "go", "", "failed to parse version", nil)
}

// ListExecutables returns Go executables relative to installPath.
func (g *GolangProvider) ListExecutables(installPath string, version string) ([]string, error) {
	executables := []string{filepath.Join("bin", "go"), filepath.Join("bin", "gofmt")}
	if runtime.GOOS == "windows" {
		for i := range executables {
			executables[i] += ".exe"
		}
	}
	return executables, nil
}

// Uninstall performs Go-specific cleanup.
func (g *GolangProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	gopath := filepath.Join(installPath, "gopath")
	if err := os.RemoveAll(gopath); err != nil {
		return NewProviderError("golang", "go", version, "failed to remove GOPATH", err)
	}
	return nil
}

// generateGoShim generates a Go-specific shim.
func (g *GolangProvider) generateGoShim(name, exePath, installPath, version string) string {
	gopath := filepath.Join(installPath, "gopath")

	if runtime.GOOS == "windows" {
		return fmt.Sprintf(`@echo off
REM UniRTM shim for %s (version %s)
set "GOPATH=%s"
"%s" %%*
`, name, version, gopath, exePath)
	}

	return fmt.Sprintf(`#!/bin/sh
# UniRTM shim for %s (version %s)
export GOPATH="%s"
exec "%s" "$@"
`, name, version, gopath, exePath)
}
