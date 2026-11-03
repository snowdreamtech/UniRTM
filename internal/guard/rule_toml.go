// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

// TomlRule validates well-formedness of TOML files using pure-Go parser
type TomlRule struct{}

func NewTomlRule() Rule {
	return &TomlRule{}
}

func (r *TomlRule) ID() string {
	return "builtin:check-toml"
}

func (r *TomlRule) Category() Category {
	return CategoryHygiene
}

func (r *TomlRule) Phase() Phase {
	return PhaseLinter
}

func (r *TomlRule) DefaultExtensions() []string {
	return []string{".toml"}
}

func (r *TomlRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		if !strings.EqualFold(filepath.Ext(rel), ".toml") {
			continue
		}

		fullPath := rel
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(opts.WorkspaceDir, rel)
		}

		fi, err := os.Lstat(fullPath)
		if err != nil || fi.IsDir() || fi.Mode()&os.ModeSymlink != 0 {
			continue
		}

		if fi.Size() > 5*1024*1024 {
			// Skip TOML > 5MB
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		var raw map[string]interface{}
		if err := toml.Unmarshal(content, &raw); err != nil {
			issues = append(issues, Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityError,
				Message:  fmt.Sprintf("Malformed TOML: %v", err),
				Fixable:  false,
			})
		}
	}

	return issues, nil
}
