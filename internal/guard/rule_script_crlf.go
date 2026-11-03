// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ScriptCrlfRule detects and automatically repairs CRLF line endings in executable scripts
// to prevent "bad interpreter" errors on Linux and in CI environments.
type ScriptCrlfRule struct {
	scriptExtensions map[string]bool
}

func NewScriptCrlfRule() Rule {
	exts := []string{".sh", ".bash", ".zsh", ".bats", ".ksh", ".csh"}
	m := make(map[string]bool, len(exts))
	for _, ext := range exts {
		m[ext] = true
	}
	return &ScriptCrlfRule{
		scriptExtensions: m,
	}
}

func (r *ScriptCrlfRule) ID() string {
	return "builtin:forbid-crlf-in-scripts"
}

func (r *ScriptCrlfRule) Category() Category {
	return CategoryHygiene
}

func (r *ScriptCrlfRule) Phase() Phase {
	return PhaseFixer
}

func (r *ScriptCrlfRule) DefaultExtensions() []string {
	return []string{".sh", ".bash", ".zsh", ".bats", ".ksh", ".csh"}
}

func (r *ScriptCrlfRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue
	maxSize := opts.MaxFileSize
	if maxSize <= 0 {
		maxSize = 2 * 1024 * 1024 // 2MB
	}

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		fullPath := rel
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(opts.WorkspaceDir, rel)
		}

		fi, err := os.Lstat(fullPath)
		if err != nil || fi.IsDir() || fi.Mode()&os.ModeSymlink != 0 || fi.Size() > maxSize {
			continue
		}

		ext := strings.ToLower(filepath.Ext(rel))
		isScript := r.scriptExtensions[ext]

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Also check files without script extension if they start with a shebang line
		if !isScript && len(content) >= 2 && bytes.HasPrefix(content, []byte("#!")) {
			isScript = true
		}

		if !isScript {
			continue
		}

		crlfCount := bytes.Count(content, []byte("\r\n"))
		if crlfCount > 0 {
			fixed := false
			if opts.Fix && !opts.DryRun {
				normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
				if err := os.WriteFile(fullPath, normalized, fi.Mode().Perm()); err == nil {
					fixed = true
				}
			}

			issues = append(issues, Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityError,
				Message:  fmt.Sprintf("Script contains %d CRLF line endings (\\r\\n), which causes '/bin/sh^M: bad interpreter' on Linux/CI. Auto-fixable to LF.", crlfCount),
				Fixable:  true,
				Fixed:    fixed,
			})
		}
	}

	return issues, nil
}
