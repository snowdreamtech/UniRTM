// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// NodeJSHandler handles the official Node.js download metadata from nodejs.org/dist/index.json.
type NodeJSHandler struct{}

type nodeVersion struct {
	Version string      `json:"version"`
	Date    string      `json:"date"`
	Files   []string    `json:"files"`
	Lts     interface{} `json:"lts"` // can be false or a string (the LTS name)
}

func (h *NodeJSHandler) Name() string {
	return "nodejs"
}

func (h *NodeJSHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	url := fmt.Sprintf("%s/index.json", strings.TrimSuffix(baseURL, "/"))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nodejs: fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	var nv []nodeVersion
	if err := json.NewDecoder(resp.Body).Decode(&nv); err != nil {
		return nil, fmt.Errorf("nodejs: decode metadata: %w", err)
	}

	var versions []VersionInfo
	for _, v := range nv {
		vi := VersionInfo{
			Version: strings.TrimPrefix(v.Version, "v"),
		}

		if s, ok := v.Lts.(string); ok {
			vi.IsLTS = true
			vi.LTSName = s
		}

		for _, f := range v.Files {
			osName, archName, isSupported := parseNodeFile(f)
			if !isSupported {
				continue
			}

			vi.Assets = append(vi.Assets, Asset{
				URL:          fmt.Sprintf("%s/%s/node-%s-%s-%s.tar.gz", strings.TrimSuffix(baseURL, "/"), v.Version, v.Version, osName, archName),
				Filename:     fmt.Sprintf("node-%s-%s-%s.tar.gz", v.Version, osName, archName),
				OS:           osName,
				Arch:         archName,
				Algo:         "sha256", // Node provides separate SHASUMS256.txt, can be handled in Verify
				SignatureURL: fmt.Sprintf("%s/%s/SHASUMS256.txt.asc", strings.TrimSuffix(baseURL, "/"), v.Version),
			})
		}

		if len(vi.Assets) > 0 {
			versions = append(versions, vi)
		}
	}

	return versions, nil
}

func parseNodeFile(f string) (string, string, bool) {
	// Node files format: os-arch (e.g., linux-x64, osx-arm64)
	parts := strings.Split(f, "-")
	if len(parts) < 2 {
		return "", "", false
	}

	osName := parts[0]
	archName := parts[1]

	// Map osx to darwin
	if osName == "osx" {
		osName = "darwin"
	}

	// Map architecture names to UniRTM standards
	switch archName {
	case "x64":
		archName = "amd64"
	case "x86":
		archName = "386"
	}

	// Skip non-tar.gz formats for now (like .msi, .pkg, .exe)
	if strings.Contains(f, "zip") || strings.Contains(f, "7z") || strings.Contains(f, "msi") || strings.Contains(f, "pkg") {
		return "", "", false
	}

	return osName, archName, true
}
