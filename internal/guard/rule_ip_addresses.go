// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// ipv4CandidateRegex matches potential IPv4 addresses
	ipv4CandidateRegex = regexp.MustCompile(`\b([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})\b`)

	// rfc5737DocBlocks are reserved documentation IP blocks safe for examples
	rfc5737Blocks = []string{
		"192.0.2.",
		"198.51.100.",
		"203.0.113.",
	}
)

// IpAddressesRule detects hardcoded private or sensitive public IPv4 addresses in source and configs
type IpAddressesRule struct{}

// NewIpAddressesRule returns a new hardcoded IP address detection rule instance
func NewIpAddressesRule() Rule {
	return &IpAddressesRule{}
}

func (r *IpAddressesRule) ID() string {
	return "builtin:forbid-ip-addresses"
}

func (r *IpAddressesRule) Category() Category {
	return CategorySecurity
}

func (r *IpAddressesRule) Phase() Phase {
	return PhaseLinter
}

func (r *IpAddressesRule) DefaultExtensions() []string {
	return []string{
		".go", ".ts", ".js", ".jsx", ".tsx",
		".py", ".java", ".rb", ".php", ".sh", ".bash",
		".yaml", ".yml", ".json", ".env", ".toml", ".ini", ".conf", ".properties",
	}
}

func (r *IpAddressesRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	allowedExts := make(map[string]bool)
	for _, ext := range r.DefaultExtensions() {
		allowedExts[strings.ToLower(ext)] = true
	}

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		ext := strings.ToLower(filepath.Ext(rel))
		if !allowedExts[ext] {
			continue
		}

		// Skip test files, mock files, and testdata
		lowerRel := strings.ToLower(filepath.ToSlash(rel))
		if strings.Contains(lowerRel, "testdata/") ||
			strings.Contains(lowerRel, "fixtures/") ||
			strings.HasSuffix(lowerRel, "_test.go") ||
			strings.HasSuffix(lowerRel, ".test.ts") ||
			strings.HasSuffix(lowerRel, ".spec.ts") ||
			strings.HasSuffix(lowerRel, ".test.js") ||
			strings.HasSuffix(lowerRel, "internal/network/defaults.go") ||
			strings.HasSuffix(lowerRel, "internal/network/proxy_test_conn.go") {
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
				Message:  fmt.Sprintf("Failed to read file for IP address scanning: %v", err),
				Fixable:  false,
			})
			continue
		}

		fileIssues := r.scanForIpAddresses(rel, data)
		issues = append(issues, fileIssues...)
	}

	return issues, nil
}

func (r *IpAddressesRule) scanForIpAddresses(rel string, data []byte) []Issue {
	var issues []Issue
	scanner := bufio.NewScanner(bytes.NewReader(data))

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		matches := ipv4CandidateRegex.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			ipStr := m[1]
			ip := net.ParseIP(ipStr)
			if ip == nil || ip.To4() == nil {
				continue
			}

			// 1. Whitelist loopback (127.0.0.0/8)
			if ip.IsLoopback() {
				continue
			}

			// 2. Whitelist unspecified (0.0.0.0)
			if ip.IsUnspecified() {
				continue
			}

			// 3. Whitelist broadcast and common subnet masks
			if ipStr == "255.255.255.255" || ipStr == "255.255.255.0" || ipStr == "255.255.0.0" || ipStr == "255.0.0.0" {
				continue
			}

			// 4. Whitelist RFC 5737 documentation blocks (192.0.2.x, 198.51.100.x, 203.0.113.x)
			isDocIp := false
			for _, prefix := range rfc5737Blocks {
				if strings.HasPrefix(ipStr, prefix) {
					isDocIp = true
					break
				}
			}
			if isDocIp {
				continue
			}

			// 5. Flag private IPs (10.x.x.x, 172.16-31.x.x, 192.168.x.x) with SeverityError
			if ip.IsPrivate() {
				issues = append(issues, Issue{
					FilePath: rel,
					Line:     lineNum,
					Column:   strings.Index(line, ipStr) + 1,
					RuleID:   r.ID(),
					Category: r.Category(),
					Phase:    r.Phase(),
					Severity: SeverityError,
					Message:  fmt.Sprintf("Hardcoded private IPv4 address forbidden: %s (use environment variable or service discovery)", ipStr),
					Fixable:  false,
				})
				continue
			}

			// 6. Flag public IPs with SeverityWarning
			issues = append(issues, Issue{
				FilePath: rel,
				Line:     lineNum,
				Column:   strings.Index(line, ipStr) + 1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("Hardcoded public IPv4 address detected: %s (externalize into configuration or environment variable)", ipStr),
				Fixable:  false,
			})
		}
	}

	return issues
}
