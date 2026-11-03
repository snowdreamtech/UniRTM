// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package version

import (
	"testing"
)

func TestParseSemVerPartial(t *testing.T) {
	tests := []struct {
		input      string
		major      int
		minor      int
		patch      int
		prerelease string
		build      string
		expectErr  bool
	}{
		{"1", 1, 0, 0, "", "", false},
		{"v1", 1, 0, 0, "", "", false},
		{"1.2", 1, 2, 0, "", "", false},
		{"1.2.3", 1, 2, 3, "", "", false},
		{"1.2.3-alpha", 1, 2, 3, "alpha", "", false},
		{"1.2.3+build", 1, 2, 3, "", "build", false},
		{"1.2.3-alpha+build", 1, 2, 3, "alpha", "build", false},
		{"invalid", 0, 0, 0, "", "", true},
		{"1.invalid", 0, 0, 0, "", "", true},
		{"1.2.invalid", 0, 0, 0, "", "", true},
		{"1.2.3.4", 0, 0, 0, "", "", true},
		{"1.2.3-invalid+format+here", 1, 2, 3, "invalid", "format+here", false}, // Still parses as semver, we just split by -, +, so prerelease is invalid, build is format+here
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, err := ParseSemVerPartial(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error for %s", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tt.input, err)
			}
			if v.Major != tt.major || v.Minor != tt.minor || v.Patch != tt.patch {
				t.Errorf("expected %d.%d.%d, got %d.%d.%d", tt.major, tt.minor, tt.patch, v.Major, v.Minor, v.Patch)
			}
			if v.Prerelease != tt.prerelease {
				t.Errorf("expected prerelease %s, got %s", tt.prerelease, v.Prerelease)
			}
			if v.Build != tt.build {
				t.Errorf("expected build %s, got %s", tt.build, v.Build)
			}
		})
	}
}

func TestSemVer_Compare_Full(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.2.3", "1.2.3", 0},
		{"1.2.4", "1.2.3", 1},
		{"1.2.3", "1.2.4", -1},
		{"1.3.0", "1.2.3", 1},
		{"1.2.3", "1.3.0", -1},
		{"2.0.0", "1.9.9", 1},
		{"1.9.9", "2.0.0", -1},

		// Prerelease comparisons
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0-beta", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0-alpha.1", "1.0.0-alpha.2", -1},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_"+tt.v2, func(t *testing.T) {
			v1, _ := ParseSemVer(tt.v1)
			v2, _ := ParseSemVer(tt.v2)

			res := v1.Compare(v2)
			if res != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, res)
			}
		})
	}

	// Test nils
	var n1 *SemVer
	var n2 *SemVer
	v, _ := ParseSemVer("1.0.0")

	if n1.Compare(n2) != 0 {
		t.Error("expected 0")
	}
	if n1.Compare(v) != -1 {
		t.Error("expected -1")
	}
	if v.Compare(n2) != 1 {
		t.Error("expected 1")
	}
}
