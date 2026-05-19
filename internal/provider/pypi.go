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

// PypiProvider implements the Provider interface for PyPI packages.
type PypiProvider struct {
}

// NewPypiProvider creates a new PyPI provider.
func NewPypiProvider() *PypiProvider {
	return &PypiProvider{}
}

func (p *PypiProvider) Name() string {
	return "pypi"
}

func (p *PypiProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	// Ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(installPath), 0755); err != nil {
		return err
	}

	// Find python executable
	pythonCmd, err := p.findPython()
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "python is required to install pypi packages but was not found", err)
	}

	// Extract extra domains from environment variables
	var extraDomains []string
	if d := DomainFromURL(env.Get("PIP_INDEX_URL")); d != "" {
		extraDomains = append(extraDomains, d)
	}
	if d := DomainFromURL(env.Get("PIP_EXTRA_INDEX_URL")); d != "" {
		extraDomains = append(extraDomains, d)
	}

	// 1. Create a virtual environment
	logger.Debug("Creating virtual environment", map[string]interface{}{"path": installPath})
	venvCmd := exec.CommandContext(ctx, pythonCmd, "-m", "venv", installPath)
	venvCmd.Env = GetNoProxyEnv(extraDomains...)
	outVenv, err := venvCmd.CombinedOutput()
	if err != nil {
		return NewProviderError(p.Name(), tool, version, fmt.Sprintf("failed to create virtual environment: %s", string(outVenv)), err)
	}

	// 2. Install the package inside the venv
	pipCmd := filepath.Join(installPath, "bin", "pip")
	if _, err := os.Stat(pipCmd); os.IsNotExist(err) {
		pipCmd = filepath.Join(installPath, "Scripts", "pip.exe") // Windows
	}

	pkgSpec := fmt.Sprintf("%s==%s", tool, version)
	logger.Debug("Installing pypi package", map[string]interface{}{"pkg": pkgSpec, "venv": installPath})
	cmd := exec.CommandContext(ctx, pipCmd, "install", pkgSpec)
	if ctx != nil && ctx.Value("quietProgress") == true {
		cmd.Stdout = nil
		cmd.Stderr = nil
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.Env = GetNoProxyEnv(extraDomains...)

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "pip install failed", err)
	}

	return nil
}

func (p *PypiProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	// Since virtual environments write absolute paths into shebangs and activate scripts,
	// and unirtm uses an atomic rename strategy (installing into <path>.unirtm-tmp first),
	// we must rewrite all occurrences of the temporary install path to the final install path
	// to make the virtual environment functional after the rename.
	finalPath := strings.TrimSuffix(installPath, ".unirtm-tmp")
	if finalPath == installPath {
		return nil
	}

	binDir := filepath.Join(installPath, "bin")
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		binDir = filepath.Join(installPath, "Scripts") // Windows
	}

	entries, err := os.ReadDir(binDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(binDir, entry.Name())
		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		// Only rewrite text files (like shell scripts or python shebang files)
		// Binary files shouldn't start with '#'
		if len(content) > 2 && content[0] == '#' && content[1] == '!' {
			// Rewrite shebang
			contentStr := string(content)
			if strings.Contains(contentStr, installPath) {
				logger.Debug("Rewriting shebang for virtualenv relocatability", map[string]interface{}{
					"file": entry.Name(),
					"from": installPath,
					"to":   finalPath,
				})
				newContent := strings.ReplaceAll(contentStr, installPath, finalPath)
				if err := os.WriteFile(filePath, []byte(newContent), 0755); err != nil {
					return fmt.Errorf("failed to rewrite shebang for %s: %w", entry.Name(), err)
				}
			}
		}
	}

	// Also rewrite activate scripts
	activateFiles := []string{"activate", "activate.sh", "activate.bat", "activate.ps1"}
	for _, act := range activateFiles {
		filePath := filepath.Join(binDir, act)
		if _, err := os.Stat(filePath); err == nil {
			content, err := os.ReadFile(filePath)
			if err == nil {
				contentStr := string(content)
				if strings.Contains(contentStr, installPath) {
					newContent := strings.ReplaceAll(contentStr, installPath, finalPath)
					_ = os.WriteFile(filePath, []byte(newContent), 0644)
				}
			}
		}
	}

	return nil
}

func (p *PypiProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	executables, err := p.ListExecutables(tool, installPath, version)
	if err != nil {
		return nil, err
	}

	shims := make(map[string]string)
	for _, exe := range executables {
		name := filepath.Base(exe)

		// Skip internal venv python/pip binaries to avoid polluting global bin
		if name == "python" || name == "python3" || name == "pip" || name == "pip3" || name == "python.exe" || name == "pip.exe" {
			continue
		}

		shims[name] = exe
	}

	return shims, nil
}

func (p *PypiProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *PypiProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "bin")
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		binDir = filepath.Join(installPath, "Scripts") // Windows
	}

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
				if info.Mode()&0111 != 0 || filepath.Ext(entry.Name()) == ".cmd" || filepath.Ext(entry.Name()) == ".exe" {
					executables = append(executables, filepath.Join(binDir, entry.Name()))
				}
			}
		}
	}

	return executables, nil
}

// GetBinPaths returns no paths for PATH injection.
// pipx/pypi tools use isolated venvs; injecting the venv's bin/ would expose
// internal python/pip interpreters and pollute the system PATH.
// Tool binaries are accessed via shims or resolved directly by their tool name.
func (p *PypiProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return []string{}, nil
}

// GetEnvVars returns the VIRTUAL_ENV environment variable.
func (p *PypiProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return map[string]string{
		"VIRTUAL_ENV": installPath,
	}, nil
}

func (p *PypiProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *PypiProvider) findPython() (string, error) {
	// 1. Try to find a UniRTM-managed Python installation first
	pythonInstallsDir := filepath.Join(env.GetInstallsDir(), "python")
	if entries, err := os.ReadDir(pythonInstallsDir); err == nil {
		var bestVer string
		var bestPath string
		for _, entry := range entries {
			if entry.IsDir() {
				verDir := filepath.Join(pythonInstallsDir, entry.Name())
				candidates := []string{
					filepath.Join(verDir, "bin", "python3"),
					filepath.Join(verDir, "bin", "python"),
					filepath.Join(verDir, "python.exe"),
					filepath.Join(verDir, "Scripts", "python.exe"),
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
	cmds := []string{"python3", "python"}
	for _, cmd := range cmds {
		if path, err := exec.LookPath(cmd); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("python not found in PATH")
}
