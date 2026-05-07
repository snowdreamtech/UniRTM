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

// PythonProvider implements the Provider interface for Python.
type PythonProvider struct {
	generic *GenericProvider
}

// NewPythonProvider creates a new Python provider.
func NewPythonProvider() *PythonProvider {
	return &PythonProvider{
		generic: NewGenericProvider(),
	}
}

// Name returns the provider identifier.
func (p *PythonProvider) Name() string {
	return "python"
}

// Install performs Python-specific installation.
func (p *PythonProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	return p.generic.Install(ctx, installPath, artifactPath, version)
}

// PostInstall creates a virtual environment.
func (p *PythonProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	pythonPath := filepath.Join(installPath, "bin", "python3")
	if runtime.GOOS == "windows" {
		pythonPath = filepath.Join(installPath, "bin", "python.exe")
	}

	venvDir := filepath.Join(installPath, "venv")
	cmd := exec.CommandContext(ctx, pythonPath, "-m", "venv", venvDir)
	if err := cmd.Run(); err != nil {
		return NewProviderError("python", "python", version, "failed to create virtual environment", err)
	}

	return nil
}

// GenerateShims generates shims for python, pip, and python3.
func (p *PythonProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
	shims := make(map[string]string)

	executables := []string{"python", "python3", "pip", "pip3"}
	for _, exe := range executables {
		exePath := filepath.Join(installPath, "bin", exe)
		if runtime.GOOS == "windows" {
			exePath += ".exe"
		}

		shimContent := p.generatePythonShim(exe, exePath, installPath, version)
		shims[exe] = shimContent
	}

	return shims, nil
}

// DetectVersion detects Python version.
func (p *PythonProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	pythonPath := filepath.Join(installPath, "bin", "python3")
	if runtime.GOOS == "windows" {
		pythonPath = filepath.Join(installPath, "bin", "python.exe")
	}

	cmd := exec.CommandContext(ctx, pythonPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", NewProviderError("python", "python", "", "failed to detect version", err)
	}

	version := strings.TrimSpace(string(output))
	version = strings.TrimPrefix(version, "Python ")
	return version, nil
}

// ListExecutables returns Python executables.
func (p *PythonProvider) ListExecutables(installPath string, version string) ([]string, error) {
	executables := []string{"python", "python3", "pip", "pip3"}
	if runtime.GOOS == "windows" {
		for i := range executables {
			executables[i] += ".exe"
		}
	}
	return executables, nil
}

// Uninstall performs Python-specific cleanup.
func (p *PythonProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	venvDir := filepath.Join(installPath, "venv")
	if err := os.RemoveAll(venvDir); err != nil {
		return NewProviderError("python", "python", version, "failed to remove virtual environment", err)
	}
	return nil
}

// generatePythonShim generates a Python-specific shim.
func (p *PythonProvider) generatePythonShim(name, exePath, installPath, version string) string {
	venvDir := filepath.Join(installPath, "venv")

	if runtime.GOOS == "windows" {
		return fmt.Sprintf(`@echo off
REM UniRTM shim for %s (version %s)
set "VIRTUAL_ENV=%s"
"%s" %%*
`, name, version, venvDir, exePath)
	}

	return fmt.Sprintf(`#!/bin/sh
# UniRTM shim for %s (version %s)
export VIRTUAL_ENV="%s"
exec "%s" "$@"
`, name, version, venvDir, exePath)
}
