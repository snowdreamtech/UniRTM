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

// RubyProvider implements the Provider interface for Ruby.
type RubyProvider struct {
	native  *NativeProvider
	asdf    *AsdfProvider
	generic *GenericProvider
}

// NewRubyProvider creates a new Ruby provider with native-first fallback logic.
func NewRubyProvider(native *NativeProvider) *RubyProvider {
	return &RubyProvider{
		native:  native,
		asdf:    NewAsdfProvider(),
		generic: NewGenericProvider(),
	}
}

// Name returns the provider identifier.
func (r *RubyProvider) Name() string {
	return "ruby"
}

// Install performs Ruby-specific installation with native-first fallback.
func (r *RubyProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	// Check if we are on a platform likely to support native binaries
	isNativeSupported := runtime.GOOS == "darwin" || strings.Contains(strings.ToLower(artifactPath), "ubuntu")

	if isNativeSupported {
		// 1. Try Native installation first
		err := r.native.Install(ctx, tool, installPath, artifactPath, version)
		if err == nil {
			// Verification: try to run ruby -v to see if the binary works on this specific Linux distro (e.g. Debian)
			if _, detectErr := r.DetectVersion(ctx, tool, installPath); detectErr == nil {
				return nil
			}
			// If binary doesn't work (e.g. glibc mismatch), cleanup and fallback
			os.RemoveAll(installPath)
		}
	}

	// 2. Fallback to ASDF (source compilation)
	// Note: ASDF provider handles its own downloading if artifactPath is not what it expects,
	// or we might need to pass an empty artifactPath to force it to resolve.
	return r.asdf.Install(ctx, tool, installPath, "", version)
}

// PostInstall performs post-installation steps.
func (r *RubyProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	gemGlobalDir := filepath.Join(installPath, "gem-global")
	if err := os.MkdirAll(gemGlobalDir, 0755); err != nil {
		return NewProviderError("ruby", "ruby", version, "failed to create GEM_HOME directory", err)
	}
	return nil
}

// GenerateShims generates shims for ruby executables.
func (r *RubyProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	shims := make(map[string]string)

	executables, err := r.ListExecutables(tool, installPath, version)
	if err != nil {
		return nil, err
	}

	for _, exe := range executables {
		// Executables could be in bin/ or in gem-global/bin/
		exePath := filepath.Join(installPath, "bin", exe)

		// If it's not in bin/, assume it's a global gem
		if _, err := os.Stat(exePath); os.IsNotExist(err) {
			exePath = filepath.Join(installPath, "gem-global", "bin", exe)
		}

		shimContent := r.generateRubyShim(tool, exe, exePath, installPath, version)
		shims[exe] = shimContent
	}

	return shims, nil
}

// DetectVersion detects Ruby version.
func (r *RubyProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	rubyPath := filepath.Join(installPath, "bin", "ruby")
	if runtime.GOOS == "windows" {
		rubyPath += ".exe"
	}

	cmd := exec.CommandContext(ctx, rubyPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", NewProviderError("ruby", "ruby", "", "failed to detect version", err)
	}

	version := strings.TrimSpace(string(output))
	parts := strings.Fields(version)
	if len(parts) >= 2 {
		// Output format is usually: ruby 3.2.2 (2023-03-30 revision e51014f9c0) ...
		return parts[1], nil
	}

	return "", NewProviderError("ruby", "ruby", "", "failed to parse version", nil)
}

// ListExecutables returns Ruby executables relative to installPath.
func (r *RubyProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	var executables []string

	// 1. Core binaries in bin/
	coreBinDir := filepath.Join(installPath, "bin")
	if entries, err := os.ReadDir(coreBinDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				executables = append(executables, filepath.Join("bin", entry.Name()))
			}
		}
	}

	// 2. Global gems in gem-global/bin/
	gemBinDir := filepath.Join(installPath, "gem-global", "bin")
	if entries, err := os.ReadDir(gemBinDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				name := entry.Name()
				// Avoid duplicates if already in bin/
				found := false
				for _, e := range executables {
					if filepath.Base(e) == name {
						found = true
						break
					}
				}
				if !found {
					executables = append(executables, filepath.Join("gem-global", "bin", name))
				}
			}
		}
	}

	return executables, nil
}

// GetBinPaths returns the absolute paths to the bin directories.
func (r *RubyProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return []string{
		filepath.Join(installPath, "bin"),
		filepath.Join(installPath, "gem-global", "bin"),
	}, nil
}

// GetEnvVars returns the GEM_HOME environment variable.
func (r *RubyProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return map[string]string{
		"GEM_HOME": filepath.Join(installPath, "gem-global"),
	}, nil
}

// Uninstall performs Ruby-specific cleanup.
func (r *RubyProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	gemGlobalDir := filepath.Join(installPath, "gem-global")
	if err := os.RemoveAll(gemGlobalDir); err != nil {
		return NewProviderError("ruby", "ruby", version, "failed to remove GEM_HOME directory", err)
	}
	return nil
}

// generateRubyShim generates a Ruby-specific shim.
func (r *RubyProvider) generateRubyShim(tool string, name, exePath, installPath, version string) string {
	gemHome := filepath.Join(installPath, "gem-global")

	if runtime.GOOS == "windows" {
		return fmt.Sprintf(`@echo off
REM UniRTM shim for %s (version %s)
set "GEM_HOME=%s"
"%s" %%*
`, name, version, gemHome, exePath)
	}

	return fmt.Sprintf(`#!/bin/sh
# UniRTM shim for %s (version %s)
export GEM_HOME="%s"
exec "%s" "$@"
`, name, version, gemHome, exePath)
}
