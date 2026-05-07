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

// GenericProvider implements the Provider interface with default behavior
// suitable for most tools that don't require special handling.
type GenericProvider struct{}

// NewGenericProvider creates a new generic provider.
func NewGenericProvider() *GenericProvider {
	return &GenericProvider{}
}

// Name returns the provider identifier.
func (g *GenericProvider) Name() string {
	return "generic"
}

// Install performs default installation by copying binaries from artifact to install path.
func (g *GenericProvider) Install(ctx context.Context, installPath string, artifactPath string, version string) error {
	// Create bin directory in install path
	binDir := filepath.Join(installPath, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return NewProviderError("generic", "unknown", version, "failed to create bin directory", err)
	}

	// Find all executable files in artifact path
	executables, err := g.findExecutables(artifactPath)
	if err != nil {
		return NewProviderError("generic", "unknown", version, "failed to find executables", err)
	}

	// Copy executables to bin directory
	for _, exe := range executables {
		srcPath := filepath.Join(artifactPath, exe)
		dstPath := filepath.Join(binDir, filepath.Base(exe))

		if err := g.copyFile(srcPath, dstPath); err != nil {
			return NewProviderError("generic", "unknown", version, fmt.Sprintf("failed to copy %s", exe), err)
		}

		// Make executable
		if err := os.Chmod(dstPath, 0755); err != nil {
			return NewProviderError("generic", "unknown", version, fmt.Sprintf("failed to chmod %s", exe), err)
		}
	}

	return nil
}

// PostInstall performs no additional steps for generic provider.
func (g *GenericProvider) PostInstall(ctx context.Context, installPath string, version string) error {
	// No post-install steps for generic provider
	return nil
}

// GenerateShims generates basic shim scripts for all executables.
func (g *GenericProvider) GenerateShims(installPath string, version string) (map[string]string, error) {
	executables, err := g.ListExecutables(installPath, version)
	if err != nil {
		return nil, err
	}

	shims := make(map[string]string)
	for _, exe := range executables {
		exePath := filepath.Join(installPath, "bin", exe)
		shimContent := g.generateShimScript(exePath, version)
		shims[exe] = shimContent
	}

	return shims, nil
}

// DetectVersion attempts to detect the version by running the executable with --version.
func (g *GenericProvider) DetectVersion(ctx context.Context, installPath string) (string, error) {
	executables, err := g.ListExecutables(installPath, "")
	if err != nil || len(executables) == 0 {
		return "", NewProviderError("generic", "unknown", "", "no executables found", err)
	}

	// Try the first executable
	exePath := filepath.Join(installPath, "bin", executables[0])

	// Try common version flags
	versionFlags := []string{"--version", "-version", "-v", "version"}
	for _, flag := range versionFlags {
		cmd := exec.CommandContext(ctx, exePath, flag)
		output, err := cmd.CombinedOutput()
		if err == nil && len(output) > 0 {
			// Extract version from output (first line, first word)
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 {
				fields := strings.Fields(lines[0])
				if len(fields) > 0 {
					return fields[0], nil
				}
			}
		}
	}

	return "", NewProviderError("generic", "unknown", "", "failed to detect version", nil)
}

// ListExecutables returns all executable files in the bin directory.
func (g *GenericProvider) ListExecutables(installPath string, version string) ([]string, error) {
	binDir := filepath.Join(installPath, "bin")

	entries, err := os.ReadDir(binDir)
	if err != nil {
		return nil, NewProviderError("generic", "unknown", version, "failed to read bin directory", err)
	}

	var executables []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if file is executable
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if g.isExecutable(info) {
			executables = append(executables, entry.Name())
		}
	}

	return executables, nil
}

// Uninstall performs no special cleanup for generic provider.
func (g *GenericProvider) Uninstall(ctx context.Context, installPath string, version string) error {
	// No special cleanup needed
	return nil
}

// findExecutables recursively finds all executable files in a directory.
func (g *GenericProvider) findExecutables(root string) ([]string, error) {
	var executables []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if g.isExecutable(info) {
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			executables = append(executables, relPath)
		}

		return nil
	})

	return executables, err
}

// isExecutable checks if a file is executable.
func (g *GenericProvider) isExecutable(info os.FileInfo) bool {
	// On Unix, check executable bit
	if runtime.GOOS != "windows" {
		return info.Mode()&0111 != 0
	}

	// On Windows, check file extension
	name := strings.ToLower(info.Name())
	return strings.HasSuffix(name, ".exe") ||
		strings.HasSuffix(name, ".bat") ||
		strings.HasSuffix(name, ".cmd")
}

// copyFile copies a file from src to dst.
func (g *GenericProvider) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := dstFile.ReadFrom(srcFile); err != nil {
		return err
	}

	return nil
}

// generateShimScript generates a shim script for an executable.
func (g *GenericProvider) generateShimScript(exePath string, version string) string {
	if runtime.GOOS == "windows" {
		return g.generateWindowsShim(exePath, version)
	}
	return g.generateUnixShim(exePath, version)
}

// generateUnixShim generates a Unix shell shim script.
func (g *GenericProvider) generateUnixShim(exePath string, version string) string {
	return fmt.Sprintf(`#!/bin/sh
# UniRTM shim for %s (version %s)
exec "%s" "$@"
`, filepath.Base(exePath), version, exePath)
}

// generateWindowsShim generates a Windows batch shim script.
func (g *GenericProvider) generateWindowsShim(exePath string, version string) string {
	return fmt.Sprintf(`@echo off
REM UniRTM shim for %s (version %s)
"%s" %%*
`, filepath.Base(exePath), version, exePath)
}
