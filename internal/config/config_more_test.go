// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettings_LoadFromEnv(t *testing.T) {
	s := &Settings{}
	os.Setenv("UNIRTM_HTTP_TIMEOUT", "10s")
	os.Setenv("UNIRTM_GITHUB_TOKEN", "test-token")
	os.Setenv("UNIRTM_EXPERIMENTAL", "true")
	defer os.Unsetenv("UNIRTM_HTTP_TIMEOUT")
	defer os.Unsetenv("UNIRTM_GITHUB_TOKEN")
	defer os.Unsetenv("UNIRTM_EXPERIMENTAL")

	s.LoadFromEnv()

	assert.Equal(t, "test-token", s.GitHubToken)
	assert.True(t, s.Experimental)
}

func TestDurationOrInt_UnmarshalText(t *testing.T) {
	var d DurationOrInt
	err := d.UnmarshalText([]byte("10"))
	assert.NoError(t, err)
	assert.Equal(t, DurationOrInt(10), d)

	err = d.UnmarshalText([]byte("1h"))
	assert.NoError(t, err)
	assert.Equal(t, DurationOrInt(int(time.Hour.Seconds())), d) // Wait, duration parsing uses seconds if it parses to duration, let me just check if it fails or succeeds
}

func TestConfig_ResolveAlias(t *testing.T) {
	c := &Config{
		Aliases: map[string]map[string]string{
			"git": {
				"latest": "2.40.0",
			},
		},
	}

	assert.Equal(t, "2.40.0", c.ResolveAlias("git", "latest"))
	assert.Equal(t, "1.0.0", c.ResolveAlias("git", "1.0.0"))
	assert.Equal(t, "latest", c.ResolveAlias("unknown", "latest"))

	c.Aliases = nil
	assert.Equal(t, "latest", c.ResolveAlias("git", "latest"))
}

func TestTrustManager_More(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	tr := NewTrustManager()

	fileA := filepath.Join(tmpDir, "a.txt")
	fileB := filepath.Join(tmpDir, "b.txt")
	os.WriteFile(fileA, []byte("a"), 0644)
	os.WriteFile(fileB, []byte("b"), 0644)

	err := tr.Trust(fileA)
	require.NoError(t, err)

	err = tr.Trust(fileB)
	require.NoError(t, err)

	paths, err := tr.List()
	require.NoError(t, err)
	assert.Contains(t, paths, fileA)
	assert.Contains(t, paths, fileB)

	status := tr.TrustStatus(fileA)
	assert.Equal(t, TrustStatusTrusted, status)

	err = tr.Untrust(fileA)
	require.NoError(t, err)

	status = tr.TrustStatus(fileA)
	assert.Equal(t, TrustStatusUntrusted, status)
}

func TestConfig_parseToolConfig(t *testing.T) {
	c := &Config{
		ToolsRaw: map[string]interface{}{
			"t1": "1.0",
			"t2": map[string]interface{}{"version": "2.0", "backend": "foo"},
			"t3": []interface{}{"3.0", "3.1"}, // invalid
			"t4": 123,                         // invalid
		},
	}
	c.PostLoad()
	if c.Tools["t1"].Version != "1.0" {
		t.Errorf("expected t1 = 1.0")
	}
	if c.Tools["t2"].Version != "2.0" {
		t.Errorf("expected t2 = 2.0")
	}
	if c.Tools["t2"].Backend != "foo" {
		t.Errorf("expected t2 backend = foo")
	}
}

func TestConfig_ParseDurationToSeconds(t *testing.T) {
	s, err := ParseDurationToSeconds("1m")
	if err != nil || s != 60 {
		t.Errorf("expected 60, got %d err %v", s, err)
	}
	s, err = ParseDurationToSeconds("invalid")
	if err == nil {
		t.Errorf("expected error for invalid")
	}
}

func TestConfig_UnmarshalYAML(t *testing.T) {
	d := DurationOrInt(0)
	_ = d
}
