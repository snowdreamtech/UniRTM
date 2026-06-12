package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

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

	phpCmd, composerPhar, err := p.findPHPAndComposer(ctx)
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "failed to find PHP and Composer", err)
	}

	pkgSpec := fmt.Sprintf("%s:%s", tool, version)
	logger.Debug("Installing composer package", map[string]interface{}{"pkg": pkgSpec, "prefix": installPath})

	// Setup custom mirror if provided
	mirror := env.Get("UNIRTM_COMPOSER_MIRROR")
	if mirror != "" {
		configCmd := exec.CommandContext(ctx, phpCmd, composerPhar, "config", "-g", "repo.packagist", "composer", mirror)
		configCmd.Env = append(os.Environ(), fmt.Sprintf("COMPOSER_HOME=%s", installPath))
		if output, err := configCmd.CombinedOutput(); err != nil {
			return NewProviderError(p.Name(), tool, version, "failed to configure composer mirror: "+string(output), err)
		}
	}

	cmdArgs := []string{composerPhar, "global", "require", pkgSpec, "--no-interaction", "--no-progress"}

	cmd := exec.CommandContext(ctx, phpCmd, cmdArgs...)
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

func (p *ComposerProvider) findPHPAndComposer(ctx context.Context) (string, string, error) {
	// 1. Find PHP
	phpCmd := ""
	for _, backendPrefix := range []string{"php-php", "github-php", "ubi-php", "native-php"} {
		phpInstallBase := filepath.Join(env.GetInstallsDir(), backendPrefix)
		if dirs, err := os.ReadDir(phpInstallBase); err == nil {
			// Find the most recent version directory or any directory
			for _, d := range dirs {
				if d.IsDir() {
					binPath := filepath.Join(phpInstallBase, d.Name(), "bin", "php")
					if runtime.GOOS == "windows" {
						binPath += ".exe"
					}
					if _, err := os.Stat(binPath); err == nil {
						phpCmd = binPath
						break
					}
					rootPath := filepath.Join(phpInstallBase, d.Name(), "php")
					if runtime.GOOS == "windows" {
						rootPath += ".exe"
					}
					if _, err := os.Stat(rootPath); err == nil {
						phpCmd = rootPath
						break
					}
				}
			}
		}
		if phpCmd != "" {
			break
		}
	}

	if phpCmd == "" {
		return "", "", fmt.Errorf("php is required but was not found natively managed by UniRTM")
	}

	// 2. Find or download composer.phar
	phpDir := filepath.Dir(phpCmd)
	composerPhar := filepath.Join(phpDir, "composer.phar")

	if _, err := os.Stat(composerPhar); os.IsNotExist(err) {
		if err := p.downloadComposerPhar(ctx, composerPhar); err != nil {
			return "", "", fmt.Errorf("failed to download composer.phar: %w", err)
		}
	}

	return phpCmd, composerPhar, nil
}

func (p *ComposerProvider) downloadComposerPhar(ctx context.Context, destPath string) error {
	githubProxy := env.Get("GITHUB_PROXY")
	if githubProxy != "" && !strings.HasSuffix(githubProxy, "/") {
		githubProxy += "/"
	}
	url := githubProxy + "https://github.com/composer/composer/releases/latest/download/composer.phar"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
