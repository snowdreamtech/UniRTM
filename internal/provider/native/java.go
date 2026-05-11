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
type JavaHandler struct{}

func (h *JavaHandler) Name() string {
	return "java"
}

type adoptiumRelease struct {
	Binary struct {
		Package struct {
			Name string `json:"name"`
			Link string `json:"link"`
		} `json:"package"`
	} `json:"binary"`
	ReleaseName string `json:"release_name"`
	Version     struct {
		OpenjdkVersion string `json:"openjdk_version"`
	} `json:"version"`
}

func (h *JavaHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// Adoptium LTS versions: 8, 11, 17, 21
	ltsVersions := []string{"8", "11", "17", "21"}
	
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

	for _, v := range ltsVersions {
		url := fmt.Sprintf("https://api.adoptium.net/v3/assets/feature_releases/%s/ga?architecture=%s&heap_size=normal&image_type=jdk&jvm_impl=hotspot&os=%s&project=jdk&vendor=eclipse", v, arch, os)
		
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
			version := rel.Version.OpenjdkVersion
			// Clean version string (e.g. 21.0.2+13 -> 21.0.2)
			if idx := strings.Index(version, "+"); idx != -1 {
				version = version[:idx]
			}
			
			assets := []Asset{
				{
					Filename: rel.Binary.Package.Name,
					URL:      rel.Binary.Package.Link,
					OS:       runtime.GOOS,
					Arch:     runtime.GOARCH,
				},
			}
			
			allVersions = append(allVersions, VersionInfo{
				Version: version,
				Assets:  assets,
			})
		}
	}

	return allVersions, nil
}
