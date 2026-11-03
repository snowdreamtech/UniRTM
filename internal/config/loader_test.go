// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestResolveEnvironment_NilEnv tests that nil Env is handled gracefully.
func TestResolveEnvironment_NilEnv(t *testing.T) {
	c := &Config{Env: nil}
	resolved, sources, redacted, err := c.ResolveEnvironment()
	assert.NoError(t, err)
	assert.Empty(t, resolved)
	assert.Empty(t, sources)
	assert.Empty(t, redacted)
}

// TestResolveEnvironment_PathString tests _.path with a plain string value.
func TestResolveEnvironment_PathString(t *testing.T) {
	c := &Config{
		Env: map[string]interface{}{
			"_.path": "/my/custom/bin",
		},
	}
	resolved, _, _, err := c.ResolveEnvironment()
	assert.NoError(t, err)
	assert.Contains(t, resolved["PATH"], "/my/custom/bin")
}

// TestResolveEnvironment_PathTemplate tests _.path with a Jinja2 template.
func TestResolveEnvironment_PathTemplate(t *testing.T) {
	t.Setenv("MY_HOME", "/home/user")
	c := &Config{
		Env: map[string]interface{}{
			"_.path": "{{ env.MY_HOME }}/bin",
		},
	}
	resolved, _, _, err := c.ResolveEnvironment()
	assert.NoError(t, err)
	assert.Contains(t, resolved["PATH"], "/home/user/bin")
}

// TestResolveEnvironment_SourceTemplate tests _.source with a Jinja2 template.
func TestResolveEnvironment_SourceTemplate(t *testing.T) {
	t.Setenv("PROFILE_DIR", "/etc")
	c := &Config{
		Env: map[string]interface{}{
			"_.source": "{{ env.PROFILE_DIR }}/profile.d/myapp.sh",
		},
	}
	_, sources, _, err := c.ResolveEnvironment()
	assert.NoError(t, err)
	assert.Contains(t, sources, "/etc/profile.d/myapp.sh")
}

// TestResolveEnvironment_PythonVenv tests _.python_venv with an existing venv directory.
func TestResolveEnvironment_PythonVenv(t *testing.T) {
	tmpDir := t.TempDir()

	binDirName := "bin"
	if runtime.GOOS == "windows" {
		binDirName = "Scripts"
	}

	// Create a mock venv structure
	venvBin := filepath.Join(tmpDir, "venv", binDirName)
	require.NoError(t, os.MkdirAll(venvBin, 0755))

	c := &Config{
		Env: map[string]interface{}{
			"_.python_venv": filepath.Join(tmpDir, "venv"),
		},
	}
	resolved, _, _, err := c.ResolveEnvironment()
	assert.NoError(t, err)
	assert.Contains(t, resolved["PATH"], venvBin)
	assert.Equal(t, filepath.Join(tmpDir, "venv"), resolved["VIRTUAL_ENV"])
}

// TestResolveEnvironment_PythonVenv_Nonexistent tests _.python_venv when the venv dir doesn't exist.
func TestResolveEnvironment_PythonVenv_Nonexistent(t *testing.T) {
	c := &Config{
		Env: map[string]interface{}{
			"_.python_venv": "/nonexistent/venv/path",
		},
	}
	resolved, _, _, err := c.ResolveEnvironment()
	assert.NoError(t, err)
	assert.Empty(t, resolved["VIRTUAL_ENV"], "VIRTUAL_ENV should not be set for non-existent venv")
}

// TestResolveEnvironment_PythonVenv_RelativePath tests _.python_venv with a relative path.
func TestResolveEnvironment_PythonVenv_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()

	binDirName := "bin"
	if runtime.GOOS == "windows" {
		binDirName = "Scripts"
	}

	venvBin := filepath.Join(tmpDir, ".venv", binDirName)
	require.NoError(t, os.MkdirAll(venvBin, 0755))

	// Change into tmpDir so the relative path resolves correctly
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)
	require.NoError(t, os.Chdir(tmpDir))

	c := &Config{
		Env: map[string]interface{}{
			"_.python_venv": ".venv",
		},
	}
	resolved, _, _, err := c.ResolveEnvironment()
	assert.NoError(t, err)
	assert.Contains(t, resolved["VIRTUAL_ENV"], ".venv")
}

// TestResolveEnvironment_File tests the _.file directive with a dotenv file.
func TestResolveEnvironment_File(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	require.NoError(t, os.WriteFile(envFile, []byte("FROM_FILE=hello\nSECOND=world\n"), 0644))

	c := &Config{
		Env: map[string]interface{}{
			"_.file": envFile,
		},
	}
	resolved, _, _, err := c.ResolveEnvironment()
	assert.NoError(t, err)
	assert.Equal(t, "hello", resolved["FROM_FILE"])
	assert.Equal(t, "world", resolved["SECOND"])
}

// TestResolveEnvironment_File_InvalidPath tests the _.file directive with a non-existent file.
func TestResolveEnvironment_File_InvalidPath(t *testing.T) {
	c := &Config{
		Env: map[string]interface{}{
			"_.file": "/nonexistent/.env",
		},
	}
	// Should not error - godotenv.Read errors are silently ignored
	_, _, _, err := c.ResolveEnvironment()
	assert.NoError(t, err)
}

