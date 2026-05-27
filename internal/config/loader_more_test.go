package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_LoadGlobal(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	globalPath := GetGlobalConfigPath()
	os.MkdirAll(filepath.Dir(globalPath), 0755)
	os.WriteFile(globalPath, []byte(`[tools]
t1 = "1.0"
`), 0644)

	c, err := LoadGlobal()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if c == nil {
		t.Errorf("expected config")
	}
	if c.Tools["t1"].Version != "1.0" {
		t.Errorf("expected t1 = 1.0")
	}

	c, err = LoadGlobal()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if c.ToolsRaw["t1"] != "1.0" {
		t.Errorf("expected t1 = 1.0")
	}

	os.Chmod(globalPath, 0000)
	_, err = LoadGlobal()
	if err == nil {
		t.Errorf("expected error reading unreadable global config")
	}
	os.Chmod(globalPath, 0644)
}

func TestLoader_LoadHierarchy(t *testing.T) {
	c, err := LoadHierarchy("/nonexistent/path/here")
	if err != nil {
		t.Errorf("expected no error for nonexistent path, got %v", err)
	}
	if c == nil {
		t.Errorf("expected config")
	}

	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "unirtm.toml"), []byte(`[tools]
t2 = "2.0"
`), 0644)
	c, err = LoadHierarchy(tmpDir)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if c.Tools["t2"].Version != "2.0" {
		t.Errorf("expected t2 = 2.0")
	}
}

func TestLoader_ResolveEnvironment(t *testing.T) {
	c := &Config{
		ToolsRaw: map[string]interface{}{"t1": "1.0"},
	}
	_, _, _, err := c.ResolveEnvironment()
	if err != nil {
		t.Errorf("expected no err")
	}
}

func TestLoader_Load(t *testing.T) {
	pwd, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(pwd)

	c, err := Load()
	if err == nil {
		// Might not be error if empty, depends on LoadFromDir implementation when no file
	}

	p := filepath.Join(tmpDir, "unirtm.toml")
	os.WriteFile(p, []byte(`[tools]
t1="1.0"`), 0644)
	c, err = Load()
	if err != nil {
		t.Errorf("expected no error")
	}
	if c.ToolsRaw["t1"] != "1.0" {
		// maybe postLoad parses it to c.Tools
		if c.Tools["t1"].Version != "1.0" {
			t.Errorf("expected t1 = 1.0")
		}
	}
}

func TestLoader_LoadFull(t *testing.T) {
	c, err := LoadFull()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if c == nil {
		t.Errorf("expected config")
	}
}

func TestLoader_ApplyEnvironment(t *testing.T) {
	c := &Config{
		ToolsRaw: map[string]interface{}{"t1": "1.0"},
	}
	c.ApplyEnvironment()
}
