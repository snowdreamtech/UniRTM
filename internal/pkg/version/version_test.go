// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSemVer(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *SemVer
		wantErr bool
	}{
		{
			name:  "basic semver",
			input: "1.2.3",
			want: &SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
		},
		{
			name:  "semver with v prefix",
			input: "v1.2.3",
			want: &SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
		},
		{
			name:  "semver with prerelease",
			input: "1.2.3-alpha.1",
			want: &SemVer{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Prerelease: "alpha.1",
			},
		},
		{
			name:  "semver with build metadata",
			input: "1.2.3+build.123",
			want: &SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
				Build: "build.123",
			},
		},
		{
			name:  "semver with prerelease and build",
			input: "1.2.3-beta.2+build.456",
			want: &SemVer{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Prerelease: "beta.2",
				Build:      "build.456",
			},
		},
		{
			name:  "zero version",
			input: "0.0.0",
			want: &SemVer{
				Major: 0,
				Minor: 0,
				Patch: 0,
			},
		},
		{
			name:    "invalid format - missing patch",
			input:   "1.2",
			wantErr: true,
		},
		{
			name:    "invalid format - non-numeric",
			input:   "1.2.x",
			wantErr: true,
		},
		{
			name:    "invalid format - missing patch",
			input:   "1.2",
			wantErr: true,
		},
		{
			name:    "invalid format - non-numeric",
			input:   "1.2.x",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSemVer(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Version
		wantErr bool
	}{
		// Exact versions
		{
			name:  "exact version",
			input: "1.20.0",
			want: &Version{
				Type: VersionTypeExact,
				Exact: &SemVer{
					Major: 1,
					Minor: 20,
					Patch: 0,
				},
			},
		},
		{
			name:  "exact version with v prefix",
			input: "v3.11.5",
			want: &Version{
				Type: VersionTypeExact,
				Exact: &SemVer{
					Major: 3,
					Minor: 11,
					Patch: 5,
				},
			},
		},
		{
			name:  "exact version with prerelease",
			input: "2.0.0-rc.1",
			want: &Version{
				Type: VersionTypeExact,
				Exact: &SemVer{
					Major:      2,
					Minor:      0,
					Patch:      0,
					Prerelease: "rc.1",
				},
			},
		},

		// Aliases
		{
			name:  "alias latest",
			input: "latest",
			want: &Version{
				Type:  VersionTypeAlias,
				Alias: VersionAliasLatest,
			},
		},
		{
			name:  "alias lts",
			input: "lts",
			want: &Version{
				Type:  VersionTypeAlias,
				Alias: VersionAliasLTS,
			},
		},
		{
			name:  "alias stable",
			input: "stable",
			want: &Version{
				Type:  VersionTypeAlias,
				Alias: VersionAliasStable,
			},
		},
		{
			name:  "alias case insensitive",
			input: "LATEST",
			want: &Version{
				Type:  VersionTypeAlias,
				Alias: VersionAliasLatest,
			},
		},

		// Range operators
		{
			name:  "range >=",
			input: ">=1.20.0",
			want: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorGTE,
				RangeVer: &SemVer{
					Major: 1,
					Minor: 20,
					Patch: 0,
				},
			},
		},
		{
			name:  "range >",
			input: ">2.0.0",
			want: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorGT,
				RangeVer: &SemVer{
					Major: 2,
					Minor: 0,
					Patch: 0,
				},
			},
		},
		{
			name:  "range <=",
			input: "<=3.0.0",
			want: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorLTE,
				RangeVer: &SemVer{
					Major: 3,
					Minor: 0,
					Patch: 0,
				},
			},
		},
		{
			name:  "range <",
			input: "<4.0.0",
			want: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorLT,
				RangeVer: &SemVer{
					Major: 4,
					Minor: 0,
					Patch: 0,
				},
			},
		},
		{
			name:  "range =",
			input: "=1.2.3",
			want: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorEQ,
				RangeVer: &SemVer{
					Major: 1,
					Minor: 2,
					Patch: 3,
				},
			},
		},

		// Caret ranges
		{
			name:  "caret range",
			input: "^3.11.0",
			want: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorCaret,
				RangeVer: &SemVer{
					Major: 3,
					Minor: 11,
					Patch: 0,
				},
			},
		},
		{
			name:  "caret range with v prefix",
			input: "^v1.2.3",
			want: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorCaret,
				RangeVer: &SemVer{
					Major: 1,
					Minor: 2,
					Patch: 3,
				},
			},
		},

		// Tilde ranges
		{
			name:  "tilde range full",
			input: "~2.7.0",
			want: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorTilde,
				RangeVer: &SemVer{
					Major: 2,
					Minor: 7,
					Patch: 0,
				},
			},
		},
		{
			name:  "tilde range minor",
			input: "~1.2",
			want: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorTilde,
				RangeVer: &SemVer{
					Major: 1,
					Minor: 2,
					Patch: 0,
				},
			},
		},
		{
			name:  "tilde range major",
			input: "~1",
			want: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorTilde,
				RangeVer: &SemVer{
					Major: 1,
					Minor: 0,
					Patch: 0,
				},
			},
		},

		// Error cases
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid semver",
			input:   "1.2.x",
			wantErr: true,
		},
		{
			name:    "invalid alias",
			input:   "unknown",
			wantErr: true,
		},
		{
			name:    "invalid range",
			input:   ">=1.2.x",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.True(t, tt.want.Equal(got), "expected %v, got %v", tt.want, got)
		})
	}
}

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   *Version
		want    string
		wantErr bool
	}{
		// Exact versions
		{
			name: "exact version",
			input: &Version{
				Type: VersionTypeExact,
				Exact: &SemVer{
					Major: 1,
					Minor: 20,
					Patch: 0,
				},
			},
			want: "1.20.0",
		},
		{
			name: "exact version with prerelease",
			input: &Version{
				Type: VersionTypeExact,
				Exact: &SemVer{
					Major:      2,
					Minor:      0,
					Patch:      0,
					Prerelease: "rc.1",
				},
			},
			want: "2.0.0-rc.1",
		},
		{
			name: "exact version with build",
			input: &Version{
				Type: VersionTypeExact,
				Exact: &SemVer{
					Major: 1,
					Minor: 2,
					Patch: 3,
					Build: "build.123",
				},
			},
			want: "1.2.3+build.123",
		},
		{
			name: "exact version with prerelease and build",
			input: &Version{
				Type: VersionTypeExact,
				Exact: &SemVer{
					Major:      1,
					Minor:      2,
					Patch:      3,
					Prerelease: "beta.2",
					Build:      "build.456",
				},
			},
			want: "1.2.3-beta.2+build.456",
		},

		// Aliases
		{
			name: "alias latest",
			input: &Version{
				Type:  VersionTypeAlias,
				Alias: VersionAliasLatest,
			},
			want: "latest",
		},
		{
			name: "alias lts",
			input: &Version{
				Type:  VersionTypeAlias,
				Alias: VersionAliasLTS,
			},
			want: "lts",
		},
		{
			name: "alias stable",
			input: &Version{
				Type:  VersionTypeAlias,
				Alias: VersionAliasStable,
			},
			want: "stable",
		},

		// Range operators
		{
			name: "range >=",
			input: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorGTE,
				RangeVer: &SemVer{
					Major: 1,
					Minor: 20,
					Patch: 0,
				},
			},
			want: ">=1.20.0",
		},
		{
			name: "range >",
			input: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorGT,
				RangeVer: &SemVer{
					Major: 2,
					Minor: 0,
					Patch: 0,
				},
			},
			want: ">2.0.0",
		},
		{
			name: "caret range",
			input: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorCaret,
				RangeVer: &SemVer{
					Major: 3,
					Minor: 11,
					Patch: 0,
				},
			},
			want: "^3.11.0",
		},
		{
			name: "tilde range",
			input: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorTilde,
				RangeVer: &SemVer{
					Major: 2,
					Minor: 7,
					Patch: 0,
				},
			},
			want: "~2.7.0",
		},

		// Error cases
		{
			name:    "nil version",
			input:   nil,
			wantErr: true,
		},
		{
			name: "nil exact version",
			input: &Version{
				Type:  VersionTypeExact,
				Exact: nil,
			},
			wantErr: true,
		},
		{
			name: "nil range version",
			input: &Version{
				Type:     VersionTypeRange,
				RangeOp:  RangeOperatorGTE,
				RangeVer: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatVersion(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVersionRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// Exact versions
		{"exact version", "1.20.0"},
		{"exact version with prerelease", "2.0.0-rc.1"},
		{"exact version with build", "1.2.3+build.123"},
		{"exact version with both", "1.2.3-beta.2+build.456"},

		// Aliases
		{"alias latest", "latest"},
		{"alias lts", "lts"},
		{"alias stable", "stable"},

		// Range operators
		{"range >=", ">=1.20.0"},
		{"range >", ">2.0.0"},
		{"range <=", "<=3.0.0"},
		{"range <", "<4.0.0"},
		{"range =", "=1.2.3"},

		// Caret and tilde
		{"caret range", "^3.11.0"},
		{"tilde range", "~2.7.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse
			parsed, err := ParseVersion(tt.input)
			require.NoError(t, err, "parsing failed")

			// Format
			formatted, err := FormatVersion(parsed)
			require.NoError(t, err, "formatting failed")

			// Parse again
			reparsed, err := ParseVersion(formatted)
			require.NoError(t, err, "re-parsing failed")

			// Verify equivalence
			assert.True(t, parsed.Equal(reparsed), "round-trip failed: %s -> %s -> %s", tt.input, formatted, reparsed.String())

			// Format again and verify string equality
			reformatted, err := FormatVersion(reparsed)
			require.NoError(t, err, "re-formatting failed")
			assert.Equal(t, formatted, reformatted, "formatted strings should be identical")
		})
	}
}

