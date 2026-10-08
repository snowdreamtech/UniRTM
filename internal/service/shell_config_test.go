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

func TestShellConfigManager_Remove_MiseWithoutUniRTMKeyword(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	sm := NewShellConfigManager(&mockFormatter{}, false)
	configPath := filepath.Join(tmpDir, ".bashrc")

	content := []byte("# unirtm mise activation\neval \"$(/usr/local/bin/mise activate bash)\"\nexport FOO=bar\n")
	os.WriteFile(configPath, content, 0644)

	err := sm.Remove(ShellBash, "mise")
	if err != nil {
		t.Fatalf("Remove mise failed: %v", err)
	}

	readContent, _ := os.ReadFile(configPath)
	str := string(readContent)
	if strings.Contains(str, "mise activate") || strings.Contains(str, "unirtm mise activation") {
		t.Errorf("expected mise activation to be completely removed, got: %s", str)
	}
	if !strings.Contains(str, "export FOO=bar") {
		t.Errorf("expected other configuration to remain untouched, got: %s", str)
	}
}

func TestShellConfigManager_Remove_WithoutComment(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	sm := NewShellConfigManager(&mockFormatter{}, false)
	configPath := filepath.Join(tmpDir, ".bashrc")

	content := []byte("export PATH=/usr/bin:$PATH\neval \"$(C:/Users/test/unirtm.exe activate bash)\"\nalias ll='ls -l'\n")
	os.WriteFile(configPath, content, 0644)

	err := sm.Remove(ShellBash, "unirtm")
	if err != nil {
		t.Fatalf("Remove unirtm without comment failed: %v", err)
	}

	readContent, _ := os.ReadFile(configPath)
	str := string(readContent)
	if strings.Contains(str, "unirtm.exe activate") {
		t.Errorf("expected unirtm command to be removed even without comment, got: %s", str)
	}
	if !strings.Contains(str, "export PATH") || !strings.Contains(str, "alias ll") {
		t.Errorf("expected unrelated lines to be preserved, got: %s", str)
	}
}

func TestShellConfigManager_Remove_CRLF(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	sm := NewShellConfigManager(&mockFormatter{}, false)
	configPath := filepath.Join(tmpDir, ".bashrc")

	content := []byte("# unirtm activation\r\n& \"C:\\unirtm.exe\" activate powershell | Out-String | Invoke-Expression\r\nWrite-Host 'hello'\r\n")
	os.WriteFile(configPath, content, 0644)

	err := sm.Remove(ShellBash, "unirtm")
	if err != nil {
		t.Fatalf("Remove CRLF failed: %v", err)
	}

	readContent, _ := os.ReadFile(configPath)
	str := string(readContent)
	if strings.Contains(str, "activate powershell") {
		t.Errorf("expected powershell activation to be removed, got: %s", str)
	}
	if !strings.Contains(str, "Write-Host 'hello'") {
		t.Errorf("expected other lines to remain, got: %s", str)
	}
	if !strings.Contains(str, "\r\n") {
		t.Errorf("expected CRLF line endings to be preserved, got: %q", str)
	}
}

func TestShellConfigManager_MultiLineBlock(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	sm := NewShellConfigManager(&mockFormatter{}, false)
	configPath := filepath.Join(tmpDir, ".bashrc")

	// 1. Initial content
	initial := "alias ll='ls -l'\n"
	err := os.WriteFile(configPath, []byte(initial), 0644)
	if err != nil {
		t.Fatal(err)
	}

	block1 := `# Ensure user private bin directory is in PATH
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    export PATH="$HOME/.local/bin:$PATH"
fi
if [[ -d "$HOME/bin" && ":$PATH:" != *":$HOME/bin:"* ]]; then
    export PATH="$HOME/bin:$PATH"
fi

# UniRTM activation
if command -v unirtm >/dev/null 2>&1; then
    eval "$(unirtm activate bash)"
fi`

	// Inject block
	err = sm.Inject(ShellBash, "unirtm", block1)
	if err != nil {
		t.Fatalf("Inject failed: %v", err)
	}

	content, _ := os.ReadFile(configPath)
	str := string(content)
	if !strings.Contains(str, "# unirtm unirtm activation") {
		t.Errorf("expected marker in content")
	}
	if !strings.Contains(str, "Ensure user private bin directory is in PATH") {
		t.Errorf("expected PATH comment in content")
	}
	if !strings.Contains(str, "alias ll='ls -l'") {
		t.Errorf("expected existing alias preserved")
	}

	// 2. Update with new block (e.g. --shims)
	block2 := `# Ensure user private bin directory is in PATH
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    export PATH="$HOME/.local/bin:$PATH"
fi
if [[ -d "$HOME/bin" && ":$PATH:" != *":$HOME/bin:"* ]]; then
    export PATH="$HOME/bin:$PATH"
fi

# UniRTM activation
if command -v unirtm >/dev/null 2>&1; then
    eval "$(unirtm activate --shims bash)"
fi`

	err = sm.Inject(ShellBash, "unirtm", block2)
	if err != nil {
		t.Fatalf("Inject update failed: %v", err)
	}

	content, _ = os.ReadFile(configPath)
	str = string(content)
	if strings.Count(str, "unirtm unirtm activation") != 1 {
		t.Errorf("expected exactly 1 marker, got %d", strings.Count(str, "unirtm unirtm activation"))
	}
	if !strings.Contains(str, "--shims") {
		t.Errorf("expected updated --shims content")
	}
	if strings.Contains(str, `eval "$(unirtm activate bash)"`) {
		t.Errorf("expected old activation without --shims to be replaced")
	}

	// 3. Remove block
	err = sm.Remove(ShellBash, "unirtm")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	content, _ = os.ReadFile(configPath)
	str = string(content)
	if strings.Contains(str, "unirtm") {
		t.Errorf("expected all unirtm logic to be removed, got: %s", str)
	}
	if strings.Contains(str, "Ensure user private bin directory") {
		t.Errorf("expected PATH comment to be removed")
	}
	if strings.Contains(str, "export PATH") {
		t.Errorf("expected PATH export to be removed")
	}
	if !strings.Contains(str, "alias ll='ls -l'") {
		t.Errorf("expected alias ll to remain untouched")
	}
}

func (m *mockFormatter) Infof(format string, a ...interface{}) {}

func (m *mockFormatter) Successf(format string, a ...interface{}) {}
func (m *mockFormatter) Warningf(format string, a ...interface{}) {}
func (m *mockFormatter) Errorf(format string, a ...interface{})   {}
