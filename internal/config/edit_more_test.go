// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEdit_ReadFileOrEmpty(t *testing.T) {
	tempDir := t.TempDir()

	// File not exist
	content, err := ReadFileOrEmpty(filepath.Join(tempDir, "none.txt"))
	if err != nil {
		t.Fatalf("expected nil err for not exist")
	}
	if content != "" {
		t.Errorf("expected empty content")
	}

	// File exist
	p := filepath.Join(tempDir, "exist.txt")
	os.WriteFile(p, []byte("hello"), 0644)
	content, err = ReadFileOrEmpty(p)
	if err != nil {
		t.Fatalf("expected nil err")
	}
	if content != "hello" {
		t.Errorf("expected hello")
	}
}

func TestEdit_RawTOML(t *testing.T) {
	tempDir := t.TempDir()
	p := filepath.Join(tempDir, "test.toml")

	// LoadNotExist
	m, err := LoadRawTOML(p)
	if err != nil {
		t.Fatalf("expected nil err")
	}
	if len(m) != 0 {
		t.Errorf("expected empty map")
	}

	// Save
	m["tools"] = map[string]interface{}{"node": "20.0.0"}
	err = SaveRawTOML(p, m)
	if err != nil {
		t.Fatalf("expected nil err")
	}

	// Load Exist
	m2, err := LoadRawTOML(p)
	if err != nil {
		t.Fatalf("expected nil err")
	}

	toolsMap := m2["tools"].(map[string]interface{})
	if toolsMap["node"] != "20.0.0" {
		t.Errorf("expected node 20.0.0")
	}
}

func TestFmt_FormatTOML(t *testing.T) {
	tomlStr := `
[tools]
node = "20"
[settings]
cache_dir = "hello"
[env]
A = "1"
`
	out, err := FormatTOML(tomlStr)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}
	if len(out) == 0 {
		t.Errorf("empty out")
	}
}
