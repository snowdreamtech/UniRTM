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

// DotnetProvider implements the Provider interface for .NET global tools.
type DotnetProvider struct {
}

// NewDotnetProvider creates a new dotnet provider.
func NewDotnetProvider() *DotnetProvider {
	return &DotnetProvider{}
}

func (p *DotnetProvider) Name() string {
	return "dotnet"
}

func (p *DotnetProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	// Extract the full tool name (including scope if present) from the install path.
	installsDir := env.GetInstallsDir()
	toolDir := filepath.Dir(installPath)
	tool, err := filepath.Rel(installsDir, toolDir)
	if err != nil {
		tool = filepath.Base(toolDir) // fallback
	}

	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	dotnetCmd, err := exec.LookPath("dotnet")
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "dotnet is required to install .NET tools but was not found in PATH", err)
	}

	logger.Debug("Installing .NET tool", map[string]interface{}{"tool": tool, "version": version, "installDir": installPath})

	// dotnet tool install <tool> --version <version> --tool-path <installPath>/bin
	args := []string{"tool", "install", tool, "--tool-path", filepath.Join(installPath, "bin")}
	if version != "latest" && version != "" {
		args = append(args, "--version", version)
	}

	cmd := exec.CommandContext(ctx, dotnetCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "dotnet tool install failed", err)
	}

	return nil
}

func (p *DotnetProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	return nil
}

func (p *DotnetProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
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

func (p *DotnetProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *DotnetProvider) ListExecutables(installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "bin")

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
				if info.Mode()&0111 != 0 || filepath.Ext(entry.Name()) == ".exe" {
					executables = append(executables, filepath.Join(binDir, entry.Name()))
				}
			}
		}
	}

	return executables, nil
}

// GetBinPaths returns the absolute path to the bin directory.
func (p *DotnetProvider) GetBinPaths(installPath string, version string) ([]string, error) {
	return []string{filepath.Join(installPath, "bin")}, nil
}

// GetEnvVars returns no special environment variables.
func (p *DotnetProvider) GetEnvVars(installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (p *DotnetProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	return nil
}
