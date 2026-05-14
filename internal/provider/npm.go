// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// NpmProvider implements the Provider interface for npm packages.
type NpmProvider struct {
}

// NewNpmProvider creates a new npm provider.
func NewNpmProvider() *NpmProvider {
	return &NpmProvider{}
}

func (p *NpmProvider) Name() string {
	return "npm"
}

func (p *NpmProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	// Extract the full tool name (including scope if present) from the install path.
	// The path structure is <installs_dir>/<tool_name>/<version>
	installsDir := env.GetInstallsDir()
	toolDir := filepath.Dir(installPath)
	tool, err := filepath.Rel(installsDir, toolDir)
	if err != nil {
		tool = filepath.Base(toolDir) // fallback
	}

	// Ensure install path exists
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	// We use the system's npm to install the package globally into the specific prefix.
	// This requires npm to be available in PATH.
	npmCmd, err := exec.LookPath("npm")
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "npm is required to install npm packages but was not found in PATH", err)
	}

	pkgSpec := fmt.Sprintf("%s@%s", tool, version)
	logger.Debug("Installing npm package", map[string]interface{}{"pkg": pkgSpec, "prefix": installPath})

	cmd := exec.CommandContext(ctx, npmCmd, "install", "-g", pkgSpec, "--prefix", installPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "npm install failed", err)
	}

	return nil
}

func (p *NpmProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	return nil
}

func (p *NpmProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
	executables, err := p.ListExecutables(installPath, version)
	if err != nil {
		return nil, err
	}

	shims := make(map[string]string)
	for _, exe := range executables {
		name := filepath.Base(exe)
		shims[name] = exe
	}

	return shims, nil
}

func (p *NpmProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *NpmProvider) ListExecutables(installPath string, version string) ([]string, error) {
	// npm installs global binaries into <prefix>/bin (on Unix) or <prefix> (on Windows)
	binDir := filepath.Join(installPath, "bin")

	// If /bin doesn't exist, check the root (common on Windows npm installs)
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		binDir = installPath
	}

	entries, err := os.ReadDir(binDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // Package might not have binaries
		}
		return nil, err
	}

	var executables []string
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				// On Unix, check executable bit. On Windows, assume .cmd/.exe are executable.
				if info.Mode()&0111 != 0 || filepath.Ext(entry.Name()) == ".cmd" || filepath.Ext(entry.Name()) == ".exe" || filepath.Ext(entry.Name()) == ".ps1" {
					executables = append(executables, filepath.Join(binDir, entry.Name()))
				}
			}
		}
	}

	return executables, nil
}

// GetBinPaths returns the absolute path to the bin directory.
func (p *NpmProvider) GetBinPaths(installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "bin")
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		return []string{installPath}, nil
	}
	return []string{binDir}, nil
}

// GetEnvVars returns no special environment variables.
func (p *NpmProvider) GetEnvVars(installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (p *NpmProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	// Let UniRTM delete the directory
	return nil
}
