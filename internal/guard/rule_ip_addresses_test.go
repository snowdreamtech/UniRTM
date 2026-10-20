// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestIpAddressesRule(t *testing.T) {
	rule := NewIpAddressesRule()
	if rule.ID() != "builtin:forbid-ip-addresses" {
		t.Fatalf("unexpected rule ID: %s", rule.ID())
	}
	if rule.Category() != CategorySecurity {
		t.Fatalf("unexpected category: %s", rule.Category())
	}

	tmpDir, err := os.MkdirTemp("", "guard_ip_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 1. Whitelisted IPs (127.0.0.1, 0.0.0.0, mask, RFC 5737 doc IP)
	whitelistedContent := `package main

const (
    LocalHost = "127.0.0.1"
    BindAll   = "0.0.0.0"
    Subnet    = "255.255.255.0"
    DocServer = "192.0.2.1"
)
`
	whitelistedPath := filepath.Join(tmpDir, "safe.go")
	if err := os.WriteFile(whitelistedPath, []byte(whitelistedContent), 0644); err != nil {
		t.Fatalf("failed to write safe.go: %v", err)
	}

	// 2. Private IPs (192.168.1.50, 10.20.30.40)
	privateContent := `server:
  host: 192.168.1.50
  internal_gateway: 10.20.30.40
`
	privatePath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(privatePath, []byte(privateContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}

	// 3. Public IP (e.g., 93.184.216.34)
	publicContent := `const apiHost = "http://93.184.216.34:8080";`
	publicPath := filepath.Join(tmpDir, "client.ts")
	if err := os.WriteFile(publicPath, []byte(publicContent), 0644); err != nil {
		t.Fatalf("failed to write client.ts: %v", err)
	}

	opts := ExecuteOptions{
		WorkspaceDir: tmpDir,
	}

	files := []string{"safe.go", "config.yaml", "client.ts"}
	issues, err := rule.Execute(context.Background(), files, opts)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	fileIssues := make(map[string][]Issue)
	for _, issue := range issues {
		fileIssues[issue.FilePath] = append(fileIssues[issue.FilePath], issue)
	}

	// Verify safe.go has no issues
	if len(fileIssues["safe.go"]) != 0 {
		t.Errorf("expected 0 issues for safe.go, got %d: %+v", len(fileIssues["safe.go"]), fileIssues["safe.go"])
	}

	// Verify config.yaml has 2 private IP errors
	if len(fileIssues["config.yaml"]) != 2 {
		t.Errorf("expected 2 issues for config.yaml, got %d: %+v", len(fileIssues["config.yaml"]), fileIssues["config.yaml"])
	} else {
		for _, iss := range fileIssues["config.yaml"] {
			if iss.Severity != SeverityError {
				t.Errorf("expected SeverityError for private IP, got %s", iss.Severity)
			}
		}
	}

	// Verify client.ts has 1 public IP warning
	if len(fileIssues["client.ts"]) != 1 {
		t.Errorf("expected 1 issue for client.ts, got %d: %+v", len(fileIssues["client.ts"]), fileIssues["client.ts"])
	} else {
		if fileIssues["client.ts"][0].Severity != SeverityWarning {
			t.Errorf("expected SeverityWarning for public IP, got %s", fileIssues["client.ts"][0].Severity)
		}
	}
}
