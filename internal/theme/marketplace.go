// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package theme

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const OpenVSXSearchURL = "https://open-vsx.org/api/-/search"

type MarketplaceExtensionFiles struct {
	Download  string `json:"download"`
	Icon      string `json:"icon,omitempty"`
	Signature string `json:"signature,omitempty"`
	Readme    string `json:"readme,omitempty"`
	Changelog string `json:"changelog,omitempty"`
	Manifest  string `json:"manifest,omitempty"`
}

type ContributedSubTheme struct {
	ID      string `json:"id,omitempty"`
	Label   string `json:"label"`
	UITheme string `json:"uiTheme,omitempty"` // "vs", "vs-dark", "hc-black", "hc-light"
	Path    string `json:"path,omitempty"`
}

type ContributedSubIconTheme struct {
	ID    string `json:"id,omitempty"`
	Label string `json:"label"`
	Path  string `json:"path,omitempty"`
}

type ExtensionContributions struct {
	Themes     []ContributedSubTheme     `json:"themes,omitempty"`
	IconThemes []ContributedSubIconTheme `json:"iconThemes,omitempty"`
}

type MarketplaceExtensionDetail struct {
	Namespace     string                    `json:"namespace"`
	Name          string                    `json:"name"`
	DisplayName   string                    `json:"displayName"`
	Version       string                    `json:"version"`
	Description   string                    `json:"description"`
	DownloadCount int64                     `json:"downloadCount"`
	AverageRating float64                   `json:"averageRating"`
	ReviewCount   int                       `json:"reviewCount"`
	Repository    string                    `json:"repository,omitempty"`
	Homepage      string                    `json:"homepage,omitempty"`
	Bugs          string                    `json:"bugs,omitempty"`
	License       string                    `json:"license,omitempty"`
	Timestamp     string                    `json:"timestamp,omitempty"`
	Tags          []string                  `json:"tags,omitempty"`
	Categories    []string                  `json:"categories,omitempty"`
	Files         MarketplaceExtensionFiles `json:"files"`
	URL           string                    `json:"url"`
	IsInstalled   bool                      `json:"is_installed"`
	PackageID     string                    `json:"package_id"`
	ReadmeContent string                    `json:"readme,omitempty"`
	Contributes   ExtensionContributions    `json:"contributes"`
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

// GetOpenVSXExtensionDetail fetches full metadata, README markdown, and contributed themes for a specific extension.
func GetOpenVSXExtensionDetail(namespace, name string, installedIDs map[string]bool) (*MarketplaceExtensionDetail, error) {
	namespace = strings.TrimSpace(namespace)
	name = strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return nil, errors.New("namespace and name are required")
	}

	apiURL := fmt.Sprintf("https://open-vsx.org/api/%s/%s", url.PathEscape(namespace), url.PathEscape(name))
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "UniRTM-ThemeMarketplace/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch extension details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Open VSX API returned status: %d", resp.StatusCode)
	}

	var detail MarketplaceExtensionDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, fmt.Errorf("failed to decode extension details: %w", err)
	}

	pkgID := detail.Name
	if pkgID == "" {
		pkgID = detail.DisplayName
	}
	pkgID = strings.ReplaceAll(pkgID, "/", "-")
	detail.PackageID = pkgID

	if installedIDs != nil {
		if installedIDs[detail.Name] || installedIDs[pkgID] || installedIDs[detail.Namespace+"."+detail.Name] {
			detail.IsInstalled = true
		}
	}

	// Concurrently fetch README and manifest if available
	var wg sync.WaitGroup
	if detail.Files.Readme != "" {
		wg.Add(1)
		go func(readmeURL string) {
			defer wg.Done()
			rReq, rErr := http.NewRequest(http.MethodGet, readmeURL, nil)
			if rErr == nil {
				rReq.Header.Set("User-Agent", "UniRTM-ThemeMarketplace/1.0")
				rResp, doErr := client.Do(rReq)
				if doErr == nil && rResp.StatusCode == http.StatusOK {
					defer rResp.Body.Close()
					limited := io.LimitReader(rResp.Body, 512*1024)
					b, _ := io.ReadAll(limited)
					detail.ReadmeContent = string(b)
				}
			}
		}(detail.Files.Readme)
	}

	if detail.Files.Manifest != "" {
		wg.Add(1)
		go func(manifestURL string) {
			defer wg.Done()
			mReq, mErr := http.NewRequest(http.MethodGet, manifestURL, nil)
			if mErr == nil {
				mReq.Header.Set("User-Agent", "UniRTM-ThemeMarketplace/1.0")
				mResp, doErr := client.Do(mReq)
				if doErr == nil && mResp.StatusCode == http.StatusOK {
					defer mResp.Body.Close()
					var pkg struct {
						Contributes struct {
							Themes []struct {
								ID      string `json:"id"`
								Label   string `json:"label"`
								UITheme string `json:"uiTheme"`
								Path    string `json:"path"`
							} `json:"themes"`
							IconThemes []struct {
								ID    string `json:"id"`
								Label string `json:"label"`
								Path  string `json:"path"`
							} `json:"iconThemes"`
						} `json:"contributes"`
					}
					limited := io.LimitReader(mResp.Body, 512*1024)
					if err := json.NewDecoder(limited).Decode(&pkg); err == nil {
						for _, t := range pkg.Contributes.Themes {
							lbl := t.Label
							if lbl == "" {
								lbl = t.ID
							}
							detail.Contributes.Themes = append(detail.Contributes.Themes, ContributedSubTheme{
								ID:      t.ID,
								Label:   lbl,
								UITheme: t.UITheme,
								Path:    t.Path,
							})
						}
						for _, it := range pkg.Contributes.IconThemes {
							lbl := it.Label
							if lbl == "" {
								lbl = it.ID
							}
							detail.Contributes.IconThemes = append(detail.Contributes.IconThemes, ContributedSubIconTheme{
								ID:    it.ID,
								Label: lbl,
								Path:  it.Path,
							})
						}
					}
				}
			}
		}(detail.Files.Manifest)
	}

	wg.Wait()
	return &detail, nil
}

