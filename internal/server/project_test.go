// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveProjectDir(t *testing.T) {
	tmpDir := t.TempDir()
	subProj := filepath.Join(tmpDir, "sub-proj")
	if err := os.MkdirAll(subProj, 0755); err != nil {
		t.Fatalf("failed to create subProj: %v", err)
	}

	// 1. Query parameter root
	req := httptest.NewRequest("GET", "/api/env?root="+subProj, nil)
	resolved := ResolveProjectDir(req)
	if resolved != subProj {
		t.Errorf("expected %q, got %q", subProj, resolved)
	}

	// 2. Query parameter project
	req = httptest.NewRequest("GET", "/api/env?project="+subProj, nil)
	resolved = ResolveProjectDir(req)
	if resolved != subProj {
		t.Errorf("expected %q, got %q", subProj, resolved)
	}

	// 3. Header X-Project-Root
	req = httptest.NewRequest("GET", "/api/env", nil)
	req.Header.Set("X-Project-Root", subProj)
	resolved = ResolveProjectDir(req)
	if resolved != subProj {
		t.Errorf("expected %q, got %q", subProj, resolved)
	}

	// 4. Fallback when nil or empty
	resolved = ResolveProjectDir(nil)
	if resolved == "" {
		t.Errorf("expected non-empty resolved directory for nil request")
	}
}
