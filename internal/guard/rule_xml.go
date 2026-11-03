// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var xmlExtensions = map[string]bool{
	".xml":   true,
	".plist": true,
	".xsd":   true,
	".xsl":   true,
	".xslt":  true,
}

// XmlRule validates well-formedness of XML and XML-based files using standard Go xml parser
type XmlRule struct{}

func NewXmlRule() Rule {
	return &XmlRule{}
}

func (r *XmlRule) ID() string {
	return "builtin:check-xml"
}

func (r *XmlRule) Category() Category {
	return CategoryHygiene
}

func (r *XmlRule) Phase() Phase {
	return PhaseLinter
}

func (r *XmlRule) DefaultExtensions() []string {
	return []string{".xml", ".plist", ".xsd", ".xsl", ".xslt"}
}

func (r *XmlRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		ext := strings.ToLower(filepath.Ext(rel))
		if !xmlExtensions[ext] {
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
			// Skip XML > 5MB
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Empty XML files are not well-formed XML documents
		trimmed := bytes.TrimSpace(content)
		if len(trimmed) == 0 {
			issues = append(issues, Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityError,
				Message:  "Malformed XML: file is empty",
				Fixable:  false,
			})
			continue
		}

		decoder := xml.NewDecoder(bytes.NewReader(content))
		for {
			_, err := decoder.Token()
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
					Message:  fmt.Sprintf("Malformed XML: %v", err),
					Fixable:  false,
				})
				break
			}
		}
	}

	return issues, nil
}
