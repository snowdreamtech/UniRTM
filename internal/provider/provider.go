// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"net/url"
	"os"
	"strings"
)

// Provider defines the interface for tool-specific installation and management logic.
// Each provider handles the unique requirements of a specific tool or tool family.
type Provider interface {
	// Name returns the unique identifier for this provider (e.g., "node", "python", "go", "generic").
	Name() string

	// Install performs tool-specific installation steps.
	// This is called after the artifact has been downloaded and extracted.
	// tool is the name of the tool being installed.
	// installPath is the directory where the tool should be installed.
	// artifactPath is the path to the downloaded and extracted artifact.
	Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error

	// PostInstall performs any post-installation steps.
	// tool is the name of the tool being installed.
	PostInstall(ctx context.Context, tool string, installPath string, version string) error

	// GenerateShims generates shim scripts for the tool's executables.
	// Returns a map of executable name to shim script content.
	GenerateShims(tool string, installPath string, version string) (map[string]string, error)

	// DetectVersion detects the version of an installed tool.
	// This is used to verify installation and for version management.
	DetectVersion(ctx context.Context, tool string, installPath string) (string, error)

	// ListExecutables returns a list of executable names provided by this tool.
	// This is used for shim generation and PATH management.
	ListExecutables(tool string, installPath string, version string) ([]string, error)

	// GetBinPaths returns a list of absolute paths to the directories containing
	// the tool's executables. These are used in PATH mode activation.
	GetBinPaths(tool string, installPath string, version string) ([]string, error)

	// GetEnvVars returns a map of environment variables that should be set
	// when this tool is active (e.g., JAVA_HOME for Java).
	GetEnvVars(tool string, installPath string, version string) (map[string]string, error)

	// Uninstall performs tool-specific cleanup before uninstallation.
	// This is called before the installation directory is removed.
	Uninstall(ctx context.Context, tool string, installPath string, version string) error
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
// GlobalNoProxy is a list of additional domains to skip proxy for, typically loaded from configuration.
var GlobalNoProxy []string

// GetNoProxyEnv returns the system environment with common mirror domains added to NO_PROXY.
// This helps prevent installation failures when using mirrors behind a proxy.
// It also accepts additional domains (e.g. dynamically extracted from mirror URLs).
func GetNoProxyEnv(extraDomains ...string) []string {
	env := os.Environ()
	
	// Collect all domains to skip proxy for
	domains := []string{
		"mirrors.aliyun.com",
		"npm.aliyun.com",
		"maven.aliyun.com",
		"registry.npmmirror.com",
		"mirrors.cloud.tencent.com",
		"mirrors.huaweicloud.com",
		"repo.huaweicloud.com",
		"mirrors.163.com",
		"mirrors.ustc.edu.cn",
		"pypi.mirrors.ustc.edu.cn",
		"mirrors.tuna.tsinghua.edu.cn",
		"pypi.tuna.tsinghua.edu.cn",
		"pypi.douban.com",
		"registry.taobao.org",
		"npm.taobao.org",
		"rsproxy.cn",
		"r.cnpmjs.org",
		"goproxy.cn",
		"goproxy.io",
		"gems.ruby-china.com",
		"sn0wdr1am.com",
	}
	
	// Add global configuration domains
	domains = append(domains, GlobalNoProxy...)
	
	// Add dynamically provided domains
	domains = append(domains, extraDomains...)
	
	// Remove duplicates and empty strings
	domainMap := make(map[string]bool)
	var finalDomains []string
	for _, d := range domains {
		d = strings.TrimSpace(d)
		if d != "" && !domainMap[d] {
			domainMap[d] = true
			finalDomains = append(finalDomains, d)
		}
	}
	
	mirrors := strings.Join(finalDomains, ",")
	
	found := false
	for i, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "NO_PROXY=") {
			// Extract the current value (case-insensitive search but preserve original case for value if needed)
			// Actually NO_PROXY value is just a string.
			parts := strings.SplitN(e, "=", 2)
			current := ""
			if len(parts) > 1 {
				current = parts[1]
			}
			
			if current != "" {
				env[i] = "NO_PROXY=" + current + "," + mirrors
			} else {
				env[i] = "NO_PROXY=" + mirrors
			}
			found = true
			break
		}
	}
	
	if !found {
		env = append(env, "NO_PROXY="+mirrors)
	}
	return env
}

// DomainFromURL extracts the domain name from a URL string.
func DomainFromURL(u string) string {
	if u == "" {
		return ""
	}
	// Add scheme if missing for parsing
	if !strings.Contains(u, "://") {
		u = "http://" + u
	}
	
	importUrl, err := url.Parse(u)
	if err != nil {
		return ""
	}
	
	host := importUrl.Host
	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host
}
