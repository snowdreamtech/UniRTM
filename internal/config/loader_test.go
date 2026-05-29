// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveEnvironment(t *testing.T) {
	// Set up basic env vars for testing
	t.Setenv("TEST_BASE", "base_val")

	c := &Config{
		Env: map[string]interface{}{
			"SIMPLE_VAR": "simple",
			"EXPAND_VAR": "expanded_${TEST_BASE}",
			"TPL_VAR":    "{{ env.TEST_BASE }}_tpl",
			"_.path":     []interface{}{"/usr/local/bin", "/opt/bin"},
			"_.source":   "/etc/profile",
			"INT_VAR":    42,
			"BOOL_VAR":   true,
			"DICT_VAR": map[string]interface{}{
				"value": "dict_val",
			},
			"RM_VAR": map[string]interface{}{
				"rm": true,
			},
			"REQ_VAR": map[string]interface{}{
				"required": true,
				"value":    "",
			},
			"REDACT_VAR": map[string]interface{}{
				"redact": true,
				"value":  "secret",
			},
		},
	}

	// Preset RM_VAR in env to test removal (we don't actually check os.Environ here,
	// but we check the returned resolved map)

	resolved, sources, redacted, err := c.ResolveEnvironment()

	if err == nil {
		t.Error("expected error for required variable REQ_VAR, but got nil")
	}

	if resolved["SIMPLE_VAR"] != "simple" {
		t.Errorf("expected SIMPLE_VAR=simple, got %s", resolved["SIMPLE_VAR"])
	}

	if resolved["EXPAND_VAR"] != "expanded_base_val" {
		t.Errorf("expected EXPAND_VAR=expanded_base_val, got %s", resolved["EXPAND_VAR"])
	}

	if resolved["TPL_VAR"] != "base_val_tpl" {
		t.Errorf("expected TPL_VAR=base_val_tpl, got %s", resolved["TPL_VAR"])
	}

	if resolved["INT_VAR"] != "42" {
		t.Errorf("expected INT_VAR=42, got %s", resolved["INT_VAR"])
	}

	if resolved["BOOL_VAR"] != "true" {
		t.Errorf("expected BOOL_VAR=true, got %s", resolved["BOOL_VAR"])
	}

	if resolved["DICT_VAR"] != "dict_val" {
		t.Errorf("expected DICT_VAR=dict_val, got %s", resolved["DICT_VAR"])
	}

	if _, ok := resolved["RM_VAR"]; ok {
		t.Error("expected RM_VAR to be removed")
	}

	if resolved["REDACT_VAR"] != "secret" {
		t.Errorf("expected REDACT_VAR=secret, got %s", resolved["REDACT_VAR"])
	}

	foundRedact := false
	for _, r := range redacted {
		if r == "REDACT_VAR" {
			foundRedact = true
			break
		}
	}
	if !foundRedact {
		t.Error("expected REDACT_VAR in redacted slice")
	}

	foundSource := false
	for _, s := range sources {
		if s == "/etc/profile" {
			foundSource = true
			break
		}
	}
	if !foundSource {
		t.Error("expected /etc/profile in sources slice")
	}

	path := resolved["PATH"]
	if !stringsContainsAll(path, "/usr/local/bin", "/opt/bin") {
		t.Errorf("expected PATH to contain /usr/local/bin and /opt/bin, got %s", path)
	}
}

func stringsContainsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !stringsContains(s, sub) {
			return false
		}
	}
	return true
}

func stringsContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || string(s[0:len(sub)]) == sub || stringsContains(s[1:], sub))
}

// strings.Contains is available in Go 1.0 but since I wrote stringsContainsAll myself I'll just use strings.Contains
func TestLoadFromDir(t *testing.T) {
	tmpDir := t.TempDir()
	tomlPath := filepath.Join(tmpDir, "unirtm.toml")

	content := `
[tools]
node = "18"
`
	if err := os.WriteFile(tomlPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir failed: %v", err)
	}

	if cfg.Tools == nil {
		t.Fatal("expected tools to be non-nil")
	}

	if cfg.Tools["node"].Version != "18" {
		t.Errorf("expected node=18, got %v", cfg.Tools["node"].Version)
	}

	// Test failure on empty dir
	emptyDir := t.TempDir()
	_, err = LoadFromDir(emptyDir)
	if err == nil {
		t.Error("expected error when loading from empty dir")
	}
}

func TestGetGlobalConfigPath(t *testing.T) {
	path := GetGlobalConfigPath()
	if path == "" {
		t.Error("expected non-empty global config path")
	}
}
