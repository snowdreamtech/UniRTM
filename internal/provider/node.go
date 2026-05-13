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
)

// NodeProvider implements the Provider interface for Node.js.
type NodeProvider struct {
	generic *GenericProvider
}

// NewNodeProvider creates a new Node.js provider.
func NewNodeProvider() *NodeProvider {
	return &NodeProvider{
		generic: NewGenericProvider(),
	}
}

// Name returns the provider identifier.
func (n *NodeProvider) Name() string {
	return "node"
}

// Install performs Node.js-specific installation.
func (n *NodeProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	// Use generic installation
	return n.generic.Install(ctx, installPath, artifactPath, version)
}

// PostInstall sets up npm global directory.
func (n *NodeProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	// Create npm global directory
	npmGlobalDir := filepath.Join(installPath, "npm-global")
	if err := os.MkdirAll(npmGlobalDir, 0755); err != nil {
		return NewProviderError("node", "node", version, "failed to create npm global directory", err)
	}

	return nil
}

// GenerateShims generates shims for node, npm, and npx.
func (n *NodeProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
	shims := make(map[string]string)

	// Generate shims for node, npm, npx
	executables := []string{"node", "npm", "npx"}
	for _, exe := range executables {
		exePath := filepath.Join(installPath, "bin", exe)
		if runtime.GOOS == "windows" {
			exePath += ".exe"
		}

		shimContent := n.generateNodeShim(exe, exePath, installPath, version)
		shims[exe] = shimContent
	}

	return shims, nil
}

// DetectVersion detects Node.js version.
func (n *NodeProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	nodePath := filepath.Join(installPath, "bin", "node")
	if runtime.GOOS == "windows" {
		nodePath += ".exe"
	}

	cmd := exec.CommandContext(ctx, nodePath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", NewProviderError("node", "node", "", "failed to detect version", err)
	}

	version := strings.TrimSpace(string(output))
	version = strings.TrimPrefix(version, "v")
	return version, nil
}

// ListExecutables returns Node.js executables relative to installPath.
func (n *NodeProvider) ListExecutables(installPath string, version string) ([]string, error) {
	executables := []string{filepath.Join("bin", "node"), filepath.Join("bin", "npm"), filepath.Join("bin", "npx")}
	if runtime.GOOS == "windows" {
		executables = []string{"node.exe", "npm.cmd", "npx.cmd"}
	}
	return executables, nil
}

// Uninstall performs Node.js-specific cleanup.
func (n *NodeProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	// Clean up npm global directory
	npmGlobalDir := filepath.Join(installPath, "npm-global")
	if err := os.RemoveAll(npmGlobalDir); err != nil {
		return NewProviderError("node", "node", version, "failed to remove npm global directory", err)
	}
	return nil
}

// generateNodeShim generates a Node.js-specific shim.
func (n *NodeProvider) generateNodeShim(name, exePath, installPath, version string) string {
	npmGlobalDir := filepath.Join(installPath, "npm-global")

	if runtime.GOOS == "windows" {
		return fmt.Sprintf(`@echo off
REM UniRTM shim for %s (version %s)
set "NPM_CONFIG_PREFIX=%s"
"%s" %%*
`, name, version, npmGlobalDir, exePath)
	}

	return fmt.Sprintf(`#!/bin/sh
# UniRTM shim for %s (version %s)
export NPM_CONFIG_PREFIX="%s"
exec "%s" "$@"
`, name, version, npmGlobalDir, exePath)
}
