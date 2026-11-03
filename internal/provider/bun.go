// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// BunProvider implements the Provider interface for Bun.
type BunProvider struct {
}

// NewBunProvider creates a new Bun provider.
func NewBunProvider() *BunProvider {
	return &BunProvider{}
}

func (p *BunProvider) Name() string {
	return "bun"
}

func (p *BunProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	logger.Debug("Installing Bun", map[string]interface{}{"version": version, "installPath": installPath, "artifactPath": artifactPath})

	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	// Bun releases are typically single binaries or zip files containing the binary.
	// The generic extractor handles the artifact extraction.
	// We just need to ensure the binary is in the correct place and executable.
	return nil
}

func (p *BunProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *BunProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
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

func (p *BunProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *BunProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	// Bun typically has 'bun' and 'bunx' (which is usually a symlink to bun)
	var executables []string

	// Search recursively for 'bun' binary in installPath
	err := filepath.Walk(installPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (info.Name() == "bun" || info.Name() == "bun.exe") {
			executables = append(executables, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return executables, nil
}

// GetBinPaths returns the absolute paths to the bin directories.
func (p *BunProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	exes, err := p.ListExecutables(tool, installPath, version)
	if err != nil {
		return nil, err
	}
	var paths []string
	seen := make(map[string]bool)
	for _, exe := range exes {
		dir := filepath.Dir(exe)
		if !seen[dir] {
			paths = append(paths, dir)
			seen[dir] = true
		}
	}
	if len(paths) == 0 {
		return []string{installPath}, nil
	}
	return paths, nil
}

// GetEnvVars returns no special environment variables.
func (p *BunProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (p *BunProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}
