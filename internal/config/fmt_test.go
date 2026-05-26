// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatTOML(t *testing.T) {
	input := `
[tools]
node = "18"

[env]
FOO = "bar"

[tasks]
build = "go build"

[min_version]
version = "1.0.0"
`

	expected := `
[min_version]
version = "1.0.0"


[env]
FOO = "bar"
[tools]
node = "18"

[tasks]
build = "go build"
`

	result, err := FormatTOML(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != expected {
		t.Errorf("FormatTOML mismatch.\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestFormatTOML_MultilineStrings(t *testing.T) {
	input := `
[tasks]
build = """
  go build
  go test
"""

[env]
FOO = "bar"
`

	expected := `

[env]
FOO = "bar"
[tasks]
build = """
  go build
  go test
"""
`
	result, err := FormatTOML(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("FormatTOML multiline mismatch.\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestFormatFile(t *testing.T) {
	tmpDir := t.TempDir()
	tomlFile := filepath.Join(tmpDir, "test.toml")
	
	input := `
[tools]
node = "18"

[env]
FOO = "bar"
`
	err := os.WriteFile(tomlFile, []byte(input), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	modified, err := FormatFile(tomlFile, false)
	if err != nil {
		t.Fatalf("FormatFile error: %v", err)
	}
	if !modified {
		t.Errorf("expected FormatFile to return modified=true")
	}

	content, _ := os.ReadFile(tomlFile)
	if !strings.Contains(string(content), "[env]\nFOO = \"bar\"\n[tools]\nnode = \"18\"") {
		t.Errorf("FormatFile didn't format correctly: %s", string(content))
	}

	// formatting again should return false (no modification)
	modified, err = FormatFile(tomlFile, false)
	if err != nil {
		t.Fatalf("FormatFile 2 error: %v", err)
	}
	if modified {
		t.Errorf("expected FormatFile 2 to return modified=false")
	}
}
