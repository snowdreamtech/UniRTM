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

func (p *GemProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	gemCmd, err := p.findGem()
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "gem is required to install rubygems but was not found", err)
	}

	logger.Debug("Installing rubygem", map[string]interface{}{"gem": tool, "version": version, "installDir": installPath})

	// gem install <tool> -v <version> --install-dir <installPath> --bindir <installPath>/bin --no-document
	args := []string{"install", tool, "--install-dir", installPath, "--bindir", filepath.Join(installPath, "bin"), "--no-document"}
	if version != "latest" && version != "" {
		args = append(args, "-v", version)
	}

	// Extract extra domains from Ruby mirror environment variables
	var extraDomains []string
	if d := DomainFromURL(env.Get("BUNDLE_MIRROR__HTTPS__RUBYGEMS__ORG")); d != "" {
		extraDomains = append(extraDomains, d)
	}

	cmd := exec.CommandContext(ctx, gemCmd, args...)
	if ctx != nil && ctx.Value("quietProgress") == true {
		cmd.Stdout = nil
		cmd.Stderr = nil
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.Env = GetNoProxyEnv(extraDomains...)

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "gem install failed", err)
	}

	return nil
}

func (p *GemProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *GemProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
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

func (p *GemProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *GemProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
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
func (p *GemProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return []string{filepath.Join(installPath, "bin")}, nil
}

// GetEnvVars returns the GEM_HOME environment variable.
func (p *GemProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return map[string]string{
		"GEM_HOME": installPath,
	}, nil
}

func (p *GemProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *GemProvider) findGem() (string, error) {
	// 1. Try to find a UniRTM-managed Ruby/Gem installation first
	rubyInstallsDir := filepath.Join(env.GetInstallsDir(), "ruby")
	if entries, err := os.ReadDir(rubyInstallsDir); err == nil {
		var bestVer string
		var bestPath string
		for _, entry := range entries {
			if entry.IsDir() {
				verDir := filepath.Join(rubyInstallsDir, entry.Name())
				candidates := []string{
					filepath.Join(verDir, "bin", "gem"),
					filepath.Join(verDir, "bin", "gem.exe"),
					filepath.Join(verDir, "gem.exe"),
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
	return exec.LookPath("gem")
}
