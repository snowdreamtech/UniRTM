// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"

	"github.com/snowdreamtech/unirtm/internal/config"
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
//  3. GH_TOKEN            (standard GitHub CLI env var)
//  4. GITHUB_API_TOKEN    (legacy alternative)
//  5. credential_command  (via UNIRTM_GITHUB_CREDENTIAL_COMMAND env var)
//  6. github_tokens.toml  (~/.config/unirtm/github_tokens.toml)
//  7. gh CLI hosts.yml
//  8. gh auth token command fallback
func resolveGitHubToken(host string) string {
	if host == "" {
		host = "github.com"
	}

	// 1. Environment variables via unified env.Get (UNIRTM_* -> MISE_* -> raw, supporting case variations)
	if token := resolveGitHubTokenFromEnv(); token != "" {
		return token
	}

	// 2. credential_command (UNIRTM_GITHUB_CREDENTIAL_COMMAND -> MISE_GITHUB_CREDENTIAL_COMMAND -> GITHUB_CREDENTIAL_COMMAND)
	if cmd := env.Get("GITHUB_CREDENTIAL_COMMAND"); cmd != "" {
		if token := runCredentialCommand(cmd, host); token != "" {
			return token
		}
	}

	// 3. Unified configuration hierarchy (workspace & user level via config.LoadFull)
	if token := readConfigGitHubToken(host); token != "" {
		return token
	}

	// 4. Workspace configuration (.unirtm.toml, unirtm.toml in cwd and parent dirs)
	if token := readWorkspaceGitHubToken(host); token != "" {
		return token
	}

	// 5. User configuration (github_tokens.toml, ~/.config/unirtm/unirtm.toml)
	if token := readGitHubTokensFile(host); token != "" {
		return token
	}

	// 6. gh CLI hosts.yml
	if token := readGhCliToken(host); token != "" {
		return token
	}

	// 7. gh CLI auth token execution fallback
	if ghAuthTokenFn != nil {
		if token := ghAuthTokenFn(host); token != "" {
			return token
		}
	}

	return ""
}

// resolveGitHubTokenFromEnv retrieves GitHub token from environment variables using unified env.Get.
// Priority: UNIRTM_* > MISE_* > Raw keys. Fully supports case-insensitivity and prefix tolerance.
func resolveGitHubTokenFromEnv() string {
	for _, key := range []string{"GITHUB_TOKEN", "GH_TOKEN", "GITHUB_API_TOKEN"} {
		if token := env.Get(key); token != "" {
			return token
		}
	}
	return ""
}

// readConfigGitHubToken loads configuration via unified config.LoadFull() and extracts github token.
func readConfigGitHubToken(host string) string {
	if !strings.EqualFold(host, "github.com") && host != "" {
		return ""
	}

	cfg, err := config.LoadFull()
	if err != nil || cfg == nil {
		return ""
	}

	// 1. Check Settings.GitHubToken
	if cfg.Settings.GitHubToken != "" {
		return cfg.Settings.GitHubToken
	}

	// 2. Check [env] section
	for k, v := range cfg.Env {
		if isGitHubTokenKey(k) {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}

	return ""
}

var ghAuthTokenFn = runGhAuthToken

// runGhAuthToken runs `gh auth token` to retrieve token from gh credential helper if available.
func runGhAuthToken(host string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gh", "auth", "token", "--hostname", host)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// runCredentialCommand executes the configured credential command and returns its stdout.
func runCredentialCommand(command, host string) string {
	var cmd *exec.Cmd
	if env.RuntimeGOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command+" "+host)
	} else {
		cmd = exec.Command("sh", "-c", command, "--", host) // #nosec G204
	}
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// isGitHubTokenKey checks if a key matches any casing or format for a GitHub token.
// Matches e.g. github_token, GITHUB_TOKEN, GitHub_Token, githubToken, github-token,
// gh_token, GH_TOKEN, unirtm_github_token, mise_github_token, unirtm_gh_token, mise_gh_token, github_api_token.
func isGitHubTokenKey(key string) bool {
	clean := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "_", ""), "-", ""))
	clean = strings.TrimPrefix(clean, "unirtm")
	clean = strings.TrimPrefix(clean, "mise")
	return clean == "githubtoken" || clean == "ghtoken" || clean == "githubapitoken"
}

