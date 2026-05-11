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
	var extractCmd *exec.Cmd
	
	// Detect if we should strip components (most official archives have a root dir)
	strip := "--strip-components=1"
	
	if strings.HasSuffix(artifactPath, ".tar.gz") || strings.HasSuffix(artifactPath, ".tgz") {
		extractCmd = exec.CommandContext(ctx, "tar", "-xzf", artifactPath, "-C", installPath, strip)
	} else if ext == ".zip" {
		extractCmd = exec.CommandContext(ctx, "unzip", "-q", "-o", artifactPath, "-d", installPath)
	} else {
		// Fallback to generic install
		return p.generic.Install(ctx, installPath, artifactPath, version)
	}

	if output, err := extractCmd.CombinedOutput(); err != nil {
		// Try without strip-components if it fails (some archives might not have a root dir)
		extractCmd = exec.CommandContext(ctx, "tar", "-xzf", artifactPath, "-C", installPath)
		if _, err2 := extractCmd.CombinedOutput(); err2 != nil {
			return fmt.Errorf("native: extraction failed: %v, output: %s", err, string(output))
		}
	}

	return nil
}

func (p *NativeProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	return p.generic.PostInstall(ctx, installPath, version)
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
