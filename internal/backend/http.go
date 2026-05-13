// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// HTTPBackend implements the Backend interface for direct HTTP downloads.
// This backend supports URL templates with placeholders for version, OS, and architecture.
type HTTPBackend struct {
	client *http.Client
}

// NewHTTPBackend creates a new HTTP backend.
func NewHTTPBackend() *HTTPBackend {
	return &HTTPBackend{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the backend identifier.
func (h *HTTPBackend) Name() string {
	return "http"
}

// HTTPConfig represents the configuration for an HTTP backend tool.
// This should be provided in the tool configuration.
type HTTPConfig struct {
	URLTemplate      string            // URL template with placeholders (e.g., "https://example.com/{{.version}}/{{.os}}-{{.arch}}.tar.gz")
	Versions         []string          // List of available versions
	ChecksumTemplate string            // Optional checksum URL template
	Replacements     map[string]string // Custom placeholder replacements
}

// ListVersions returns all available versions for an HTTP backend tool.
// Note: HTTP backend requires versions to be explicitly configured.
func (h *HTTPBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// HTTP backend requires explicit configuration
	// In a real implementation, this would read from configuration
	return nil, NewBackendError("http", tool, "HTTP backend requires explicit version configuration", nil)
}

// ResolveVersion resolves a version request to a concrete version.
func (h *HTTPBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	// For HTTP backend, we assume the version is explicitly provided
	// and construct the download URL from the template
	return h.GetDownloadInfo(ctx, tool, versionRequest, platform)
}

// GetDownloadInfo retrieves download information for a specific version.
// This constructs the download URL from the configured template.
func (h *HTTPBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	// In a real implementation, this would read the HTTPConfig from configuration
	// For now, we return an error indicating configuration is required
	return nil, NewBackendError("http", tool, "HTTP backend requires URL template configuration", nil)
}

// GetDownloadInfoWithConfig retrieves download information using explicit configuration.
func (h *HTTPBackend) GetDownloadInfoWithConfig(ctx context.Context, tool string, version string, platform Platform, config HTTPConfig) (*VersionInfo, error) {
	if config.URLTemplate == "" {
		return nil, NewBackendError("http", tool, "URL template is required", nil)
	}

	// Build download URL from template
	downloadURL := h.buildURL(config.URLTemplate, version, platform, config.Replacements)

	// Verify URL is accessible (HEAD request)
	if !ProbeURL(ctx, h.client, downloadURL) {
		return nil, NewBackendError("http", tool, fmt.Sprintf("failed to verify URL: %s", downloadURL), nil)
	}

	// Build checksum URL if template provided
	checksumURL := ""
	if config.ChecksumTemplate != "" {
		checksumURL = h.buildURL(config.ChecksumTemplate, version, platform, config.Replacements)
	}

	// Fetch checksum if URL provided
	checksum := ""
	if checksumURL != "" {
		checksumMap := FetchAndParseChecksumFile(ctx, h.client, checksumURL)
		if checksumMap != nil {
			// Try to find by filename or just take the first one if it's a single-file checksum
			fileName := h.extractFileName(downloadURL)
			if c, ok := checksumMap[fileName]; ok {
				checksum = c
			} else if len(checksumMap) == 1 {
				for _, c := range checksumMap {
					checksum = c
					break
				}
			}
		}
	}

	// Try to auto-detect GPG signature if not provided
	gpgSigURL := ""
	// Try appending .asc or .sig to the download URL
	if ProbeURL(ctx, h.client, downloadURL+".asc") {
		gpgSigURL = downloadURL + ".asc"
	} else if ProbeURL(ctx, h.client, downloadURL+".sig") {
		gpgSigURL = downloadURL + ".sig"
	}

	return &VersionInfo{
		Version:     version,
		DownloadURL: downloadURL,
		Checksum:    checksum,
		Platform:    platform,
		Metadata: map[string]string{
			"backend":           "http",
			"url_template":      config.URLTemplate,
			"gpg_signature_url": gpgSigURL,
		},
	}, nil
}

func (h *HTTPBackend) extractFileName(rawURL string) string {
	parts := strings.Split(rawURL, "/")
	return parts[len(parts)-1]
}

// SupportsChecksum indicates whether this backend provides checksums.
func (h *HTTPBackend) SupportsChecksum() bool {
	return true
}

// SupportsGPG indicates whether this backend supports GPG signatures.
func (h *HTTPBackend) SupportsGPG() bool {
	return true
}

func (h *HTTPBackend) AttestationType() string {
	return ""
}

// buildURL constructs a URL from a template with placeholders.
func (h *HTTPBackend) buildURL(template string, version string, platform Platform, customReplacements map[string]string) string {
	// Standard replacements
	replacements := map[string]string{
		"version": version,
		"Version": version,
		"os":      platform.OS,
		"arch":    platform.Arch,
		"OS":      platform.OS,
		"ARCH":    platform.Arch,
	}

	// Add custom replacements
	for k, v := range customReplacements {
		replacements[k] = v
	}

	// Apply OS-specific naming conventions
	osName := platform.OS
	switch platform.OS {
	case "darwin":
		replacements["os_alt"] = "macos"
		replacements["OS_ALT"] = "macOS"
	case "windows":
		replacements["os_alt"] = "win"
		replacements["OS_ALT"] = "Win"
	default:
		replacements["os_alt"] = osName
		replacements["OS_ALT"] = strings.Title(osName)
	}

	// Apply architecture-specific naming conventions
	archName := platform.Arch
	switch platform.Arch {
	case "amd64":
		replacements["arch_alt"] = "x86_64"
		replacements["ARCH_ALT"] = "x86_64"
	case "386":
		replacements["arch_alt"] = "i386"
		replacements["ARCH_ALT"] = "i386"
	default:
		replacements["arch_alt"] = archName
		replacements["ARCH_ALT"] = archName
	}

	// Replace placeholders
	url := template
	for k, v := range replacements {
		url = strings.ReplaceAll(url, "{{."+k+"}}", v)
		url = strings.ReplaceAll(url, "${"+k+"}", v)
	}

	return url
}



func (h *HTTPBackend) IsRecommended() bool {
	return true
}

func (h *HTTPBackend) IsScriptless() bool {
	return true
}

func (h *HTTPBackend) GetReach() string {
	return "Small"
}

func (h *HTTPBackend) IsStable() bool {
	return true
}

func (h *HTTPBackend) SupportsOffline() bool {
	return false
}
