// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// JavaHandler handles Java distributions via Adoptium (Temurin) API.
type JavaHandler struct {
	ImageType string // "jdk" or "jre"
}

func (h *JavaHandler) Name() string {
	return "java"
}

type adoptiumRelease struct {
	Binaries []struct {
		Package struct {
			Name string `json:"name"`
			Link string `json:"link"`
		} `json:"package"`
		SignatureLink string `json:"signature_link"`
	} `json:"binaries"`
	ReleaseName string `json:"release_name"`
	VersionData struct {
		OpenjdkVersion string `json:"openjdk_version"`
	} `json:"version_data"`
}

func (h *JavaHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// Adoptium versions: 23 (GA), 21 (LTS), 17 (LTS), 11 (LTS), 8 (LTS)
	majorVersions := []string{"23", "21", "17", "11", "8"}
	
	var allVersions []VersionInfo
	
	// Map OS/Arch to Adoptium values
	os := runtime.GOOS
	if os == "darwin" {
		os = "mac"
	}
	
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x64"
	} else if arch == "arm64" {
		arch = "aarch64"
	}

	for _, v := range majorVersions {
		imageType := h.ImageType
		if imageType == "" {
			imageType = "jdk"
		}
		url := fmt.Sprintf("https://api.adoptium.net/v3/assets/feature_releases/%s/ga?architecture=%s&heap_size=normal&image_type=%s&jvm_impl=hotspot&os=%s&project=jdk&vendor=eclipse", v, arch, imageType, os)
		
		client := &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		var releases []adoptiumRelease
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			continue
		}

		for _, rel := range releases {
			version := rel.VersionData.OpenjdkVersion
			// Clean version string (e.g. 21.0.2+13-LTS -> 21.0.2)
			// Remove -LTS if present
			version = strings.ReplaceAll(version, "-LTS", "")
			if idx := strings.Index(version, "+"); idx != -1 {
				version = version[:idx]
			}
			
			for _, bin := range rel.Binaries {
				assets := []Asset{
					{
						Filename:     bin.Package.Name,
						URL:          bin.Package.Link,
						SignatureURL: bin.SignatureLink,
						OS:           runtime.GOOS,
						Arch:         runtime.GOARCH,
					},
				}
				
				allVersions = append(allVersions, VersionInfo{
					Version: version,
					Assets:  assets,
				})
				// Just take the first binary that matches our query filters
				break
			}
		}
	}

	return allVersions, nil
}
