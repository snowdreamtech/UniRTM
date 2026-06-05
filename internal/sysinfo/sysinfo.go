// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package sysinfo

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	isMuslCached bool
	isMuslOnce   sync.Once
)

// IsMusl detects if the underlying Linux system uses the musl libc (e.g. Alpine Linux).
func IsMusl() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	isMuslOnce.Do(func() {
		isMuslCached = checkMusl()
	})
	return isMuslCached
}

func checkMusl() bool {
	// 1. Check for Alpine release file
	if _, err := os.Stat("/etc/alpine-release"); err == nil {
		return true
	}

	// 2. Check for common musl dynamic linker paths
	matches, err := filepath.Glob("/lib/ld-musl-*.so.1")
	if err == nil && len(matches) > 0 {
		return true
	}

	// 3. Check ldd output
	cmd := exec.Command("ldd", "--version")
	out, err := cmd.CombinedOutput()
	// ldd --version might exit with 1 on musl, but it still prints to stdout/stderr
	if strings.Contains(strings.ToLower(string(out)), "musl") {
		return true
	}

	return false
}
