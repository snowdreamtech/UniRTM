// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package service provides business logic for UniRTM operations.
//
// Shim scripts are thin wrapper scripts placed in the shims directory
// that intercept calls to tools and delegate to the correct version based
// on the active environment settings.
//
// Validates Requirements: 14.1, 14.2, 14.3, 14.4, 14.5, 14.6, 14.7
package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// Generator creates shim scripts for installed tools.
type Generator struct {
	// shimsDir is the directory where shim scripts are written.
	shimsDir string
	// installsDir is the root directory where tools are installed.
	installsDir string
}

// NewGenerator creates a new shim Generator.
func NewGenerator(shimsDir, installsDir string) *Generator {
	return &Generator{
		shimsDir:    shimsDir,
		installsDir: installsDir,
	}
}

// GenerateShim creates a shim script for the given tool.
//
// On Unix systems it creates a bash/sh script; on Windows it creates
// both a .cmd batch file and a .ps1 PowerShell script.
//
// Validates Requirements: 14.1, 14.2, 14.3, 14.4
func (g *Generator) GenerateShim(ctx context.Context, tool string, executables ...string) error {
	if err := os.MkdirAll(g.shimsDir, 0755); err != nil {
		return fmt.Errorf("create shims directory: %w", err)
	}

	if len(executables) == 0 {
		executables = []string{tool}
	}

	// Expand executables with standard aliases so both canonical and alias names are shimmed
	allExecutables := make([]string, 0, len(executables)*2)
	seen := make(map[string]bool)
	for _, exe := range executables {
		clean := strings.ToLower(filepath.Base(exe))
		if ext := filepath.Ext(clean); ext == ".exe" || ext == ".cmd" || ext == ".bat" || ext == ".ps1" {
			clean = clean[:len(clean)-len(ext)]
		}
		if !seen[clean] {
			seen[clean] = true
			allExecutables = append(allExecutables, exe)
		}
		if aliases, ok := standardExecutableAliases[clean]; ok {
			for _, alias := range aliases {
				if !seen[alias] {
					seen[alias] = true
					allExecutables = append(allExecutables, alias)
				}
			}
		}
	}

	for _, exe := range allExecutables {
		// Flatten shim directory by always using filepath.Base for the filename.
		// This ensures consistency with mise and avoids nested directories in shims/.
		shimName := filepath.Base(exe)
		if shimName == "." || shimName == "/" {
			continue
		}

		switch env.RuntimeGOOS {
		case "windows":
			if err := g.generateWindowsShim(tool, shimName); err != nil {
				return err
			}
		default:
			if err := g.generateUnixShim(tool, shimName); err != nil {
				return err
			}
		}
	}
	return nil
}

// RemoveShim removes the shim script(s) for the given tool.
func (g *Generator) RemoveShim(ctx context.Context, tool string) error {
	paths := g.shimPaths(tool)
	var errs []string
	for _, p := range paths {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("remove shim: %s", strings.Join(errs, "; "))
	}
	return nil
}

// ShimExists reports whether a shim script exists for the given tool.
func (g *Generator) ShimExists(tool string) bool {
	paths := g.shimPaths(tool)
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

// ListShims returns all tool names that have shim scripts.
func (g *Generator) ListShims() ([]string, error) {
	entries, err := os.ReadDir(g.shimsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list shims: %w", err)
	}

	seen := make(map[string]bool)
	var tools []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// Strip Windows-specific extensions
		tool := name
		if env.RuntimeGOOS == "windows" {
			ext := strings.ToLower(filepath.Ext(name))
			if ext == ".exe" || ext == ".cmd" || ext == ".bat" || ext == ".ps1" {
				tool = name[:len(name)-len(ext)]
			}
		} else {
			tool = strings.TrimSuffix(strings.TrimSuffix(name, ".cmd"), ".ps1")
		}
		if !seen[tool] {
			seen[tool] = true
			tools = append(tools, tool)
		}
	}
	return tools, nil
}

