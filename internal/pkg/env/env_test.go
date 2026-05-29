// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package env

import (
	"os"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	// Setup
	t.Setenv("UNIRTM_TEST_KEY_1", "val1")

	t.Setenv("MISE_TEST_KEY_2", "val2")

	t.Setenv("TEST_KEY_3", "val3")

	t.Setenv("UNIRTM_TEST_KEY_4", "val4_unirtm")
	t.Setenv("MISE_TEST_KEY_4", "val4_mise")
	t.Setenv("TEST_KEY_4", "val4_raw")

	tests := []struct {
		key      string
		expected string
	}{
		{"TEST_KEY_1", "val1"},
		{"TEST_KEY_2", "val2"},
		{"TEST_KEY_3", "val3"},
		{"TEST_KEY_4", "val4_unirtm"},
		{"TEST_KEY_NONEXISTENT", ""},
		{"PATH", os.Getenv("PATH")},
	}

	for _, tc := range tests {
		t.Run(tc.key, func(t *testing.T) {
			result := Get(tc.key)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestRandomString(t *testing.T) {
	str, err := RandomString(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(str) != 10 {
		t.Errorf("expected length 10, got %d", len(str))
	}

	// Test randomness/uniqueness (basic)
	str2, _ := RandomString(10)
	if str == str2 {
		t.Errorf("expected random strings to be different, got %s twice", str)
	}
}

func TestGetFSToolName(t *testing.T) {
	tests := []struct {
		tool     string
		backend  string
		expected string
	}{
		{"npm", "", "npm"},
		{"npm", "native", "npm"},
		{"prettier", "npm", "npm-prettier"},
		{"npm-prettier", "npm", "npm-prettier"},
		{"npm:prettier", "npm", "npm-prettier"},
		{"org/pkg", "github", "github-org-pkg"},
		{"tool@v1", "github", "github-toolv1"},
	}

	for _, tc := range tests {
		t.Run(tc.tool+"_"+tc.backend, func(t *testing.T) {
			result := GetFSToolName(tc.tool, tc.backend)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestDirFunctions(t *testing.T) {
	// Just test that they return strings and don't panic

	configDir := GetConfigDir()
	if configDir == "" {
		t.Error("expected non-empty config dir")
	}

	dataDir := GetDataDir()
	if dataDir == "" {
		t.Error("expected non-empty data dir")
	}

	if !strings.HasPrefix(GetDatabasePath(), dataDir) {
		t.Error("expected database path to be inside data dir")
	}

	if !strings.HasPrefix(GetShimsDir(), dataDir) {
		t.Error("expected shims dir to be inside data dir")
	}

	if !strings.HasPrefix(GetInstallsDir(), dataDir) {
		t.Error("expected installs dir to be inside data dir")
	}

	if !strings.HasPrefix(GetDownloadsDir(), dataDir) {
		t.Error("expected downloads dir to be inside data dir")
	}

	if !strings.HasPrefix(GetPluginsDir(), dataDir) {
		t.Error("expected plugins dir to be inside data dir")
	}

	cacheDir := GetCacheDir()
	if cacheDir == "" {
		t.Error("expected non-empty cache dir")
	}

	lockFile := GetLockFilePath()
	if lockFile == "" {
		t.Error("expected non-empty lock file path")
	}

	if !strings.HasPrefix(GetGlobalConfigPath(), configDir) {
		t.Error("expected global config path to be inside config dir")
	}
}
