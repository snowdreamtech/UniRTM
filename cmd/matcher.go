// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/service"
)

// isInstallationMatch checks if an installation matches a user query (name, alias, executable, or canonical key).
func isInstallationMatch(inst *repository.Installation, query string, targetCanonical string, cfg *config.Config) bool {
	qClean := strings.ToLower(query)
	instToolClean := strings.ToLower(inst.Tool)

	// 1. Direct match with query or canonical key
	if instToolClean == qClean || inst.Tool == query {
		return true
	}
	if targetCanonical != "" {
		tcClean := strings.ToLower(targetCanonical)
		if instToolClean == tcClean || inst.Tool == targetCanonical {
			return true
		}
		// Match stripped prefix: e.g. targetCanonical is "go:mvdan.cc/sh/v3/cmd/shfmt" -> "mvdan.cc/sh/v3/cmd/shfmt"
		if idx := strings.Index(targetCanonical, ":"); idx >= 0 {
			noPrefix := strings.ToLower(targetCanonical[idx+1:])
			if instToolClean == noPrefix {
				return true
			}
		}
		// Match backend-prefixed inst.Tool: e.g. inst.Backend + ":" + inst.Tool == targetCanonical
		if strings.ToLower(inst.Backend+":"+inst.Tool) == tcClean {
			return true
		}
	}

	// 2. Binary / base name match: e.g. "mvdan.cc/sh/v3/cmd/shfmt" -> "shfmt"
	if strings.EqualFold(filepath.Base(inst.Tool), qClean) {
		return true
	}
	if strings.EqualFold(provider.ResolveToolBinaryName(inst.Tool), qClean) {
		return true
	}

	// 3. Config alias match
	if cfg != nil {
		if tc, ok := cfg.Tools[inst.Tool]; ok && tc.Alias != "" {
			if strings.EqualFold(tc.Alias, qClean) {
				return true
			}
		}
		if tc, ok := cfg.Tools[inst.Backend+":"+inst.Tool]; ok && tc.Alias != "" {
			if strings.EqualFold(tc.Alias, qClean) {
				return true
			}
		}
		if _, tc, ok := cfg.FindToolConfig(inst.Tool); ok && tc.Alias != "" {
			if strings.EqualFold(tc.Alias, qClean) {
				return true
			}
		}
	}

	// 4. Standard executable aliases or provider executables
	for _, alias := range service.GetExecutableAliases(inst.Tool) {
		if strings.EqualFold(alias, qClean) {
			return true
		}
	}

	// 5. Provider executables match
	if p := provider.DefaultRegistry.GetWithBackend(inst.Tool, inst.Backend); p != nil {
		if execs, err := p.ListExecutables(inst.Tool, inst.InstallPath, inst.Version); err == nil {
			for _, exe := range execs {
				if strings.EqualFold(filepath.Base(exe), qClean) {
					return true
				}
			}
		}
	}

	return false
}

// isToolKeyMatch checks if a config tool key matches a user filter or canonical target.
func isToolKeyMatch(key string, filter string, targetCanonical string, cfg *config.Config) bool {
	if strings.EqualFold(key, filter) || strings.EqualFold(key, targetCanonical) {
		return true
	}
	kNoPrefix := key
	if idx := strings.Index(key, ":"); idx >= 0 {
		kNoPrefix = key[idx+1:]
	}
	fNoPrefix := filter
	if idx := strings.Index(filter, ":"); idx >= 0 {
		fNoPrefix = filter[idx+1:]
	}
	if strings.EqualFold(kNoPrefix, fNoPrefix) {
		return true
	}
	if strings.EqualFold(filepath.Base(kNoPrefix), fNoPrefix) {
		return true
	}
	if cfg != nil {
		if tc, ok := cfg.Tools[key]; ok && tc.Alias != "" {
			if strings.EqualFold(tc.Alias, filter) {
				return true
			}
		}
	}
	return false
}

// resolveToolArg resolves a tool argument (which might be an alias, short name, or backend:tool@version)
// against configuration and parses it into (backendName, toolName, version, explicitVersion).
func resolveToolArg(im *service.InstallationManager, cfg *config.Config, input string) (backendName, toolName, version string, explicitVersion bool) {
	if im == nil {
		return "", input, "", false
	}
	bName, tName, ver, explicit := im.ParseToolSpec(input)
	backendName = bName
	toolName = tName
	version = ver
	explicitVersion = explicit

	if cfg != nil {
		if canonicalKey, tc, found := cfg.FindToolConfig(toolName); found {
			cBackend, cTool, cVer, cExplicit := im.ParseToolSpec(canonicalKey)
			if cBackend != "" {
				backendName = cBackend
			}
			toolName = cTool
			if !explicitVersion {
				if tc.Version != "" {
					version = tc.Version
					explicitVersion = true
				} else if cExplicit && cVer != "" {
					version = cVer
					explicitVersion = true
				}
			}
			if backendName == "" && tc.Backend != "" {
				backendName = tc.Backend
			}
		}
	}

	if backendName == "" && toolName != "" {
		backendName = im.AutoDetectBackend(toolName)
	}

	return backendName, toolName, version, explicitVersion
}
