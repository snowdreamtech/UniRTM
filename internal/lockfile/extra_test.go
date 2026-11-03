// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package lockfile

import (
	"path/filepath"
	"testing"
)

func TestParsePlatformKey(t *testing.T) {
	tests := []struct {
		key      string
		goos     string
		goarch   string
		musl     bool
		hasError bool
	}{
		{"linux-amd64", "linux", "amd64", false, false},
		{"macos-arm64", "darwin", "arm64", false, false},
		{"windows-amd64", "windows", "amd64", false, false},
		{"linux-amd64-musl", "linux", "amd64", true, false},
		{"unknown-unknown", "unknown", "unknown", false, false},
		{"invalid", "", "", false, true},
		{"linux-amd64-invalid", "", "", false, true},
		{"macos-386", "darwin", "386", false, false},
		{"linux-arm", "linux", "arm", false, false},
	}

	for _, tc := range tests {
		t.Run(tc.key, func(t *testing.T) {
			g, a, m, err := ParsePlatformKey(tc.key)
			if tc.hasError {
				if err == nil {
					t.Errorf("expected error for %q, got nil", tc.key)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for %q: %v", tc.key, err)
				}
				if g != tc.goos || a != tc.goarch || m != tc.musl {
					t.Errorf("ParsePlatformKey(%q) = (%q, %q, %v); want (%q, %q, %v)", tc.key, g, a, m, tc.goos, tc.goarch, tc.musl)
				}
			}
		})
	}
}

func TestPath(t *testing.T) {
	lf := &LockFile{
		path: "/test/unirtm.lock",
	}
	if lf.Path() != "/test/unirtm.lock" {
		t.Errorf("expected Path() to be /test/unirtm.lock, got %s", lf.Path())
	}
}

func TestSortedMapEntries(t *testing.T) {
	m := map[string]string{
		"windows-amd64": "c",
		"linux-amd64":   "a",
		"macos-arm64":   "b",
	}

	keys := sortedMapEntries(m)
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	if keys[0].k != "linux-amd64" || keys[1].k != "macos-arm64" || keys[2].k != "windows-amd64" {
		t.Errorf("expected linux-amd64, macos-arm64, windows-amd64; got %v", keys)
	}
}

func TestSave_EmptyTools(t *testing.T) {
	tmpDir := t.TempDir()
	lf := New(filepath.Join(tmpDir, "unirtm.lock"))
	// no tools
	err := lf.Save()
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}
}
