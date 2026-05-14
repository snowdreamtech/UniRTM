// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// ZigProvider implements the Provider interface for Zig.
type ZigProvider struct {
}

// NewZigProvider creates a new Zig provider.
func NewZigProvider() *ZigProvider {
	return &ZigProvider{}
}

func (p *ZigProvider) Name() string {
	return "zig"
}

func (p *ZigProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	logger.Debug("Installing Zig", map[string]interface{}{"version": version, "installPath": installPath, "artifactPath": artifactPath})

	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	return nil
}

func (p *ZigProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}

func (p *ZigProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
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

func (p *ZigProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *ZigProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	var executables []string
	
	err := filepath.Walk(installPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (info.Name() == "zig" || info.Name() == "zig.exe") {
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
func (p *ZigProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
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
func (p *ZigProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (p *ZigProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return nil
}
