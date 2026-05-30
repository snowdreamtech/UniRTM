// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"github.com/snowdreamtech/unirtm/internal/pkg/version"
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
	// On Windows, npm generates .cmd wrappers that reference node.exe relative to the .cmd
	// file's own directory (using %dp0%\node.exe). Since UniRTM installs Node.js and npm
	// packages in separate, isolated directories, node.exe is never adjacent to these .cmd
	// files. This causes all node-based tools (prettier, eslint, taplo, etc.) to fail with
	// '"node"' is not recognized as an internal or external command'.
	//
	// Fix: rewrite every .cmd in the install directory to use the absolute path to the
	// UniRTM-managed node.exe, replacing the broken relative-path fallback.
	if runtime.GOOS != "windows" {
		return nil
	}
	return p.fixWindowsCmdWrappers(installPath)
}

// fixWindowsCmdWrappers rewrites all npm-generated .cmd files in installPath to use
// an absolute path to the UniRTM-managed node.exe.
func (p *NpmProvider) fixWindowsCmdWrappers(installPath string) error {
	nodePath, err := p.findNodeExe()
	if err != nil || nodePath == "" {
		// Cannot locate node.exe — skip rewrite silently.
		// The .cmd files may still work if node is on the system PATH.
		return nil
	}

	entries, err := os.ReadDir(installPath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".cmd") {
			continue
		}
		cmdFile := filepath.Join(installPath, entry.Name())
		if rewriteErr := p.rewriteCmdNodePath(cmdFile, nodePath); rewriteErr != nil {
			logger.Debug("npm: failed to rewrite .cmd wrapper", map[string]interface{}{
				"file":  cmdFile,
				"error": rewriteErr.Error(),
			})
		}
	}
	return nil
}

// findNodeExe returns the absolute path to node.exe in UniRTM's managed node installation.
// It prefers the highest-version managed install and falls back to the system PATH.
func (p *NpmProvider) findNodeExe() (string, error) {
	nodeInstallsDir := filepath.Join(env.GetInstallsDir(), "node")
	entries, err := os.ReadDir(nodeInstallsDir)
	if err == nil {
		var bestVer, bestPath string
		for _, entry := range entries {
			if !entry.IsDir() || strings.HasSuffix(entry.Name(), ".unirtm-tmp") {
				continue
			}
			nodePath := filepath.Join(nodeInstallsDir, entry.Name(), "node.exe")
			if info, statErr := os.Stat(nodePath); statErr == nil && !info.IsDir() {
				if bestVer == "" || version.CompareVersions(entry.Name(), bestVer) > 0 {
					bestVer = entry.Name()
					bestPath = nodePath
				}
			}
		}
		if bestPath != "" {
			return bestPath, nil
		}
	}
	// Fallback: search system PATH
	return exec.LookPath("node.exe")
}

