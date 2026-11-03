// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package addlicense is a port of github.com/google/addlicense v1.2.0.
// It ensures source code files have copyright license headers.
// Original authors: Google LLC (https://github.com/google/addlicense)
// Adapted for internal use by SnowdreamTech — package renamed from main to addlicense.
package addlicense

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil" //nolint:staticcheck // ioutil kept for parity with upstream
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	doublestar "github.com/bmatcuk/doublestar/v4"
	"golang.org/x/sync/errgroup"
)

// Options controls the behaviour of AddLicenseToFiles / CheckLicenseInFiles.
type Options struct {
	// License type (e.g. "MIT", "Apache-2.0"). Mutually exclusive with TemplateFile.
	License string
	// TemplateFile is a path to a custom license-header template file.
	TemplateFile string
	// Holder is the copyright holder name.
	Holder string
	// Year is the copyright year string (e.g. "2026").
	Year string
	// SPDX controls SPDX-style header generation.
	SPDX SpdxFlag
	// IgnorePatterns is a list of doublestar glob patterns to skip.
	IgnorePatterns []string
	// SkipExtensions is a list of file extensions to skip (without leading dot,
	// e.g. "rb", "py"). Comparison is case-insensitive. Equivalent to the
	// upstream -skip flag in google/addlicense.
	SkipExtensions []string
	// Verbose enables per-file logging.
	Verbose bool
}

// SpdxFlag controls whether SPDX identifiers are added.
type SpdxFlag uint8

const (
	SpdxOff  SpdxFlag = iota // No SPDX identifier
	SpdxOn                   // Append SPDX identifier to recognized license
	SpdxOnly                 // Only SPDX identifier, no full license text
)

// AddLicenseToFiles adds license headers to all matching source files under paths.
// Returns the number of modified files and any error encountered.
func AddLicenseToFiles(paths []string, opts Options) (int, error) {
	return processFiles(paths, opts, false)
}

// CheckLicenseInFiles checks that all matching source files under paths have a
// license header. Returns the number of files missing a header and any error.
func CheckLicenseInFiles(paths []string, opts Options) (int, error) {
	return processFiles(paths, opts, true)
}

// processFiles is the shared implementation for add and check modes.
func processFiles(paths []string, opts Options, checkOnly bool) (int, error) {
	data := licenseData{
		Year:   opts.Year,
		Holder: opts.Holder,
		SPDXID: opts.License,
	}

	tpl, err := fetchTemplate(opts.License, opts.TemplateFile, opts.SPDX)
	if err != nil {
		return 0, err
	}
	t, err := template.New("").Parse(tpl)
	if err != nil {
		return 0, err
	}

	// validate ignore patterns up-front
	for _, p := range opts.IgnorePatterns {
		if !doublestar.ValidatePattern(p) {
			return 0, fmt.Errorf("invalid ignore pattern: %q", p)
		}
	}

	type result struct {
		path    string
		changed bool
		err     error
	}

	ch := make(chan *file, 1000)
	results := make(chan result, 1000)

	// producer: walk all paths
	go func() {
		for _, p := range paths {
			if err := walk(ch, p, opts.IgnorePatterns, opts.SkipExtensions, opts.Verbose); err != nil {
				results <- result{err: err}
			}
		}
		close(ch)
	}()

	// consumer: process files concurrently
	var wg errgroup.Group
	go func() {
		for f := range ch {
			f := f
			wg.Go(func() error {
				if checkOnly {
					lic, err := licenseHeader(f.path, t, data)
					if err != nil {
						results <- result{path: f.path, err: err}
						return err
					}
					if lic == nil {
						return nil // unknown extension, skip
					}
					hasLic, err := fileHasLicense(f.path)
					if err != nil {
						results <- result{path: f.path, err: err}
						return err
					}
					if !hasLic {
						results <- result{path: f.path, changed: true}
						return errors.New("missing license header")
					}
				} else {
					modified, err := addLicense(f.path, f.mode, t, data)
					if err != nil {
						results <- result{path: f.path, err: err}
						return err
					}
					if modified {
						results <- result{path: f.path, changed: true}
					}
				}
				return nil
			})
		}
		_ = wg.Wait()
		close(results)
	}()

	count := 0
	var firstErr error
	for r := range results {
		if r.changed {
			count++
			if checkOnly {
				fmt.Printf("missing license header: %s\n", r.path)
			} else if opts.Verbose {
				fmt.Printf("modified: %s\n", r.path)
			}
		}
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
	}

	return count, firstErr
}

// ---- internal helpers (ported verbatim from google/addlicense v1.2.0) ------

type file struct {
	path string
	mode os.FileMode
}

func walk(ch chan<- *file, start string, ignorePatterns []string, skipExts []string, verbose bool) error {
	// Build a normalized set of extensions to skip (lowercase, no leading dot).
	skipSet := make(map[string]struct{}, len(skipExts))
	for _, e := range skipExts {
		skipSet[strings.ToLower(strings.TrimPrefix(e, "."))] = struct{}{}
	}

	return filepath.Walk(start, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.IsDir() {
			return nil
		}
		if fileMatches(path, ignorePatterns) {
			if verbose {
				fmt.Printf("skipping: %s\n", path)
			}
			return nil
		}
		if len(skipSet) > 0 {
			ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
			if _, skip := skipSet[ext]; skip {
				if verbose {
					fmt.Printf("skipping (ext): %s\n", path)
				}
				return nil
			}
		}
		ch <- &file{path, fi.Mode()}
		return nil
	})
}

