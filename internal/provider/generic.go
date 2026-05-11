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

	// Extract artifact if it is an archive
	if err := g.extractArtifact(ctx, artifactPath, installPath); err != nil {
		// If it's not an archive, we just copy it to binDir
		dstPath := filepath.Join(binDir, filepath.Base(artifactPath))
		if err := g.copyFile(artifactPath, dstPath); err != nil {
			return NewProviderError("generic", "unknown", version, "failed to copy executable", err)
		}
		if err := os.Chmod(dstPath, 0755); err != nil {
			return NewProviderError("generic", "unknown", version, "failed to chmod executable", err)
		}
	} else {
		// Determine tool name from installPath to help identify the main binary
		toolName := filepath.Base(filepath.Dir(installPath))

		// Find and score all executable files in the extracted path
		allExecs, err := g.findExecutables(installPath)
		if err != nil {
			return NewProviderError("generic", toolName, version, "failed to find executables", err)
		}

		// Pick the best executables based on scoring
		executables := g.pickBestExecutables(allExecs, toolName)

		// Ensure executables have +x and link them to binDir
		for _, exe := range executables {
			exePath := filepath.Join(installPath, exe)
			if err := os.Chmod(exePath, 0755); err != nil {
				return NewProviderError("generic", toolName, version, fmt.Sprintf("failed to chmod %s", exe), err)
			}

			dstPath := filepath.Join(binDir, filepath.Base(exe))
			if filepath.Dir(exePath) != binDir {
				// Remove existing symlink if it exists
				os.Remove(dstPath)
				os.Symlink(exePath, dstPath)
			}
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

// pickBestExecutables filters the list of executables to find the most relevant ones.
func (g *GenericProvider) pickBestExecutables(execs []string, toolName string) []string {
	if len(execs) <= 1 {
		return execs
	}

	type scoredExe struct {
		path  string
		score int
	}

	var scored []scoredExe
	maxScore := -1000

	for _, exe := range execs {
		score := g.calculateExeScore(exe, toolName)
		scored = append(scored, scoredExe{exe, score})
		if score > maxScore {
			maxScore = score
		}
	}

	var results []string
	// If we have a clear winner, only take the high-scoring ones
	for _, s := range scored {
		// Heuristic: if a file scores much higher than others, it's probably the main binary.
		// We take anything that is within 30 points of the max score.
		if s.score >= maxScore-30 {
			results = append(results, s.path)
		}
	}

	return results
}

// calculateExeScore evaluates how likely an executable is the primary tool binary.
func (g *GenericProvider) calculateExeScore(relPath string, toolName string) int {
	filename := filepath.Base(relPath)
	nameLower := strings.ToLower(filename)
	toolLower := strings.ToLower(toolName)
	score := 0

	// 1. Name Match
	if nameLower == toolLower || nameLower == toolLower+".exe" {
		score += 100
	} else if strings.HasPrefix(nameLower, toolLower) {
		score += 50
	} else if strings.Contains(nameLower, toolLower) {
		score += 20
	}

	// 2. Location
	dir := filepath.ToSlash(filepath.Dir(relPath))
	if dir == "." || dir == "bin" {
		score += 30
	} else if strings.Contains(dir, "examples") || strings.Contains(dir, "tests") || strings.Contains(dir, "scripts") {
		score -= 50
	}

	// 3. Format
	ext := strings.ToLower(filepath.Ext(filename))
	if runtime.GOOS != "windows" {
		if ext == "" {
			score += 20 // Prefer extensionless binaries on Unix
		} else if ext == ".sh" || ext == ".py" || ext == ".pl" || ext == ".rb" {
			score -= 20 // Scripts are less likely to be the main binary if a binary exists
		}
	} else {
		if ext == ".exe" {
			score += 20
		} else if ext == ".bat" || ext == ".cmd" {
			score -= 10
		}
	}

	return score
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

// extractArtifact attempts to extract an archive to the destination directory.
// Returns an error if the file is not a supported archive or extraction fails.
func (g *GenericProvider) extractArtifact(ctx context.Context, artifactPath string, dstDir string) error {
	ext := strings.ToLower(filepath.Ext(artifactPath))
	if ext == ".gz" || ext == ".tgz" {
		cmd := exec.CommandContext(ctx, "tar", "-xzf", artifactPath, "-C", dstDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("tar extract failed: %v, output: %s", err, string(output))
		}
		return nil
	} else if ext == ".zip" {
		cmd := exec.CommandContext(ctx, "unzip", "-q", "-o", artifactPath, "-d", dstDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("unzip failed: %v, output: %s", err, string(output))
		}
		return nil
	} else if ext == ".tar" {
		cmd := exec.CommandContext(ctx, "tar", "-xf", artifactPath, "-C", dstDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("tar extract failed: %v, output: %s", err, string(output))
		}
		return nil
	}
	return fmt.Errorf("unsupported archive type: %s", ext)
}
