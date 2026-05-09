// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// UbiProvider implements the Provider interface for downloading GitHub releases using the Ubi Universal Binary Installer.
type UbiProvider struct {
}

// NewUbiProvider creates a new Ubi provider.
func NewUbiProvider() *UbiProvider {
	return &UbiProvider{}
}

func (p *UbiProvider) Name() string {
	return "ubi"
}

func (p *UbiProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	installsDir := env.GetInstallsDir()
	relPath, err := filepath.Rel(installsDir, filepath.Dir(installPath))
	if err != nil {
		return NewProviderError(p.Name(), "unknown", version, "failed to determine tool name from install path", err)
	}
	tool := filepath.ToSlash(relPath)

	binDir := filepath.Join(installPath, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	ubiCmd, err := exec.LookPath("ubi")
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "ubi is required to install this tool but was not found in PATH", err)
	}

	logger.Debug("Installing tool via ubi", map[string]interface{}{"tool": tool, "version": version, "binDir": binDir})

	args := []string{"--project", tool, "--in", binDir}
	tag := version
	if version != "latest" && version != "" {
		if !strings.HasPrefix(tag, "v") {
			tag = "v" + version
		}
		args = append(args, "--tag", tag)
	}

	cmd := exec.CommandContext(ctx, ubiCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if tag != version {
			logger.Debug("Retrying ubi without 'v' prefix", map[string]interface{}{"tool": tool, "version": version})
			args[len(args)-1] = version
			cmd2 := exec.CommandContext(ctx, ubiCmd, args...)
			cmd2.Stdout = os.Stdout
			cmd2.Stderr = os.Stderr
			if err2 := cmd2.Run(); err2 != nil {
				return NewProviderError(p.Name(), tool, version, "ubi install failed (tried both with and without 'v' prefix)", err2)
			}
		} else {
			return NewProviderError(p.Name(), tool, version, "ubi install failed", err)
		}
	}

	return nil
}

func (p *UbiProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	return nil
}

func (p *UbiProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
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

func (p *UbiProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *UbiProvider) ListExecutables(installPath string, version string) ([]string, error) {
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

func (p *UbiProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	return nil
}