// rewriteCmdNodePath rewrites a single npm-generated .cmd file so that it uses
// the given absolute nodePath instead of the relative %dp0%\node.exe that npm embeds.
//
// npm 7+ generates a conditional block:
//
//	IF EXIST "%dp0%\node.exe" (
//	  SET "_prog=%dp0%\node.exe"
//	) ELSE (
//	  SET "_prog=node"
//	  SET PATHEXT=%PATHEXT:;.JS;=;%
//	)
//
// We replace that entire block with:
//
//	SET "_prog=<absolute nodePath>"
//
// Older npm versions use a simpler one-liner:
//
//	"%~dp0\node.exe" "...\script.js" %*
//
// We replace "%~dp0\node.exe" with "<absolute nodePath>".
func (p *NpmProvider) rewriteCmdNodePath(cmdPath, nodePath string) error {
	data, err := os.ReadFile(cmdPath)
	if err != nil {
		return err
	}
	content := string(data)

	// Bail out quickly if neither known pattern is present.
	if !strings.Contains(content, "node.exe") {
		return nil
	}

	replacement := fmt.Sprintf(`SET "_prog=%s"`, nodePath)

	// Pattern 1 — npm 7+ conditional block (CRLF or LF tolerant).
	// Match:
	// IF EXIST "%dp0%\node.exe" (
	//   ...
	// ) ELSE (
	//   ...
	// )
	reNpm7 := regexp.MustCompile(`(?i)IF EXIST "%~?dp0%?\\node\.exe" \([\s\S]*?\) ELSE \([\s\S]*?\)`)
	modified := false
	if reNpm7.MatchString(content) {
		content = reNpm7.ReplaceAllString(content, replacement)
		modified = true
	}

	// Pattern 2 — older npm one-liner: "%~dp0\node.exe" …
	newContent := strings.ReplaceAll(
		content,
		`"%~dp0\node.exe"`,
		fmt.Sprintf(`"%s"`, nodePath),
	)
	if newContent != content {
		content = newContent
		modified = true
	}

	if !modified {
		return nil
	}

	return os.WriteFile(cmdPath, []byte(content), 0644)
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

// GetBinPaths returns the absolute path(s) to the directories containing the
// tool's executables.
//
// On Windows, we also append the directory containing node.exe from the
// UniRTM-managed Node.js installation. This provides a second safety net:
// even if PostInstall's .cmd rewrite is skipped for any reason, the npm .cmd
// fallback path (SET "_prog=node") will still resolve because node.exe's
// directory is on PATH.
func (p *NpmProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	var paths []string

	// Primary bin directory for the npm package itself.
	binDir := filepath.Join(installPath, "bin")
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		paths = append(paths, installPath)
	} else {
		paths = append(paths, binDir)
	}

	// Windows only: also expose the node.exe directory so that npm-generated
	// .cmd wrappers can find the node runtime via PATH as a last resort.
	if runtime.GOOS == "windows" {
		if nodeBin := p.findNodeBinDir(); nodeBin != "" {
			paths = append(paths, nodeBin)
		}
	}

	return paths, nil
}

// findNodeBinDir returns the directory containing node.exe in the
// UniRTM-managed installation. On Windows, node.exe lives directly in the
// version directory (no bin/ subdirectory).
func (p *NpmProvider) findNodeBinDir() string {
	nodePath, err := p.findNodeExe()
	if err != nil || nodePath == "" {
		return ""
	}
	return filepath.Dir(nodePath)
}

// GetEnvVars returns environment variables that should be set when this npm
// tool is active.
//
//   - NODE_PATH: lets Node.js resolve globally-installed peer plugins
//     (e.g. @commitlint/config-conventional for @commitlint/cli).
//   - NPM_CONFIG_PREFIX (Windows-only): aligns npm's prefix with the package
//     install directory so any subsequent `npm` invocations inside the tool
//     do not create files in unexpected locations.
func (p *NpmProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	envVars := make(map[string]string)

	// Set NODE_PATH to the node_modules directory so that global plugins
	// (like @commitlint/config-conventional for @commitlint/cli) can be resolved.
	nodeModulesDir := filepath.Join(installPath, "lib", "node_modules")
	if _, err := os.Stat(nodeModulesDir); os.IsNotExist(err) {
		// Fallback for Windows: npm with --prefix puts modules directly under prefix.
		nodeModulesDir = filepath.Join(installPath, "node_modules")
	}

	if info, err := os.Stat(nodeModulesDir); err == nil && info.IsDir() {
		envVars["NODE_PATH"] = nodeModulesDir
	}

	// Windows-only: set NPM_CONFIG_PREFIX so that npm resolves the correct
	// global prefix when called within the context of this tool. Without this,
	// npm would fall back to a system-wide or user-home prefix.
	if runtime.GOOS == "windows" {
		envVars["NPM_CONFIG_PREFIX"] = installPath
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
				var candidates []string
				if runtime.GOOS == "windows" {
					candidates = []string{
						filepath.Join(verDir, "npm.cmd"),
						filepath.Join(verDir, "bin", "npm.cmd"),
						filepath.Join(verDir, "npm"),
					}
				} else {
					candidates = []string{
						filepath.Join(verDir, "bin", "npm"),
						filepath.Join(verDir, "npm"),
					}
				}
				for _, cand := range candidates {
					if info, err := os.Stat(cand); err == nil && !info.IsDir() {
						if bestVer == "" || version.CompareVersions(entry.Name(), bestVer) > 0 {
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
