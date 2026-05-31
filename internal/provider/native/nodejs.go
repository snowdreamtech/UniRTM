// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
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
	// Support Node.js Mirrors
	mirrorURL := env.Get("MISE_NODE_MIRROR_URL")
	if mirrorURL == "" {
		mirrorURL = env.Get("NODEJS_ORG_MIRROR")
	}
	if mirrorURL != "" {
		baseURL = mirrorURL
	}

	url := fmt.Sprintf("%s/index.json", strings.TrimSuffix(baseURL, "/"))
	client := pkgHttp.NewClientWithTimeout(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "unirtm/"+env.GitTag)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nodejs: fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	var nv []nodeVersion
	if err := json.NewDecoder(resp.Body).Decode(&nv); err != nil {
		return nil, fmt.Errorf("nodejs: decode metadata: %w", err)
	}

	flavor := env.Get("MISE_NODE_FLAVOR")

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
			osName, archName, rawArch, ext, isSupported := parseNodeFile(f)
			if !isSupported {
				continue
			}

			downloadURL := fmt.Sprintf("%s/%s/node-%s-%s-%s%s", strings.TrimSuffix(baseURL, "/"), v.Version, v.Version, osName, rawArch, ext)
			if flavor == "musl" {
				// unofficial-builds naming convention: node-vX.Y.Z-linux-ARCH-musl.tar.gz
				downloadURL = fmt.Sprintf("%s/%s/node-%s-%s-%s-musl%s", strings.TrimSuffix(baseURL, "/"), v.Version, v.Version, osName, rawArch, ext)
			}

			vi.Assets = append(vi.Assets, Asset{
				URL:          downloadURL,
				Filename:     filepath.Base(downloadURL),
				OS:           osName,
				Arch:         archName,
				Algo:         "sha256",
				SignatureURL: fmt.Sprintf("%s/%s/SHASUMS256.txt.asc", strings.TrimSuffix(baseURL, "/"), v.Version),
				Metadata: map[string]string{
					"flavor": flavor,
				},
			})
		}

		if len(vi.Assets) > 0 {
			versions = append(versions, vi)
		}
	}

	return versions, nil
}

func parseNodeFile(f string) (string, string, string, string, bool) {
	// Node files format: os-arch (e.g., linux-x64, osx-arm64, win-x64-zip)
	parts := strings.Split(f, "-")
	if len(parts) < 2 {
		return "", "", "", "", false
	}

	osName := parts[0]
	rawArch := parts[1]
	ext := ".tar.gz"

	if len(parts) >= 3 {
		format := parts[2]
		switch format {
		case "zip":
			ext = ".zip"
		case "7z":
			ext = ".7z"
		case "msi":
			ext = ".msi"
		case "pkg":
			ext = ".pkg"
		case "exe":
			ext = ".exe"
		case "tar":
			ext = ".tar.gz"
		}
	}

	// Map osx to darwin
	if osName == "osx" {
		osName = "darwin"
	}

	// Map architecture names to UniRTM standards
	archName := rawArch
	switch rawArch {
	case "x64":
		archName = "amd64"
	case "x86":
		archName = "386"
	}

	// Skip non-tar.gz and non-zip formats for now
	if ext == ".7z" || ext == ".msi" || ext == ".pkg" || ext == ".exe" {
		return "", "", "", "", false
	}

	return osName, archName, rawArch, ext, true
}
