// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package lockfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate_Errors(t *testing.T) {
	lf := &LockFile{
		Tools: map[string][]*ToolLockEntry{
			"empty": {},
			"bad_version": {
				{Version: ""},
				{Version: "1.0"},
				{Version: "1.0"}, // duplicate
			},
			"bad_platform": {
				{
					Version: "2.0",
					Platforms: map[string]*PlatformEntry{
						"invalid-key":   {},
						"linux-amd64":   nil,
						"windows-amd64": {Checksum: "invalid", Size: -1, URL: "ftp://bad"},
					},
				},
			},
		},
	}

	err := lf.Validate()
	assert.Error(t, err)

	ve, ok := err.(*ValidationError)
	assert.True(t, ok)

	errMsg := ve.Error()
	assert.Contains(t, errMsg, "tool \"empty\": empty entry list")
	assert.Contains(t, errMsg, "tool \"bad_version\": entry has empty version")
	assert.Contains(t, errMsg, "tool \"bad_version\": duplicate version \"1.0\"")
	assert.Contains(t, errMsg, "unknown platform key \"invalid-key\"")
	assert.Contains(t, errMsg, "nil entry")
	assert.Contains(t, errMsg, "must be a valid hex string or start with a supported algorithm prefix (e.g., sha256:)")
	assert.Contains(t, errMsg, "size must be ≥ 0, got -1")
	assert.Contains(t, errMsg, "url \"ftp://bad\" does not look like a valid HTTP URL")
}

