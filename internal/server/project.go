// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ResolveProjectDir extracts the active project or sub-root directory for the given HTTP request.
// It checks in priority order:
//  1. ?root= / ?project= query parameter
//  2. X-Project-Root header
//  3. Fallback: if current wd does not contain a .unirtm.toml or .git, it scans discovered workspace roots
//     and returns the first sub-project directory that contains a .unirtm.toml (or first sub-root).
//  4. Default: wd
func ResolveProjectDir(r *http.Request) string {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	if r == nil {
		return wd
	}

	rootParam := strings.TrimSpace(r.URL.Query().Get("root"))
	if rootParam == "" {
		rootParam = strings.TrimSpace(r.URL.Query().Get("project"))
	}
	if rootParam == "" {
		rootParam = strings.TrimSpace(r.Header.Get("X-Project-Root"))
	}

	if rootParam != "" {
		// 1. Direct absolute path
		if filepath.IsAbs(rootParam) {
			if fi, err := os.Stat(rootParam); err == nil && fi.IsDir() {
				return rootParam
			}
		}

		// 2. Relative to wd
		rel := filepath.Join(wd, rootParam)
		if fi, err := os.Stat(rel); err == nil && fi.IsDir() {
			return rel
		}

		// 3. Match against workspaceRoots by Name, ID, or Path
		roots := loadWorkspaceRoots(wd)
		for _, rt := range roots {
			if rt.Name == rootParam || rt.ID == rootParam || rt.Path == rootParam {
				target := rt.Path
				if !filepath.IsAbs(target) {
					target = filepath.Join(wd, target)
				}
				if fi, err := os.Stat(target); err == nil && fi.IsDir() {
					return target
				}
			}
		}
	}

	// Fallback when current wd does not contain .unirtm.toml
	if _, err := os.Stat(filepath.Join(wd, ".unirtm.toml")); os.IsNotExist(err) {
		roots := loadWorkspaceRoots(wd)
		// Check for first subproject with .unirtm.toml
		for _, rt := range roots {
			target := rt.Path
			if !filepath.IsAbs(target) {
				target = filepath.Join(wd, target)
			}
			if target != wd && target != "." {
				if _, err := os.Stat(filepath.Join(target, ".unirtm.toml")); err == nil {
					return target
				}
			}
		}
		// If none have .unirtm.toml, pick first non-empty root
		for _, rt := range roots {
			target := rt.Path
			if !filepath.IsAbs(target) {
				target = filepath.Join(wd, target)
			}
			if target != wd && target != "." {
				return target
			}
		}
	}

	return wd
}
