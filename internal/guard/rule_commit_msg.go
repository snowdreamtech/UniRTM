// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// conventionalCommitRegex validates standard Conventional Commits header: <type>(<scope>): <subject>
	conventionalCommitRegex = regexp.MustCompile(`^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9_\-\.\/ ]+\))?(!)?: (.+)$`)
	// allowedSpecialCommitRegex allows standard git merges, reverts, and interactive fixup commits
	allowedSpecialCommitRegex = regexp.MustCompile(`^(Merge |Revert |fixup! |squash! |amend! )`)
)

// AllowedCommitTypes lists all supported standard Conventional Commit types
var AllowedCommitTypes = []string{
	"feat", "fix", "docs", "style", "refactor",
	"perf", "test", "build", "ci", "chore", "revert",
}

// ValidateCommitMessage validates raw commit message string against Conventional Commits specification
func ValidateCommitMessage(msg string) (bool, string) {
	lines := strings.Split(strings.ReplaceAll(msg, "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		return false, "Commit message cannot be empty"
	}

	// Filter out Git comment lines starting with #
	var contentLines []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if !strings.HasPrefix(trimmed, "#") {
			contentLines = append(contentLines, l)
		}
	}

	if len(contentLines) == 0 || strings.TrimSpace(contentLines[0]) == "" {
		return false, "Commit message cannot be empty"
	}

	header := strings.TrimSpace(contentLines[0])

	// Allow standard Git merge or fixup commits
	if allowedSpecialCommitRegex.MatchString(header) {
		return true, ""
	}

	matches := conventionalCommitRegex.FindStringSubmatch(header)
	if len(matches) == 0 {
		return false, fmt.Sprintf(
			"Invalid commit header format %q. Expected: <type>(<scope>): <subject> (valid types: %s)",
			header, strings.Join(AllowedCommitTypes, ", "),
		)
	}

	subject := strings.TrimSpace(matches[4])
	if subject == "" {
		return false, "Commit message subject cannot be empty"
	}

	if strings.HasSuffix(subject, ".") {
		return false, "Commit subject must not end with a period '.'"
	}

	// Check if body separator is a blank line if body exists
	if len(contentLines) > 1 {
		secondLine := strings.TrimSpace(contentLines[1])
		if secondLine != "" {
			return false, "Second line of commit message must be empty (separating header from body)"
		}
	}

	return true, ""
}

// CommitMsgRule validates git commit messages against Conventional Commits specification
type CommitMsgRule struct{}

func NewCommitMsgRule() Rule {
	return &CommitMsgRule{}
}

func (r *CommitMsgRule) ID() string {
	return "builtin:commit-msg"
}

func (r *CommitMsgRule) Category() Category {
	return CategoryPolicy
}

func (r *CommitMsgRule) Phase() Phase {
	return PhaseLinter
}

func (r *CommitMsgRule) DefaultExtensions() []string {
	return []string{"COMMIT_EDITMSG"}
}

func (r *CommitMsgRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		baseName := filepath.Base(rel)
		if baseName != "COMMIT_EDITMSG" && !strings.HasSuffix(rel, "COMMIT_EDITMSG") {
			continue
		}

		fullPath := rel
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(opts.WorkspaceDir, rel)
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		ok, reason := ValidateCommitMessage(string(content))
		if !ok {
			issues = append(issues, Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityError,
				Message:  reason,
				Fixable:  false,
			})
		}
	}

	return issues, nil
}
