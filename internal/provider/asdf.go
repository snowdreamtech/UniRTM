// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// AsdfProvider implements the Provider interface for asdf plugins.
type AsdfProvider struct {
	dataDir     string
	pluginsPath string
}

// NewAsdfProvider creates a new asdf provider.
func NewAsdfProvider() *AsdfProvider {
	return &AsdfProvider{
		pluginsPath: env.GetPluginsDir(),
	}
}

func (p *AsdfProvider) Name() string {
	return "asdf"
}

func (p *AsdfProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	tool = backend.ResolveAsdfToolName(tool)
	pluginDir := filepath.Join(p.pluginsPath, tool)

	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return NewProviderError(p.Name(), tool, version, "plugin not found (run backend resolve first)", err)
	}

	// Prepare environment variables required by asdf plugins
	downloadPath := filepath.Join(installPath, "download")
	if err := os.MkdirAll(downloadPath, 0755); err != nil {
		return err
	}
	defer os.RemoveAll(downloadPath)

	// 2. Create a temporary 'asdf' binary symlink to unirtm in the path to handle reshim calls.
	// Many asdf plugins call 'asdf reshim' at the end of installation.
	tmpBinDir, err := os.MkdirTemp("", "unirtm-asdf-bin-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpBinDir)

	asdfStubPath := filepath.Join(tmpBinDir, "asdf")
	selfExe, err := os.Executable()
	if err != nil {
		return err
	}
	if err := os.Symlink(selfExe, asdfStubPath); err != nil {
		// If symlink fails (e.g. on Windows without permissions), fallback to a simple script/executable
		// but for now we assume Unix-like system as requested by user.
		return fmt.Errorf("failed to create asdf symlink: %w", err)
	}

	// Extract extra domains from common ASDF mirror environment variables
	var extraDomains []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "ASDF_") && strings.Contains(e, "_MIRROR_URL=") {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) > 1 {
				if d := DomainFromURL(parts[1]); d != "" {
					extraDomains = append(extraDomains, d)
				}
			}
		}
	}

	cmdEnv := GetNoProxyEnv(extraDomains...)
	cmdEnv = append(cmdEnv,
		"ASDF_INSTALL_TYPE=version",
		"ASDF_INSTALL_VERSION="+version,
		"ASDF_INSTALL_PATH="+installPath,
		"ASDF_DOWNLOAD_PATH="+downloadPath,
		"ASDF_CONCURRENCY=4", // reasonable default
		"PATH="+tmpBinDir+string(os.PathListSeparator)+env.Get("PATH"),
	)

	// Ensure install path exists
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	// 1. Run bin/download if it exists
	downloadScript := filepath.Join(pluginDir, "bin", "download")
	if stat, err := os.Stat(downloadScript); err == nil && !stat.IsDir() {
		logger.Debug("Running asdf plugin download script", map[string]interface{}{"tool": tool, "version": version})
		cmd := exec.CommandContext(ctx, downloadScript)
		cmd.Env = cmdEnv
		cmd.Dir = installPath
		if ctx != nil && ctx.Value("quietProgress") == true {
			cmd.Stdout = nil
			cmd.Stderr = nil
		} else {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		if err := cmd.Run(); err != nil {
			return NewProviderError(p.Name(), tool, version, "bin/download failed", err)
		}
	}

	// 2. Run bin/install
	installScript := filepath.Join(pluginDir, "bin", "install")
	if stat, err := os.Stat(installScript); err == nil && !stat.IsDir() {
		logger.Debug("Running asdf plugin install script", map[string]interface{}{"tool": tool, "version": version})
		cmd := exec.CommandContext(ctx, installScript)
		cmd.Env = cmdEnv
		cmd.Dir = installPath
		if ctx != nil && ctx.Value("quietProgress") == true {
			cmd.Stdout = nil
			cmd.Stderr = nil
		} else {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		if err := cmd.Run(); err != nil {
			return NewProviderError(p.Name(), tool, version, "bin/install failed", err)
		}
	} else {
		return NewProviderError(p.Name(), tool, version, "bin/install script not found in plugin", err)
	}

	return nil
}

func (p *AsdfProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	// Execute bin/post-install if it exists
	tool = backend.ResolveAsdfToolName(tool)
	pluginDir := filepath.Join(p.pluginsPath, tool)

	postInstallScript := filepath.Join(pluginDir, "bin", "post-install")
	if stat, err := os.Stat(postInstallScript); err == nil && !stat.IsDir() {
		cmd := exec.CommandContext(ctx, postInstallScript)
		cmd.Env = append(GetNoProxyEnv(),
			"ASDF_INSTALL_TYPE=version",
			"ASDF_INSTALL_VERSION="+version,
			"ASDF_INSTALL_PATH="+installPath,
		)
		cmd.Dir = installPath
		if ctx != nil && ctx.Value("quietProgress") == true {
			cmd.Stdout = nil
			cmd.Stderr = nil
		} else {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		if err := cmd.Run(); err != nil {
			return NewProviderError(p.Name(), tool, version, "bin/post-install failed", err)
		}
	}
	return nil
}

func (p *AsdfProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	executables, err := p.ListExecutables(tool, installPath, version)
	if err != nil {
		return nil, err
	}

	shims := make(map[string]string)
	// We just return an empty mapping because UniRTM's shim generator will create
	// the actual shell scripts. We just need to specify the executable names mapped to their full paths.
	for _, exe := range executables {
		name := filepath.Base(exe)
		shims[name] = exe
	}

	return shims, nil
}

func (p *AsdfProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	// Tool version should match the directory name
	return filepath.Base(installPath), nil
}

func (p *AsdfProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	binPaths, err := p.getRelBinPaths(tool, installPath, version)
	if err != nil {
		return nil, err
	}

	var executables []string
	for _, relPath := range binPaths {
		dir := filepath.Join(installPath, relPath)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // Directory might not exist, which is fine
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				// Only include executable files
				info, err := entry.Info()
				if err == nil && info.Mode()&0111 != 0 {
					executables = append(executables, filepath.Join(dir, entry.Name()))
				}
			}
		}
	}

	return executables, nil
}

// GetBinPaths returns the absolute paths to the bin directories.
func (p *AsdfProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	relPaths, err := p.getRelBinPaths(tool, installPath, version)
	if err != nil {
		return nil, err
	}
	var absPaths []string
	for _, rel := range relPaths {
		absPaths = append(absPaths, filepath.Join(installPath, rel))
	}
	return absPaths, nil
}

// GetEnvVars returns no special environment variables.
func (p *AsdfProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (p *AsdfProvider) getRelBinPaths(tool string, installPath string, version string) ([]string, error) {
	tool = backend.ResolveAsdfToolName(tool)
	pluginDir := filepath.Join(p.pluginsPath, tool)

	binPaths := []string{"bin"} // Default

	// Run bin/list-bin-paths if it exists
	listBinPathsScript := filepath.Join(pluginDir, "bin", "list-bin-paths")
	if stat, err := os.Stat(listBinPathsScript); err == nil && !stat.IsDir() {
		cmd := exec.Command(listBinPathsScript)
		cmd.Env = append(GetNoProxyEnv(),
			"ASDF_INSTALL_TYPE=version",
			"ASDF_INSTALL_VERSION="+version,
			"ASDF_INSTALL_PATH="+installPath,
		)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err == nil {
			output := strings.TrimSpace(out.String())
			if output != "" {
				binPaths = strings.Split(output, " ")
			}
		}
	}
	return binPaths, nil
}

func (p *AsdfProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	// asdf plugins don't have a specific uninstall script for versions.
	// We just let UniRTM delete the directory.
	return nil
}