// shimPaths returns all shim file paths for a tool (platform-dependent).
// It ensures that the shim name is flat by using filepath.Base(tool).
func (g *Generator) shimPaths(tool string) []string {
	// Flatten tool name for lookup to match flat shims directory
	flatName := filepath.Base(tool)
	if env.RuntimeGOOS == "windows" {
		ext := strings.ToLower(filepath.Ext(flatName))
		if ext == ".exe" || ext == ".cmd" || ext == ".bat" || ext == ".ps1" {
			flatName = flatName[:len(flatName)-len(ext)]
		}
		return []string{
			filepath.Join(g.shimsDir, flatName+".exe"),
			filepath.Join(g.shimsDir, flatName+".cmd"),
			filepath.Join(g.shimsDir, flatName+".ps1"),
		}
	}
	return []string{filepath.Join(g.shimsDir, flatName)}
}

// generateUnixShim creates a POSIX-compatible shim script.
//
// The script:
//  1. Detects the active version from UNIRTM_<TOOL>_VERSION or defaults to "current"
//  2. Resolves the tool binary path in the installs directory
//  3. Delegates execution with all original arguments preserved
//  4. Preserves exit codes (Requirement 14.4)
//  5. Provides helpful error when no version is active (Requirement 14.7)
func (g *Generator) generateUnixShim(tool, executable string) error {
	shimPath := filepath.Join(g.shimsDir, executable)

	// 1. Get the absolute path to the current UniRTM executable
	unirtmPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get unirtm executable path: %w", err)
	}

	// Resolve symlinks to prevent pointing to an entrypoint wrapper
	if realPath, err := filepath.EvalSymlinks(unirtmPath); err == nil {
		unirtmPath = realPath
	}

	// Safety check: Prevent self-referential shimming
	if isSelfReferential(shimPath, unirtmPath, tool, executable) {
		return fmt.Errorf("refusing to create self-referential shim for %s at %s", executable, shimPath)
	}

	// 2. Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(shimPath), 0755); err != nil {
		return fmt.Errorf("create shim directory for %s: %w", tool, err)
	}

	// 3. Remove existing shim (might be an old script or symlink)
	_ = os.Remove(shimPath)

	// 4. Create symlink pointing to the current UniRTM binary
	if err := os.Symlink(unirtmPath, shimPath); err != nil {
		// Fallback to minimal wrapper script if symlink fails (rare on Unix).
		// No recursion guard needed here: UniRTM's invokeShimModeWithArgs sets
		// a per-executable guard (_UNIRTM_SHIM_GUARD_<TOOL>) and detects
		// self-referential paths before executing the resolved binary.
		content := fmt.Sprintf("#!/bin/sh\nexec %q shim \"$0\" \"$@\"\n", unirtmPath)
		if err := os.WriteFile(shimPath, []byte(content), 0755); err != nil {
			return fmt.Errorf("failed to create shim for %s: %w", tool, err)
		}
	}

	return nil
}

