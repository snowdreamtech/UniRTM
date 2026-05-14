// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// SwiftProvider implements the Provider interface for Swift.
type SwiftProvider struct {
}

// NewSwiftProvider creates a new Swift provider.
func NewSwiftProvider() *SwiftProvider {
	return &SwiftProvider{}
}

func (p *SwiftProvider) Name() string {
	return "swift"
}

func (p *SwiftProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	logger.Debug("Installing Swift", map[string]interface{}{"version": version, "installPath": installPath, "artifactPath": artifactPath})

	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	return nil
}

func (p *SwiftProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	return nil
}

func (p *SwiftProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
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

func (p *SwiftProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	return filepath.Base(installPath), nil
}

func (p *SwiftProvider) ListExecutables(installPath string, version string) ([]string, error) {
	var executables []string
	
	err := filepath.Walk(installPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (info.Name() == "swift" || info.Name() == "swift.exe") {
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
func (p *SwiftProvider) GetBinPaths(installPath string, version string) ([]string, error) {
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
		return []string{filepath.Join(installPath, "usr", "bin")}, nil
	}
	return paths, nil
}

// GetEnvVars returns no special environment variables.
func (p *SwiftProvider) GetEnvVars(installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (p *SwiftProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	return nil
}
