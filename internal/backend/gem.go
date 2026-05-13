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

// GemBackend implements the Backend interface for RubyGems.
type GemBackend struct {
	client *http.Client
}

// NewGemBackend creates a new gem backend.
func NewGemBackend() *GemBackend {
	return &GemBackend{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (b *GemBackend) Name() string {
	return "gem"
}

type gemVersion struct {
	Number string `json:"number"`
}

func (b *GemBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	url := fmt.Sprintf("https://rubygems.org/api/v1/versions/%s.json", tool)

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
		return nil, NewBackendError(b.Name(), tool, "gem not found", nil)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, NewBackendError(b.Name(), tool, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	var gemVersions []gemVersion
	if err := json.NewDecoder(resp.Body).Decode(&gemVersions); err != nil {
		return nil, NewBackendError(b.Name(), tool, "decode response", err)
	}

	var versions []VersionInfo
	for _, v := range gemVersions {
		versions = append(versions, VersionInfo{
			Version:  v.Number,
			Platform: platform,
		})
	}

	return versions, nil
}

func (b *GemBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "latest" {
		url := fmt.Sprintf("https://rubygems.org/api/v1/gems/%s.json", tool)
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

func (b *GemBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	// Provider handles actual downloading via gem cli
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *GemBackend) SupportsChecksum() bool {
	return false
}

func (b *GemBackend) SupportsGPG() bool {
	return false
}

func (b *GemBackend) SupportsAttestation() bool {
	return false
}
