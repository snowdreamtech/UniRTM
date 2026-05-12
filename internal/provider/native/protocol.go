// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
)

// Asset represents a downloadable file for a specific version.
type Asset struct {
	URL          string
	Filename     string
	OS           string
	Arch         string
	Checksum     string
	Algo         string // sha256, sha1, etc.
	SignatureURL string // URL to the GPG signature (.asc, .sig)
}

// VersionInfo represents a tool version and its associated assets.
type VersionInfo struct {
	Version string
	IsLTS   bool
	LTSName string
	Assets  []Asset
}

// ProtocolHandler defines the interface for parsing upstream tool metadata.
type ProtocolHandler interface {
	// Name returns the identifier of the protocol handler.
	Name() string
	// ResolveVersions fetches and parses the upstream metadata to get all available versions.
	ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error)
}

// Recipe defines the configuration for a native tool.
type Recipe struct {
	ID      string
	Handler ProtocolHandler
	BaseURL string
	Aliases map[string]string
	GPGKeys []string // List of trusted GPG fingerprints for this tool
}
