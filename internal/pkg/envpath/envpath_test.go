// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package envpath

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinForOS(t *testing.T) {
	paths := []string{"dir1", "dir2"}
	result := JoinForOS(paths)
	expected := "dir1" + string(os.PathListSeparator) + "dir2"
	assert.Equal(t, expected, result)
}

func TestJoinForPosix(t *testing.T) {
	paths := []string{"C:\\foo\\bin", "D:\\bar\\bin"}
	result := JoinForPosix(paths)
	if runtime.GOOS == "windows" {
		assert.Equal(t, "C:/foo/bin:D:/bar/bin", result)
	} else {
		assert.Equal(t, "C:\\foo\\bin:D:\\bar\\bin", result)
	}
}

func TestFormatDirForPosix(t *testing.T) {
	dir := "C:\\Users\\test\\shims"
	result := FormatDirForPosix(dir)
	if runtime.GOOS == "windows" {
		assert.Equal(t, filepath.ToSlash(dir), result)
	} else {
		assert.Equal(t, dir, result)
	}
}

func TestJoinForPowerShell(t *testing.T) {
	paths := []string{"dir1", "dir2"}
	result := JoinForPowerShell(paths)
	expected := "dir1" + string(os.PathListSeparator) + "dir2"
	assert.Equal(t, expected, result)
}

func TestDeduplicateOSPaths(t *testing.T) {
	// 1. Empty string
	assert.Equal(t, "", DeduplicateOSPaths(""))

	sep := string(os.PathListSeparator)

	// 2. No duplicates
	input := "dir1" + sep + "dir2"
	assert.Equal(t, "dir1"+sep+"dir2", DeduplicateOSPaths(input))

	// 3. Exact duplicates
	input = "dir1" + sep + "dir2" + sep + "dir1"
	assert.Equal(t, "dir1"+sep+"dir2", DeduplicateOSPaths(input))

	// 4. Consecutive separators / empty parts
	input = "dir1" + sep + sep + "dir2"
	assert.Equal(t, "dir1"+sep+"dir2", DeduplicateOSPaths(input))

	// 5. Case sensitivity testing
	if runtime.GOOS == "windows" {
		// On Windows, case-insensitive
		input = "C:\\dir1" + sep + "c:\\DIR1" + sep + "D:\\dir2"
		assert.Equal(t, "C:\\dir1"+sep+"D:\\dir2", DeduplicateOSPaths(input))
	} else {
		// On POSIX, case-sensitive
		input = "/dir1" + sep + "/DIR1" + sep + "/dir2"
		assert.Equal(t, "/dir1"+sep+"/DIR1"+sep+"/dir2", DeduplicateOSPaths(input))
	}
}