// extractGitHubTokenFromToml extracts a GitHub token from unmarshalled TOML bytes,
// case-insensitively inspecting [tokens], [env], [settings], and root-level keys.
func extractGitHubTokenFromToml(data []byte, host string) string {
	if host == "" {
		host = "github.com"
	}

	var raw map[string]interface{}
	if err := toml.Unmarshal(data, &raw); err != nil {
		return ""
	}

	// 1. Check [tokens] table (e.g., [tokens."github.com"] token = "...")
	for k, v := range raw {
		if strings.EqualFold(k, "tokens") {
			if tokensMap, ok := v.(map[string]interface{}); ok {
				for hostKey, hostVal := range tokensMap {
					if strings.EqualFold(hostKey, host) {
						if entryMap, ok := hostVal.(map[string]interface{}); ok {
							for ek, ev := range entryMap {
								cleanEk := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(ek, "_", ""), "-", ""))
								if cleanEk == "token" || isGitHubTokenKey(ek) {
									if s, ok := ev.(string); ok && strings.TrimSpace(s) != "" {
										return strings.TrimSpace(s)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// For non-github.com hosts, host-specific tokens table is the primary target
	if !strings.EqualFold(host, "github.com") {
		return ""
	}

	// 2. Check [env] section with case-insensitive section and key matching
	for k, v := range raw {
		if strings.EqualFold(k, "env") {
			if envMap, ok := v.(map[string]interface{}); ok {
				for ek, ev := range envMap {
					if isGitHubTokenKey(ek) {
						if s, ok := ev.(string); ok && strings.TrimSpace(s) != "" {
							return strings.TrimSpace(s)
						}
					}
				}
			}
		}
	}

	// 3. Check [settings] section with case-insensitive section and key matching
	for k, v := range raw {
		if strings.EqualFold(k, "settings") {
			if settingsMap, ok := v.(map[string]interface{}); ok {
				for sk, sv := range settingsMap {
					if isGitHubTokenKey(sk) {
						if s, ok := sv.(string); ok && strings.TrimSpace(s) != "" {
							return strings.TrimSpace(s)
						}
					}
				}
			}
		}
	}

	// 4. Check root-level keys
	for k, v := range raw {
		if isGitHubTokenKey(k) {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}

	return ""
}

// readWorkspaceGitHubToken searches for a GitHub token in workspace configuration files
// (.unirtm.local.toml, .unirtm.toml, unirtm.toml, .mise.local.toml, .mise.toml)
// from the current working directory walking up to the root directory.
func readWorkspaceGitHubToken(host string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	curr := cwd
	for {
		candidates := []string{
			filepath.Join(curr, ".unirtm.local.toml"),
			filepath.Join(curr, ".unirtm.toml"),
			filepath.Join(curr, "unirtm.toml"),
			filepath.Join(curr, ".mise.local.toml"),
			filepath.Join(curr, ".mise.toml"),
		}
		for _, path := range candidates {
			if data, err := os.ReadFile(path); err == nil {
				if tok := extractGitHubTokenFromToml(data, host); tok != "" {
					return tok
				}
			}
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	return ""
}

// readGitHubTokensFile reads the per-host token from user configuration:
// 1. ~/.config/unirtm/github_tokens.toml
// 2. ~/.config/unirtm/unirtm.toml
// 3. ~/.unirtm.toml
func readGitHubTokensFile(host string) string {
	configDir := env.Get("CONFIG_DIR")
	if configDir == "" {
		xdg := env.Get("XDG_CONFIG_HOME")
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

	userFiles := []string{
		filepath.Join(configDir, "github_tokens.toml"),
		filepath.Join(configDir, "unirtm.toml"),
		filepath.Join(configDir, ".unirtm.toml"),
	}

	if home, err := os.UserHomeDir(); err == nil {
		userFiles = append(userFiles, filepath.Join(home, ".unirtm.toml"))
	}

	for _, file := range userFiles {
		if data, err := os.ReadFile(file); err == nil {
			if tok := extractGitHubTokenFromToml(data, host); tok != "" {
				return tok
			}
		}
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
	if ghConfigDir := env.Get("GH_CONFIG_DIR"); ghConfigDir != "" {
		p := filepath.Join(ghConfigDir, "hosts.yml")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// 2. $XDG_CONFIG_HOME/gh/hosts.yml
	xdg := env.Get("XDG_CONFIG_HOME")
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
	if err == nil {
		macosPath := filepath.Join(home, "Library", "Application Support", "gh", "hosts.yml")
		if _, err := os.Stat(macosPath); err == nil {
			return macosPath
		}
	}

	// 4. Windows: %APPDATA%\GitHub CLI\hosts.yml
	if appData := os.Getenv("APPDATA"); appData != "" {
		winPath := filepath.Join(appData, "GitHub CLI", "hosts.yml")
		if _, err := os.Stat(winPath); err == nil {
			return winPath
		}
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
