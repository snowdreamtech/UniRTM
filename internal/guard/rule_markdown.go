// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// atxHeadingRegex matches Markdown ATX headings like "# Title" or "### Subtitle"
	atxHeadingRegex = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	// markdownLinkRegex matches [text](url) and ![alt](url)
	markdownLinkRegex = regexp.MustCompile(`!?\[([^\]]*)\]\(([^)]*)\)`)
	// codeFenceRegex matches ``` or ~~~ code fences
	codeFenceRegex = regexp.MustCompile(`^(\s*)(` + "```" + `+|~~~+)(.*)$`)
)

// MarkdownRule validates Markdown file structure, code fence balance, heading hierarchy, and link integrity
type MarkdownRule struct{}

// NewMarkdownRule returns a new Markdown validation rule instance
func NewMarkdownRule() Rule {
	return &MarkdownRule{}
}

func (r *MarkdownRule) ID() string {
	return "builtin:check-markdown"
}

func (r *MarkdownRule) Category() Category {
	return CategoryLint
}

func (r *MarkdownRule) Phase() Phase {
	return PhaseLinter
}

func (r *MarkdownRule) DefaultExtensions() []string {
	return []string{".md", ".markdown", ".mdown", ".mkd"}
}

func (r *MarkdownRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		ext := strings.ToLower(filepath.Ext(rel))
		if ext != ".md" && ext != ".markdown" && ext != ".mdown" && ext != ".mkd" {
			continue
		}

		fullPath := rel
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(opts.WorkspaceDir, rel)
		}

		data, err := os.ReadFile(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			issues = append(issues, Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityError,
				Message:  fmt.Sprintf("Failed to read Markdown file: %v", err),
				Fixable:  false,
			})
			continue
		}

		// Skip empty files
		if len(bytes.TrimSpace(data)) == 0 {
			continue
		}

		fileIssues := r.validateMarkdown(rel, data)
		issues = append(issues, fileIssues...)
	}

	return issues, nil
}

func (r *MarkdownRule) validateMarkdown(rel string, data []byte) []Issue {
	var issues []Issue
	scanner := bufio.NewScanner(bytes.NewReader(data))

	inCodeBlock := false
	codeFenceChar := ""
	codeFenceLen := 0
	codeFenceStartLine := 0
	inHtmlComment := false

	prevHeadingLevel := 0
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// 1. Handle HTML comment blocks
		if inHtmlComment {
			if strings.Contains(line, "-->") {
				inHtmlComment = false
			}
			continue
		}
		if strings.HasPrefix(trimmed, "<!--") && !strings.Contains(trimmed, "-->") {
			inHtmlComment = true
			continue
		}

		// 2. Handle code fences
		fenceMatches := codeFenceRegex.FindStringSubmatch(line)
		if len(fenceMatches) > 2 {
			fenceStr := fenceMatches[2]
			char := string(fenceStr[0])
			length := len(fenceStr)

			if !inCodeBlock {
				inCodeBlock = true
				codeFenceChar = char
				codeFenceLen = length
				codeFenceStartLine = lineNum
				continue
			} else {
				// Check if closing fence matches opening fence type and length
				if char == codeFenceChar && length >= codeFenceLen {
					inCodeBlock = false
					codeFenceChar = ""
					codeFenceLen = 0
					codeFenceStartLine = 0
					continue
				}
			}
		}

		// If inside a code fence, skip markdown structural checks
		if inCodeBlock {
			continue
		}

		// 3. Heading level hierarchy checks (MD001)
		if headingMatches := atxHeadingRegex.FindStringSubmatch(trimmed); len(headingMatches) > 2 {
			level := len(headingMatches[1])
			if prevHeadingLevel == 0 {
				if level > 2 {
					issues = append(issues, Issue{
						FilePath: rel,
						Line:     lineNum,
						Column:   1,
						RuleID:   r.ID(),
						Category: r.Category(),
						Phase:    r.Phase(),
						Severity: SeverityWarning,
						Message:  fmt.Sprintf("First heading level should start at h1 or h2, found h%d", level),
						Fixable:  false,
					})
				}
			} else if level > prevHeadingLevel+1 {
				issues = append(issues, Issue{
					FilePath: rel,
					Line:     lineNum,
					Column:   1,
					RuleID:   r.ID(),
					Category: r.Category(),
					Phase:    r.Phase(),
					Severity: SeverityWarning,
					Message:  fmt.Sprintf("Heading level skipped: jumped from h%d to h%d (expected at most h%d)", prevHeadingLevel, level, prevHeadingLevel+1),
					Fixable:  false,
				})
			}
			prevHeadingLevel = level
		}

		// 4. Link & Image URL integrity checks
		linkMatches := markdownLinkRegex.FindAllStringSubmatch(line, -1)
		for _, m := range linkMatches {
			if len(m) > 2 {
				text := strings.TrimSpace(m[1])
				target := strings.TrimSpace(m[2])

				if target == "" {
					issues = append(issues, Issue{
						FilePath: rel,
						Line:     lineNum,
						Column:   1,
						RuleID:   r.ID(),
						Category: r.Category(),
						Phase:    r.Phase(),
						Severity: SeverityWarning,
						Message:  fmt.Sprintf("Empty link target: [%s]()", text),
						Fixable:  false,
					})
				} else if target == "#" {
					issues = append(issues, Issue{
						FilePath: rel,
						Line:     lineNum,
						Column:   1,
						RuleID:   r.ID(),
						Category: r.Category(),
						Phase:    r.Phase(),
						Severity: SeverityWarning,
						Message:  fmt.Sprintf("Placeholder hash link target: [%s](#)", text),
						Fixable:  false,
					})
				} else if strings.EqualFold(target, "todo") || strings.EqualFold(target, "tbd") {
					issues = append(issues, Issue{
						FilePath: rel,
						Line:     lineNum,
						Column:   1,
						RuleID:   r.ID(),
						Category: r.Category(),
						Phase:    r.Phase(),
						Severity: SeverityWarning,
						Message:  fmt.Sprintf("Unresolved placeholder link target: [%s](%s)", text, target),
						Fixable:  false,
					})
				}
			}
		}
	}

	// 5. Unclosed code fence check
	if inCodeBlock {
		issues = append(issues, Issue{
			FilePath: rel,
			Line:     codeFenceStartLine,
			Column:   1,
			RuleID:   r.ID(),
			Category: r.Category(),
			Phase:    r.Phase(),
			Severity: SeverityError,
			Message:  fmt.Sprintf("Unclosed code fence block starting at line %d", codeFenceStartLine),
			Fixable:  false,
		})
	}

	return issues
}
