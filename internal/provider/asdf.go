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

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// AsdfProvider implements the Provider interface for asdf plugins.
type AsdfProvider struct {
	dataDir     string
	pluginsPath string
}

// NewAsdfProvider creates a new asdf provider.
func NewAsdfProvider() *AsdfProvider {
	dataDir := os.Getenv("UNIRTM_DATA_DIR")
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".local", "share", "unirtm")
	}

	asdfDir := filepath.Join(dataDir, "asdf")
	return &AsdfProvider{
		dataDir:     asdfDir,
		pluginsPath: filepath.Join(asdfDir, "plugins"),
	}
}

func (p *AsdfProvider) Name() string {
	return "asdf"
}

func (p *AsdfProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	// asdf plugins don't use the artifactPath (they download it themselves).
	// We need to extract the tool name from the installPath.
	// installPath format: ~/.local/share/unirtm/installs/<tool>/<version>
	tool := filepath.Base(filepath.Dir(installPath))
	pluginDir := filepath.Join(p.pluginsPath, tool)

	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return NewProviderError(p.Name(), tool, version, "plugin not found (run backend resolve first)", err)
	}

	// Prepare environment variables required by asdf plugins
	env := os.Environ()
	env = append(env,
		"ASDF_INSTALL_TYPE=version",
		"ASDF_INSTALL_VERSION="+version,
		"ASDF_INSTALL_PATH="+installPath,
		"ASDF_CONCURRENCY=4", // reasonable default
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
		cmd.Env = env
		cmd.Dir = installPath
		out, err := cmd.CombinedOutput()
		if err != nil {
			return NewProviderError(p.Name(), tool, version, fmt.Sprintf("bin/download failed: %s", string(out)), err)
		}
	}

	// 2. Run bin/install
	installScript := filepath.Join(pluginDir, "bin", "install")
	if stat, err := os.Stat(installScript); err == nil && !stat.IsDir() {
		logger.Debug("Running asdf plugin install script", map[string]interface{}{"tool": tool, "version": version})
		cmd := exec.CommandContext(ctx, installScript)
		cmd.Env = env
		cmd.Dir = installPath
		out, err := cmd.CombinedOutput()
		if err != nil {
			return NewProviderError(p.Name(), tool, version, fmt.Sprintf("bin/install failed: %s", string(out)), err)
		}
	} else {
		return NewProviderError(p.Name(), tool, version, "bin/install script not found in plugin", err)
	}

	return nil
}

func (p *AsdfProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	// Execute bin/post-install if it exists
	tool := filepath.Base(filepath.Dir(installPath))
	pluginDir := filepath.Join(p.pluginsPath, tool)

	postInstallScript := filepath.Join(pluginDir, "bin", "post-install")
	if stat, err := os.Stat(postInstallScript); err == nil && !stat.IsDir() {
		cmd := exec.CommandContext(ctx, postInstallScript)
		cmd.Env = append(os.Environ(),
			"ASDF_INSTALL_TYPE=version",
			"ASDF_INSTALL_VERSION="+version,
			"ASDF_INSTALL_PATH="+installPath,
		)
		cmd.Dir = installPath
		out, err := cmd.CombinedOutput()
		if err != nil {
			return NewProviderError(p.Name(), tool, version, fmt.Sprintf("bin/post-install failed: %s", string(out)), err)
		}
	}
	return nil
}

func (p *AsdfProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
	executables, err := p.ListExecutables(installPath, version)
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

func (p *AsdfProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	// Tool version should match the directory name
	return filepath.Base(installPath), nil
}

func (p *AsdfProvider) ListExecutables(installPath string, version string) ([]string, error) {
	tool := filepath.Base(filepath.Dir(installPath))
	pluginDir := filepath.Join(p.pluginsPath, tool)

	binPaths := []string{"bin"} // Default

	// Run bin/list-bin-paths if it exists
	listBinPathsScript := filepath.Join(pluginDir, "bin", "list-bin-paths")
	if stat, err := os.Stat(listBinPathsScript); err == nil && !stat.IsDir() {
		cmd := exec.Command(listBinPathsScript)
		cmd.Env = append(os.Environ(),
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

func (p *AsdfProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	// asdf plugins don't have a specific uninstall script for versions.
	// We just let UniRTM delete the directory.
	return nil
}
