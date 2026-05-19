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

// VfoxProvider implements the Provider interface for vfox plugins.
type VfoxProvider struct {
}

// NewVfoxProvider creates a new vfox provider.
func NewVfoxProvider() *VfoxProvider {
	return &VfoxProvider{}
}

func (p *VfoxProvider) Name() string {
	return "vfox"
}

func (p *VfoxProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	if err := os.MkdirAll(installPath, 0755); err != nil {
	}

	vfoxCmd, err := exec.LookPath("vfox")
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "vfox is required to install vfox plugins but was not found in PATH", err)
	}

	logger.Debug("Installing via vfox", map[string]interface{}{"tool": tool, "version": version, "installDir": installPath})

	// To correctly use vfox within unirtm install paths, we'd ideally configure vfox home.
	// For now, we simulate an install by calling vfox and relying on standard vfox mechanisms,
	// or telling it to install into our path if possible.
	// The basic wrap: `vfox install tool@version`

	pkgSpec := tool
	if version != "latest" && version != "" {
		pkgSpec = tool + "@" + version
	}

	args := []string{"install", pkgSpec}

	cmd := exec.CommandContext(ctx, vfoxCmd, args...)
	if ctx != nil && ctx.Value("quietProgress") == true {
		cmd.Stdout = nil
		cmd.Stderr = nil
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.Env = GetNoProxyEnv()

	// Optional: force vfox to use installPath as its base via environment variables if vfox supports it.
	// VFOX_HOME or something similar. For simplicity, we just run it.

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "vfox install failed", err)
	}

	return nil
}

func (p *VfoxProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *VfoxProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
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

func (p *VfoxProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *VfoxProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	// Typically vfox installs to its own global directory if not overridden.
	// Assuming it's in our `installPath/bin` or `installPath/...` based on how we run it.
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
func (p *VfoxProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return []string{filepath.Join(installPath, "bin")}, nil
}

// GetEnvVars returns no special environment variables.
func (p *VfoxProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (p *VfoxProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}
