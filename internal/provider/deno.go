// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// DenoProvider implements the Provider interface for Deno.
type DenoProvider struct {
}

// NewDenoProvider creates a new Deno provider.
func NewDenoProvider() *DenoProvider {
	return &DenoProvider{}
}

func (p *DenoProvider) Name() string {
	return "deno"
}

func (p *DenoProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	logger.Debug("Installing Deno", map[string]interface{}{"version": version, "installPath": installPath, "artifactPath": artifactPath})

	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	return nil
}

func (p *DenoProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	return nil
}

func (p *DenoProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
	executables, err := p.ListExecutables(installPath, version)
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

func (p *DenoProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *DenoProvider) ListExecutables(installPath string, version string) ([]string, error) {
	var executables []string
	
	err := filepath.Walk(installPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (info.Name() == "deno" || info.Name() == "deno.exe") {
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
func (p *DenoProvider) GetBinPaths(installPath string, version string) ([]string, error) {
	exes, err := p.ListExecutables(installPath, version)
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
func (p *DenoProvider) GetEnvVars(installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (p *DenoProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	return nil
}
