// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"strings"
	"testing"
)

func TestUpsertEnvVar(t *testing.T) {
	// 1. Empty content
	content := ""
	res := UpsertEnvVar(content, "FOO", "bar")
	if !strings.Contains(res, "[env]") || !strings.Contains(res, "FOO = \"bar\"") {
		t.Errorf("UpsertEnvVar failed on empty content: %s", res)
	}

	// 2. Existing [env] but no FOO
	content = "[env]\nBAZ = \"qux\"\n[tools]\nnode = \"18\""
	res = UpsertEnvVar(content, "FOO", "bar")
	if !strings.Contains(res, "FOO = \"bar\"") {
		t.Errorf("UpsertEnvVar failed to add to existing [env]: %s", res)
	}
	// FOO should be before [tools]
	if strings.Index(res, "FOO") > strings.Index(res, "[tools]") {
		t.Errorf("UpsertEnvVar added key after next section: %s", res)
	}

	// 3. Existing [env] with FOO
	content = "[env]\nFOO = \"old\"\n[tools]\nnode = \"18\""
	res = UpsertEnvVar(content, "FOO", "new")
	if !strings.Contains(res, "FOO = \"new\"") || strings.Contains(res, "FOO = \"old\"") {
		t.Errorf("UpsertEnvVar failed to update existing key: %s", res)
	}

	// 4. Lowercase or mixed-case key input is normalized to uppercase
	content = ""
	res = UpsertEnvVar(content, "github_token", "ghp_123")
	if !strings.Contains(res, "GITHUB_TOKEN = \"ghp_123\"") || strings.Contains(res, "github_token") {
		t.Errorf("UpsertEnvVar failed to normalize lowercase key to uppercase: %s", res)
	}

	// 5. Existing lowercase key in [env] is replaced and migrated to uppercase
	content = "[env]\ngithub_token = \"old_token\"\n"
	res = UpsertEnvVar(content, "GITHUB_TOKEN", "new_token")
	if !strings.Contains(res, "GITHUB_TOKEN = \"new_token\"") || strings.Contains(res, "github_token") {
		t.Errorf("UpsertEnvVar failed to migrate existing lowercase key to uppercase: %s", res)
	}

	// 6. Deduplicates multiple casing variants if they existed
	content = "[env]\ngithub_token = \"tok1\"\nGITHUB_TOKEN = \"tok2\"\n"
	res = UpsertEnvVar(content, "GitHub_Token", "tok3")
	if strings.Count(res, "GITHUB_TOKEN") != 1 || strings.Contains(res, "github_token") || !strings.Contains(res, "tok3") {
		t.Errorf("UpsertEnvVar failed to deduplicate casing variants: %s", res)
	}
}

func TestUnsetEnvVar(t *testing.T) {
	// 1. Key exists
	content := "[env]\nFOO = \"bar\"\nBAZ = \"qux\"\n"
	res, changed := UnsetEnvVar(content, "FOO")
	if !changed {
		t.Errorf("UnsetEnvVar should return changed=true")
	}
	if strings.Contains(res, "FOO") {
		t.Errorf("UnsetEnvVar failed to remove key: %s", res)
	}

	// 2. Key does not exist
	res, changed = UnsetEnvVar(content, "NONEXISTENT")
	if changed {
		t.Errorf("UnsetEnvVar should return changed=false for nonexistent key")
	}

	// 3. Case-insensitive removal
	content = "[env]\nGITHUB_TOKEN = \"ghp_xxx\"\n"
	res, changed = UnsetEnvVar(content, "github_token")
	if !changed || strings.Contains(res, "GITHUB_TOKEN") {
		t.Errorf("UnsetEnvVar failed to remove case-insensitively: %s", res)
	}
}

func TestUpsertToolVersion(t *testing.T) {
	// 1. Empty content
	content := ""
	res := UpsertToolVersion(content, "node", "18")
	if !strings.Contains(res, "[tools]") || !strings.Contains(res, "node = \"18\"") {
		t.Errorf("UpsertToolVersion failed on empty content: %s", res)
	}

	// 2. Existing [tools]
	content = "[tools]\npython = \"3.10\"\n[env]\nFOO = \"bar\""
	res = UpsertToolVersion(content, "node", "18")
	if !strings.Contains(res, "node = \"18\"") {
		t.Errorf("UpsertToolVersion failed to add to existing [tools]: %s", res)
	}
	if strings.Index(res, "node = \"18\"") > strings.Index(res, "[env]") {
		t.Errorf("UpsertToolVersion added key after next section: %s", res)
	}

	// 3. Update existing tool
	content = "[tools]\nnode = \"16\"\npython = \"3.10\""
	res = UpsertToolVersion(content, "node", "18")
	if !strings.Contains(res, "node = \"18\"") || strings.Contains(res, "node = \"16\"") {
		t.Errorf("UpsertToolVersion failed to update existing tool: %s", res)
	}
}

