// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// singleLineCommentRegex matches single-line comments in JSONC
	singleLineCommentRegex = regexp.MustCompile(`(?m)^\s*//.*$|([^\":])\s*//.*$`)
	// multiLineCommentRegex matches multi-line comments in JSONC
	multiLineCommentRegex = regexp.MustCompile(`/\*[\s\S]*?\*/`)
)

// JsonRule validates well-formedness of JSON and JSONC files
type JsonRule struct{}

func NewJsonRule() Rule {
	return &JsonRule{}
}

func (r *JsonRule) ID() string {
	return "builtin:check-json"
}

func (r *JsonRule) Category() Category {
	return CategoryHygiene
}

func (r *JsonRule) Phase() Phase {
	return PhaseLinter
}

func (r *JsonRule) DefaultExtensions() []string {
	return []string{".json", ".jsonc", ".json5"}
}

func (r *JsonRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		ext := strings.ToLower(filepath.Ext(rel))
		if ext != ".json" && ext != ".jsonc" && ext != ".json5" {
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
			// Skip JSON > 5MB
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Check if it is a JSONC / JSON5 file or VS Code / TS config allowing comments
		slashRel := filepath.ToSlash(rel)
		isJsonc := ext == ".jsonc" || ext == ".json5" ||
			strings.Contains(slashRel, ".vscode/") ||
			strings.Contains(slashRel, "tsconfig") ||
			strings.Contains(slashRel, "devcontainer")

		payloadToValidate := content
		if isJsonc {
			payloadToValidate = stripJsonComments(content)
		}

		var raw interface{}
		decoder := json.NewDecoder(bytes.NewReader(payloadToValidate))
		if err := decoder.Decode(&raw); err != nil {
			issues = append(issues, Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityError,
				Message:  fmt.Sprintf("Malformed JSON: %v", err),
				Fixable:  false,
			})
		}
	}

	return issues, nil
}

func stripJsonComments(input []byte) []byte {
	// Strip multi-line comments first
	cleaned := multiLineCommentRegex.ReplaceAll(input, []byte(""))
	// Strip single-line comments
	cleaned = singleLineCommentRegex.ReplaceAll(cleaned, []byte("$1"))
	return cleaned
}
