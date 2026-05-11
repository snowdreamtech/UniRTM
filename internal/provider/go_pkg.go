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

func (p *GoPkgProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	// For Go packages, tool name is derived from the path
	installsDir := env.GetInstallsDir()
	toolDir := filepath.Dir(installPath)
	tool, err := filepath.Rel(installsDir, toolDir)
	if err != nil {
		tool = filepath.Base(toolDir)
	}

	// Ensure install path exists
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	// We use the system's go to install the package.
	goCmd, err := exec.LookPath("go")
	if err != nil {
		return NewProviderError(p.Name(), tool, version, "go is required to install go packages but was not found in PATH", err)
	}

	// Construct package spec: path@version
	pkgSpec := tool
	if version != "" && version != "latest" {
		pkgSpec = fmt.Sprintf("%s@%s", tool, version)
	} else {
		pkgSpec = fmt.Sprintf("%s@latest", tool)
	}

	logger.Info("Installing Go package", map[string]interface{}{"pkg": pkgSpec, "GOBIN": installPath})

	// Use GOBIN to install the binary into our specific versioned directory
	cmd := exec.CommandContext(ctx, goCmd, "install", pkgSpec)
	cmd.Env = append(os.Environ(), "GOBIN="+installPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return NewProviderError(p.Name(), tool, version, "go install failed", err)
	}

	return nil
}

func (p *GoPkgProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	return nil
}

func (p *GoPkgProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
	executables, err := p.ListExecutables(installPath, version)
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

func (p *GoPkgProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *GoPkgProvider) ListExecutables(installPath string, version string) ([]string, error) {
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

func (p *GoPkgProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	return nil
}

func (p *GoPkgProvider) Verify(ctx context.Context, tool, version, path string) error {
	return nil
}
