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

func TestGet_Validation(t *testing.T) {
	// GITHUB_PROXY validation
	t.Setenv("UNIRTM_GITHUB_PROXY", "ftp://invalid-url")
	if Get("GITHUB_PROXY") != "" {
		t.Errorf("expected empty string for invalid GITHUB_PROXY, got %q", Get("GITHUB_PROXY"))
	}
	t.Setenv("UNIRTM_GITHUB_PROXY", "https://valid.proxy.com/")
	if Get("GITHUB_PROXY") != "https://valid.proxy.com/" {
		t.Errorf("expected valid GITHUB_PROXY to pass, got %q", Get("GITHUB_PROXY"))
	}
	t.Setenv("UNIRTM_GITHUB_PROXY", "direct")
	if Get("GITHUB_PROXY") != "direct" {
		t.Errorf("expected direct GITHUB_PROXY to pass, got %q", Get("GITHUB_PROXY"))
	}

	// JOBS validation
	t.Setenv("UNIRTM_JOBS", "invalid")
	if Get("JOBS") != "" {
		t.Errorf("expected empty string for non-numeric JOBS")
	}
	t.Setenv("UNIRTM_JOBS", "0")
	if Get("JOBS") != "" {
		t.Errorf("expected empty string for JOBS < 1")
	}
	t.Setenv("UNIRTM_JOBS", "999")
	if Get("JOBS") != "" {
		t.Errorf("expected empty string for JOBS > 256")
	}
	t.Setenv("UNIRTM_JOBS", "8")
	if Get("JOBS") != "8" {
		t.Errorf("expected valid JOBS to pass, got %q", Get("JOBS"))
	}

	// HTTP2 validation
	t.Setenv("UNIRTM_HTTP2", "2")
	if Get("HTTP2") != "" {
		t.Errorf("expected empty string for HTTP2 other than 0 or 1")
	}
	t.Setenv("UNIRTM_HTTP2", "0")
	if Get("HTTP2") != "0" {
		t.Errorf("expected HTTP2=0 to pass")
	}
	t.Setenv("UNIRTM_HTTP2", "1")
	if Get("HTTP2") != "1" {
		t.Errorf("expected HTTP2=1 to pass")
	}
}

func TestEnvManager(t *testing.T) {
	_ = EnvManager{}
}

func TestSmartCache_DynamicUpdates(t *testing.T) {
	const key = "TEST_SMART_CACHE_VAR"

	// 1. Initially unset
	_ = Unset(key)
	if val := Get(key); val != "" {
		t.Fatalf("expected empty, got %q", val)
	}

	// 2. Set via env.Set
	if err := Set(key, "value_1"); err != nil {
		t.Fatalf("failed to Set: %v", err)
	}
	if val := Get(key); val != "value_1" {
		t.Fatalf("expected 'value_1', got %q", val)
	}

	// 3. Update via env.Set
	if err := Set(key, "value_2"); err != nil {
		t.Fatalf("failed to update Set: %v", err)
	}
	if val := Get(key); val != "value_2" {
		t.Fatalf("expected 'value_2', got %q", val)
	}

	// 4. Update externally via os.Setenv (simulating runtime/child process change)
	_ = os.Setenv(key, "value_external")
	if val := Get(key); val != "value_external" {
		t.Fatalf("smart cache failed to detect external os.Setenv change: expected 'value_external', got %q", val)
	}

	// 5. Unset via env.Unset
	if err := Unset(key); err != nil {
		t.Fatalf("failed to Unset: %v", err)
	}
	if val := Get(key); val != "" {
		t.Fatalf("expected empty after Unset, got %q", val)
	}
	if Has(key) {
		t.Fatalf("expected Has(key) to be false after Unset")
	}

	// 6. External removal via os.Unsetenv
	_ = os.Setenv(key, "temp_val")
	if val := Get(key); val != "temp_val" {
		t.Fatalf("expected 'temp_val', got %q", val)
	}
	_ = os.Unsetenv(key)
	if val := Get(key); val != "" {
		t.Fatalf("smart cache failed to detect external os.Unsetenv removal: expected empty, got %q", val)
	}
}

func TestSmartCache_PrefixAndCaseTolerance(t *testing.T) {
	// Clean up before and after
	defer func() {
		_ = os.Unsetenv("UNIRTM_CUSTOM_API_KEY")
		_ = os.Unsetenv("MISE_CUSTOM_API_KEY")
		_ = os.Unsetenv("CUSTOM_API_KEY")
		ResetCache()
	}()

	// 1. Raw key with mixed case
	_ = os.Setenv("CUSTOM_API_KEY", "raw_val")
	if val := Get("custom_api_key"); val != "raw_val" {
		t.Errorf("expected 'raw_val', got %q", val)
	}
	if val := Get("Custom-Api-Key"); val != "raw_val" {
		t.Errorf("expected 'raw_val' for dashed key, got %q", val)
	}

	// 2. MISE_ prefix takes precedence over raw key
	_ = os.Setenv("MISE_CUSTOM_API_KEY", "mise_val")
	if val := Get("CUSTOM_API_KEY"); val != "mise_val" {
		t.Errorf("expected 'mise_val' (MISE precedence), got %q", val)
	}

	// 3. UNIRTM_ prefix takes highest precedence
	_ = os.Setenv("UNIRTM_CUSTOM_API_KEY", "unirtm_val")
	if val := Get("CUSTOM_API_KEY"); val != "unirtm_val" {
		t.Errorf("expected 'unirtm_val' (UNIRTM precedence), got %q", val)
	}

	// 4. Test Unset clears all variants
	if err := Unset("CUSTOM_API_KEY"); err != nil {
		t.Fatalf("failed to Unset: %v", err)
	}
	if val := Get("CUSTOM_API_KEY"); val != "" {
		t.Errorf("expected empty after Unset all variants, got %q", val)
	}
}
