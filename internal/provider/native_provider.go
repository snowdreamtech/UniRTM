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

func (p *NativeProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	if artifactPath == "" {
		return fmt.Errorf("native: no artifact path provided")
	}

	// Create bin directory in install path
	binDir := filepath.Join(installPath, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(artifactPath))
	
	// Handle single binary files (like kubectl)
	if ext != ".zip" && !strings.HasSuffix(artifactPath, ".tar.gz") && !strings.HasSuffix(artifactPath, ".tgz") {
		return p.generic.Install(ctx, installPath, artifactPath, version)
	}

	// Handle archives
	var extractCmd *exec.Cmd
	strip := "--strip-components=1"
	
	if strings.HasSuffix(artifactPath, ".tar.gz") || strings.HasSuffix(artifactPath, ".tgz") {
		extractCmd = exec.CommandContext(ctx, "tar", "-xzf", artifactPath, "-C", installPath, strip)
	} else if ext == ".zip" {
		extractCmd = exec.CommandContext(ctx, "unzip", "-q", "-o", artifactPath, "-d", installPath)
	}

	if extractCmd != nil {
		if output, err := extractCmd.CombinedOutput(); err != nil {
			// Try without strip-components if it fails (some archives might not have a root dir)
			if strings.Contains(strings.Join(extractCmd.Args, " "), "tar") {
				extractCmd = exec.CommandContext(ctx, "tar", "-xzf", artifactPath, "-C", installPath)
				if _, err2 := extractCmd.CombinedOutput(); err2 != nil {
					return fmt.Errorf("native: extraction failed: %v, output: %s", err, string(output))
				}
			} else {
				return fmt.Errorf("native: extraction failed: %v, output: %s", err, string(output))
			}
		}
	}

	return nil
}

func (p *NativeProvider) PostInstall(ctx context.Context, installPath string, version string) error {
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

	return p.generic.PostInstall(ctx, installPath, version)
}

// Helper to avoid duplicate logic from generic.go for now
func isExecutable(info os.FileInfo) bool {
	if runtime.GOOS != "windows" {
		return info.Mode()&0111 != 0
	}
	name := strings.ToLower(info.Name())
	return strings.HasSuffix(name, ".exe") || strings.HasSuffix(name, ".bat") || strings.HasSuffix(name, ".cmd")
}

func (p *NativeProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
	return p.generic.GenerateShims(installPath, version)
}

func (p *NativeProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	return p.generic.DetectVersion(ctx, installPath)
}

func (p *NativeProvider) ListExecutables(installPath string, version string) ([]string, error) {
	return p.generic.ListExecutables(installPath, version)
}

func (p *NativeProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	return p.generic.Uninstall(ctx, installPath, version)
}

func (p *NativeProvider) Verify(ctx context.Context, tool, version, path string) error {
	return nil
}