func fileMatches(path string, patterns []string) bool {
	path = filepath.ToSlash(path)
	for _, p := range patterns {
		if match, _ := doublestar.Match(p, path); match {
			return true
		}
	}
	return false
}

func addLicense(path string, fmode os.FileMode, tmpl *template.Template, data licenseData) (bool, error) {
	lic, err := licenseHeader(path, tmpl, data)
	if err != nil || lic == nil {
		return false, err
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}
	if hasLicense(b) || isGenerated(b) {
		return false, nil
	}

	line := hashBang(b)
	if len(line) > 0 {
		b = b[len(line):]
		if line[len(line)-1] != '\n' {
			line = append(line, '\n')
		}
		lic = append(line, lic...)
	}
	b = append(lic, b...)
	return true, ioutil.WriteFile(path, b, fmode)
}

func fileHasLicense(path string) (bool, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}
	return hasLicense(b) || isGenerated(b), nil
}

func licenseHeader(path string, tmpl *template.Template, data licenseData) ([]byte, error) {
	base := strings.ToLower(filepath.Base(path))

	switch fileExtension(base) {
	case ".c", ".h", ".gv", ".java", ".kt", ".kts", ".scala":
		return executeTemplate(tmpl, data, "/*", " * ", " */")
	case ".css", ".scss", ".sass", ".less",
		".js", ".mjs", ".cjs", ".jsx",
		".ts", ".tsx":
		return executeTemplate(tmpl, data, "/**", " * ", " */")
	case ".cc", ".cpp", ".hh", ".hpp",
		".cs", ".dart", ".go", ".groovy", ".gradle",
		".hcl", ".m", ".mm", ".php", ".proto",
		".rs", ".swift", ".v", ".sv":
		return executeTemplate(tmpl, data, "", "// ", "")
	case ".awk", ".buckconfig", "buck",
		".bzl", ".bazel", "build", ".build",
		".dockerfile", "dockerfile",
		".ex", ".exs", ".graphql", ".jl", ".nix",
		".pl", ".pp", ".py", ".pyx", ".pxd", ".raku",
		".rb", ".ru", "gemfile",
		".sh", ".bash", ".zsh",
		".tcl", ".tf", ".toml", ".yaml", ".yml":
		return executeTemplate(tmpl, data, "", "# ", "")
	case ".el", ".lisp", ".scm":
		return executeTemplate(tmpl, data, "", ";; ", "")
	case ".erl":
		return executeTemplate(tmpl, data, "", "% ", "")
	case ".hs", ".lua", ".sql", ".sdl":
		return executeTemplate(tmpl, data, "", "-- ", "")
	case ".html", ".htm", ".vue", ".wxi", ".wxl", ".wxs", ".xml":
		return executeTemplate(tmpl, data, "<!--", " ", "-->")
	case ".j2":
		return executeTemplate(tmpl, data, "{#", "", "#}")
	case ".ml", ".mli", ".mll", ".mly":
		return executeTemplate(tmpl, data, "(**", "   ", "*)")
	case ".ps1", ".psm1":
		return executeTemplate(tmpl, data, "<#", " ", "#>")
	case ".vim":
		return executeTemplate(tmpl, data, "", `" `, "")
	default:
		if base == "cmakelists.txt" || strings.HasSuffix(base, ".cmake.in") || strings.HasSuffix(base, ".cmake") {
			return executeTemplate(tmpl, data, "", "# ", "")
		}
	}
	return nil, nil
}

func fileExtension(name string) string {
	if v := filepath.Ext(name); v != "" {
		return v
	}
	return name
}

var head = []string{
	"#!",
	"<?xml",
	"<!doctype",
	"# encoding:",
	"# frozen_string_literal:",
	`#\`,
	"<?php",
	"# escape",
	"# syntax",
}

func hashBang(b []byte) []byte {
	var line []byte
	for _, c := range b {
		line = append(line, c)
		if c == '\n' {
			break
		}
	}
	first := strings.ToLower(string(line))
	for _, h := range head {
		if strings.HasPrefix(first, h) {
			return line
		}
	}
	return nil
}

var goGenerated = regexp.MustCompile(`(?m)^.{1,2} Code generated .* DO NOT EDIT\.$`)
var cargoRazeGenerated = regexp.MustCompile(`(?m)^DO NOT EDIT! Replaced on runs of cargo-raze$`)

func isGenerated(b []byte) bool {
	return goGenerated.Match(b) || cargoRazeGenerated.Match(b)
}

func hasLicense(b []byte) bool {
	n := 1000
	if len(b) < 1000 {
		n = len(b)
	}
	return bytes.Contains(bytes.ToLower(b[:n]), []byte("copyright")) ||
		bytes.Contains(bytes.ToLower(b[:n]), []byte("mozilla public")) ||
		bytes.Contains(bytes.ToLower(b[:n]), []byte("spdx-license-identifier"))
}
