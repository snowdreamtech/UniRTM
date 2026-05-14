// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/provider/native"
)

// NativeProvider is a generic provider that uses hardcoded protocol handlers
// and recipes to install tools directly from official sources.
type NativeProvider struct {
	recipes map[string]native.Recipe
	generic *GenericProvider
}

func NewNativeProvider() *NativeProvider {
	return &NativeProvider{
		recipes: native.GetBuiltinRecipes(),
		generic: NewGenericProvider(),
	}
}

func (p *NativeProvider) Name() string {
	return "native"
}

func (p *NativeProvider) ListVersions(ctx context.Context, tool string) ([]string, error) {
	recipe, ok := p.recipes[tool]
	if !ok {
		return nil, fmt.Errorf("native: no recipe for tool: %s", tool)
	}

	versions, err := recipe.Handler.ResolveVersions(ctx, recipe.BaseURL)
	if err != nil {
		return nil, err
	}

	var res []string
	for _, v := range versions {
		res = append(res, v.Version)
	}
	return res, nil
}

func (p *NativeProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	if artifactPath == "" {
		return fmt.Errorf("native: no artifact path provided")
	}

	// Delegate to generic provider which handles extraction, flattening, and binDir creation correctly
	return p.generic.Install(ctx, tool, installPath, artifactPath, version)
}

func (p *NativeProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	// Ensure all executables are in bin/ directory (UniRTM standard)
	binDir := filepath.Join(installPath, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	// If bin/ is already populated, we're likely good
	entries, _ := os.ReadDir(binDir)
	if len(entries) > 0 {
		// Some tools like Node.js have their main executables in bin/ but might need symlinks for others
		// We'll let it be for now if it looks correct
	}

	// Scan the whole installPath for executables that should be in bin/
	// This handles tools like Deno/Bun that extract to root.
	err := filepath.Walk(installPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Skip files already in bin/
		if strings.HasPrefix(path, binDir) {
			return nil
		}

		// Check if it is an executable
		if isExecutable(info) {
			// Move or link to bin/
			dstPath := filepath.Join(binDir, filepath.Base(path))
			if _, err := os.Stat(dstPath); os.IsNotExist(err) {
				// For single binary tools that extract to root, moving is cleaner
				if filepath.Dir(path) == installPath {
					os.Rename(path, dstPath)
				} else {
					// For others, symlink
					os.Symlink(path, dstPath)
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return p.generic.PostInstall(ctx, tool, installPath, version)
}

// Helper to avoid duplicate logic from generic.go for now
func isExecutable(info os.FileInfo) bool {
	if runtime.GOOS != "windows" {
		return info.Mode()&0111 != 0
	}
	name := strings.ToLower(info.Name())
	return strings.HasSuffix(name, ".exe") || strings.HasSuffix(name, ".bat") || strings.HasSuffix(name, ".cmd")
}

func (p *NativeProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	return p.generic.GenerateShims(tool, installPath, version)
}

func (p *NativeProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	return p.generic.DetectVersion(ctx, tool, installPath)
}

func (p *NativeProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	return p.generic.ListExecutables(tool, installPath, version)
}

// GetBinPaths returns the absolute path to the bin directory.
func (p *NativeProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return p.generic.GetBinPaths(tool, installPath, version)
}

// GetEnvVars returns the environment variables for the tool.
func (p *NativeProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return p.generic.GetEnvVars(tool, installPath, version)
}

func (p *NativeProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	return p.generic.Uninstall(ctx, tool, installPath, version)
}

func (p *NativeProvider) Verify(ctx context.Context, tool, version, path string) error {
	return nil
}
