// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// ComposerProvider implements the Provider interface for PHP Composer packages.
type ComposerProvider struct {
}

// NewComposerProvider creates a new composer provider.
func NewComposerProvider() *ComposerProvider {
	return &ComposerProvider{}
}

func (p *ComposerProvider) Name() string {
	return "composer"
}

func (p *ComposerProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	composerCmd, err := exec.LookPath("composer")
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "composer is required to install php packages but was not found", err)
	}

	pkgSpec := fmt.Sprintf("%s:%s", tool, version)
	logger.Debug("Installing composer package", map[string]interface{}{"pkg": pkgSpec, "prefix": installPath})

	// Setup custom mirror if provided
	mirror := env.Get("UNIRTM_COMPOSER_MIRROR")
	if mirror != "" {
		configCmd := exec.CommandContext(ctx, composerCmd, "config", "-g", "repo.packagist", "composer", mirror)
		configCmd.Env = append(os.Environ(), fmt.Sprintf("COMPOSER_HOME=%s", installPath))
		if output, err := configCmd.CombinedOutput(); err != nil {
			return NewProviderError(p.Name(), tool, version, "failed to configure composer mirror: "+string(output), err)
		}
	}

	cmdArgs := []string{"global", "require", pkgSpec, "--no-interaction", "--no-progress"}

	cmd := exec.CommandContext(ctx, composerCmd, cmdArgs...)
	if ctx != nil && ctx.Value("quietProgress") == true {
		cmd.Stdout = nil
		cmd.Stderr = nil
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	noProxyEnv := GetNoProxyEnv()
	var finalEnv []string
	for _, e := range noProxyEnv {
		if !strings.HasPrefix(e, "COMPOSER_HOME=") {
			finalEnv = append(finalEnv, e)
		}
	}
	finalEnv = append(finalEnv, fmt.Sprintf("COMPOSER_HOME=%s", installPath))
	cmd.Env = finalEnv

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "composer install failed", err)
	}

	return nil
}

func (p *ComposerProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *ComposerProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
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

func (p *ComposerProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *ComposerProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "vendor", "bin")

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
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if info.Mode()&0111 != 0 || ext == ".bat" || ext == ".cmd" || ext == ".exe" {
					executables = append(executables, filepath.Join(binDir, entry.Name()))
				}
			}
		}
	}

	return executables, nil
}

func (p *ComposerProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "vendor", "bin")
	return []string{binDir}, nil
}

func (p *ComposerProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	envVars := make(map[string]string)
	envVars["COMPOSER_HOME"] = installPath
	return envVars, nil
}

func (p *ComposerProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}
