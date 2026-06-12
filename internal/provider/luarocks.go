// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"runtime"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// LuaRocksProvider implements the Provider interface for LuaRocks packages.
type LuaRocksProvider struct{}

// NewLuaRocksProvider creates a new luarocks provider.
func NewLuaRocksProvider() *LuaRocksProvider {
	return &LuaRocksProvider{}
}

func (p *LuaRocksProvider) Name() string { return "luarocks" }

func (p *LuaRocksProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	lrCmd, err := p.findLuarocks()
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "luarocks is required to install lua packages but was not found", err)
	}

	args := []string{"install", tool, "--tree=" + installPath}
	if version != "" && version != "latest" {
		args = append(args, version)
	}

	// Support custom mirror via environment variable
	server := env.Get("UNIRTM_LUAROCKS_SERVER")
	if server != "" {
		args = append(args, "--server="+server)
	}

	logger.Debug("Installing luarocks package", map[string]interface{}{"pkg": tool, "version": version, "prefix": installPath})

	cmd := exec.CommandContext(ctx, lrCmd, args...)
	if ctx != nil && ctx.Value("quietProgress") == true {
		cmd.Stdout = nil
		cmd.Stderr = nil
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cmd.Env = GetNoProxyEnv()

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "luarocks install failed", err)
	}

	return nil
}

func (p *LuaRocksProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *LuaRocksProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
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

func (p *LuaRocksProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *LuaRocksProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
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

func (p *LuaRocksProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "bin")
	return []string{binDir}, nil
}

func (p *LuaRocksProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return nil, nil
}

func (p *LuaRocksProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *LuaRocksProvider) findLuarocks() (string, error) {
	luaInstallsDir := filepath.Join(env.GetInstallsDir(), "lua")
	entries, err := os.ReadDir(luaInstallsDir)
	if err == nil {
		var bestPath string
		for _, entry := range entries {
			if entry.IsDir() {
				verDir := filepath.Join(luaInstallsDir, entry.Name())
				var candidates []string
				if runtime.GOOS == "windows" {
					candidates = []string{
						filepath.Join(verDir, "bin", "luarocks.bat"),
						filepath.Join(verDir, "luarocks.bat"),
						filepath.Join(verDir, "bin", "luarocks.cmd"),
						filepath.Join(verDir, "luarocks.cmd"),
						filepath.Join(verDir, "bin", "luarocks"),
						filepath.Join(verDir, "luarocks"),
					}
				} else {
					candidates = []string{
						filepath.Join(verDir, "bin", "luarocks"),
						filepath.Join(verDir, "luarocks"),
					}
				}
				for _, cand := range candidates {
					if info, err := os.Stat(cand); err == nil && !info.IsDir() {
						bestPath = cand
						break
					}
				}
			}
		}
		if bestPath != "" {
			return bestPath, nil
		}
	}

	return exec.LookPath("luarocks")
}
