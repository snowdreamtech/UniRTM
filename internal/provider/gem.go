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

// GemProvider implements the Provider interface for RubyGems.
type GemProvider struct {
}

// NewGemProvider creates a new gem provider.
func NewGemProvider() *GemProvider {
	return &GemProvider{}
}

func (p *GemProvider) Name() string {
	return "gem"
}

func (p *GemProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
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

	gemCmd, err := exec.LookPath("gem")
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "gem is required to install rubygems but was not found in PATH", err)
	}

	logger.Debug("Installing rubygem", map[string]interface{}{"gem": tool, "version": version, "installDir": installPath})

	// gem install <tool> -v <version> --install-dir <installPath> --bindir <installPath>/bin --no-document
	args := []string{"install", tool, "--install-dir", installPath, "--bindir", filepath.Join(installPath, "bin"), "--no-document"}
	if version != "latest" && version != "" {
		args = append(args, "-v", version)
	}

	cmd := exec.CommandContext(ctx, gemCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "gem install failed", err)
	}

	return nil
}

func (p *GemProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	return nil
}

func (p *GemProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
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

func (p *GemProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *GemProvider) ListExecutables(installPath string, version string) ([]string, error) {
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
				if info.Mode()&0111 != 0 || filepath.Ext(entry.Name()) == ".bat" || filepath.Ext(entry.Name()) == ".cmd" || filepath.Ext(entry.Name()) == ".exe" {
					executables = append(executables, filepath.Join(binDir, entry.Name()))
				}
			}
		}
	}

	return executables, nil
}

// GetBinPaths returns the absolute path to the bin directory.
func (p *GemProvider) GetBinPaths(installPath string, version string) ([]string, error) {
	return []string{filepath.Join(installPath, "bin")}, nil
}

// GetEnvVars returns the GEM_HOME environment variable.
func (p *GemProvider) GetEnvVars(installPath string, version string) (map[string]string, error) {
	return map[string]string{
		"GEM_HOME": installPath,
	}, nil
}

func (p *GemProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	return nil
}