// TestResolveEnvironment_RequiredWithHelp tests that the help text is included in the error message.
func TestResolveEnvironment_RequiredWithHelp(t *testing.T) {
	c := &Config{
		Env: map[string]interface{}{
			"API_KEY": map[string]interface{}{
				"required": true,
				"value":    "",
				"help":     "Set your API key from https://example.com/api",
			},
		},
	}
	_, _, _, err := c.ResolveEnvironment()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API_KEY")
	assert.Contains(t, err.Error(), "https://example.com/api")
}

// TestResolveEnvironment_IsDefined tests the 'is defined' Jinja2-like syntax.
func TestResolveEnvironment_IsDefined(t *testing.T) {
	t.Setenv("EXISTING_VAR", "exists")
	os.Unsetenv("MISSING_VAR")

	c := &Config{
		Env: map[string]interface{}{
			// 'is defined' syntax: if EXISTING_VAR is defined, use 'yes'
			"DEFINED_RESULT": "{% if env.EXISTING_VAR is defined %}yes{% else %}no{% endif %}",
			// 'is not defined' syntax
			"NOT_DEFINED_RESULT": "{% if env.MISSING_VAR is not defined %}missing{% else %}present{% endif %}",
		},
	}
	resolved, _, _, err := c.ResolveEnvironment()
	assert.NoError(t, err)
	assert.Equal(t, "yes", resolved["DEFINED_RESULT"])
	assert.Equal(t, "missing", resolved["NOT_DEFINED_RESULT"])
}

// TestResolveEnvironment_DictRmFalse tests dict env var with rm=false and value field.
func TestResolveEnvironment_DictRmFalse(t *testing.T) {
	c := &Config{
		Env: map[string]interface{}{
			"STAY_VAR": map[string]interface{}{
				"value": "stay",
				"rm":    false,
			},
		},
	}
	resolved, _, _, err := c.ResolveEnvironment()
	assert.NoError(t, err)
	assert.Equal(t, "stay", resolved["STAY_VAR"])
}

// TestResolveEnvironment_PathPrependsExistingPATH tests that existing PATH is prepended.
func TestResolveEnvironment_PathPrependsExistingPATH(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/existing/path")
	defer t.Setenv("PATH", origPath)

	c := &Config{
		Env: map[string]interface{}{
			"_.path": "/new/bin",
		},
	}
	resolved, _, _, err := c.ResolveEnvironment()
	assert.NoError(t, err)
	path := resolved["PATH"]
	assert.True(t, strings.HasPrefix(path, "/new/bin"), "new path should be prepended, got: %s", path)
	assert.Contains(t, path, "/existing/path")
}

// TestApplyEnvironment tests that ApplyEnvironment sets env vars on the process.
func TestApplyEnvironment(t *testing.T) {
	c := &Config{
		Env: map[string]interface{}{
			"APPLY_TEST_VAR": "applied_value",
		},
	}

	os.Unsetenv("APPLY_TEST_VAR")
	c.ApplyEnvironment()
	assert.Equal(t, "applied_value", os.Getenv("APPLY_TEST_VAR"))
	os.Unsetenv("APPLY_TEST_VAR")
}

// TestApplyEnvironment_WithRequired tests ApplyEnvironment still applies vars even if required fails.
func TestApplyEnvironment_WithRequired(t *testing.T) {
	c := &Config{
		Env: map[string]interface{}{
			"PARTIAL_VAR": "partial_value",
			"MISSING_REQ": map[string]interface{}{
				"required": true,
				"value":    "",
			},
		},
	}

	os.Unsetenv("PARTIAL_VAR")
	// ApplyEnvironment should not panic or crash even if resolution has errors
	c.ApplyEnvironment()
	assert.Equal(t, "partial_value", os.Getenv("PARTIAL_VAR"))
	os.Unsetenv("PARTIAL_VAR")
}

// TestLoadFull_InEmptyDir tests LoadFull in a directory with no config files.
func TestLoadFull_InEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	require.NoError(t, os.Chdir(tmpDir))

	cfg, err := LoadFull()
	// With no config file, should succeed with empty config
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

// TestLoadGlobal_FileNotFound tests LoadGlobal when the global config file doesn't exist.
func TestLoadGlobal_FileNotFound(t *testing.T) {
	origReadFile := OsReadFile
	defer func() { OsReadFile = origReadFile }()
	OsReadFile = func(filename string) ([]byte, error) {
		return nil, os.ErrNotExist
	}

	cfg, err := LoadGlobal()
	// When file not found, should return empty config with error
	assert.Error(t, err)
	assert.NotNil(t, cfg)
}

// TestLoadFromDir_InvalidTOML tests that invalid TOML in a config file returns an error.
func TestLoadFromDir_InvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "unirtm.toml"),
		[]byte("[invalid toml syntax{{{{"),
		0644,
	))
	_, err := LoadFromDir(tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal config")
}

// TestGetGlobalConfigPath_HomeDir tests GetGlobalConfigPath format.
func TestGetGlobalConfigPath_HomeDir(t *testing.T) {
	path := GetGlobalConfigPath()
	assert.True(t, strings.HasSuffix(path, "unirtm.toml"),
		fmt.Sprintf("expected path to end with unirtm.toml, got: %s", path))
	assert.Contains(t, path, "unirtm")
}
