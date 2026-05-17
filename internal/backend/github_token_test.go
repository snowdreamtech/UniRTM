// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"os"
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
		host  string
		want  string
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
