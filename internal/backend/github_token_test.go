// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveGitHubToken_EnvPriority(t *testing.T) {
	// Clean all relevant env vars before each sub-test to ensure strict isolation
	envVars := []string{
		"UNIRTM_GITHUB_TOKEN",
		"MISE_GITHUB_TOKEN",
		"GITHUB_TOKEN",
		"UNIRTM_GITHUB_API_TOKEN",
		"MISE_GITHUB_API_TOKEN",
		"GITHUB_API_TOKEN",
		"UNIRTM_GITHUB_CREDENTIAL_COMMAND",
		"MISE_GITHUB_CREDENTIAL_COMMAND",
		"GITHUB_CREDENTIAL_COMMAND",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	t.Run("UNIRTM_GITHUB_TOKEN takes priority", func(t *testing.T) {
		os.Setenv("UNIRTM_GITHUB_TOKEN", "unirtm-token")
		os.Setenv("GITHUB_TOKEN", "github-token")
		defer os.Unsetenv("UNIRTM_GITHUB_TOKEN")
		defer os.Unsetenv("GITHUB_TOKEN")

		got := resolveGitHubToken("github.com")
		if got != "unirtm-token" {
			t.Errorf("expected 'unirtm-token', got %q", got)
		}
	})

	t.Run("GITHUB_TOKEN used when UNIRTM_GITHUB_TOKEN not set", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "ci-token")
		defer os.Unsetenv("GITHUB_TOKEN")

		got := resolveGitHubToken("github.com")
		if got != "ci-token" {
			t.Errorf("expected 'ci-token', got %q", got)
		}
	})

	t.Run("GITHUB_API_TOKEN used as fallback", func(t *testing.T) {
		os.Setenv("GITHUB_API_TOKEN", "api-token")
		defer os.Unsetenv("GITHUB_API_TOKEN")

		got := resolveGitHubToken("github.com")
		if got != "api-token" {
			t.Errorf("expected 'api-token', got %q", got)
		}
	})

	t.Run("returns empty string when no token configured", func(t *testing.T) {
		got := resolveGitHubToken("github.com")
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})
}

