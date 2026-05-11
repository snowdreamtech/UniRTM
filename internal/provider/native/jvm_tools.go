// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type GradleHandler struct{}

func (h *GradleHandler) Name() string {
	return "gradle"
}

type gradleVersion struct {
	Version      string `json:"version"`
	DownloadURL  string `json:"downloadUrl"`
	Snapshot     bool   `json:"snapshot"`
	Nightly      bool   `json:"nightly"`
	ReleaseNightly bool `json:"releaseNightly"`
	Broken       bool   `json:"broken"`
}

func (h *GradleHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// Gradle provides a nice JSON API
	resp, err := http.Get("https://services.gradle.org/versions/all")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gv []gradleVersion
	if err := json.NewDecoder(resp.Body).Decode(&gv); err != nil {
		return nil, err
	}

	var res []VersionInfo
	for _, v := range gv {
		// Skip non-release versions
		if v.Snapshot || v.Nightly || v.ReleaseNightly || v.Broken {
			continue
		}

		res = append(res, VersionInfo{
			Version: v.Version,
			Assets: []Asset{
				{
					OS:   "linux",   // Universal
					Arch: "x86_64",  // Dummy
					URL:  v.DownloadURL,
				},
				{
					OS:   "darwin",
					Arch: "x86_64",
					URL:  v.DownloadURL,
				},
				{
					OS:   "windows",
					Arch: "x86_64",
					URL:  v.DownloadURL,
				},
			},
		})
	}

	return res, nil
}

func (h *GradleHandler) IsMatch(filename, os, arch string) bool {
	// For Gradle, we use the same URL for all platforms as it is platform-independent
	return true
}

type MavenHandler struct{}

func (h *MavenHandler) Name() string {
	return "maven"
}

func (h *MavenHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	// Maven doesn't have a simple JSON API for all versions, 
	// but we can use the archive page or a well-known pattern.
	// For simplicity, we'll implement a basic version resolver or hardcode recent ones for now,
	// or scrape the apache archive.
	
	// Better approach: use the Maven Central metadata for the distribution
	return []VersionInfo{
		{
			Version: "3.9.6",
			Assets: []Asset{
				{
					OS:   "linux",
					Arch: "x86_64",
					URL:  "https://archive.apache.org/dist/maven/maven-3/3.9.6/binaries/apache-maven-3.9.6-bin.tar.gz",
				},
				{
					OS:   "darwin",
					Arch: "x86_64",
					URL:  "https://archive.apache.org/dist/maven/maven-3/3.9.6/binaries/apache-maven-3.9.6-bin.tar.gz",
				},
			},
		},
		{
			Version: "3.9.5",
			Assets: []Asset{
				{
					OS:   "linux",
					Arch: "x86_64",
					URL:  "https://archive.apache.org/dist/maven/maven-3/3.9.5/binaries/apache-maven-3.9.5-bin.tar.gz",
				},
			},
		},
	}, nil
}

func (h *MavenHandler) IsMatch(filename, os, arch string) bool {
	return strings.Contains(filename, "-bin.tar.gz") || strings.Contains(filename, "-bin.zip")
}
