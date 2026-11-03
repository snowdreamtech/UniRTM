// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	svgScriptTagRegex   = regexp.MustCompile(`(?i)<\s*script[\s>]`)
	svgInlineEventRegex = regexp.MustCompile(`(?i)\son\w+\s*=`)
	svgJavascriptUri    = regexp.MustCompile(`(?i)(href|xlink:href)\s*=\s*["']\s*javascript:`)
)

type SvgSafetyRule struct{}

func NewSvgSafetyRule() Rule {
	return &SvgSafetyRule{}
}

func (r *SvgSafetyRule) ID() string {
	return "builtin:check-svg-safety"
}

func (r *SvgSafetyRule) Category() Category {
	return CategorySecurity
}

func (r *SvgSafetyRule) Phase() Phase {
	return PhaseLinter
}

func (r *SvgSafetyRule) DefaultExtensions() []string {
	return []string{".svg"}
}

func (r *SvgSafetyRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		if !strings.EqualFold(filepath.Ext(rel), ".svg") {
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
			// Skip SVGs > 5MB
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		var reason string
		if svgScriptTagRegex.Match(content) {
			reason = "contains <script> tag"
		} else if svgInlineEventRegex.Match(content) {
			reason = "contains inline event handler (on* attribute)"
		} else if svgJavascriptUri.Match(content) {
			reason = "contains javascript: URI in link"
		} else if bytes.Contains(bytes.ToLower(content), []byte("script")) && bytes.Contains(content, []byte("<")) {
			// Suspicious obfuscated script tag
			if bytes.Contains(bytes.ToLower(content), []byte("<script")) {
				reason = "contains script element"
			}
		}

		if reason != "" {
			issues = append(issues, Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityError,
				Message:  "Unsafe SVG detected (potential XSS script injection: " + reason + ")",
				Fixable:  false,
			})
		}
	}

	return issues, nil
}
