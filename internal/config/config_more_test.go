package config

import (
	"encoding/json"
	"testing"
)

func TestConfig_UnmarshalJSON(t *testing.T) {
	jsonData := []byte(`{
		"ToolsRaw": {
			"node": "20.0.0"
		},
		"env": {
			"NODE_ENV": "production"
		}
	}`)
	
	var cfg Config
	err := json.Unmarshal(jsonData, &cfg)
	if err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	cfg.PostLoad()
	
	if cfg.Tools["node"].Version != "20.0.0" {
		t.Errorf("expected tool version 20.0.0, got %s", cfg.Tools["node"].Version)
	}
	
	if cfg.Env["NODE_ENV"] != "production" {
		t.Errorf("expected env NODE_ENV to be production, got %v", cfg.Env["NODE_ENV"])
	}
	
	// Duration parsing
	var d DurationOrInt
	err = json.Unmarshal([]byte(`"10s"`), &d)
	if err != nil {
		t.Errorf("expected valid duration string parsing")
	}
	err = json.Unmarshal([]byte(`10`), &d)
	if err != nil {
		t.Errorf("expected valid duration int parsing")
	}
	err = json.Unmarshal([]byte(`"invalid"`), &d)
	if err == nil {
		t.Errorf("expected error for invalid duration")
	}
	
	_, err = ParseDurationToSeconds("30s")
	if err != nil {
		t.Errorf("ParseDurationToSeconds failed: %v", err)
	}
	_, err = ParseDurationToSeconds("invalid")
	if err == nil {
		t.Error("ParseDurationToSeconds invalid duration")
	}
}

func TestConfig_MarshalTOML(t *testing.T) {
	cfg := Config{
		Tools: map[string]ToolConfig{
			"node": {Version: "20.0.0"},
		},
		Env: map[string]interface{}{
			"TEST_ENV": "test",
		},
	}
	
	// Actually MarshalTOML is a method on ToolMap, not Config
	data, err := cfg.Tools.MarshalTOML()
	if err != nil {
		t.Fatalf("MarshalTOML failed: %v", err)
	}
	if data == nil {
		t.Error("expected TOML data, got empty")
	}
}

func TestConfig_Merge(t *testing.T) {
	c1 := &Config{
		Tools: map[string]ToolConfig{
			"node": {Version: "18.0.0"},
			"go":   {Version: "1.20"},
		},
		Env: map[string]interface{}{
			"A": "1",
		},
		Tasks: map[string]Task{
			"build": {Run: "make"},
		},
		Environments: map[string]EnvironmentConfig{
			"prod": {Env: map[string]interface{}{"P": "1"}},
		},
		Aliases: map[string]map[string]string{
			"npm": {"i": "install"},
		},
	}
	
	c2 := &Config{
		Tools: map[string]ToolConfig{
			"node": {Version: "20.0.0"}, // c1 takes precedence based on comment
			"python": {Version: "3.10"},
		},
		Env: map[string]interface{}{
			"A": "2",
			"B": "3",
		},
		Tasks: map[string]Task{
			"build": {Run: "make2"},
			"test": {Run: "test"},
		},
		Environments: map[string]EnvironmentConfig{
			"prod": {Env: map[string]interface{}{"P": "2"}},
			"dev": {Env: map[string]interface{}{"D": "1"}},
		},
		Aliases: map[string]map[string]string{
			"npm": {"ci": "ci"},
			"yarn": {"add": "add"},
		},
	}
	
	c1.Merge(c2)
	
	// c1 takes precedence according to doc comment:
	// "The current configuration takes precedence over the other one (deep merge)."
	if c1.Tools["node"].Version != "18.0.0" {
		t.Errorf("expected node 18.0.0, got %s", c1.Tools["node"].Version)
	}
	if c1.Tools["python"].Version != "3.10" {
		t.Errorf("expected python 3.10, got %s", c1.Tools["python"].Version)
	}
	if c1.Env["A"] != "1" {
		t.Errorf("expected A=1, got %v", c1.Env["A"])
	}
	if c1.Env["B"] != "3" {
		t.Errorf("expected B=3, got %v", c1.Env["B"])
	}
	
	// c1 nil Merge
	var c3 Config
	c3.Merge(nil)
}
