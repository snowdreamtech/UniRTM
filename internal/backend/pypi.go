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

// PypiBackend implements the Backend interface for PyPI packages.
type PypiBackend struct {
	client *http.Client
}

// NewPypiBackend creates a new PyPI backend.
func NewPypiBackend() *PypiBackend {
	return &PypiBackend{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (b *PypiBackend) Name() string {
	return "pypi"
}

type pypiRegistryResponse struct {
	Info struct {
		Version string `json:"version"`
	} `json:"info"`
	Releases map[string]interface{} `json:"releases"`
}

func (b *PypiBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", tool)

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

	var registry pypiRegistryResponse
	if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
		return nil, NewBackendError(b.Name(), tool, "decode response", err)
	}

	var versions []VersionInfo
	for v := range registry.Releases {
		versions = append(versions, VersionInfo{
			Version:  v,
			Platform: platform,
		})
	}

	return versions, nil
}

func (b *PypiBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "latest" {
		url := fmt.Sprintf("https://pypi.org/pypi/%s/json", tool)
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

		var registry pypiRegistryResponse
		if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
			return nil, err
		}

		return &VersionInfo{
			Version:  registry.Info.Version,
			Platform: platform,
		}, nil
	}

	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *PypiBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	return &VersionInfo{
		Version:  version,
		Platform: platform,
	}, nil
}

func (b *PypiBackend) SupportsChecksum() bool {
	return false
}

func (b *PypiBackend) SupportsGPG() bool {
	return false
}

func (b *PypiBackend) AttestationType() string {
	return ""
}
