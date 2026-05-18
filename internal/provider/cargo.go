// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
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

func (p *CargoProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	// Ensure install path exists
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	// We use cargo to install the package globally into the specific prefix.
	cargoCmd, err := p.findCargo()
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "cargo is required to install crates but was not found", err)
	}

	logger.Debug("Installing cargo crate", map[string]interface{}{"crate": tool, "version": version, "root": installPath})

	// Extract extra domains from Rust mirror environment variables
	var extraDomains []string
	if d := DomainFromURL(env.Get("RUSTUP_DIST_SERVER")); d != "" {
		extraDomains = append(extraDomains, d)
	}
	if d := DomainFromURL(env.Get("RUSTUP_UPDATE_ROOT")); d != "" {
		extraDomains = append(extraDomains, d)
	}

	cmd := exec.CommandContext(ctx, cargoCmd, "install", tool, "--version", version, "--root", installPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = GetNoProxyEnv(extraDomains...)
	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "cargo install failed", err)
	}

	return nil
}

func (p *CargoProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *CargoProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	executables, err := p.ListExecutables(tool, installPath, version)
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

func (p *CargoProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *CargoProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
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

// GetBinPaths returns the absolute path to the bin directory.
func (p *CargoProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return []string{filepath.Join(installPath, "bin")}, nil
}

// GetEnvVars returns no special environment variables.
func (p *CargoProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (p *CargoProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	// Let UniRTM delete the directory
	return nil
}

func (p *CargoProvider) findCargo() (string, error) {
	// 1. Try to find a UniRTM-managed Rust/Cargo installation first
	rustInstallsDir := filepath.Join(env.GetInstallsDir(), "rust")
	if entries, err := os.ReadDir(rustInstallsDir); err == nil {
		var bestVer string
		var bestPath string
		for _, entry := range entries {
			if entry.IsDir() {
				verDir := filepath.Join(rustInstallsDir, entry.Name())
				candidates := []string{
					filepath.Join(verDir, "bin", "cargo"),
					filepath.Join(verDir, "bin", "cargo.exe"),
					filepath.Join(verDir, "cargo.exe"),
				}
				for _, cand := range candidates {
					if info, err := os.Stat(cand); err == nil && !info.IsDir() {
						if bestVer == "" || entry.Name() > bestVer {
							bestVer = entry.Name()
							bestPath = cand
						}
						break
					}
				}
			}
		}
		if bestPath != "" {
			return bestPath, nil
		}
	}

	// 2. Fallback to system PATH
	return exec.LookPath("cargo")
}
