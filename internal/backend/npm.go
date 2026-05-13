// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// NpmBackend implements the Backend interface for npm packages.
type NpmBackend struct {
	client *http.Client
}

// NewNpmBackend creates a new npm backend.
func NewNpmBackend() *NpmBackend {
	return &NpmBackend{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (b *NpmBackend) Name() string {
	return "npm"
}

type npmRegistryResponse struct {
	Versions map[string]interface{} `json:"versions"`
	Time     map[string]string      `json:"time"`
}

func (b *NpmBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	url := fmt.Sprintf("https://registry.npmjs.org/%s", tool)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "create request", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, NewBackendError(b.Name(), tool, "execute request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, NewBackendError(b.Name(), tool, "package not found", nil)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	var registry npmRegistryResponse
	if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
		return nil, NewBackendError(b.Name(), tool, "decode response", err)
	}

	var versions []VersionInfo
	for v := range registry.Versions {
		versions = append(versions, VersionInfo{
			Version:  v,
			Platform: platform,
		})
	}

	// npm registry doesn't strictly order keys in maps, so we should ideally sort them.
	// But since this is just a listing and version manager usually sorts them semantically,
	// returning them as-is is acceptable for now.

	return versions, nil
}

func (b *NpmBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "latest" {
		url := fmt.Sprintf("https://registry.npmjs.org/%s/latest", tool)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := b.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, NewBackendError(b.Name(), tool, "latest version not found", nil)
		}

		var latest struct {
			Version string `json:"version"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&latest); err != nil {
			return nil, err
		}

		return &VersionInfo{
			Version:  latest.Version,
			Platform: platform,
		}, nil
	}

	// Assume explicit version is correct; provider will fail if it doesn't exist
	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *NpmBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	// Provider handles actual downloading via npm cli
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *NpmBackend) SupportsChecksum() bool {
	return true
}

func (b *NpmBackend) SupportsGPG() bool {
	return false
}

func (b *NpmBackend) AttestationType() string {
	return ""
}
