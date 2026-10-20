// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DefaultExcludedDirs lists common heavy, dependency, and artifact directories that must be skipped immediately
var DefaultExcludedDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	"dist":         true,
	"build":        true,
	"out":          true,
	"target":       true,
	".venv":        true,
	"venv":         true,
	"env":          true,
	".tmp":         true,
	"tmp":          true,
	".cache":       true,
	"coverage":     true,
	".next":        true,
	".nuxt":        true,
	".output":      true,
	"testdata":     true,
	".specify":     true,
	".agent":       true,
	".agents":      true,
}

// WalkOptions specifies target mode and filtering criteria
type WalkOptions struct {
	WorkspaceDir string
	Mode         string   // "staged", "changed", "all"
	Extensions   []string // e.g. [".go", ".ts"]
	IncludeDirs  []string // e.g. ["src", "cmd"]
	ExcludeDirs  []string // e.g. ["node_modules", "dist"]
	ExcludeFiles []string // glob patterns e.g. ["*.min.js", "*.lock"]
}

// CollectFiles gathers candidate files based on WalkOptions
func CollectFiles(ctx context.Context, opts WalkOptions) ([]string, error) {
	wsDir := opts.WorkspaceDir
	if wsDir == "" {
		wsDir = "."
	}
	absWs, err := filepath.Abs(wsDir)
	if err != nil {
		return nil, err
	}

	mode := strings.ToLower(opts.Mode)
	if mode == "" {
		mode = "changed"
	}

	// Prepare exclude dirs set
	excludeDirMap := make(map[string]bool)
	for k, v := range DefaultExcludedDirs {
		excludeDirMap[k] = v
	}
	for _, ed := range opts.ExcludeDirs {
		ed = strings.TrimSpace(ed)
		if ed != "" {
			excludeDirMap[ed] = true
			excludeDirMap[filepath.Clean(ed)] = true
		}
	}

	var rawFiles []string
	switch mode {
	case "staged":
		rawFiles, _ = getGitStagedFiles(ctx, absWs)
	case "changed":
		rawFiles, _ = getGitChangedFiles(ctx, absWs)
		if len(rawFiles) == 0 && !isGitRepo(absWs) {
			// Fallback to all if not inside a git repository
			return walkAllFiles(absWs, excludeDirMap, opts)
		}
	case "all":
		if isGitRepo(absWs) {
			files, err := getGitAllFiles(ctx, absWs)
			if err == nil && len(files) > 0 {
				rawFiles = files
			} else {
				return walkAllFiles(absWs, excludeDirMap, opts)
			}
		} else {
			return walkAllFiles(absWs, excludeDirMap, opts)
		}
	default:
		rawFiles, _ = getGitChangedFiles(ctx, absWs)
	}

	return filterFiles(absWs, rawFiles, excludeDirMap, opts)
}

func getGitAllFiles(ctx context.Context, dir string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-files", "--cached", "--others", "--exclude-standard")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseLines(out), nil
}

func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	fi, err := os.Stat(gitDir)
	return err == nil && fi.IsDir()
}

func getGitStagedFiles(ctx context.Context, dir string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--name-only", "--diff-filter=ACMR")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseLines(out), nil
}

func getGitChangedFiles(ctx context.Context, dir string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain", "-uall")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var files []string
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if len(line) < 4 {
			continue
		}
		statusCode := line[:2]
		if statusCode == "D " || statusCode == " D" || statusCode == "DD" {
			// Deleted file, skip
			continue
		}
		filePath := strings.TrimSpace(line[3:])
		// Handle rename format: "old -> new"
		if idx := strings.Index(filePath, " -> "); idx != -1 {
			filePath = filePath[idx+4:]
		}
		// Strip quotes if any
		filePath = strings.Trim(filePath, "\"")
		if filePath != "" {
			files = append(files, filePath)
		}
	}
	return files, nil
}

func parseLines(b []byte) []string {
	var res []string
	lines := strings.Split(string(b), "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			res = append(res, l)
		}
	}
	return res
}

func walkAllFiles(absWs string, excludeDirMap map[string]bool, opts WalkOptions) ([]string, error) {
	var collected []string

	err := filepath.Walk(absWs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if excludeDirMap[base] {
				return filepath.SkipDir
			}
			rel, _ := filepath.Rel(absWs, path)
			if rel != "." && excludeDirMap[rel] {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(absWs, path)
		if err != nil {
			return nil
		}
		collected = append(collected, rel)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return filterFiles(absWs, collected, excludeDirMap, opts)
}

func filterFiles(absWs string, files []string, excludeDirMap map[string]bool, opts WalkOptions) ([]string, error) {
	// Prepare extensions map
	extMap := make(map[string]bool)
	for _, ext := range opts.Extensions {
		ext = strings.ToLower(strings.TrimSpace(ext))
		if ext != "" {
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			extMap[ext] = true
		}
	}

	var results []string
	for _, rel := range files {
		rel = filepath.Clean(rel)
		fullPath := filepath.Join(absWs, rel)

		// Check if file exists on disk
		fi, err := os.Lstat(fullPath)
		if err != nil || fi.IsDir() {
			continue
		}

		// Check directory exclusion along relative path parts
		parts := strings.Split(filepath.ToSlash(rel), "/")
		skip := false
		for _, part := range parts[:len(parts)-1] {
			if excludeDirMap[part] {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Check IncludeDirs if specified
		if len(opts.IncludeDirs) > 0 {
			matchedInclude := false
			for _, id := range opts.IncludeDirs {
				idClean := filepath.Clean(id)
				if rel == idClean || strings.HasPrefix(rel, idClean+string(filepath.Separator)) || strings.HasPrefix(rel, idClean+"/") {
					matchedInclude = true
					break
				}
			}
			if !matchedInclude {
				continue
			}
		}

		// Check Extensions if specified
		if len(extMap) > 0 {
			ext := strings.ToLower(filepath.Ext(rel))
			if !extMap[ext] {
				continue
			}
		}

		// Check ExcludeFiles globs
		if isExcludedFile(rel, opts.ExcludeFiles) {
			continue
		}

		results = append(results, rel)
	}

	return results, nil
}

func isExcludedFile(rel string, globs []string) bool {
	base := filepath.Base(rel)
	slashRel := filepath.ToSlash(rel)
	for _, pattern := range globs {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		// Match against base name (e.g. *.lock)
		if m, _ := filepath.Match(pattern, base); m {
			return true
		}
		// Match against relative path (e.g. internal/**)
		if m, _ := filepath.Match(pattern, slashRel); m {
			return true
		}
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "/**")
			if strings.HasPrefix(slashRel, prefix+"/") || slashRel == prefix {
				return true
			}
		}
	}
	return false
}
