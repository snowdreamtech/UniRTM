// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// AsdfBackend implements the Backend interface for asdf plugins.
type AsdfBackend struct {
	mu           sync.Mutex
	client       *http.Client
	registryPath string
	pluginsPath  string
}

// NewAsdfBackend creates a new asdf backend.
func NewAsdfBackend() *AsdfBackend {
	return &AsdfBackend{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		registryPath: filepath.Join(env.GetDataDir(), "asdf", "registry"),
		pluginsPath:  env.GetPluginsDir(),
	}
}

func (b *AsdfBackend) Name() string {
	return "asdf"
}

// asdfAliases maps common tool names to their official asdf plugin names.
var asdfAliases = map[string]string{
	"node": "nodejs",
	"go":   "golang",
}

// ResolveAsdfToolName maps common tool names to their official asdf plugin names.
func ResolveAsdfToolName(tool string) string {
	if alias, ok := asdfAliases[tool]; ok {
		return alias
	}
	return tool
}

func (b *AsdfBackend) resolveToolName(tool string) string {
	return ResolveAsdfToolName(tool)
}

func (b *AsdfBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	tool = b.resolveToolName(tool)
	pluginDir, err := b.ensurePlugin(ctx, tool)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "ensure plugin", err)
	}

	listAllScript := filepath.Join(pluginDir, "bin", "list-all")
	if _, err := os.Stat(listAllScript); os.IsNotExist(err) {
		return nil, NewBackendError(b.Name(), tool, "plugin does not support list-all", err)
	}

	cmd := exec.CommandContext(ctx, listAllScript)
	cmd.Dir = pluginDir
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, NewBackendError(b.Name(), tool, "execute list-all", err)
	}

	lines := strings.Split(out.String(), "\n")
	var versions []VersionInfo

	// asdf plugins usually return versions oldest to newest, separated by spaces or newlines.
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		// Some plugins split versions by space on a single line
		parts := strings.Fields(line)
		for j := len(parts) - 1; j >= 0; j-- {
			v := parts[j]
			if v != "" {
				versions = append(versions, VersionInfo{
					Version:  v,
					Platform: platform,
				})
			}
		}
	}

	return versions, nil
}

func (b *AsdfBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	tool = b.resolveToolName(tool)
	if versionRequest == "latest" {
		versions, err := b.ListVersions(ctx, tool, platform)
		if err != nil {
			return nil, err
		}
		if len(versions) == 0 {
			return nil, NewBackendError(b.Name(), tool, "no versions found", nil)
		}
		// Assuming ListVersions returns newest first
		return &versions[0], nil
	}

	// For specific version requests, we just return it.
	// The provider will fail during installation if it's invalid.
	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *AsdfBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	tool = b.resolveToolName(tool)
	// asdf plugins don't provide download info without actually downloading.
	// We need to ensure the plugin is present for the provider.
	if _, err := b.ensurePlugin(ctx, tool); err != nil {
		return nil, NewBackendError(b.Name(), tool, "ensure plugin", err)
	}

	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *AsdfBackend) SupportsChecksum() bool {
	return false
}

func (b *AsdfBackend) SupportsGPG() bool {
	return false
}

func (b *AsdfBackend) AttestationType() string {
	return ""
}

// ensurePlugin ensures the plugin repository is cloned locally.
func (b *AsdfBackend) ensurePlugin(ctx context.Context, tool string) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	pluginDir := filepath.Join(b.pluginsPath, tool)
	if _, err := os.Stat(pluginDir); err == nil {
		return pluginDir, nil // already cloned
	}

	if err := b.updateRegistry(ctx); err != nil {
		logger.Warn("Failed to update asdf registry (will try fallback)", map[string]interface{}{"error": err.Error()})
	}

	repoURL, err := b.lookupPluginURL(tool)
	if err != nil {
		// Fallback heuristics
		repoURL = fmt.Sprintf("https://github.com/asdf-community/asdf-%s.git", tool)
	}

	logger.Info("Cloning asdf plugin", map[string]interface{}{"tool": tool, "url": repoURL})

	if err := os.MkdirAll(b.pluginsPath, 0755); err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, pluginDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		// Try alternative fallback
		if strings.Contains(repoURL, "asdf-community") {
			altURL := fmt.Sprintf("https://github.com/asdf-vm/asdf-%s.git", tool)
			logger.Info("Fallback clone asdf plugin", map[string]interface{}{"tool": tool, "url": altURL})
			cmd = exec.CommandContext(ctx, "git", "clone", altURL, pluginDir)
			if _, err := cmd.CombinedOutput(); err != nil {
				return "", fmt.Errorf("git clone failed: %s", string(out))
			}
		} else {
			return "", fmt.Errorf("git clone failed: %s", string(out))
		}
	}

	return pluginDir, nil
}

// updateRegistry clones or fetches the central asdf-plugins registry.
func (b *AsdfBackend) updateRegistry(ctx context.Context) error {
	if _, err := os.Stat(b.registryPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(b.registryPath), 0755); err != nil {
			return err
		}
		cmd := exec.CommandContext(ctx, "git", "clone", "https://github.com/asdf-vm/asdf-plugins.git", b.registryPath)
		return cmd.Run()
	}

	// Only update once every 24 hours to avoid spamming
	stat, err := os.Stat(filepath.Join(b.registryPath, ".git", "FETCH_HEAD"))
	if err == nil && time.Since(stat.ModTime()) < 24*time.Hour {
		return nil
	}

	cmd := exec.CommandContext(ctx, "git", "-C", b.registryPath, "pull")
	return cmd.Run()
}

// lookupPluginURL reads the repository URL from the central registry.
func (b *AsdfBackend) lookupPluginURL(tool string) (string, error) {
	pluginFile := filepath.Join(b.registryPath, "plugins", tool)
	data, err := os.ReadFile(pluginFile)
	if err != nil {
		return "", err
	}
	url := strings.TrimSpace(string(data))
	if strings.HasPrefix(url, "repository =") {
		url = strings.TrimSpace(strings.TrimPrefix(url, "repository ="))
	}
	if url == "" {
		return "", errors.New("empty repository url")
	}
	return url, nil
}

func (b *AsdfBackend) IsRecommended() bool {
	return false
}
