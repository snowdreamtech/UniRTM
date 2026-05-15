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
func (g *GenericProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	// 1. Ensure install path exists
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return NewProviderError("generic", "unknown", version, "failed to create install directory", err)
	}

	// 2. Extract artifact if it is an archive
	if err := g.extractArtifact(ctx, artifactPath, installPath); err != nil {
		// If it's not an archive, we just copy it to a new bin directory
		binDir := filepath.Join(installPath, "bin")
		if err := os.MkdirAll(binDir, 0755); err != nil {
			return NewProviderError("generic", "unknown", version, "failed to create bin directory", err)
		}

		dstPath := filepath.Join(binDir, filepath.Base(artifactPath))
		if err := g.copyFile(artifactPath, dstPath); err != nil {
			return NewProviderError("generic", "unknown", version, "failed to copy executable", err)
		}
		if err := os.Chmod(dstPath, 0755); err != nil {
			return NewProviderError("generic", "unknown", version, "failed to chmod executable", err)
		}
	} else {
		// 3. Flatten the directory if it contains only one top-level directory
		// This must happen BEFORE creating the bin directory.
		if err := g.flattenDirectory(installPath); err != nil {
			// Log error but continue, not a fatal failure
			fmt.Printf("⚠️  failed to flatten directory: %v\n", err)
		}

		// 3.1 Relativize all symlinks in the extracted directory.
		// Some archives (like Node.js) might contain absolute symlinks or links 
		// that become absolute during extraction/moving. We convert them to 
		// relative paths so they remain valid after the atomic rename from 
		// .unirtm-tmp to the final versioned path.
		if err := g.relativizeAllSymlinks(installPath); err != nil {
			fmt.Printf("⚠️  failed to relativize symlinks: %v\n", err)
		}

		// 3.5 Validate security (Zip Slip, dangerous symlinks)
		if err := g.validateInstallDir(installPath); err != nil {
			return NewProviderError("generic", "unknown", version, "security validation failed", err)
		}

		// 4. Now create bin directory for UniRTM standardization
		binDir := filepath.Join(installPath, "bin")
		if err := os.MkdirAll(binDir, 0755); err != nil {
			return NewProviderError("generic", "unknown", version, "failed to create bin directory", err)
		}

		// Determine tool name from installPath to help identify the main binary
		toolName := filepath.Base(filepath.Dir(installPath))

		// 5. Find and score all executable files in the extracted path
		allExecs, err := g.findExecutables(installPath)
		if err != nil {
			return NewProviderError("generic", toolName, version, "failed to find executables", err)
		}

		// Pick the best executables based on scoring
		executables := g.pickBestExecutables(allExecs, toolName)

		// 6. Ensure executables have +x and link them to binDir
		for _, exe := range executables {
			exePath := filepath.Join(installPath, exe)
			
			// Skip internal dependencies like node_modules
			if strings.Contains(filepath.ToSlash(exe), "/node_modules/") {
				continue
			}

			if err := os.Chmod(exePath, 0755); err != nil {
				return NewProviderError("generic", toolName, version, fmt.Sprintf("failed to chmod %s", exe), err)
			}

			dstPath := filepath.Join(binDir, filepath.Base(exe))
			if filepath.Dir(exePath) != binDir {
				// If a file already exists at dstPath, don't overwrite it
				// This preserves original symlinks/binaries from the archive
				if _, err := os.Stat(dstPath); err == nil {
					continue
				}

				// Remove existing symlink if it exists (e.g. broken link)
				os.Remove(dstPath)
				
				// Use relative symlink to avoid broken links after directory rename
				relPath, err := filepath.Rel(binDir, exePath)
				if err != nil {
					return NewProviderError("generic", toolName, version, fmt.Sprintf("failed to get relative path for %s", exe), err)
				}
				os.Symlink(relPath, dstPath)
			}
		}
	}

	return nil
}

// PostInstall performs no additional steps for generic provider.
func (g *GenericProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	// No post-install steps for generic provider
	return nil
}

// GenerateShims generates basic shim scripts for all executables.
func (g *GenericProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	executables, err := g.ListExecutables(tool, installPath, version)
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
func (g *GenericProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	executables, err := g.ListExecutables(tool, installPath, "")
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

// ListExecutables returns all executable files in the root and bin directory, relative to installPath.
func (g *GenericProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	searchDirs := []string{"", "bin"}
	var executables []string

	for _, relDir := range searchDirs {
		dir := filepath.Join(installPath, relDir)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

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
				// Return path relative to installPath
				execPath := entry.Name()
				if relDir != "" {
					execPath = filepath.Join(relDir, entry.Name())
				}
				executables = append(executables, execPath)
			}
		}
	}

	if len(executables) == 0 {
		return nil, NewProviderError("generic", "unknown", version, "no executables found in root or bin directory", nil)
	}

	return g.pickBestExecutables(executables, tool), nil
}

// GetBinPaths returns the absolute paths to the root and bin directory.
func (g *GenericProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	paths := []string{installPath, filepath.Join(installPath, "bin")}
	var existing []string
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			existing = append(existing, p)
		}
	}
	return existing, nil
}

// GetEnvVars returns no environment variables by default.
func (g *GenericProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	return make(map[string]string), nil
}