// generateWindowsShim creates Windows-compatible shim using Hard Links (Scoop Style variant).
// This provides binary performance and compatibility without requiring admin privileges or extra disk space.
func (g *Generator) generateWindowsShim(tool, executable string) error {
	unirtmPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get unirtm executable path: %w", err)
	}

	// Resolve symlinks to get real binary path
	if realPath, err := filepath.EvalSymlinks(unirtmPath); err == nil {
		unirtmPath = realPath
	}

	baseName := executable
	ext := strings.ToLower(filepath.Ext(executable))
	if ext == ".exe" || ext == ".cmd" || ext == ".bat" || ext == ".ps1" {
		baseName = executable[:len(executable)-len(ext)]
	}

	shimPath := filepath.Join(g.shimsDir, baseName+".exe")

	// Safety check: Prevent self-referential shimming
	if isSelfReferential(shimPath, unirtmPath, tool, baseName) {
		return fmt.Errorf("refusing to create self-referential shim for %s at %s", baseName, shimPath)
	}

	// 1. Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(shimPath), 0755); err != nil {
		return fmt.Errorf("create shim directory for %s: %w", tool, err)
	}

	// 2. Clean up existing shims (might be old .cmd, .ps1, .exe, or legacy .exe.exe)
	_ = os.Remove(shimPath)
	_ = os.Remove(filepath.Join(g.shimsDir, baseName+".cmd"))
	_ = os.Remove(filepath.Join(g.shimsDir, baseName+".ps1"))
	_ = os.Remove(filepath.Join(g.shimsDir, baseName+".exe.exe"))
	_ = os.Remove(filepath.Join(g.shimsDir, baseName+".exe.cmd"))

	// 3. Create a Hard Link to the UniRTM binary
	// On Windows, you cannot delete a hard link to a running executable.
	// In tests, unirtmPath is the test runner (.test.exe), so creating a hard link
	// causes t.TempDir() cleanup to fail with "Access is denied".
	// We fallback to .cmd if we are in a test environment.
	isTest := strings.HasSuffix(unirtmPath, ".test.exe") || strings.HasSuffix(unirtmPath, ".test")

	if !isTest {
		_ = os.Link(unirtmPath, shimPath)
	}

	// Always generate minimal wrapper script (.cmd) as well, ensuring tools or scripts
	// explicitly calling <tool>.cmd or invoking via cmd.exe can always resolve it.
	cmdContent := fmt.Sprintf("@echo off\n\"%s\" shim \"%%~n0\" %%*\n", unirtmPath)
	cmdPath := filepath.Join(g.shimsDir, baseName+".cmd")
	return os.WriteFile(cmdPath, []byte(cmdContent), 0644)
}

// isSelfReferential checks if creating a shim for the given executable would result
// in a self-referential execution loop or overwrite the unirtm binary itself.
func isSelfReferential(shimPath, unirtmPath, tool, executable string) bool {
	// 1. Check if tool or executable base name is unirtm or related version manager entry points
	exeBase := strings.ToLower(strings.TrimSuffix(filepath.Base(executable), filepath.Ext(executable)))
	toolBase := strings.ToLower(strings.TrimSuffix(filepath.Base(tool), filepath.Ext(tool)))
	blockedNames := map[string]bool{
		"unirtm": true,
		"mise":   true,
		"rtx":    true,
		"asdf":   true,
	}
	if blockedNames[exeBase] || blockedNames[toolBase] {
		return true
	}

	// 2. Check if clean paths are identical
	cleanShim := filepath.Clean(shimPath)
	cleanUnirtm := filepath.Clean(unirtmPath)
	if cleanShim == cleanUnirtm {
		return true
	}

	// 3. Check if evaluated symlinks point to the same physical file
	evalShim, err1 := filepath.EvalSymlinks(cleanShim)
	evalUnirtm, err2 := filepath.EvalSymlinks(cleanUnirtm)
	if err1 == nil && err2 == nil && evalShim == evalUnirtm {
		return true
	}

	return false
}

// toolVersionEnvVar returns the environment variable name for a tool's active version.
// e.g. "node" → "UNIRTM_NODE_VERSION"
func toolVersionEnvVar(tool string) string {
	upper := strings.ToUpper(strings.ReplaceAll(tool, "-", "_"))
	return fmt.Sprintf("UNIRTM_%s_VERSION", upper)
}

// ExecuteBinary executes a binary with arguments, replacing the current process on Unix.
func ExecuteBinary(binPath string, args []string) error {
	if env.RuntimeGOOS == "windows" {
		// On Windows, we must use exec.Command because syscall.Exec is not available.
		// If binPath is a .cmd or .bat file, wrap it with cmd.exe /c to ensure reliable execution.
		var cmd *exec.Cmd
		ext := strings.ToLower(filepath.Ext(binPath))
		if ext == ".cmd" || ext == ".bat" {
			cmdArgs := append([]string{"/c", binPath}, args[1:]...)
			cmd = exec.Command("cmd.exe", cmdArgs...)
		} else {
			cmd = exec.Command(binPath, args[1:]...)
		}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// On Unix, replace the current process
	return syscall.Exec(binPath, args, os.Environ())
}
