package config

import (
	"os"
	"testing"
)

func TestLoader_TopLevelMethods(t *testing.T) {
	tempDir := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(tempDir)
	
	cfgContent := `
[tools]
node = "18.0.0"
`
	os.WriteFile(".unirtm.toml", []byte(cfgContent), 0644)
	
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Tools["node"].Version != "18.0.0" {
		t.Errorf("expected node 18.0.0")
	}
	
	cfg, err = LoadFull()
	if err != nil {
		t.Fatalf("LoadFull failed: %v", err)
	}
	if cfg.Tools["node"].Version != "18.0.0" {
		t.Errorf("expected node 18.0.0")
	}
	
	_, err = LoadGlobal()
	if err != nil {
		// Might fail if user home is not set/doesn't have config, we don't care
		t.Logf("LoadGlobal passed/failed: %v", err)
	}
	
	cfg2, err := LoadHierarchy(tempDir)
	if err != nil {
		t.Fatalf("LoadHierarchy failed: %v", err)
	}
	if cfg2 == nil {
		t.Errorf("expected hierarchy config, got nil")
	}
	
	// ApplyEnvironment test
	cfg.Env = map[string]interface{}{
		"TEST_ENV_VAR": "HELLO",
	}
	cfg.ApplyEnvironment()
	if os.Getenv("TEST_ENV_VAR") != "HELLO" {
		t.Errorf("expected TEST_ENV_VAR to be HELLO")
	}
}
