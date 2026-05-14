// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// githubTokensFile represents the structure of github_tokens.toml.
type githubTokensFile struct {
	Tokens map[string]struct {
		Token string `toml:"token"`
	} `toml:"tokens"`
}

// ResolveGitHubTokenPublic is the exported wrapper for use outside the backend package.
func ResolveGitHubTokenPublic(host string) string {
	return resolveGitHubToken(host)
}

// resolveGitHubToken resolves the best available GitHub token for a given host.
// Priority (mirrors mise behavior):
//  1. UNIRTM_GITHUB_TOKEN (UniRTM-specific override)
//  2. GITHUB_TOKEN        (standard CI env var, e.g. GitHub Actions)
//  3. GITHUB_API_TOKEN    (legacy alternative)
//  4. credential_command  (via UNIRTM_GITHUB_CREDENTIAL_COMMAND env var)
//  5. github_tokens.toml  (~/.config/unirtm/github_tokens.toml)
//  6. gh CLI hosts.yml
func resolveGitHubToken(host string) string {
	if host == "" {
		host = "github.com"
	}

	// 1. GITHUB_TOKEN (UNIRTM_GITHUB_TOKEN -> MISE_GITHUB_TOKEN -> GITHUB_TOKEN)
	if token := env.Get("GITHUB_TOKEN"); token != "" {
		return token
	}

	// 2. GITHUB_API_TOKEN (legacy fallback)
	if token := os.Getenv("GITHUB_API_TOKEN"); token != "" {
		return token
	}

	// 3. credential_command (UNIRTM_GITHUB_CREDENTIAL_COMMAND -> MISE_GITHUB_CREDENTIAL_COMMAND -> GITHUB_CREDENTIAL_COMMAND)
	if cmd := env.Get("GITHUB_CREDENTIAL_COMMAND"); cmd != "" {
		if token := runCredentialCommand(cmd, host); token != "" {
			return token
		}
	}

	// 5. github_tokens.toml (~/.config/unirtm/github_tokens.toml)
	if token := readGitHubTokensFile(host); token != "" {
		return token
	}

	// 6. gh CLI hosts.yml
	if token := readGhCliToken(host); token != "" {
		return token
	}

	return ""
}

// runCredentialCommand executes the configured credential command and returns its stdout.
func runCredentialCommand(command, host string) string {
	cmd := exec.Command("sh", "-c", command, "--", host) // #nosec G204
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// readGitHubTokensFile reads the per-host token from ~/.config/unirtm/github_tokens.toml.
func readGitHubTokensFile(host string) string {
	configDir := os.Getenv("UNIRTM_CONFIG_DIR")
	if configDir == "" {
		xdg := os.Getenv("XDG_CONFIG_HOME")
		if xdg != "" {
			configDir = filepath.Join(xdg, "unirtm")
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return ""
			}
			configDir = filepath.Join(home, ".config", "unirtm")
		}
	}

	tokenFile := filepath.Join(configDir, "github_tokens.toml")
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return ""
	}

	var tokens githubTokensFile
	if err := toml.Unmarshal(data, &tokens); err != nil {
		return ""
	}

	if entry, ok := tokens.Tokens[host]; ok && entry.Token != "" {
		return entry.Token
	}

	return ""
}

// readGhCliToken attempts to read a token from the gh CLI's hosts.yml config file.
func readGhCliToken(host string) string {
	hostsFile := findGhHostsFile()
	if hostsFile == "" {
		return ""
	}

	data, err := os.ReadFile(hostsFile)
	if err != nil {
		return ""
	}

	// Parse the hosts.yml manually (avoid importing a YAML library).
	// Format:
	//   github.com:
	//     oauth_token: ghp_xxxx
	//     user: you
	return parseGhHostsYml(string(data), host)
}

// findGhHostsFile locates the gh CLI hosts.yml file.
func findGhHostsFile() string {
	// 1. $GH_CONFIG_DIR/hosts.yml
	if ghConfigDir := os.Getenv("GH_CONFIG_DIR"); ghConfigDir != "" {
		p := filepath.Join(ghConfigDir, "hosts.yml")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// 2. $XDG_CONFIG_HOME/gh/hosts.yml
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		xdg = filepath.Join(home, ".config")
	}
	p := filepath.Join(xdg, "gh", "hosts.yml")
	if _, err := os.Stat(p); err == nil {
		return p
	}

	// 3. macOS: ~/Library/Application Support/gh/hosts.yml
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	macosPath := filepath.Join(home, "Library", "Application Support", "gh", "hosts.yml")
	if _, err := os.Stat(macosPath); err == nil {
		return macosPath
	}

	return ""
}

// parseGhHostsYml parses the gh CLI hosts.yml and extracts the oauth_token for the given host.
// This is a minimal line-by-line parser that avoids importing a YAML library.
func parseGhHostsYml(content, targetHost string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	inTargetHost := false

	for scanner.Scan() {
		line := scanner.Text()

		// Top-level key (host entry, e.g. "github.com:")
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			key := strings.TrimSuffix(strings.TrimSpace(line), ":")
			inTargetHost = (key == targetHost)
			continue
		}

		// Inside the target host block
		if inTargetHost {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "oauth_token:") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}

	return ""
}
