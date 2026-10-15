// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package theme

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const OpenVSXSearchURL = "https://open-vsx.org/api/-/search"

type MarketplaceExtensionFiles struct {
	Download  string `json:"download"`
	Icon      string `json:"icon,omitempty"`
	Signature string `json:"signature,omitempty"`
}

type MarketplaceExtension struct {
	Namespace     string                    `json:"namespace"`
	Name          string                    `json:"name"`
	DisplayName   string                    `json:"displayName"`
	Version       string                    `json:"version"`
	Description   string                    `json:"description"`
	DownloadCount int64                     `json:"downloadCount"`
	AverageRating float64                   `json:"averageRating"`
	Files         MarketplaceExtensionFiles `json:"files"`
	URL           string                    `json:"url"`
	IsInstalled   bool                      `json:"is_installed"`
	PackageID     string                    `json:"package_id"`
}

type MarketplaceSearchResponse struct {
	Offset     int                    `json:"offset"`
	TotalSize  int                    `json:"totalSize"`
	Extensions []MarketplaceExtension `json:"extensions"`
}

// SearchOpenVSXThemes queries Open VSX Registry strictly filtered to category=Themes.
func SearchOpenVSXThemes(query string, size, offset int, installedIDs map[string]bool) (*MarketplaceSearchResponse, error) {
	if size <= 0 {
		size = 20
	}
	if size > 50 {
		size = 50
	}
	if offset < 0 {
		offset = 0
	}

	endpoint, err := url.Parse(OpenVSXSearchURL)
	if err != nil {
		return nil, err
	}

	q := endpoint.Query()
	q.Set("category", "Themes")
	q.Set("size", fmt.Sprintf("%d", size))
	q.Set("offset", fmt.Sprintf("%d", offset))
	if strings.TrimSpace(query) != "" {
		q.Set("query", strings.TrimSpace(query))
	}
	endpoint.RawQuery = q.Encode()

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "UniRTM-ThemeMarketplace/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Open VSX API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Open VSX API returned status: %d", resp.StatusCode)
	}

	var result MarketplaceSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Open VSX search response: %w", err)
	}

	// Mark already installed extensions
	for i := range result.Extensions {
		ext := &result.Extensions[i]
		pkgID := ext.Name
		if pkgID == "" {
			pkgID = ext.DisplayName
		}
		pkgID = strings.ReplaceAll(pkgID, "/", "-")
		ext.PackageID = pkgID

		if installedIDs != nil {
			if installedIDs[ext.Name] || installedIDs[pkgID] || installedIDs[ext.Namespace+"."+ext.Name] {
				ext.IsInstalled = true
			}
		}
	}

	return &result, nil
}

// InstallFromURL streams a .vsix archive from a remote URL and unpacks it into user themes.
func (m *Manager) InstallFromURL(downloadURL string) ([]ThemeItem, error) {
	if downloadURL == "" {
		return nil, errors.New("download URL is required")
	}

	parsed, err := url.Parse(downloadURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, errors.New("invalid download URL scheme (http/https required)")
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "UniRTM-ThemeMarketplace/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download server returned HTTP %d", resp.StatusCode)
	}

	return m.InstallVSIX(resp.Body)
}
