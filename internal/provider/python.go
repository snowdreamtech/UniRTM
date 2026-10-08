// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
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

// SkipAtomicRename indicates that this provider requires installing directly into the final path.
// This is necessary because Python virtualenvs hardcode absolute paths in executable PE wrappers on Windows.
func (p *PythonProvider) SkipAtomicRename() bool {
	return true
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
	if env.RuntimeGOOS == "windows" {
		// Prefer binaries directly in the install root or install/ folder next to DLLs
		candidates := []string{
			filepath.Join(installPath, "python.exe"),
			filepath.Join(installPath, "install", "python.exe"),
			filepath.Join(installPath, "python3.exe"),
			filepath.Join(installPath, "install", "python3.exe"),
			filepath.Join(installPath, "python_d.exe"),
			filepath.Join(installPath, "install", "python_d.exe"),
		}
		for _, cand := range candidates {
			if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
				return cand
			}
		}

		binPy := filepath.Join(installPath, "bin", "python.exe")
		if realPy, err := filepath.EvalSymlinks(binPy); err == nil {
			if fi, err := os.Stat(realPy); err == nil && !fi.IsDir() {
				return realPy
			}
		}

		binPyD := filepath.Join(installPath, "bin", "python_d.exe")
		if realPy, err := filepath.EvalSymlinks(binPyD); err == nil {
			if fi, err := os.Stat(realPy); err == nil && !fi.IsDir() {
				return realPy
			}
		}

		// Fallback for edge cases
		return binPy
	}

	candidates := []string{
		filepath.Join(installPath, "bin", "python3"),
		filepath.Join(installPath, "bin", "python"),
		filepath.Join(installPath, "install", "bin", "python3"),
		filepath.Join(installPath, "install", "bin", "python"),
		filepath.Join(installPath, "python3"),
		filepath.Join(installPath, "python"),
	}

	// If installPath has a version folder name like "3.14.7", directly check "python3.14" without directory scanning
	ver := filepath.Base(installPath)
	if parts := strings.Split(ver, "."); len(parts) >= 2 {
		majorMinor := parts[0] + "." + parts[1]
		candidates = append(candidates,
			filepath.Join(installPath, "bin", "python"+majorMinor),
			filepath.Join(installPath, "install", "bin", "python"+majorMinor),
		)
	}

	for _, cand := range candidates {
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			return cand
		}
	}

	return filepath.Join(installPath, "bin", "python3")
}

