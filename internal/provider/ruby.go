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
	generic *GenericProvider
}

// NewRubyProvider creates a new Ruby provider.
func NewRubyProvider() *RubyProvider {
	return &RubyProvider{
		generic: NewGenericProvider(),
	}
}

// Name returns the provider identifier.
func (r *RubyProvider) Name() string {
	return "ruby"
}

// Install performs Ruby-specific installation.
func (r *RubyProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	return r.generic.Install(ctx, installPath, artifactPath, version)
}

// PostInstall performs post-installation steps.
func (r *RubyProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	gemGlobalDir := filepath.Join(installPath, "gem-global")
	if err := os.MkdirAll(gemGlobalDir, 0755); err != nil {
		return NewProviderError("ruby", "ruby", version, "failed to create GEM_HOME directory", err)
	}
	return nil
}

// GenerateShims generates shims for ruby executables.
func (r *RubyProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
	shims := make(map[string]string)

	executables, err := r.ListExecutables(installPath, version)
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

		shimContent := r.generateRubyShim(exe, exePath, installPath, version)
		shims[exe] = shimContent
	}

	return shims, nil
}

// DetectVersion detects Ruby version.
func (r *RubyProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
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

// ListExecutables returns Ruby executables including installed gems.
func (r *RubyProvider) ListExecutables(installPath string, version string) ([]string, error) {
	var executables []string
	
	// Default binaries
	executables = append(executables, "ruby", "gem", "irb", "bundle", "bundler", "erb", "rake", "rdoc", "ri")
	
	// Read global gems
	gemBinDir := filepath.Join(installPath, "gem-global", "bin")
	if entries, err := os.ReadDir(gemBinDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				// Avoid duplicates
				found := false
				for _, e := range executables {
					if e == entry.Name() {
						found = true
						break
					}
				}
				if !found {
					executables = append(executables, entry.Name())
				}
			}
		}
	}

	if runtime.GOOS == "windows" {
		for i := range executables {
			if !strings.HasSuffix(executables[i], ".exe") && !strings.HasSuffix(executables[i], ".bat") && !strings.HasSuffix(executables[i], ".cmd") {
				executables[i] += ".exe"
			}
		}
	}
	return executables, nil
}

// Uninstall performs Ruby-specific cleanup.
func (r *RubyProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	gemGlobalDir := filepath.Join(installPath, "gem-global")
	if err := os.RemoveAll(gemGlobalDir); err != nil {
		return NewProviderError("ruby", "ruby", version, "failed to remove GEM_HOME directory", err)
	}
	return nil
}

// generateRubyShim generates a Ruby-specific shim.
func (r *RubyProvider) generateRubyShim(name, exePath, installPath, version string) string {
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