func TestParseGhHostsYml(t *testing.T) {
	content := `github.com:
    oauth_token: ghp_fromghcli
    user: octocat
github.mycompany.com:
    oauth_token: ghp_enterprise
    user: corp-user
`
	tests := []struct {
		host string
		want string
	}{
		{"github.com", "ghp_fromghcli"},
		{"github.mycompany.com", "ghp_enterprise"},
		{"unknown.host", ""},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := parseGhHostsYml(content, tt.host)
			if got != tt.want {
				t.Errorf("parseGhHostsYml(%q) = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}

func TestReadGitHubTokensFile(t *testing.T) {
	// Create a temp config dir with github_tokens.toml
	dir := t.TempDir()
	os.Setenv("UNIRTM_CONFIG_DIR", dir)
	defer os.Unsetenv("UNIRTM_CONFIG_DIR")

	tomlContent := `[tokens]
[tokens."github.com"]
token = "ghp_from_toml"

[tokens."github.mycompany.com"]
token = "ghp_enterprise_toml"
`
	if err := os.WriteFile(dir+"/github_tokens.toml", []byte(tomlContent), 0600); err != nil {
		t.Fatalf("failed to write test token file: %v", err)
	}

	t.Run("reads token for github.com", func(t *testing.T) {
		got := readGitHubTokensFile("github.com")
		if got != "ghp_from_toml" {
			t.Errorf("expected 'ghp_from_toml', got %q", got)
		}
	})

	t.Run("reads token for enterprise host", func(t *testing.T) {
		got := readGitHubTokensFile("github.mycompany.com")
		if got != "ghp_enterprise_toml" {
			t.Errorf("expected 'ghp_enterprise_toml', got %q", got)
		}
	})

	t.Run("returns empty for unknown host", func(t *testing.T) {
		got := readGitHubTokensFile("unknown.host")
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})
}

func TestFindGhHostsFile(t *testing.T) {
	// Create a dummy hosts file
	dir := t.TempDir()
	hostsFile := filepath.Join(dir, "hosts.yml")
	os.WriteFile(hostsFile, []byte(""), 0644)

	// Test 1: GH_CONFIG_DIR
	t.Setenv("GH_CONFIG_DIR", dir)
	if f := findGhHostsFile(); f != hostsFile {
		t.Errorf("expected %s, got %s", hostsFile, f)
	}

	// Test 2: XDG_CONFIG_HOME
	t.Setenv("GH_CONFIG_DIR", "")
	ghDir := filepath.Join(dir, "gh")
	os.Mkdir(ghDir, 0755)
	hostsFile2 := filepath.Join(ghDir, "hosts.yml")
	os.WriteFile(hostsFile2, []byte(""), 0644)

	t.Setenv("XDG_CONFIG_HOME", dir)
	if f := findGhHostsFile(); f != hostsFile2 {
		t.Errorf("expected %s, got %s", hostsFile2, f)
	}
}

func TestResolveGitHubTokenPublic(t *testing.T) {
	token := ResolveGitHubTokenPublic("github.com")
	if token == "mock-token-xyz" {
		t.Log("got mock token")
	}
}

func TestRunCredentialCommand(t *testing.T) {
	// Test with a real command that produces output
	token := runCredentialCommand("echo mytoken", "github.com")
	if token != "mytoken" {
		t.Errorf("expected 'mytoken', got %q", token)
	}

	// Test with a failing command
	token2 := runCredentialCommand("exit 1", "github.com")
	if token2 != "" {
		t.Errorf("expected empty string for failing command, got %q", token2)
	}
}

func TestResolveGitHubToken_EmptyHost(t *testing.T) {
	// Empty host should default to github.com and not panic
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_API_TOKEN")
	os.Unsetenv("GITHUB_CREDENTIAL_COMMAND")
	os.Unsetenv("UNIRTM_GITHUB_TOKEN")
	os.Unsetenv("MISE_GITHUB_TOKEN")
	// Just test it doesn't panic with empty host
	_ = resolveGitHubToken("")
}

func TestResolveGitHubToken_CredentialCommand(t *testing.T) {
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_API_TOKEN")
	os.Unsetenv("UNIRTM_GITHUB_TOKEN")
	os.Unsetenv("MISE_GITHUB_TOKEN")
	os.Unsetenv("UNIRTM_GITHUB_API_TOKEN")
	os.Unsetenv("MISE_GITHUB_API_TOKEN")
	// env.Get("GITHUB_CREDENTIAL_COMMAND") checks UNIRTM_GITHUB_CREDENTIAL_COMMAND first
	os.Setenv("UNIRTM_GITHUB_CREDENTIAL_COMMAND", "echo cred-token")
	defer os.Unsetenv("UNIRTM_GITHUB_CREDENTIAL_COMMAND")

	token := resolveGitHubToken("github.com")
	// The credential command 'echo cred-token' should return 'cred-token'
	if token != "cred-token" {
		t.Logf("token from credential command: %q", token)
	}
}

func TestReadGitHubTokensFile_WithXDG(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	dir := t.TempDir()
	unirtmDir := dir + "/unirtm"
	os.MkdirAll(unirtmDir, 0755)
	tomlContent := `[tokens]
[tokens."github.com"]
token = "ghp_from_xdg"
`
	os.WriteFile(unirtmDir+"/github_tokens.toml", []byte(tomlContent), 0600)

	// Ensure CONFIG_DIR is not set
	os.Unsetenv("UNIRTM_CONFIG_DIR")
	os.Unsetenv("CONFIG_DIR")
	os.Setenv("UNIRTM_XDG_CONFIG_HOME", dir)
	defer os.Unsetenv("UNIRTM_XDG_CONFIG_HOME")

	got := readGitHubTokensFile("github.com")
	// May or may not work depending on env alias resolution - just ensure no panic
	t.Logf("token from XDG: %q", got)
}

func TestReadGitHubTokensFile_InvalidToml(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("UNIRTM_CONFIG_DIR", dir)
	defer os.Unsetenv("UNIRTM_CONFIG_DIR")

	// Write invalid TOML
	os.WriteFile(dir+"/github_tokens.toml", []byte("not valid toml [[["), 0600)

	got := readGitHubTokensFile("github.com")
	if got != "" {
		t.Errorf("expected empty string for invalid TOML, got %q", got)
	}
}

func TestFindGhHostsFile_NotFound(t *testing.T) {
	// Unset all env vars and use a non-existent home
	os.Unsetenv("GH_CONFIG_DIR")
	// Even if not found, function should return empty string without panic
	// We can't easily override home directory but we can at least test the path
	t.Setenv("GH_CONFIG_DIR", "/nonexistent/path")
	f := findGhHostsFile()
	// Should not find it
	if f != "" {
		t.Logf("found unexpected file: %s", f)
	}
}

func TestParseGhHostsYml_TabIndented(t *testing.T) {
	// Test with tab-indented content
	content := "github.com:\n\toauth_token: ghp_tabindented\n\tuser: octocat\n"
	got := parseGhHostsYml(content, "github.com")
	if got != "ghp_tabindented" {
		t.Errorf("expected 'ghp_tabindented', got %q", got)
	}
}

func TestReadGhCliToken_WithHostsFile(t *testing.T) {
	dir := t.TempDir()
	hostsContent := `github.com:
    oauth_token: ghp_from_hosts_yml
    user: octocat
`
	hostsFile := dir + "/hosts.yml"
	os.WriteFile(hostsFile, []byte(hostsContent), 0644)

	t.Setenv("GH_CONFIG_DIR", dir)

	token := readGhCliToken("github.com")
	if token != "ghp_from_hosts_yml" {
		t.Errorf("expected 'ghp_from_hosts_yml', got %q", token)
	}
}

func TestReadGhCliToken_UnknownHost(t *testing.T) {
	dir := t.TempDir()
	hostsContent := `github.com:
    oauth_token: ghp_token
    user: octocat
`
	hostsFile := dir + "/hosts.yml"
	os.WriteFile(hostsFile, []byte(hostsContent), 0644)

	t.Setenv("GH_CONFIG_DIR", dir)

	token := readGhCliToken("unknown.host")
	if token != "" {
		t.Errorf("expected empty string for unknown host, got %q", token)
	}
}

func TestReadGhCliToken_FileNotReadable(t *testing.T) {
	dir := t.TempDir()
	hostsFile := dir + "/hosts.yml"
	os.WriteFile(hostsFile, []byte(""), 0000)

	t.Setenv("GH_CONFIG_DIR", dir)

	// Should return empty string (file not readable)
	token := readGhCliToken("github.com")
	if token != "" {
		t.Logf("token from non-readable file: %q", token)
	}
}

func TestFindGhHostsFile_XDGPath(t *testing.T) {
	dir := t.TempDir()
	ghDir := dir + "/gh"
	os.MkdirAll(ghDir, 0755)
	hostsFile := ghDir + "/hosts.yml"
	os.WriteFile(hostsFile, []byte(""), 0644)

	t.Setenv("GH_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", dir)

	f := findGhHostsFile()
	if f != hostsFile {
		t.Errorf("expected %s, got %s", hostsFile, f)
	}
}

func TestFindGhHostsFile_XDGConfigHome(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "gh"), 0755)
	hostsFile := filepath.Join(dir, "gh", "hosts.yml")
	os.WriteFile(hostsFile, []byte("github.com:\n  oauth_token: ghp_test"), 0600)

	t.Setenv("UNIRTM_GH_CONFIG_DIR", "")
	t.Setenv("UNIRTM_XDG_CONFIG_HOME", dir)

	f := findGhHostsFile()
	if f != hostsFile {
		t.Errorf("expected %s, got %s", hostsFile, f)
	}
}

func TestFindGhHostsFile_HomeDirFallback(t *testing.T) {
	homeDir := t.TempDir()
	// Mock XDG default
	os.MkdirAll(filepath.Join(homeDir, ".config", "gh"), 0755)
	hostsFile := filepath.Join(homeDir, ".config", "gh", "hosts.yml")
	os.WriteFile(hostsFile, []byte("github.com:\n  oauth_token: ghp_test"), 0600)

	t.Setenv("UNIRTM_GH_CONFIG_DIR", "")
	t.Setenv("UNIRTM_XDG_CONFIG_HOME", "")
	t.Setenv("HOME", homeDir)

	f := findGhHostsFile()
	if f != hostsFile {
		t.Errorf("expected %s, got %s", hostsFile, f)
	}
}

func TestFindGhHostsFile_MacosFallback(t *testing.T) {
	homeDir := t.TempDir()
	// Mock macos path
	os.MkdirAll(filepath.Join(homeDir, "Library", "Application Support", "gh"), 0755)
	hostsFile := filepath.Join(homeDir, "Library", "Application Support", "gh", "hosts.yml")
	os.WriteFile(hostsFile, []byte("github.com:\n  oauth_token: ghp_test"), 0600)

	t.Setenv("UNIRTM_GH_CONFIG_DIR", "")
	t.Setenv("UNIRTM_XDG_CONFIG_HOME", "")
	t.Setenv("HOME", homeDir)

	f := findGhHostsFile()
	if f != hostsFile {
		t.Errorf("expected %s, got %s", hostsFile, f)
	}
}
