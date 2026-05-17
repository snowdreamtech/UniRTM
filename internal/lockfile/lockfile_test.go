// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package lockfile

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ─── LockFile read/write round-trip ──────────────────────────────────────────

func TestLockFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "unirtm.lock")

	lf := New(path)
	lf.UpsertEntry("github:cli/cli", &ToolLockEntry{
		Version: "2.72.0",
		Backend: "github",
		Platforms: map[string]*PlatformEntry{
			"linux-amd64": {
				Checksum: "sha256:abc123",
				Size:     12345678,
				URL:      "https://example.com/gh_2.72.0_linux_amd64.tar.gz",
				URLAPI:   "https://api.github.com/repos/cli/cli/releases/assets/1",
			},
			"macos-arm64": {
				Checksum: "sha256:def456",
				URL:      "https://example.com/gh_2.72.0_macOS_arm64.zip",
			},
		},
	})

	if err := lf.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify the file exists and contains the header.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "@generated") {
		t.Error("saved file missing @generated header")
	}

	// Reload and verify.
	lf2, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	pe := lf2.GetPlatform("github:cli/cli", "2.72.0", "linux-amd64")
	if pe == nil {
		t.Fatal("GetPlatform linux-amd64: got nil")
	}
	if pe.Checksum != "sha256:abc123" {
		t.Errorf("Checksum: got %q, want sha256:abc123", pe.Checksum)
	}
	if pe.Size != 12345678 {
		t.Errorf("Size: got %d, want 12345678", pe.Size)
	}
}

func TestLockFile_LoadMissing(t *testing.T) {
	lf, err := Load("/nonexistent/path/unirtm.lock")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if !lf.IsEmpty() {
		t.Error("expected empty lockfile for missing path")
	}
}

// ─── ToolKey ──────────────────────────────────────────────────────────────────

func TestToolKey(t *testing.T) {
	tests := []struct {
		backend, tool, want string
	}{
		{"github", "cli/cli", "github:cli/cli"},
		{"npm", "typescript", "npm:typescript"},
		{"ubi", "BurntSushi/ripgrep", "ubi:BurntSushi/ripgrep"},
	}
	for _, tt := range tests {
		got := ToolKey(tt.backend, tt.tool)
		if got != tt.want {
			t.Errorf("ToolKey(%q,%q) = %q, want %q", tt.backend, tt.tool, got, tt.want)
		}
	}
}

// ─── UpsertEntry / GetEntry ───────────────────────────────────────────────────

func TestLockFile_UpsertAndGet(t *testing.T) {
	lf := New("/tmp/test.lock")
	key := "github:cli/cli"

	lf.UpsertEntry(key, &ToolLockEntry{Version: "2.70.0", Backend: "github"})
	lf.UpsertEntry(key, &ToolLockEntry{Version: "2.72.0", Backend: "github"})

	if got := lf.GetEntry(key, "2.72.0"); got == nil {
		t.Error("GetEntry 2.72.0: got nil")
	}
	if got := lf.GetEntry(key, "9.9.9"); got != nil {
		t.Error("GetEntry non-existent: expected nil")
	}

	// Updating an existing version should replace, not append.
	lf.UpsertEntry(key, &ToolLockEntry{Version: "2.72.0", Backend: "github-v2"})
	e := lf.GetEntry(key, "2.72.0")
	if e == nil || e.Backend != "github-v2" {
		t.Errorf("upsert did not update in place: %+v", e)
	}
}

// ─── Platform helpers ─────────────────────────────────────────────────────────

func TestPlatformKey(t *testing.T) {
	tests := []struct {
		goos, goarch string
		musl         bool
		want         string
	}{
		{"linux", "amd64", false, "linux-amd64"},
		{"linux", "amd64", true, "linux-amd64-musl"},
		{"darwin", "arm64", false, "macos-arm64"},
		{"windows", "amd64", false, "windows-amd64"},
	}
	for _, tt := range tests {
		got := PlatformKey(tt.goos, tt.goarch, tt.musl)
		if got != tt.want {
			t.Errorf("PlatformKey(%q,%q,%v) = %q, want %q", tt.goos, tt.goarch, tt.musl, got, tt.want)
		}
	}
}

