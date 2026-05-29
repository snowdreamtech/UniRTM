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
func (p *PythonProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	return p.generic.Install(ctx, tool, installPath, artifactPath, version)
}

// getRealPythonPath resolves the actual Python binary path.
// This is necessary on Windows because generic.Install creates symlinks in bin/,
// and executing a symlink on Windows causes DLL resolution failures (0xc0000135)
// since vcruntime140.dll is next to the real binary, not the symlink.
func (p *PythonProvider) getRealPythonPath(installPath string) string {
	if runtime.GOOS == "windows" {
		binPy := filepath.Join(installPath, "bin", "python.exe")
		if realPy, err := filepath.EvalSymlinks(binPy); err == nil {
			return realPy
		}

		// python-build-standalone on Windows often extracts python.exe to the root
		rootPy := filepath.Join(installPath, "python.exe")
		if _, err := os.Stat(rootPy); err == nil {
			return rootPy
		}

		// Fallback for some standalone builds
		installPy := filepath.Join(installPath, "install", "python.exe")
		if _, err := os.Stat(installPy); err == nil {
			return installPy
		}

		// Try bin just in case
		return binPy
	}
	return filepath.Join(installPath, "bin", "python3")
}

// PostInstall creates a virtual environment.
func (p *PythonProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	pythonPath := p.getRealPythonPath(installPath)

	venvDir := filepath.Join(installPath, "venv")
	cmd := exec.CommandContext(ctx, pythonPath, "-m", "venv", venvDir)

	// Ensure the DLLs in installPath are discoverable by the newly created venv
	// python executable during ensurepip.
	env := os.Environ()
	pathVar := "PATH"
	for i, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			env[i] = "PATH=" + installPath + string(os.PathListSeparator) + e[5:]
			pathVar = ""
			break
		}
	}
	if pathVar != "" {
		env = append(env, "PATH="+installPath)
	}
	cmd.Env = env

	if err := cmd.Run(); err != nil {
		return NewProviderError("python", "python", version, "failed to create virtual environment", err)
	}

	return nil
}

// GenerateShims generates shims for python, pip, and python3.
func (p *PythonProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	shims := make(map[string]string)

	executables := []string{"python", "python3", "pip", "pip3"}
	for _, exe := range executables {
		var exePath string
		// Point the shim directly to the venv executables.
		// This natively solves the Windows symlink DLL resolution issue
		// and ensures the tool inherently uses its isolated environment.
		if runtime.GOOS == "windows" {
			exePath = filepath.Join(installPath, "venv", "Scripts", exe+".exe")
		} else {
			exePath = filepath.Join(installPath, "venv", "bin", exe)
		}

		shimContent := p.generatePythonShim(exe, exePath, installPath, version)
		shims[exe] = shimContent
	}

	return shims, nil
}

// DetectVersion detects Python version.
func (p *PythonProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	pythonPath := p.getRealPythonPath(installPath)

	cmd := exec.CommandContext(ctx, pythonPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", NewProviderError("python", "python", "", "failed to detect version", err)
	}

	version := strings.TrimSpace(string(output))
	version = strings.TrimPrefix(version, "Python ")
	return version, nil
}

// ListExecutables returns Python executables relative to installPath.
func (p *PythonProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	// Expose the venv executables so they can be discovered by generic logic if needed
	var executables []string
	if runtime.GOOS == "windows" {
		executables = []string{
			filepath.Join("venv", "Scripts", "python.exe"),
			filepath.Join("venv", "Scripts", "python3.exe"),
			filepath.Join("venv", "Scripts", "pip.exe"),
			filepath.Join("venv", "Scripts", "pip3.exe"),
		}
	} else {
		executables = []string{
			filepath.Join("venv", "bin", "python"),
			filepath.Join("venv", "bin", "python3"),
			filepath.Join("venv", "bin", "pip"),
			filepath.Join("venv", "bin", "pip3"),
		}
	}
	return executables, nil
}

// GetBinPaths returns the absolute path to the bin directory and the venv bin directory.
func (p *PythonProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "bin")
	venvBin := filepath.Join(installPath, "venv", "bin")
	if runtime.GOOS == "windows" {
		venvBin = filepath.Join(installPath, "venv", "Scripts")
		return []string{binDir, venvBin, installPath}, nil
	}
	return []string{binDir, venvBin}, nil
}

// GetEnvVars returns the VIRTUAL_ENV environment variable.
func (p *PythonProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	venvDir := filepath.Join(installPath, "venv")
	return map[string]string{
		"VIRTUAL_ENV": venvDir,
	}, nil
}

// Uninstall performs Python-specific cleanup.
func (p *PythonProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
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
set "PATH=%s;%%PATH%%"
"%s" %%*
`, name, version, venvDir, installPath, exePath)
	}

	return fmt.Sprintf(`#!/bin/sh
# UniRTM shim for %s (version %s)
export VIRTUAL_ENV="%s"
exec "%s" "$@"
`, name, version, venvDir, exePath)
}
