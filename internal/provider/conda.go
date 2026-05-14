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

// CondaProvider implements the Provider interface for Conda environments.
type CondaProvider struct {
}

// NewCondaProvider creates a new conda provider.
func NewCondaProvider() *CondaProvider {
	return &CondaProvider{}
}

func (p *CondaProvider) Name() string {
	return "conda"
}

func (p *CondaProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	condaCmd, err := exec.LookPath("conda")
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "conda is required to install conda packages but was not found in PATH", err)
	}

	logger.Debug("Installing Conda package", map[string]interface{}{"tool": tool, "version": version, "installDir": installPath})

	// conda create -y -p <installPath> <tool>=<version>
	args := []string{"create", "-y", "-p", installPath}
	
	pkgSpec := tool
	if version != "latest" && version != "" {
		pkgSpec = tool + "=" + version
	}
	args = append(args, pkgSpec)

	cmd := exec.CommandContext(ctx, condaCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "conda create failed", err)
	}

	return nil
}

func (p *CondaProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *CondaProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
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

func (p *CondaProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *CondaProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "bin")

	// Check bin for unix, Scripts for windows
	dirsToCheck := []string{binDir, filepath.Join(installPath, "Scripts")}
	
	var executables []string
	
	for _, dir := range dirsToCheck {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				info, err := entry.Info()
				if err == nil {
					if info.Mode()&0111 != 0 || filepath.Ext(entry.Name()) == ".exe" || filepath.Ext(entry.Name()) == ".bat" {
						executables = append(executables, filepath.Join(dir, entry.Name()))
					}
				}
			}
		}
	}

	return executables, nil
}

// GetBinPaths returns the absolute paths to the bin directories.
func (p *CondaProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return []string{
		filepath.Join(installPath, "bin"),
		filepath.Join(installPath, "Scripts"),
	}, nil
}

// GetEnvVars returns the CONDA_PREFIX environment variable.
func (p *CondaProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return map[string]string{
		"CONDA_PREFIX": installPath,
	}, nil
}

func (p *CondaProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}
