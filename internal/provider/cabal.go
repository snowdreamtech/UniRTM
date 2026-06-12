package provider

import (
	"context"
	"fmt"
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

// CabalProvider implements the Provider interface for Haskell Cabal packages.
type CabalProvider struct{}

// NewCabalProvider creates a new cabal provider.
func NewCabalProvider() *CabalProvider {
	return &CabalProvider{}
}

func (p *CabalProvider) Name() string { return "cabal" }

func (p *CabalProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	binDir := filepath.Join(installPath, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	cabalCmd, err := p.findCabal()
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "failed to find native cabal", err)
	}

	pkgSpec := tool
	if version != "" && version != "latest" {
		pkgSpec = tool + "==" + version
	}

	logger.Debug("Installing cabal package", map[string]interface{}{"pkg": pkgSpec, "prefix": binDir})

	cmdArgs := []string{"install", pkgSpec, "--installdir=" + binDir, "--overwrite-policy=always"}

	cmd := exec.CommandContext(ctx, cabalCmd, cmdArgs...)
	if ctx != nil && ctx.Value("quietProgress") == true {
		cmd.Stdout = nil
		cmd.Stderr = nil
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cmd.Env = GetNoProxyEnv()

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "cabal install failed", err)
	}

	return nil
}

func (p *CabalProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *CabalProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	executables, err := p.ListExecutables(tool, installPath, version)
	if err != nil {
		return nil, err
	}
	shims := make(map[string]string)
	for _, exe := range executables {
		shims[filepath.Base(exe)] = exe
	}
	return shims, nil
}

func (p *CabalProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *CabalProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
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
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if info.Mode()&0111 != 0 || ext == ".bat" || ext == ".cmd" || ext == ".exe" {
					executables = append(executables, filepath.Join(binDir, entry.Name()))
				}
			}
		}
	}

	return executables, nil
}

func (p *CabalProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "bin")
	return []string{binDir}, nil
}

func (p *CabalProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return nil, nil
}

func (p *CabalProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *CabalProvider) findCabal() (string, error) {
	cabalCmd := ""
	for _, backendPrefix := range []string{"haskell-haskell", "github-haskell", "ubi-haskell", "native-haskell"} {
		haskellInstallBase := filepath.Join(env.GetInstallsDir(), backendPrefix)
		if dirs, err := os.ReadDir(haskellInstallBase); err == nil {
			for _, d := range dirs {
				if d.IsDir() {
					binPath := filepath.Join(haskellInstallBase, d.Name(), "bin", "cabal")
					if runtime.GOOS == "windows" {
						binPath += ".exe"
					}
					if _, err := os.Stat(binPath); err == nil {
						cabalCmd = binPath
						break
					}
					rootPath := filepath.Join(haskellInstallBase, d.Name(), "cabal")
					if runtime.GOOS == "windows" {
						rootPath += ".exe"
					}
					if _, err := os.Stat(rootPath); err == nil {
						cabalCmd = rootPath
						break
					}
				}
			}
		}
		if cabalCmd != "" {
			break
		}
	}

	if cabalCmd == "" {
		return "", fmt.Errorf("cabal is required but was not found natively managed by UniRTM")
	}

	return cabalCmd, nil
}