// Uninstall performs no special cleanup for generic provider.
func (g *GenericProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
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
	// Inclusive Strategy:
	// 1. Keep everything in a "bin" directory (high confidence)
	// 2. Keep anything within a reasonable range of the max score
	for _, s := range scored {
		dir := filepath.Dir(s.path)
		isBinDir := filepath.Base(dir) == "bin" || dir == "bin"
		
		if isBinDir || s.score >= maxScore-50 {
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
	
	// Filter out common non-executable files even if they have +x bit
	nonExecExts := map[string]bool{
		".pyc":  true,
		".pyo":  true,
		".txt":  true,
		".md":   true,
		".json": true,
		".yaml": true,
		".yml":  true,
		".html": true,
		".css":  true,
		".js":   true, // Some JS files have +x but should be run via node
		".ts":   true,
		".go":   true,
		".c":    true,
		".h":    true,
		".cpp":  true,
		".hpp":  true,
		".rs":   true,
		".toml": true,
		".xml":  true,
		".pdf":  true,
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
		".svg":  true,
		".ico":  true,
		".map":  true,
		".log":  true,
		".tmp":  true,
		".bak":  true,
		".sql":  true,
		".db":   true,
	}
	if nonExecExts[ext] {
		return -100 // Very low score for these
	}

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
	
	// Skip common non-binary files
	if strings.HasPrefix(name, ".") || 
	   strings.HasSuffix(name, ".txt") || 
	   strings.HasSuffix(name, ".md") ||
	   strings.HasSuffix(name, "__init__.py") ||
	   strings.HasSuffix(name, "_test.go") ||
	   strings.Contains(name, "test") {
		return false
	}

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

// flattenDirectory checks if the directory contains only one sub-directory and no files.
// If so, it moves everything from that sub-directory up one level.
func (g *GenericProvider) flattenDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Filter out hidden files (like .DS_Store) and metadata directories
	var visibleEntries []os.DirEntry
	fmt.Printf("ℹ checking directory content for flattening: %s\n", dir)
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, ".") && name != "__MACOSX" {
			visibleEntries = append(visibleEntries, entry)
			fmt.Printf("  - found: %s (isDir: %v)\n", name, entry.IsDir())
		}
	}

	// If there's exactly one visible entry and it's a directory, flatten it
	if len(visibleEntries) == 1 && visibleEntries[0].IsDir() {
		subDirName := visibleEntries[0].Name()

		// DO NOT flatten standard directories
		standardDirs := map[string]bool{
			"bin":     true,
			"lib":     true,
			"include": true,
			"share":   true,
			"etc":     true,
			"man":     true,
		}
		if standardDirs[strings.ToLower(subDirName)] {
			return nil
		}

		fmt.Printf("ℹ flattening redundant directory: %s\n", subDirName)
		subDir := filepath.Join(dir, subDirName)
		subEntries, err := os.ReadDir(subDir)
		if err != nil {
			return err
		}

		// Move everything from subDir to dir
		for _, entry := range subEntries {
			oldPath := filepath.Join(subDir, entry.Name())
			newPath := filepath.Join(dir, entry.Name())
			
			// If newPath already exists (unlikely in a clean install but possible on retry),
			// remove it first to avoid rename errors.
			if _, err := os.Stat(newPath); err == nil {
				os.RemoveAll(newPath)
			}
			
			if err := os.Rename(oldPath, newPath); err != nil {
				return err
			}
		}

		// Remove the now-empty subDir
		if err := os.Remove(subDir); err != nil {
			// If removal fails, it might not be empty (hidden files?), just log and continue
			fmt.Printf("⚠️  failed to remove empty directory %s: %v\n", subDir, err)
		}

		// Recursive call to handle double-nested directories (e.g. tool-v1/tool-v1/bin)
		return g.flattenDirectory(dir)
	}

	return nil
}

// validateInstallDir ensures that all files and symlinks in the directory are safe.
// It prevents Zip Slip (path traversal) and dangerous symlinks pointing outside the install dir.
func (g *GenericProvider) validateInstallDir(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	return filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 1. Path traversal check (Zip Slip)
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(absPath, absDir) {
			return fmt.Errorf("security violation: file %s is outside of install directory %s", path, dir)
		}

		// 2. Symlink safety check
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}

			// If it's an absolute path, it must start with absDir
			if filepath.IsAbs(target) {
				absTarget, err := filepath.Abs(target)
				if err != nil {
					return err
				}
				if !strings.HasPrefix(absTarget, absDir) {
					return fmt.Errorf("security violation: symlink %s points outside of install directory: %s", path, target)
				}
			} else {
				// For relative paths, we need to check the resolved path
				absTarget, err := filepath.Abs(filepath.Join(filepath.Dir(path), target))
				if err != nil {
					return err
				}
				if !strings.HasPrefix(absTarget, absDir) {
					return fmt.Errorf("security violation: relative symlink %s points outside of install directory: %s", path, target)
				}
			}
		}

		return nil
	})
}

// relativizeAllSymlinks finds all symlinks in the directory and makes them relative
// if they point to a path within the same directory.
func (g *GenericProvider) relativizeAllSymlinks(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	return filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}

			var absTarget string
			if filepath.IsAbs(target) {
				absTarget = target
			} else {
				absTarget = filepath.Clean(filepath.Join(filepath.Dir(path), target))
			}

			// If the target is within our install directory, make it relative
			if strings.HasPrefix(absTarget, absDir) {
				relTarget, err := filepath.Rel(filepath.Dir(path), absTarget)
				if err != nil {
					return err
				}

				// If it's already the same as what we have (and it was relative), skip
				if !filepath.IsAbs(target) && target == relTarget {
					return nil
				}

				// Re-create the symlink as relative
				if err := os.Remove(path); err != nil {
					return err
				}
				if err := os.Symlink(relTarget, path); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
