// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// CargoProvider implements the Provider interface for Cargo packages.
type CargoProvider struct {
}

// NewCargoProvider creates a new Cargo provider.
func NewCargoProvider() *CargoProvider {
	return &CargoProvider{}
}

func (p *CargoProvider) Name() string {
	return "cargo"
}

func (p *CargoProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	tool := filepath.Base(filepath.Dir(installPath))

	// Ensure install path exists
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	// We use the system's cargo to install the package globally into the specific prefix.
	// This requires cargo to be available in PATH.
	cargoCmd, err := exec.LookPath("cargo")
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "cargo is required to install crates but was not found in PATH", err)
	}

	logger.Debug("Installing cargo crate", map[string]interface{}{"crate": tool, "version": version, "root": installPath})

	cmd := exec.CommandContext(ctx, cargoCmd, "install", tool, "--version", version, "--root", installPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "cargo install failed", err)
	}

	return nil
}

func (p *CargoProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	return nil
}

func (p *CargoProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
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

func (p *CargoProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *CargoProvider) ListExecutables(installPath string, version string) ([]string, error) {
	// cargo installs binaries into <root>/bin
	binDir := filepath.Join(installPath, "bin")

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
				// On Unix, check executable bit. On Windows, assume .exe are executable.
				if info.Mode()&0111 != 0 || filepath.Ext(entry.Name()) == ".exe" {
					executables = append(executables, filepath.Join(binDir, entry.Name()))
				}
			}
		}
	}

	return executables, nil
}

func (p *CargoProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	// Let UniRTM delete the directory
	return nil
}
