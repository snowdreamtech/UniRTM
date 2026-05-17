// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

// KubectlHandler handles kubectl distribution via dl.k8s.io.
type KubectlHandler struct{}

func (h *KubectlHandler) Name() string {
	return "kubectl"
}

func (h *KubectlHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// Fetch latest stable version
	stableURL := "https://dl.k8s.io/release/stable.txt"

	client := pkgHttp.NewClientWithTimeout(10 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", stableURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kubectl api: returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	latestVersion := strings.TrimSpace(string(body))

	// For now, just return the latest version.
	// In a full implementation, we could fetch historical versions.
	versions := []VersionInfo{
		{
			Version: latestVersion,
			Assets:  h.generateAssets(latestVersion),
		},
	}

	return versions, nil
}

func (h *KubectlHandler) generateAssets(version string) []Asset {
	var assets []Asset

	// Supported OS/Arch combinations
	platforms := []struct{ os, arch string }{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
	}

	for _, p := range platforms {
		ext := ""
		if p.os == "windows" {
			ext = ".exe"
		}

		url := fmt.Sprintf("https://dl.k8s.io/release/%s/bin/%s/%s/kubectl%s", version, p.os, p.arch, ext)

		assets = append(assets, Asset{
			Filename: "kubectl" + ext,
			URL:      url,
			OS:       p.os,
			Arch:     p.arch,
		})
	}

	return assets
}
