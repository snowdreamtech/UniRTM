// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

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

// GoPkgProvider implements the Provider interface for Go packages (via go install).
type GoPkgProvider struct {
}

// NewGoPkgProvider creates a new Go package provider.
func NewGoPkgProvider() *GoPkgProvider {
	return &GoPkgProvider{}
}

func (p *GoPkgProvider) Name() string {
	return "go"
}

func (p *GoPkgProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	// Ensure install path exists
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	// We use go to install the package.
	goCmd, err := p.findGo()
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "go is required to install go packages but was not found", err)
	}

	// Construct package spec: path@version
	pkgSpec := tool
	if version != "" && version != "latest" {
		pkgSpec = fmt.Sprintf("%s@%s", tool, version)
	} else {
		pkgSpec = fmt.Sprintf("%s@latest", tool)
	}

	logger.Info("Installing Go package", map[string]interface{}{"pkg": pkgSpec, "GOBIN": installPath})

	// Extract extra domains from GOPROXY
	var extraDomains []string
	if goproxy := env.Get("GOPROXY"); goproxy != "" {
		for _, proxy := range strings.Split(goproxy, ",") {
			proxy = strings.TrimSpace(proxy)
			if proxy == "direct" || proxy == "off" || proxy == "" {
				continue
			}
			if d := DomainFromURL(proxy); d != "" {
				extraDomains = append(extraDomains, d)
			}
		}
	}

	// Use GOBIN to install the binary into our specific versioned directory
	cmd := exec.CommandContext(ctx, goCmd, "install", pkgSpec)
	cmdEnv := append(GetNoProxyEnv(extraDomains...), "GOBIN="+installPath)
	if gosumdb := env.Get("GOSUMDB"); gosumdb != "" {
		cmdEnv = append(cmdEnv, "GOSUMDB="+gosumdb)
	}
	if gonosumdb := env.Get("GONOSUMDB"); gonosumdb != "" {
		cmdEnv = append(cmdEnv, "GONOSUMDB="+gonosumdb)
	}
	if goprivate := env.Get("GOPRIVATE"); goprivate != "" {
		cmdEnv = append(cmdEnv, "GOPRIVATE="+goprivate)
	}
	cmd.Env = cmdEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "go install failed", err)
	}

	return nil
}

func (p *GoPkgProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *GoPkgProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	executables, err := p.ListExecutables(tool, installPath, version)
	if err != nil {
		return nil, err
	}

	shims := make(map[string]string)
	for _, exe := range executables {
		name := filepath.Base(exe)
		// On Windows, remove extension for the shim name
		if runtime.GOOS == "windows" {
			name = strings.TrimSuffix(name, ".exe")
		}
		shims[name] = exe
	}

	return shims, nil
}

func (p *GoPkgProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *GoPkgProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	entries, err := os.ReadDir(installPath)
	if err != nil {
		return nil, err
	}

	var executables []string
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				if info.Mode()&0111 != 0 || strings.HasSuffix(strings.ToLower(entry.Name()), ".exe") {
					executables = append(executables, filepath.Join(installPath, entry.Name()))
				}
			}
		}
	}

	return executables, nil
}

// GetBinPaths returns the absolute path to the directory containing the binaries.
func (p *GoPkgProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return []string{installPath}, nil
}

// GetEnvVars returns no special environment variables.
func (p *GoPkgProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (p *GoPkgProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *GoPkgProvider) Verify(ctx context.Context, tool, version, path string) error {
	return nil
}

func (p *GoPkgProvider) findGo() (string, error) {
	// 1. Try to find a UniRTM-managed Go installation first
	goInstallsDir := filepath.Join(env.GetInstallsDir(), "go")
	if entries, err := os.ReadDir(goInstallsDir); err == nil {
		var bestVer string
		var bestPath string
		for _, entry := range entries {
			if entry.IsDir() {
				verDir := filepath.Join(goInstallsDir, entry.Name())
				candidates := []string{
					filepath.Join(verDir, "bin", "go"),
					filepath.Join(verDir, "bin", "go.exe"),
					filepath.Join(verDir, "go.exe"),
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
	return exec.LookPath("go")
}

