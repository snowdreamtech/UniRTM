// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// YamlRule validates well-formedness of YAML files using pure-Go parser
type YamlRule struct{}

func NewYamlRule() Rule {
	return &YamlRule{}
}

func (r *YamlRule) ID() string {
	return "builtin:check-yaml"
}

func (r *YamlRule) Category() Category {
	return CategoryHygiene
}

func (r *YamlRule) Phase() Phase {
	return PhaseLinter
}

func (r *YamlRule) DefaultExtensions() []string {
	return []string{".yaml", ".yml"}
}

func (r *YamlRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		ext := strings.ToLower(filepath.Ext(rel))
		if ext != ".yaml" && ext != ".yml" {
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
			// Skip YAML > 5MB
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Empty YAML is valid
		if len(bytes.TrimSpace(content)) == 0 {
			continue
		}

		dec := yaml.NewDecoder(bytes.NewReader(content))
		for {
			var doc interface{}
			err := dec.Decode(&doc)
			if err == io.EOF {
				break
			}
			if err != nil {
				issues = append(issues, Issue{
					FilePath: rel,
					Line:     1,
					RuleID:   r.ID(),
					Category: r.Category(),
					Phase:    r.Phase(),
					Severity: SeverityError,
					Message:  fmt.Sprintf("Malformed YAML: %v", err),
					Fixable:  false,
				})
				break
			}
		}
	}

	return issues, nil
}
