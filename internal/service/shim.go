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
	"runtime"
	"strings"
	"syscall"
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

	for _, exe := range executables {
		// Flatten shim directory by always using filepath.Base for the filename.
		// This ensures consistency with mise and avoids nested directories in shims/.
		shimName := filepath.Base(exe)
		if shimName == "." || shimName == "/" {
			continue
		}
		
		switch runtime.GOOS {
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
		tool := strings.TrimSuffix(strings.TrimSuffix(name, ".cmd"), ".ps1")
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
	if runtime.GOOS == "windows" {
		return []string{
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

	// 2. Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(shimPath), 0755); err != nil {
		return fmt.Errorf("create shim directory for %s: %w", tool, err)
	}

	// 3. Remove existing shim (might be an old script or symlink)
	_ = os.Remove(shimPath)

	// 4. Create symlink pointing to the current UniRTM binary
	if err := os.Symlink(unirtmPath, shimPath); err != nil {
		// Fallback to minimal wrapper script if symlink fails (rare on Unix)
		content := fmt.Sprintf("#!/bin/sh\nexec %q shim \"$0\" \"$@\"\n", unirtmPath)
		if err := os.WriteFile(shimPath, []byte(content), 0755); err != nil {
			return fmt.Errorf("failed to create shim for %s: %w", tool, err)
		}
	}

	return nil
}

// generateWindowsShim creates Windows-compatible shim scripts (batch + PowerShell).
//
// Validates Requirements: 14.1, 14.2, 14.3, 14.4, 14.5, 14.6
func (g *Generator) generateWindowsShim(tool, executable string) error {
	unirtmPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get unirtm executable path: %w", err)
	}

	cmdContent := fmt.Sprintf(`@echo off
REM UniRTM shim for %[1]s
REM Generated by UniRTM — do not edit manually
"%[2]s" shim "%%~n0" %%*
`,
		executable,
		unirtmPath,
	)

	cmdPath := filepath.Join(g.shimsDir, executable+".cmd")
	if err := os.MkdirAll(filepath.Dir(cmdPath), 0755); err != nil {
		return fmt.Errorf("create shim directory for %s: %w", tool, err)
	}
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		return fmt.Errorf("write .cmd shim for %s: %w", tool, err)
	}

	// ── PowerShell (.ps1) shim ───────────────────────────────────────────────
	ps1Content := fmt.Sprintf(`
# UniRTM shim for %[1]s
# Generated by UniRTM — do not edit manually
& "%[2]s" shim $MyInvocation.MyCommand.Name @args
`,
		executable,
		unirtmPath,
	)

	ps1Path := filepath.Join(g.shimsDir, executable+".ps1")
	if err := os.MkdirAll(filepath.Dir(ps1Path), 0755); err != nil {
		return fmt.Errorf("create shim directory for %s: %w", tool, err)
	}
	if err := os.WriteFile(ps1Path, []byte(ps1Content), 0644); err != nil {
		return fmt.Errorf("write .ps1 shim for %s: %w", tool, err)
	}

	return nil
}

// toolVersionEnvVar returns the environment variable name for a tool's active version.
// e.g. "node" → "UNIRTM_NODE_VERSION"
func toolVersionEnvVar(tool string) string {
	upper := strings.ToUpper(strings.ReplaceAll(tool, "-", "_"))
	return fmt.Sprintf("UNIRTM_%s_VERSION", upper)
}

// ExecuteBinary executes a binary with arguments, replacing the current process on Unix.
func ExecuteBinary(binPath string, args []string) error {
	if runtime.GOOS == "windows" {
		// On Windows, we must use exec.Command because syscall.Exec is not available
		cmd := exec.Command(binPath, args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// On Unix, replace the current process
	return syscall.Exec(binPath, args, os.Environ())
}