func TestCheckStrict_Errors(t *testing.T) {
	lf := &LockFile{
		Tools: map[string][]*ToolLockEntry{
			"go": {
				{
					Version: "1.20",
					Backend: "go",
					Platforms: map[string]*PlatformEntry{
						"linux-amd64": {},
					},
				},
			},
			"github:cli/cli": {
				{
					Version: "2.0",
					Backend: "github",
					Platforms: map[string]*PlatformEntry{
						"linux-amd64": {URL: ""},
					},
				},
			},
		},
	}

	// Missing tool
	err := lf.CheckStrict([]LockRequirement{
		{ToolKey: "missing", Version: "1.0"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no locked entry for tool=\"missing\" version=\"1.0\"")

	// Missing platform
	err = lf.CheckStrict([]LockRequirement{
		{ToolKey: "go", Version: "1.20", PlatformKey: "darwin-arm64"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no locked platform \"darwin-arm64\" for tool=\"go\"")

	// Missing URL for backend that requires it
	err = lf.CheckStrict([]LockRequirement{
		{ToolKey: "github:cli/cli", Version: "2.0", PlatformKey: "linux-amd64"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no locked URL for tool=\"github:cli/cli\"")

	// Valid URL for backend that doesn't require it
	err = lf.CheckStrict([]LockRequirement{
		{ToolKey: "go", Version: "1.20", PlatformKey: "linux-amd64"},
	})
	assert.NoError(t, err)
}

func TestValidate_EmptyURLForBinaryBackend(t *testing.T) {
	lf := &LockFile{
		Tools: map[string][]*ToolLockEntry{
			"github:cli/cli": {
				{
					Version: "2.72.0",
					Backend: "github",
					Platforms: map[string]*PlatformEntry{
						"linux-amd64": {URL: ""},
					},
				},
			},
			"npm:prettier": {
				{
					Version: "3.9.6",
					Backend: "npm",
					Platforms: map[string]*PlatformEntry{
						"linux-amd64": {URL: ""},
					},
				},
			},
		},
	}

	err := lf.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "URL is empty for binary download backend")
}

func TestBackendNeedsURL(t *testing.T) {
	// Package manager / ecosystem backends (no URL or checksum required)
	assert.False(t, BackendNeedsURL("go:golang.org/x/vuln/cmd/govulncheck", "go-pkg"))
	assert.False(t, BackendNeedsURL("go:golang.org/x/tools/cmd/goimports", "go"))
	assert.False(t, BackendNeedsURL("go:mvdan.cc/sh/v3/cmd/shfmt", ""))
	assert.False(t, BackendNeedsURL("npm:prettier", "npm"))
	assert.False(t, BackendNeedsURL("npm:prettier", ""))
	assert.False(t, BackendNeedsURL("npm:@commitlint/cli", ""))
	assert.False(t, BackendNeedsURL("npm:@commitlint/config-conventional", "npm"))
	assert.False(t, BackendNeedsURL("pipx:clang-format", "pipx"))
	assert.False(t, BackendNeedsURL("pipx:clang-format", ""))
	assert.False(t, BackendNeedsURL("pipx:pre-commit", "pipx"))
	assert.False(t, BackendNeedsURL("cargo:ripgrep", "cargo"))
	assert.False(t, BackendNeedsURL("cargo:ripgrep", ""))
	assert.False(t, BackendNeedsURL("gem:rubocop", "gem"))
	assert.False(t, BackendNeedsURL("composer:squizlabs/php_codesniffer", "composer"))
	assert.False(t, BackendNeedsURL("composer:squizlabs/php_codesniffer", ""))
	assert.False(t, BackendNeedsURL("pub:stagehand", "pub"))
	assert.False(t, BackendNeedsURL("docker:ghcr.io/aquasec/trivy", "docker"))
	assert.False(t, BackendNeedsURL("docker:ghcr.io/aquasec/trivy", ""))
	assert.False(t, BackendNeedsURL("podman:alpine", ""))
	assert.False(t, BackendNeedsURL("deno:fmt", "deno"))
	assert.False(t, BackendNeedsURL("dotnet:csharp-ls", "dotnet"))
	assert.False(t, BackendNeedsURL("cabal:hlint", "cabal"))
	assert.False(t, BackendNeedsURL("lua:inspect", "lua"))
	assert.False(t, BackendNeedsURL("luarocks:luacheck", "luarocks"))

	// Binary download backends (URL and checksum required)
	assert.True(t, BackendNeedsURL("github:cli/cli", "github"))
	assert.True(t, BackendNeedsURL("github:cli/cli", ""))
	assert.True(t, BackendNeedsURL("cli/cli", ""))
	assert.True(t, BackendNeedsURL("gitlab:gitlab-org/cli", "gitlab"))
	assert.True(t, BackendNeedsURL("forgejo:forgejo/forgejo", "forgejo"))
	assert.True(t, BackendNeedsURL("http:example.com/tool.tar.gz", "http"))
	assert.True(t, BackendNeedsURL("https:example.com/tool.tar.gz", ""))
	assert.True(t, BackendNeedsURL("s3:mybucket/mytool", "s3"))
	assert.True(t, BackendNeedsURL("aqua:aquaproj/aqua", "aqua"))

	// Verify BackendNeedsChecksum matches BackendNeedsURL
	assert.False(t, BackendNeedsChecksum("npm:prettier", "npm"))
	assert.False(t, BackendNeedsChecksum("go:golang.org/x/vuln/cmd/govulncheck", "go"))
	assert.False(t, BackendNeedsChecksum("pipx:clang-format", "pipx"))
	assert.True(t, BackendNeedsChecksum("github:cli/cli", "github"))
}

func TestCheckStrict_EcosystemNoURLNoChecksum(t *testing.T) {
	lf := &LockFile{
		Tools: map[string][]*ToolLockEntry{
			"npm:prettier": {
				{
					Version: "3.9.9",
					Backend: "npm",
					Platforms: map[string]*PlatformEntry{
						"linux-amd64": {URL: "", Checksum: ""},
					},
				},
			},
			"go:golang.org/x/vuln/cmd/govulncheck": {
				{
					Version: "v1.8.0",
					Backend: "go",
					Platforms: map[string]*PlatformEntry{
						"linux-amd64": {URL: "", Checksum: ""},
					},
				},
			},
			"pipx:clang-format": {
				{
					Version: "23.1.1",
					Backend: "pipx",
					Platforms: map[string]*PlatformEntry{
						"linux-amd64": {URL: "", Checksum: ""},
					},
				},
			},
		},
	}

	// In strict mode, npm/go/pipx entries with empty URL and empty Checksum MUST pass!
	err := lf.CheckStrict([]LockRequirement{
		{ToolKey: "npm:prettier", Version: "3.9.9", PlatformKey: "linux-amd64"},
		{ToolKey: "go:golang.org/x/vuln/cmd/govulncheck", Version: "v1.8.0", PlatformKey: "linux-amd64"},
		{ToolKey: "pipx:clang-format", Version: "23.1.1", PlatformKey: "linux-amd64"},
	})
	assert.NoError(t, err)
}