func TestCurrentPlatformKey(t *testing.T) {
	key := CurrentPlatformKey()
	if key == "" {
		t.Error("CurrentPlatformKey returned empty string")
	}
	// Must contain the current GOOS (mapped).
	wantOS := osNames[runtime.GOOS]
	if wantOS == "" {
		wantOS = runtime.GOOS
	}
	if !strings.HasPrefix(key, wantOS) {
		t.Errorf("CurrentPlatformKey %q does not start with %q", key, wantOS)
	}
}

func TestNormalizePlatformKey(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"linux-amd64", "linux-amd64", false},
		{"linux-x64", "linux-amd64", false}, // mise-style alias
		{"macos-x64", "macos-amd64", false},
		{"LINUX-AMD64", "linux-amd64", false}, // case-insensitive
		{"invalid-os", "", true},
	}
	for _, tt := range tests {
		got, err := NormalizePlatformKey(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("NormalizePlatformKey(%q): expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("NormalizePlatformKey(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("NormalizePlatformKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParsePlatformKeys(t *testing.T) {
	got, err := ParsePlatformKeys("linux-amd64,macos-arm64,linux-x64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// linux-x64 should be de-duped with linux-amd64.
	if len(got) != 2 {
		t.Errorf("expected 2 unique keys, got %d: %v", len(got), got)
	}
}

// ─── Validation ───────────────────────────────────────────────────────────────

func TestValidate_Valid(t *testing.T) {
	lf := New("/tmp/test.lock")
	lf.UpsertEntry("github:cli/cli", &ToolLockEntry{
		Version: "2.72.0",
		Backend: "github",
		Platforms: map[string]*PlatformEntry{
			"linux-amd64": {Checksum: "sha256:abc", URL: "https://example.com/file.tar.gz"},
		},
	})
	if err := lf.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidate_BadChecksum(t *testing.T) {
	lf := New("/tmp/test.lock")
	lf.UpsertEntry("github:cli/cli", &ToolLockEntry{
		Version: "2.72.0",
		Backend: "github",
		Platforms: map[string]*PlatformEntry{
			"linux-amd64": {Checksum: "md5:abc123"},
		},
	})
	err := lf.Validate()
	if err == nil {
		t.Error("expected validation error for bad checksum format")
	}
}

func TestValidate_EmptyVersion(t *testing.T) {
	lf := New("/tmp/test.lock")
	lf.Tools["github:cli/cli"] = []*ToolLockEntry{{Backend: "github"}} // no Version
	err := lf.Validate()
	if err == nil {
		t.Error("expected validation error for empty version")
	}
}

// ─── Strict mode ──────────────────────────────────────────────────────────────

func TestCheckStrict_Pass(t *testing.T) {
	lf := New("/tmp/test.lock")
	lf.UpsertPlatform("github:cli/cli", "2.72.0", "linux-amd64", &PlatformEntry{
		URL: "https://example.com/file.tar.gz",
	})
	err := lf.CheckStrict([]LockRequirement{
		{ToolKey: "github:cli/cli", Version: "2.72.0", PlatformKey: "linux-amd64"},
	})
	if err != nil {
		t.Errorf("CheckStrict: unexpected error: %v", err)
	}
}

func TestCheckStrict_Fail(t *testing.T) {
	lf := New("/tmp/test.lock")
	err := lf.CheckStrict([]LockRequirement{
		{ToolKey: "github:cli/cli", Version: "2.72.0", PlatformKey: "linux-amd64"},
	})
	if err == nil {
		t.Error("CheckStrict: expected error for missing URL")
	}
}

// ─── RemoveEntry ──────────────────────────────────────────────────────────────

func TestRemoveEntry(t *testing.T) {
	lf := New("/tmp/test.lock")
	lf.UpsertEntry("github:cli/cli", &ToolLockEntry{Version: "2.72.0", Backend: "github"})
	lf.RemoveEntry("github:cli/cli")
	if !lf.IsEmpty() {
		t.Error("expected empty after RemoveEntry")
	}
}

// ─── HasURL ───────────────────────────────────────────────────────────────────

func TestHasURL(t *testing.T) {
	lf := New("/tmp/test.lock")
	lf.UpsertPlatform("github:cli/cli", "2.72.0", "macos-arm64", &PlatformEntry{
		URL: "https://example.com/file.zip",
	})
	if !lf.HasURL("github:cli/cli", "2.72.0", "macos-arm64") {
		t.Error("HasURL: expected true")
	}
	if lf.HasURL("github:cli/cli", "2.72.0", "linux-amd64") {
		t.Error("HasURL: expected false for unknown platform")
	}
}