func TestLoadRawTOML(t *testing.T) {
	dir := t.TempDir()

	// 1. File does not exist
	m, err := LoadRawTOML(dir + "/missing.toml")
	if err != nil {
		t.Errorf("Expected nil error for missing file, got %v", err)
	}
	if len(m) != 0 {
		t.Errorf("Expected empty map for missing file, got %v", m)
	}

	// 2. Invalid TOML
	invalidPath := dir + "/invalid.toml"
	os.WriteFile(invalidPath, []byte("invalid \n toml"), 0644)
	_, err = LoadRawTOML(invalidPath)
	if err == nil {
		t.Errorf("Expected error for invalid TOML, got nil")
	}
}

func TestSaveRawTOML(t *testing.T) {
	dir := t.TempDir()

	// 1. Successful save
	path := dir + "/config.toml"
	m := map[string]interface{}{
		"env": map[string]interface{}{
			"FOO": "bar",
		},
	}
	err := SaveRawTOML(path, m)
	if err != nil {
		t.Errorf("Expected nil error for successful save, got %v", err)
	}

	// 2. Unencodable map
	err = SaveRawTOML(path, map[string]interface{}{
		"bad": make(chan int),
	})
	if err == nil {
		t.Errorf("Expected error for unencodable map, got nil")
	}

	// 3. Write error (directory as file)
	err = SaveRawTOML(dir, m)
	if err == nil {
		t.Errorf("Expected error for write error, got nil")
	}
}

func TestValidateTOML(t *testing.T) {
	// Valid TOML
	validCases := []string{
		"",
		"# Comment only\n",
		`[env]` + "\n" + `FOO = "bar"`,
		`[tools]` + "\n" + `node = "20.0.0"`,
	}
	for _, tc := range validCases {
		if err := ValidateTOML(tc); err != nil {
			t.Errorf("ValidateTOML expected nil for valid TOML, got %v for %q", err, tc)
		}
	}

	// Invalid TOML
	invalidCases := []string{
		`key = [unclosed array`,
		`invalid \n toml`,
		`FOO = "unclosed string`,
		`[broken table`,
	}
	for _, tc := range invalidCases {
		if err := ValidateTOML(tc); err == nil {
			t.Errorf("ValidateTOML expected error for invalid TOML %q, got nil", tc)
		}
	}
}

func TestWriteTOMLFile(t *testing.T) {
	dir := t.TempDir()

	// 1. Valid TOML writes successfully and creates parent directory
	filePath := dir + "/nested/config.toml"
	validContent := []byte("[env]\nFOO = \"bar\"\n")
	if err := WriteTOMLFile(filePath, validContent); err != nil {
		t.Fatalf("WriteTOMLFile failed on valid TOML: %v", err)
	}

	read, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}
	if string(read) != string(validContent) {
		t.Errorf("Content mismatch: expected %q, got %q", string(validContent), string(read))
	}

	// 2. Invalid TOML is rejected before touching disk (new file should not be created)
	badFilePath := dir + "/invalid.toml"
	invalidContent := []byte("key = [broken")
	if err := WriteTOMLFile(badFilePath, invalidContent); err == nil {
		t.Fatalf("WriteTOMLFile expected validation error on invalid TOML, got nil")
	}
	if _, err := os.Stat(badFilePath); !os.IsNotExist(err) {
		t.Errorf("Invalid TOML file should not have been written to disk")
	}

	// 3. Invalid TOML does not overwrite existing valid file
	if err := WriteTOMLFile(filePath, invalidContent); err == nil {
		t.Fatalf("WriteTOMLFile expected validation error, got nil")
	}
	// Verify original content was preserved
	readAfterBad, err := os.ReadFile(filePath)
	if err != nil || string(readAfterBad) != string(validContent) {
		t.Errorf("Existing file was corrupted or modified by failed WriteTOMLFile")
	}
}

func TestUpsertAndRemoveTask(t *testing.T) {
	content := "[tasks.build]\nrun = \"echo building\"\n\n[tasks.test]\nrun = \"echo testing\"\n"

	// 1. Upsert new task
	updated := UpsertTask(content, "lint", "Check code style", "npm run lint", []string{"build"})
	if !strings.Contains(updated, "[tasks.lint]") || !strings.Contains(updated, "description = \"Check code style\"") {
		t.Fatalf("UpsertTask failed to add new task: %s", updated)
	}

	// 2. Update existing task
	updated2 := UpsertTask(updated, "build", "Updated build", "vite build", nil)
	if !strings.Contains(updated2, "run = \"vite build\"") || strings.Contains(updated2, "echo building") {
		t.Fatalf("UpsertTask failed to update existing task: %s", updated2)
	}

	// 3. Remove task
	removed, ok := RemoveTask(updated2, "lint")
	if !ok || strings.Contains(removed, "[tasks.lint]") {
		t.Fatalf("RemoveTask failed to remove task: %s", removed)
	}
	if !strings.Contains(removed, "[tasks.build]") || !strings.Contains(removed, "[tasks.test]") {
		t.Fatalf("RemoveTask accidentally deleted other tasks: %s", removed)
	}
}
