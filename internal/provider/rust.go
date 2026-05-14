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

// RustProvider implements the Provider interface for Rust.
type RustProvider struct {
	generic *GenericProvider
}

// NewRustProvider creates a new Rust provider.
func NewRustProvider() *RustProvider {
	return &RustProvider{
		generic: NewGenericProvider(),
	}
}

// Name returns the provider identifier.
func (r *RustProvider) Name() string {
	return "rust"
}

// Install performs Rust-specific installation.
func (r *RustProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	return r.generic.Install(ctx, tool, installPath, artifactPath, version)
}

// PostInstall performs post-installation steps.
func (r *RustProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	cargoHome := filepath.Join(installPath, "cargo")
	rustupHome := filepath.Join(installPath, "rustup")
	
	if err := os.MkdirAll(cargoHome, 0755); err != nil {
		return NewProviderError("rust", "rust", version, "failed to create CARGO_HOME directory", err)
	}
	if err := os.MkdirAll(rustupHome, 0755); err != nil {
		return NewProviderError("rust", "rust", version, "failed to create RUSTUP_HOME directory", err)
	}
	return nil
}

// GenerateShims generates shims for rust executables.
func (r *RustProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	shims := make(map[string]string)

	executables, err := r.ListExecutables(tool, installPath, version)
	if err != nil {
		return nil, err
	}

	for _, exe := range executables {
		// Try bin/ directory first
		exePath := filepath.Join(installPath, "bin", exe)
		
		// If not in bin/, try cargo/bin/
		if _, err := os.Stat(exePath); os.IsNotExist(err) {
			exePath = filepath.Join(installPath, "cargo", "bin", exe)
		}

		shimContent := r.generateRustShim(exe, exePath, installPath, version)
		shims[exe] = shimContent
	}

	return shims, nil
}

// DetectVersion detects Rust version.
func (r *RustProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	rustcPath := filepath.Join(installPath, "bin", "rustc")
	if runtime.GOOS == "windows" {
		rustcPath += ".exe"
	}

	cmd := exec.CommandContext(ctx, rustcPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", NewProviderError("rust", "rustc", "", "failed to detect version", err)
	}

	versionStr := strings.TrimSpace(string(output))
	parts := strings.Fields(versionStr)
	if len(parts) >= 2 {
		// Output format is usually: rustc 1.70.0 (90c541806 2023-05-31)
		return parts[1], nil
	}

	return "", NewProviderError("rust", "rustc", "", "failed to parse version", nil)
}

// ListExecutables returns Rust executables.
func (r *RustProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	var executables []string
	
	// Default binaries
	executables = append(executables, "rustc", "cargo", "rustdoc", "rustfmt", "rustup", "clippy-driver", "cargo-clippy", "cargo-fmt")
	
	// Read cargo/bin
	cargoBinDir := filepath.Join(installPath, "cargo", "bin")
	if entries, err := os.ReadDir(cargoBinDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
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

// GetBinPaths returns the absolute paths to the bin directories.
func (r *RustProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	return []string{
		filepath.Join(installPath, "bin"),
		filepath.Join(installPath, "cargo", "bin"),
	}, nil
}

// GetEnvVars returns the CARGO_HOME and RUSTUP_HOME environment variables.
func (r *RustProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return map[string]string{
		"CARGO_HOME":  filepath.Join(installPath, "cargo"),
		"RUSTUP_HOME": filepath.Join(installPath, "rustup"),
	}, nil
}

// Uninstall performs Rust-specific cleanup.
func (r *RustProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	cargoHome := filepath.Join(installPath, "cargo")
	rustupHome := filepath.Join(installPath, "rustup")
	
	_ = os.RemoveAll(cargoHome)
	_ = os.RemoveAll(rustupHome)
	
	return nil
}

// generateRustShim generates a Rust-specific shim.
func (r *RustProvider) generateRustShim(name, exePath, installPath, version string) string {
	cargoHome := filepath.Join(installPath, "cargo")
	rustupHome := filepath.Join(installPath, "rustup")

	if runtime.GOOS == "windows" {
		return fmt.Sprintf(`@echo off
REM UniRTM shim for %s (version %s)
set "CARGO_HOME=%s"
set "RUSTUP_HOME=%s"
"%s" %%*
`, name, version, cargoHome, rustupHome, exePath)
	}

	return fmt.Sprintf(`#!/bin/sh
# UniRTM shim for %s (version %s)
export CARGO_HOME="%s"
export RUSTUP_HOME="%s"
exec "%s" "$@"
`, name, version, cargoHome, rustupHome, exePath)
}
