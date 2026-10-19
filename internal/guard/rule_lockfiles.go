// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

)

// LockfileCoexistenceRule ensures only one Node.js package manager lockfile exists per folder
type LockfileCoexistenceRule struct{}

func NewLockfileCoexistenceRule() Rule {
	return &LockfileCoexistenceRule{}
}

func (r *LockfileCoexistenceRule) ID() string {
	return "builtin:check-lockfile-coexistence"
}

func (r *LockfileCoexistenceRule) Category() Category {
	return CategoryHygiene
}

func (r *LockfileCoexistenceRule) Phase() Phase {
	return PhaseLinter
}

func (r *LockfileCoexistenceRule) DefaultExtensions() []string {
	return nil
}

func (r *LockfileCoexistenceRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue
	// Group node lockfiles by directory
	dirLockfiles := make(map[string][]string)

	knownNodeLockfiles := map[string]bool{
		"package-lock.json": true,
		"pnpm-lock.yaml":    true,
		"yarn.lock":         true,
		"bun.lockb":         true,
	}

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		base := filepath.Base(rel)
		if knownNodeLockfiles[base] {
			dir := filepath.Dir(rel)
			dirLockfiles[dir] = append(dirLockfiles[dir], base)
		}
	}

	for dir, lockfiles := range dirLockfiles {
		// Unique check
		unique := make(map[string]bool)
		for _, lf := range lockfiles {
			unique[lf] = true
		}
		if len(unique) > 1 {
			var names []string
			for k := range unique {
				names = append(names, k)
			}
			issues = append(issues, Issue{
				FilePath: filepath.Join(dir, names[0]),
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityError,
				Message: fmt.Sprintf("Multiple conflicting Node.js lockfiles found in %q: %v. Keep only one.",
					dir, names),
				Fixable: false,
			})
		}
	}

	return issues, nil
}

// LockfileConflictsRule checks conflict markers in large lockfiles without line-length limits
type LockfileConflictsRule struct{}

func NewLockfileConflictsRule() Rule {
	return &LockfileConflictsRule{}
}

func (r *LockfileConflictsRule) ID() string {
	return "builtin:check-lockfile-conflicts"
}

func (r *LockfileConflictsRule) Category() Category {
	return CategoryHygiene
}

func (r *LockfileConflictsRule) Phase() Phase {
	return PhaseLinter
}

func (r *LockfileConflictsRule) DefaultExtensions() []string {
	return nil
}

func (r *LockfileConflictsRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	isLockfile := func(path string) bool {
		lower := strings.ToLower(path)
		return strings.HasSuffix(lower, "lock.json") ||
			strings.HasSuffix(lower, "lock.yaml") ||
			strings.HasSuffix(lower, "cargo.lock") ||
			strings.HasSuffix(lower, "go.sum") ||
			strings.HasSuffix(lower, "yarn.lock") ||
			strings.HasSuffix(lower, "bun.lockb")
	}

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		if !isLockfile(rel) {
			continue
		}

		fullPath := rel; if !filepath.IsAbs(fullPath) { fullPath = filepath.Join(opts.WorkspaceDir, rel) }
		f, err := os.Open(fullPath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			if strings.HasPrefix(line, "<<<<<<< ") ||
				strings.HasPrefix(line, "=======") && len(strings.TrimSpace(line)) == 7 ||
				strings.HasPrefix(line, ">>>>>>> ") {
				issues = append(issues, Issue{
					FilePath: rel,
					Line:     lineNum,
					RuleID:   r.ID(),
					Category: r.Category(),
					Phase:    r.Phase(),
					Severity: SeverityError,
					Message:  "Lockfile contains unresolved Git conflict marker: " + strings.TrimSpace(line),
					Fixable:  false,
				})
				break
			}
		}
		f.Close()
	}

	return issues, nil
}