func TestVersionEqual(t *testing.T) {
	tests := []struct {
		name  string
		v1    *Version
		v2    *Version
		equal bool
	}{
		{
			name: "equal exact versions",
			v1: &Version{
				Type: VersionTypeExact,
				Exact: &SemVer{
					Major: 1,
					Minor: 2,
					Patch: 3,
				},
			},
			v2: &Version{
				Type: VersionTypeExact,
				Exact: &SemVer{
					Major: 1,
					Minor: 2,
					Patch: 3,
				},
			},
			equal: true,
		},
		{
			name: "different exact versions",
			v1: &Version{
				Type: VersionTypeExact,
				Exact: &SemVer{
					Major: 1,
					Minor: 2,
					Patch: 3,
				},
			},
			v2: &Version{
				Type: VersionTypeExact,
				Exact: &SemVer{
					Major: 1,
					Minor: 2,
					Patch: 4,
				},
			},
			equal: false,
		},
		{
			name: "equal aliases",
			v1: &Version{
				Type:  VersionTypeAlias,
				Alias: VersionAliasLatest,
			},
			v2: &Version{
				Type:  VersionTypeAlias,
				Alias: VersionAliasLatest,
			},
			equal: true,
		},
		{
			name: "different aliases",
			v1: &Version{
				Type:  VersionTypeAlias,
				Alias: VersionAliasLatest,
			},
			v2: &Version{
				Type:  VersionTypeAlias,
				Alias: VersionAliasLTS,
			},
			equal: false,
		},
		{
			name: "equal ranges",
			v1: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorGTE,
				RangeVer: &SemVer{
					Major: 1,
					Minor: 2,
					Patch: 3,
				},
			},
			v2: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorGTE,
				RangeVer: &SemVer{
					Major: 1,
					Minor: 2,
					Patch: 3,
				},
			},
			equal: true,
		},
		{
			name: "different range operators",
			v1: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorGTE,
				RangeVer: &SemVer{
					Major: 1,
					Minor: 2,
					Patch: 3,
				},
			},
			v2: &Version{
				Type:    VersionTypeRange,
				RangeOp: RangeOperatorGT,
				RangeVer: &SemVer{
					Major: 1,
					Minor: 2,
					Patch: 3,
				},
			},
			equal: false,
		},
		{
			name:  "both nil",
			v1:    nil,
			v2:    nil,
			equal: true,
		},
		{
			name: "one nil",
			v1: &Version{
				Type:  VersionTypeAlias,
				Alias: VersionAliasLatest,
			},
			v2:    nil,
			equal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.equal, tt.v1.Equal(tt.v2))
		})
	}
}

func TestSemVerEqual(t *testing.T) {
	tests := []struct {
		name  string
		s1    *SemVer
		s2    *SemVer
		equal bool
	}{
		{
			name: "equal basic versions",
			s1: &SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			s2: &SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			equal: true,
		},
		{
			name: "equal with prerelease",
			s1: &SemVer{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Prerelease: "alpha.1",
			},
			s2: &SemVer{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Prerelease: "alpha.1",
			},
			equal: true,
		},
		{
			name: "different prerelease",
			s1: &SemVer{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Prerelease: "alpha.1",
			},
			s2: &SemVer{
				Major:      1,
				Minor:      2,
				Patch:      3,
				Prerelease: "alpha.2",
			},
			equal: false,
		},
		{
			name: "equal with build",
			s1: &SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
				Build: "build.123",
			},
			s2: &SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
				Build: "build.123",
			},
			equal: true,
		},
		{
			name:  "both nil",
			s1:    nil,
			s2:    nil,
			equal: true,
		},
		{
			name: "one nil",
			s1: &SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			s2:    nil,
			equal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.equal, tt.s1.Equal(tt.s2))
		})
	}
}
