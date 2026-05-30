// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package envpath

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// JoinForOS joins multiple paths into a single string using the native operating system's
// list separator (e.g. ';' on Windows, ':' on Posix). This must be used when injecting
// PATH values into Go's native OS API execution contexts (os.Setenv, exec.Cmd).
func JoinForOS(paths []string) string {
	return strings.Join(paths, string(os.PathListSeparator))
}

// JoinForPosix joins multiple paths into a single string meant for POSIX shell scripts
// (Bash, Zsh, Sh). It strictly uses the POSIX separator ':' regardless of the underlying OS,
// and sanitizes Windows backslashes into forward slashes to prevent shell escape bugs.
func JoinForPosix(paths []string) string {
	var posixPaths []string
	for _, p := range paths {
		posixPaths = append(posixPaths, FormatDirForPosix(p))
	}
	return strings.Join(posixPaths, ":")
}

// FormatDirForPosix ensures that a single directory path is safe for injection into
// POSIX shell scripts (Bash, Zsh). On Windows, it converts backslashes to forward slashes.
func FormatDirForPosix(dir string) string {
	if runtime.GOOS == "windows" {
		return filepath.ToSlash(dir)
	}
	return dir
}

// JoinForPowerShell joins multiple paths into a single string meant for PowerShell scripts.
// PowerShell fundamentally requires the OS native PathListSeparator.
func JoinForPowerShell(paths []string) string {
	return JoinForOS(paths)
}

// DeduplicateOSPaths takes a raw native PATH string and removes duplicates,
// correctly handling case-insensitivity on Windows.
func DeduplicateOSPaths(pathStr string) string {
	if pathStr == "" {
		return ""
	}
	parts := strings.Split(pathStr, string(os.PathListSeparator))
	seen := make(map[string]bool)
	var result []string
	for _, p := range parts {
		if p == "" {
			continue
		}
		key := p
		if runtime.GOOS == "windows" {
			key = strings.ToLower(p)
		}
		if !seen[key] {
			seen[key] = true
			result = append(result, p)
		}
	}
	return JoinForOS(result)
}
