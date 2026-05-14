// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// PypiProvider implements the Provider interface for PyPI packages.
type PypiProvider struct {
}

// NewPypiProvider creates a new PyPI provider.
func NewPypiProvider() *PypiProvider {
	return &PypiProvider{}
}

func (p *PypiProvider) Name() string {
	return "pypi"
}

func (p *PypiProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	// Ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(installPath), 0755); err != nil {
		return err
	}

	// Find python executable
	pythonCmd, err := p.findPython()
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "python is required to install pypi packages but was not found", err)
	}

	// 1. Create a virtual environment
	logger.Debug("Creating virtual environment", map[string]interface{}{"path": installPath})
	venvCmd := exec.CommandContext(ctx, pythonCmd, "-m", "venv", installPath)
	venvCmd.Env = GetNoProxyEnv()
	outVenv, err := venvCmd.CombinedOutput()
	if err != nil {
		return NewProviderError(p.Name(), tool, version, fmt.Sprintf("failed to create virtual environment: %s", string(outVenv)), err)
	}

	// 2. Install the package inside the venv
	pipCmd := filepath.Join(installPath, "bin", "pip")
	if _, err := os.Stat(pipCmd); os.IsNotExist(err) {
		pipCmd = filepath.Join(installPath, "Scripts", "pip.exe") // Windows
	}

	pkgSpec := fmt.Sprintf("%s==%s", tool, version)
	logger.Debug("Installing pypi package", map[string]interface{}{"pkg": pkgSpec, "venv": installPath})
	cmd := exec.CommandContext(ctx, pipCmd, "install", pkgSpec)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = GetNoProxyEnv()

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "pip install failed", err)
	}

	return nil
}

func (p *PypiProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *PypiProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	executables, err := p.ListExecutables(tool, installPath, version)
	if err != nil {
		return nil, err
	}

	shims := make(map[string]string)
	for _, exe := range executables {
		name := filepath.Base(exe)

		// Skip internal venv python/pip binaries to avoid polluting global bin
		if name == "python" || name == "python3" || name == "pip" || name == "pip3" || name == "python.exe" || name == "pip.exe" {
			continue
		}

		shims[name] = exe
	}

	return shims, nil
}

func (p *PypiProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *PypiProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "bin")
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		binDir = filepath.Join(installPath, "Scripts") // Windows
	}

	entries, err := os.ReadDir(binDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var executables []string
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				if info.Mode()&0111 != 0 || filepath.Ext(entry.Name()) == ".cmd" || filepath.Ext(entry.Name()) == ".exe" {
					executables = append(executables, filepath.Join(binDir, entry.Name()))
				}
			}
		}
	}

	return executables, nil
}

// GetBinPaths returns the absolute path to the bin directory.
func (p *PypiProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "bin")
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		binDir = filepath.Join(installPath, "Scripts")
	}
	return []string{binDir}, nil
}

// GetEnvVars returns the VIRTUAL_ENV environment variable.
func (p *PypiProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return map[string]string{
		"VIRTUAL_ENV": installPath,
	}, nil
}

func (p *PypiProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *PypiProvider) findPython() (string, error) {
	cmds := []string{"python3", "python"}
	for _, cmd := range cmds {
		if path, err := exec.LookPath(cmd); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("python not found in PATH")
}
