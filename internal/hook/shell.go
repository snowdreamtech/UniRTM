// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hook

import (
	"os/exec"
	"runtime"
)

// GetShell returns the appropriate shell executable for the current OS.
// On Windows, it tries to find sh.exe (provided by Git for Windows / Git Bash).
// On Unix-like systems (Linux, macOS), it returns "/bin/sh".
func GetShell() string {
	if runtime.GOOS == "windows" {
		return getWindowsShell()
	}
	return "/bin/sh"
}

// getWindowsShell attempts to locate sh.exe as provided by Git for Windows.
// Falls back to plain "sh" if a full path cannot be resolved, letting the OS
// PATH resolution handle it (e.g. in Git Bash / MSYS2 environments).
func getWindowsShell() string {
	// Git for Windows ships sh.exe alongside git.exe; look for it in PATH.
	if path, err := exec.LookPath("sh"); err == nil {
		return path
	}
	// Last resort: hope it's on PATH. This will fail gracefully with a clear
	// "executable not found" error rather than a cryptic sh-injection error.
	return "sh"
}
