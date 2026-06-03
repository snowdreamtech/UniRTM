// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package env

import (
	"errors"
	"os"
	"testing"
)

func TestEnvPaths_CoverageMore(t *testing.T) {
	// 1. Test GetConfigDir windows path
	oldOS := RuntimeGOOS
	RuntimeGOOS = "windows"

	oldOsUserConfigDir := OsUserConfigDir
	OsUserConfigDir = func() (string, error) { return "C:\\Users\\test\\AppData\\Roaming", nil }
	GetConfigDir()

	OsUserConfigDir = func() (string, error) { return "", errors.New("err") }
	GetConfigDir()

	// 2. Test GetDataDir windows
	os.Setenv("UNIRTM_LOCALAPPDATA", "C:\\AppData\\Local")
	GetDataDir()
	os.Unsetenv("UNIRTM_LOCALAPPDATA")
	GetDataDir()

	// 3. Test GetCacheDir darwin
	RuntimeGOOS = "darwin"
	GetCacheDir()

	// 4. Test GetCacheDir windows
	RuntimeGOOS = "windows"
	GetCacheDir()

	// 5. Test GetLockFilePath error
	oldOsGetwd := OsGetwd
	OsGetwd = func() (string, error) { return "", errors.New("err") }
	GetLockFilePath()

	// Reset
	RuntimeGOOS = oldOS
	OsUserConfigDir = oldOsUserConfigDir
	OsGetwd = oldOsGetwd
}
