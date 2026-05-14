// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"runtime"
)

// Platform represents the operating system and architecture information.
type Platform struct {
	OS   string // e.g., "linux", "darwin", "windows"
	Arch string // e.g., "amd64", "arm64", "386"
}

// CurrentPlatform returns the platform information for the current system.
func CurrentPlatform() Platform {
	return Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// String returns a string representation of the platform (e.g., "linux-amd64").
func (p Platform) String() string {
	return p.OS + "-" + p.Arch
}

// VersionInfo contains metadata about a specific tool version.
type VersionInfo struct {
	Version      string            // The version string (e.g., "1.20.0")
	DownloadURL  string            // URL to download the artifact
	Checksum     string            // SHA-256 checksum of the artifact
	SignatureURL string            // URL to download the GPG signature (.asc, .sig)
	GPGSignature string            // Raw GPG signature content
	GPGKeys      []string          // Trusted GPG fingerprints for this version
	Platform     Platform          // Target platform for this artifact
	Metadata     map[string]string // Additional metadata (e.g., release date, notes)
}

// Backend defines the interface for tool version management backends.
// Each backend is responsible for listing available versions, resolving
// version requests, and providing download information for tools.
type Backend interface {
	// Name returns the unique identifier for this backend (e.g., "github", "aqua", "http").
	Name() string

	// ListVersions returns all available versions for a tool.
	// The versions are returned in descending order (newest first).
	ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error)

	// ResolveVersion resolves a version request (e.g., "latest", "1.20", "^1.19") to a concrete version.
	// Returns the resolved VersionInfo or an error if the version cannot be resolved.
	ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error)

	// GetDownloadInfo retrieves download information for a specific tool version.
	// This is useful when you already know the exact version you want.
	GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error)

	// SupportsChecksum indicates whether this backend provides checksum verification.
	SupportsChecksum() bool

	// SupportsGPG indicates whether this backend supports GPG signature verification.
	SupportsGPG() bool

	// AttestationType returns the type of build provenance or attestation 
	// verification supported (e.g., "SLSA", "GitHub", "Cosign").
	// Returns an empty string if not supported.
	AttestationType() string

	// IsRecommended indicates whether this backend is recommended for use.
	IsRecommended() bool

	// IsScriptless indicates whether this backend performs installation without executing arbitrary scripts.
	IsScriptless() bool

	// GetReach returns the relative coverage/reach of this backend (e.g., "Small", "Medium", "Large", "Huge").
	GetReach() string

	// IsStable indicates whether this backend has high stability and low bit-rot.
	IsStable() bool

	// SupportsOffline indicates whether this backend supports offline mode or private mirrors.
	SupportsOffline() bool

	// Dependencies returns a list of backend names or tools that this backend depends on.
	// This is used to determine the installation order.
	Dependencies() []string
}

// BackendError represents an error from a backend operation.
type BackendError struct {
	Backend string // The backend that produced the error
	Tool    string // The tool being operated on
	Message string // Error message
	Cause   error  // Underlying error, if any
}

// Error implements the error interface.
func (e *BackendError) Error() string {
	if e.Cause != nil {
		return e.Backend + " backend error for " + e.Tool + ": " + e.Message + ": " + e.Cause.Error()
	}
	return e.Backend + " backend error for " + e.Tool + ": " + e.Message
}

// Unwrap returns the underlying error.
func (e *BackendError) Unwrap() error {
	return e.Cause
}

// NewBackendError creates a new BackendError.
func NewBackendError(backend, tool, message string, cause error) *BackendError {
	return &BackendError{
		Backend: backend,
		Tool:    tool,
		Message: message,
		Cause:   cause,
	}
}
