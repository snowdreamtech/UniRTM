// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
)

// Provider defines the interface for tool-specific installation and management logic.
// Each provider handles the unique requirements of a specific tool or tool family.
type Provider interface {
	// Name returns the unique identifier for this provider (e.g., "node", "python", "go", "generic").
	Name() string

	// Install performs tool-specific installation steps.
	// This is called after the artifact has been downloaded and extracted.
	// installPath is the directory where the tool should be installed.
	// artifactPath is the path to the downloaded and extracted artifact.
	Install(ctx context.Context, installPath string, artifactPath string, version string) error

	// PostInstall performs any post-installation steps (e.g., setting up virtual environments,
	// installing additional dependencies, configuring tool-specific settings).
	PostInstall(ctx context.Context, installPath string, version string) error

	// GenerateShims generates shim scripts for the tool's executables.
	// Returns a map of executable name to shim script content.
	GenerateShims(installPath string, version string) (map[string]string, error)

	// DetectVersion detects the version of an installed tool.
	// This is used to verify installation and for version management.
	DetectVersion(ctx context.Context, installPath string) (string, error)

	// ListExecutables returns a list of executable names provided by this tool.
	// This is used for shim generation and PATH management.
	ListExecutables(installPath string, version string) ([]string, error)

	// Uninstall performs tool-specific cleanup before uninstallation.
	// This is called before the installation directory is removed.
	Uninstall(ctx context.Context, installPath string, version string) error
}

// ProviderError represents an error from a provider operation.
type ProviderError struct {
	Provider string // The provider that produced the error
	Tool     string // The tool being operated on
	Version  string // The version being operated on
	Message  string // Error message
	Cause    error  // Underlying error, if any
}

// Error implements the error interface.
func (e *ProviderError) Error() string {
	if e.Cause != nil {
		return e.Provider + " provider error for " + e.Tool + " " + e.Version + ": " + e.Message + ": " + e.Cause.Error()
	}
	return e.Provider + " provider error for " + e.Tool + " " + e.Version + ": " + e.Message
}

// Unwrap returns the underlying error.
func (e *ProviderError) Unwrap() error {
	return e.Cause
}

// NewProviderError creates a new ProviderError.
func NewProviderError(provider, tool, version, message string, cause error) *ProviderError {
	return &ProviderError{
		Provider: provider,
		Tool:     tool,
		Version:  version,
		Message:  message,
		Cause:    cause,
	}
}

// ShimConfig contains configuration for shim generation.
type ShimConfig struct {
	ExecutableName string            // Name of the executable
	ExecutablePath string            // Full path to the executable
	Version        string            // Tool version
	Environment    map[string]string // Additional environment variables
}
