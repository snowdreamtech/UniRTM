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

func (p *NpmProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	// Ensure install path exists
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	// We use npm to install the package globally into the specific prefix.
	npmCmd, err := p.findNpm()
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "npm is required to install npm packages but was not found", err)
	}

	pkgSpec := fmt.Sprintf("%s@%s", tool, version)
	logger.Debug("Installing npm package", map[string]interface{}{"pkg": pkgSpec, "prefix": installPath})

	cmd := exec.CommandContext(ctx, npmCmd, "install", "-g", pkgSpec, "--prefix", installPath)
	if ctx != nil && ctx.Value("quietProgress") == true {
		cmd.Stdout = nil
		cmd.Stderr = nil
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// Extract extra domain from environment variables
	var extraDomains []string
	if d := DomainFromURL(env.Get("NPM_CONFIG_REGISTRY")); d != "" {
		extraDomains = append(extraDomains, d)
	}

	cmd.Env = GetNoProxyEnv(extraDomains...)

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "npm install failed", err)
	}

	return nil
}

func (p *NpmProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *NpmProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
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

func (p *NpmProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *NpmProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
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
func (p *NpmProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "bin")
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		return []string{installPath}, nil
	}
	return []string{binDir}, nil
}

// GetEnvVars returns NODE_PATH for npm tools so plugins can be resolved.
func (p *NpmProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	envVars := make(map[string]string)
	
	// Set NODE_PATH to the node_modules directory so that global plugins
	// (like @commitlint/config-conventional for @commitlint/cli) can be resolved.
	nodeModulesDir := filepath.Join(installPath, "lib", "node_modules")
	if _, err := os.Stat(nodeModulesDir); os.IsNotExist(err) {
		// Fallback for Windows
		nodeModulesDir = filepath.Join(installPath, "node_modules")
	}
	
	if info, err := os.Stat(nodeModulesDir); err == nil && info.IsDir() {
		envVars["NODE_PATH"] = nodeModulesDir
	}
	
	return envVars, nil
}

func (p *NpmProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	// Let UniRTM delete the directory
	return nil
}

func (p *NpmProvider) findNpm() (string, error) {
	// 1. Try to find a UniRTM-managed Node/npm installation first
	nodeInstallsDir := filepath.Join(env.GetInstallsDir(), "node")
	if entries, err := os.ReadDir(nodeInstallsDir); err == nil {
		var bestVer string
		var bestPath string
		for _, entry := range entries {
			if entry.IsDir() {
				verDir := filepath.Join(nodeInstallsDir, entry.Name())
				candidates := []string{
					filepath.Join(verDir, "bin", "npm"),
					filepath.Join(verDir, "npm"),
					filepath.Join(verDir, "npm.cmd"),
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
	return exec.LookPath("npm")
}
