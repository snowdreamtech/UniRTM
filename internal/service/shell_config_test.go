// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

type mockFormatter struct{}

func (m *mockFormatter) Info(message string, fields ...map[string]interface{})    {}
func (m *mockFormatter) Success(message string, fields ...map[string]interface{}) {}
func (m *mockFormatter) Warning(message string, fields ...map[string]interface{}) {}
func (m *mockFormatter) Error(message string, fields ...map[string]interface{})   {}
func (m *mockFormatter) Data(data interface{})                                    {}
func (m *mockFormatter) Table(headers []string, rows [][]string)                  {}
func (m *mockFormatter) SetWriter(w io.Writer)                                    {}

func TestShellConfigManager_GetConfigPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	sm := NewShellConfigManager(&mockFormatter{}, false)

	tests := []struct {
		shell    ShellType
		expected string
	}{
		{ShellZsh, filepath.Join(home, ".zshrc")},
		{ShellBash, filepath.Join(home, ".bashrc")},
		{ShellFish, filepath.Join(home, ".config/fish/config.fish")},
		{ShellPowerShell, filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")},
	}

	for _, tt := range tests {
		t.Run(string(tt.shell), func(t *testing.T) {
			path, err := sm.GetConfigPath(tt.shell)
			if err != nil {
				t.Errorf("GetConfigPath failed: %v", err)
			}
			if path != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, path)
			}
		})
	}
}

func TestShellConfigManager_GetConfigPath_Unsupported(t *testing.T) {
	sm := NewShellConfigManager(&mockFormatter{}, false)
	_, err := sm.GetConfigPath(ShellType("unknown"))
	if err == nil {
		t.Error("expected error for unsupported shell")
	}
}

func TestShellConfigManager_GetConfigPath_PowerShellProfile(t *testing.T) {
	t.Setenv("PROFILE", "/custom/profile.ps1")
	sm := NewShellConfigManager(&mockFormatter{}, false)
	path, err := sm.GetConfigPath(ShellPowerShell)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/custom/profile.ps1" {
		t.Errorf("expected /custom/profile.ps1, got %s", path)
	}
}

func TestShellConfigManager_InjectAndRemove(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	if env.RuntimeGOOS == "windows" {
		t.Setenv("USERPROFILE", tmpDir)
	}

	sm := NewShellConfigManager(&mockFormatter{}, false)

	// Inject
	err := sm.Inject(ShellBash, "test", `eval "$(unirtm test activate bash)"`)
	if err != nil {
		t.Fatalf("Inject failed: %v", err)
	}

	configPath := filepath.Join(tmpDir, ".bashrc")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}
	contentStr := string(content)

	if !strings.Contains(contentStr, "unirtm test activation") {
		t.Errorf("expected marker in content")
	}
	if !strings.Contains(contentStr, `eval "$(unirtm test activate bash)"`) {
		t.Errorf("expected content in config")
	}

	// Inject again (should not duplicate)
	err = sm.Inject(ShellBash, "test", `eval "$(unirtm test activate bash)"`)
	if err != nil {
		t.Fatalf("Inject twice failed: %v", err)
	}
	content, _ = os.ReadFile(configPath)
	contentStr = string(content)
	count := strings.Count(contentStr, "unirtm test activation")
	if count != 1 {
		t.Errorf("expected 1 marker, got %d", count)
	}

	// Inject with different content (update)
	err = sm.Inject(ShellBash, "test", `eval "$(unirtm test activate bash --updated)"`)
	if err != nil {
		t.Fatalf("Inject update failed: %v", err)
	}
	content, _ = os.ReadFile(configPath)
	contentStr = string(content)
	if !strings.Contains(contentStr, "--updated") {
		t.Errorf("expected updated content")
	}
	if strings.Contains(contentStr, `eval "$(unirtm test activate bash)"`+"\n") {
		t.Errorf("expected old content to be removed")
	}

	// Remove
	err = sm.Remove(ShellBash, "test")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	content, _ = os.ReadFile(configPath)
	contentStr = string(content)
	if strings.Contains(contentStr, "unirtm test activation") {
		t.Errorf("expected marker to be removed")
	}
}

func TestShellConfigManager_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	if env.RuntimeGOOS == "windows" {
		t.Setenv("USERPROFILE", tmpDir)
	}

	sm := NewShellConfigManager(&mockFormatter{}, true)

	err := sm.Inject(ShellZsh, "test", `eval "$(unirtm test activate zsh)"`)
	if err != nil {
		t.Fatalf("Inject dry-run failed: %v", err)
	}

	configPath := filepath.Join(tmpDir, ".zshrc")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Errorf("expected config file to not be created in dry run")
	}

	// Create file for remove dry-run test
	os.WriteFile(configPath, []byte("# unirtm test activation\neval\n"), 0644)

	err = sm.Remove(ShellZsh, "test")
	if err != nil {
		t.Fatalf("Remove dry-run failed: %v", err)
	}

	content, _ := os.ReadFile(configPath)
	if !strings.Contains(string(content), "eval") {
		t.Errorf("expected config file to not be modified in dry run")
	}
}

func TestShellConfigManager_Remove_Fallback(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	if env.RuntimeGOOS == "windows" {
		t.Setenv("USERPROFILE", tmpDir)
	}

	sm := NewShellConfigManager(&mockFormatter{}, false)
	configPath := filepath.Join(tmpDir, ".bashrc")

	// Create file with old pattern
	content := []byte("# unirtm test\neval \"$(unirtm test activate bash)\"\n")
	os.WriteFile(configPath, content, 0644)

	err := sm.Remove(ShellBash, "test")
	if err != nil {
		t.Fatalf("Remove fallback failed: %v", err)
	}

	readContent, _ := os.ReadFile(configPath)
	if strings.Contains(string(readContent), "unirtm test") {
		t.Errorf("expected old pattern to be removed")
	}
}

func TestShellConfigManager_Inject_DryRun_Update(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	if env.RuntimeGOOS == "windows" {
		t.Setenv("USERPROFILE", tmpDir)
	}
	sm := NewShellConfigManager(&mockFormatter{}, true)
	configPath := filepath.Join(tmpDir, ".bashrc")

	// Create file with block
	content := []byte("# unirtm test activation\nold_content\n")
	os.WriteFile(configPath, content, 0644)

	err := sm.Inject(ShellBash, "test", "new_content")
	if err != nil {
		t.Fatalf("Inject dry-run failed: %v", err)
	}

	readContent, _ := os.ReadFile(configPath)
	if !strings.Contains(string(readContent), "old_content") {
		t.Errorf("expected dry-run to not update file")
	}
}

func (m *mockFormatter) Infof(format string, a ...interface{})    {}
func (m *mockFormatter) Successf(format string, a ...interface{}) {}
func (m *mockFormatter) Warningf(format string, a ...interface{}) {}
func (m *mockFormatter) Errorf(format string, a ...interface{})   {}
