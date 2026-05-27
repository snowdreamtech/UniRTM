package config

import (
	"os"
	"testing"
)

func TestSettings_LoadFromEnv2(t *testing.T) {
	// Set a bunch of env vars
	envVars := map[string]string{
		"CACHE_DIR":            "/tmp/cache",
		"DATA_DIR":             "/tmp/data",
		"CACHE_TTL":            "1h",
		"LOCKFILE":             "true",
		"LOCKED":               "1",
		"GITHUB_PROXY":         "proxy",
		"HTTP_PROXY":           "http",
		"HTTPS_PROXY":          "https",
		"GITHUB_TOKEN":         "token",
		"HTTP_TIMEOUT":         "30s",
		"TASK_TIMEOUT":         "60m",
		"TASK_OUTPUT":          "prefix",
		"EXPERIMENTAL":         "true",
		"AUTO_INSTALL":         "1",
		"COLOR":                "always",
		"EDITOR":               "vim",
		"VISUAL":               "nano",
		"SHELL":                "bash",
		"ALWAYS_KEEP_DOWNLOAD": "true",
		"CEILING_PATHS":        "/a:/b",
		"TRUSTED_CONFIG_PATHS": "/c:/d",
		"GPG_VERIFY":           "strict",
		"GPG_KEYS":             "key1,key2",
		"VERIFY_METADATA":      "true",
		"NO_PROXY":             "localhost,127.0.0.1",
		"JOBS":                 "4",
		"MINIMUM_RELEASE_AGE":  "7d",
	}

	for k, v := range envVars {
		os.Setenv("UNIRTM_"+k, v)
		defer os.Unsetenv("UNIRTM_"+k)
	}

	s := &Settings{}
	s.LoadFromEnv()

	if s.CacheDir != "/tmp/cache" {
		t.Errorf("expected cache dir /tmp/cache, got %s", s.CacheDir)
	}
	if s.Jobs != 4 {
		t.Errorf("expected jobs 4, got %d", s.Jobs)
	}
	if s.CacheTTL != 3600 {
		t.Errorf("expected cache ttl 3600, got %d", s.CacheTTL)
	}

	// Also test integer durations and integer parse failures
	os.Setenv("UNIRTM_CACHE_TTL", "300")
	os.Setenv("UNIRTM_HTTP_TIMEOUT", "60")
	os.Setenv("UNIRTM_TASK_TIMEOUT", "120")
	s2 := &Settings{}
	s2.LoadFromEnv()
	if s2.CacheTTL != 300 {
		t.Errorf("expected 300, got %d", s2.CacheTTL)
	}
}

func TestConfig_Merge(t *testing.T) {
	c1 := &Config{
		Tools: ToolMap{"t1": ToolConfig{Version: "1.0"}},
		Env:   map[string]interface{}{"E1": "V1"},
		Tasks: map[string]Task{"task1": {Run: "echo 1"}},
	}
	c2 := &Config{
		Tools: ToolMap{"t1": ToolConfig{Version: "2.0"}, "t2": ToolConfig{Version: "2.0"}},
		Env:   map[string]interface{}{"E1": "V2", "E2": "V2"},
		Tasks: map[string]Task{"task1": {Run: "echo 2"}, "task2": {Run: "echo 2"}},
	}

	c1.Merge(c2)

	if c1.Tools["t1"].Version != "1.0" {
		t.Errorf("expected t1 version 1.0, got %v", c1.Tools["t1"].Version)
	}
	if c1.Tools["t2"].Version != "2.0" {
		t.Errorf("expected t2 version 2.0")
	}

	if c1.Env["E1"] != "V1" {
		t.Errorf("expected E1 V1")
	}
	if c1.Env["E2"] != "V2" {
		t.Errorf("expected E2 V2")
	}

	if c1.Tasks["task1"].Run != "echo 1" {
		t.Errorf("expected task1 Script echo 1")
	}
	if c1.Tasks["task2"].Run != "echo 2" {
		t.Errorf("expected task2 Script echo 2")
	}

	// Test nil merge
	c1.Merge(nil)
	var cNil *Config
	cNil.Merge(c2)
}

func TestDurationOrInt_JSON(t *testing.T) {
	var d DurationOrInt
	err := d.UnmarshalJSON([]byte("3600"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if d != 3600 {
		t.Errorf("expected 3600, got %v", d)
	}

	err = d.UnmarshalJSON([]byte(`"1h"`))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if d != 3600 {
		t.Errorf("expected 3600, got %v", d)
	}

	err = d.UnmarshalJSON([]byte(`"invalid"`))
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestToolMap_MarshalTOML(t *testing.T) {
	tm := ToolMap{
		"t1": ToolConfig{Version: "1.0"},
		"t2": ToolConfig{Version: "2.0", Backend: "foo"},
	}
	v, err := tm.MarshalTOML()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		t.Errorf("expected map[string]interface{}")
	}
	if m["t1"] != "1.0" {
		t.Errorf("expected 1.0, got %v", m["t1"])
	}
	if tc, ok := m["t2"].(ToolConfig); !ok || tc.Version != "2.0" {
		t.Errorf("expected t2 ToolConfig version 2.0")
	}
}

func TestParseDurationToSeconds(t *testing.T) {
	if sec, err := parseDurationToSeconds("1h"); err != nil || sec != 3600 {
		t.Errorf("expected 3600, got %v, %v", sec, err)
	}
}

func TestToolConfig_Parse(t *testing.T) {
	c := &Config{
		ToolsRaw: map[string]interface{}{
			"str": "1.0",
			"map": map[string]interface{}{"version": "2.0"},
		},
	}
	c.PostLoad()
	if c.Tools["str"].Version != "1.0" {
		t.Errorf("expected 1.0, got %v", c.Tools["str"].Version)
	}
	if c.Tools["map"].Version != "2.0" {
		t.Errorf("expected 2.0, got %v", c.Tools["map"].Version)
	}
}