// PostInstall creates a virtual environment.
func (p *PythonProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	pythonPath := p.getRealPythonPath(installPath)
	if _, err := os.Stat(pythonPath); err != nil {
		return NewProviderError("python", "python", version, fmt.Sprintf("python executable not found at %s", pythonPath), err)
	}

	if env.RuntimeGOOS == "windows" {
		// python-build-standalone on Windows ships with a python*._pth file
		// which puts Python in isolated mode, breaking venv and pip.
		// We must delete it to restore standard behavior.
		pythonDir := filepath.Dir(pythonPath)
		files, _ := os.ReadDir(pythonDir)
		for _, file := range files {
			name := file.Name()
			if strings.HasPrefix(name, "python") && strings.HasSuffix(name, "._pth") {
				os.Remove(filepath.Join(pythonDir, name))
			}
		}
	}

	venvDir := filepath.Join(installPath, "venv")
	cmd := exec.CommandContext(ctx, pythonPath, "-m", "venv", venvDir)

	// Ensure binaries and shared libraries across platforms (Windows DLLs, Linux/macOS SOs/dylibs)
	// are discoverable by the newly created venv during ensurepip and venv setup.
	binPaths, _ := p.GetBinPaths(tool, installPath, version)
	pathPrefix := strings.Join(binPaths, string(os.PathListSeparator))

	envVars := os.Environ()
	pathFound := false
	for i, e := range envVars {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			envVars[i] = "PATH=" + pathPrefix + string(os.PathListSeparator) + e[5:]
			pathFound = true
			break
		}
	}
	if !pathFound {
		envVars = append(envVars, "PATH="+pathPrefix)
	}

	libDir := filepath.Join(installPath, "lib")
	if env.RuntimeGOOS == "linux" {
		ldFound := false
		for i, e := range envVars {
			if strings.HasPrefix(e, "LD_LIBRARY_PATH=") {
				envVars[i] = "LD_LIBRARY_PATH=" + libDir + ":" + e[16:]
				ldFound = true
				break
			}
		}
		if !ldFound {
			envVars = append(envVars, "LD_LIBRARY_PATH="+libDir)
		}
	} else if env.RuntimeGOOS == "darwin" {
		dyFound := false
		for i, e := range envVars {
			if strings.HasPrefix(e, "DYLD_LIBRARY_PATH=") {
				envVars[i] = "DYLD_LIBRARY_PATH=" + libDir + ":" + e[18:]
				dyFound = true
				break
			}
		}
		if !dyFound {
			envVars = append(envVars, "DYLD_LIBRARY_PATH="+libDir)
		}
	}
	cmd.Env = envVars

	if err := cmd.Run(); err != nil {
		return NewProviderError("python", "python", version, "failed to create virtual environment", err)
	}

	if env.RuntimeGOOS == "windows" {
		scriptsDir := filepath.Join(venvDir, "Scripts")
		copyIfMissing := func(src, dst string) {
			if _, err := os.Stat(dst); os.IsNotExist(err) {
				if srcInfo, err := os.Stat(src); err == nil && !srcInfo.IsDir() {
					if linkErr := os.Link(src, dst); linkErr != nil {
						if data, readErr := os.ReadFile(src); readErr == nil {
							_ = os.WriteFile(dst, data, 0755)
						}
					}
				}
			}
		}
		copyIfMissing(filepath.Join(scriptsDir, "python_d.exe"), filepath.Join(scriptsDir, "python.exe"))
		copyIfMissing(filepath.Join(scriptsDir, "python.exe"), filepath.Join(scriptsDir, "python3.exe"))
		copyIfMissing(filepath.Join(scriptsDir, "pip.exe"), filepath.Join(scriptsDir, "pip3.exe"))
		copyIfMissing(filepath.Join(installPath, "python_d.exe"), filepath.Join(installPath, "python.exe"))
		copyIfMissing(filepath.Join(installPath, "python.exe"), filepath.Join(installPath, "python3.exe"))
	} else {
		linkIfMissing := func(src, dst string) {
			if _, err := os.Lstat(dst); os.IsNotExist(err) {
				if _, err := os.Stat(src); err == nil {
					_ = os.Symlink(filepath.Base(src), dst)
				}
			}
		}
		binDir := filepath.Join(venvDir, "bin")
		linkIfMissing(filepath.Join(binDir, "python3"), filepath.Join(binDir, "python"))
		linkIfMissing(filepath.Join(binDir, "python"), filepath.Join(binDir, "python3"))
		linkIfMissing(filepath.Join(binDir, "pip3"), filepath.Join(binDir, "pip"))
		linkIfMissing(filepath.Join(binDir, "pip"), filepath.Join(binDir, "pip3"))

		installBin := filepath.Join(installPath, "bin")
		linkIfMissing(filepath.Join(installBin, "python3"), filepath.Join(installBin, "python"))
		linkIfMissing(filepath.Join(installBin, "python"), filepath.Join(installBin, "python3"))
		linkIfMissing(filepath.Join(installBin, "pip3"), filepath.Join(installBin, "pip"))
		linkIfMissing(filepath.Join(installBin, "pip"), filepath.Join(installBin, "pip3"))
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
		if env.RuntimeGOOS == "windows" {
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
	if env.RuntimeGOOS == "windows" {
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
	if env.RuntimeGOOS == "windows" {
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

	if env.RuntimeGOOS == "windows" {
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
