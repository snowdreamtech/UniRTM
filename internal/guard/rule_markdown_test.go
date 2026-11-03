// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMarkdownRule(t *testing.T) {
	rule := NewMarkdownRule()
	if rule.ID() != "builtin:check-markdown" {
		t.Fatalf("unexpected rule ID: %s", rule.ID())
	}
	if rule.Phase() != PhaseLinter {
		t.Fatalf("unexpected phase: %s", rule.Phase())
	}

	tmpDir, err := os.MkdirTemp("", "guard_md_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 1. Valid markdown
	validContent := `# Main Title

## Section 1

Here is a paragraph with a [valid link](https://example.com) and an image ![logo](https://example.com/logo.png).

` + "```go\nfunc main() {\n    # Not a heading in code\n    url := \"[not a link]()\"\n}\n```\n" + `
### Subsection 1.1

All good here!
`
	validPath := filepath.Join(tmpDir, "valid.md")
	if err := os.WriteFile(validPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("failed to write valid.md: %v", err)
	}

	// 2. Heading level skip (h1 to h3)
	skipHeadingContent := `# Main Title

### Skipped Section Directly to H3

Content here.
`
	skipHeadingPath := filepath.Join(tmpDir, "skip_heading.md")
	if err := os.WriteFile(skipHeadingPath, []byte(skipHeadingContent), 0644); err != nil {
		t.Fatalf("failed to write skip_heading.md: %v", err)
	}

	// 3. Unclosed code fence
	unclosedFenceContent := `# Code Guide

Here is an example:
` + "```javascript\nconsole.log('hello');\n" + `
Missing closing fence!
`
	unclosedFencePath := filepath.Join(tmpDir, "unclosed_fence.md")
	if err := os.WriteFile(unclosedFencePath, []byte(unclosedFenceContent), 0644); err != nil {
		t.Fatalf("failed to write unclosed_fence.md: %v", err)
	}

	// 4. Broken/Empty links
	emptyLinkContent := `# Links Test

Please check [empty target]() or [hash placeholder](#) or [pending](TODO).
`
	emptyLinkPath := filepath.Join(tmpDir, "empty_link.md")
	if err := os.WriteFile(emptyLinkPath, []byte(emptyLinkContent), 0644); err != nil {
		t.Fatalf("failed to write empty_link.md: %v", err)
	}

	opts := ExecuteOptions{
		WorkspaceDir: tmpDir,
	}

	files := []string{"valid.md", "skip_heading.md", "unclosed_fence.md", "empty_link.md"}
	issues, err := rule.Execute(context.Background(), files, opts)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Group issues by file
	fileIssues := make(map[string][]Issue)
	for _, issue := range issues {
		fileIssues[issue.FilePath] = append(fileIssues[issue.FilePath], issue)
	}

	// Verify valid.md has no issues
	if len(fileIssues["valid.md"]) != 0 {
		t.Errorf("expected 0 issues for valid.md, got %d: %+v", len(fileIssues["valid.md"]), fileIssues["valid.md"])
	}

	// Verify skip_heading.md has heading skipped issue
	if len(fileIssues["skip_heading.md"]) != 1 {
		t.Errorf("expected 1 issue for skip_heading.md, got %d: %+v", len(fileIssues["skip_heading.md"]), fileIssues["skip_heading.md"])
	} else {
		if fileIssues["skip_heading.md"][0].Line != 3 {
			t.Errorf("expected line 3 for heading skip, got %d", fileIssues["skip_heading.md"][0].Line)
		}
	}

	// Verify unclosed_fence.md has unclosed code fence issue
	if len(fileIssues["unclosed_fence.md"]) != 1 {
		t.Errorf("expected 1 issue for unclosed_fence.md, got %d: %+v", len(fileIssues["unclosed_fence.md"]), fileIssues["unclosed_fence.md"])
	} else {
		if fileIssues["unclosed_fence.md"][0].Severity != SeverityError {
			t.Errorf("expected SeverityError for unclosed fence, got %s", fileIssues["unclosed_fence.md"][0].Severity)
		}
	}

	// Verify empty_link.md has 3 empty/placeholder link issues
	if len(fileIssues["empty_link.md"]) != 3 {
		t.Errorf("expected 3 issues for empty_link.md, got %d: %+v", len(fileIssues["empty_link.md"]), fileIssues["empty_link.md"])
	}
}
