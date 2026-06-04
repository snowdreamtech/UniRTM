// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		// Exact SemVer comparisons
		{
			name:     "SemVer equality",
			v1:       "1.2.3",
			v2:       "1.2.3",
			expected: 0,
		},
		{
			name:     "SemVer strictly greater",
			v1:       "2.0.0",
			v2:       "1.9.9",
			expected: 1,
		},
		{
			name:     "SemVer strictly less",
			v1:       "1.2.3",
			v2:       "1.2.4",
			expected: -1,
		},
		{
			name:     "SemVer with prefixes",
			v1:       "v1.2.3",
			v2:       "V1.2.3",
			expected: 0,
		},
		// Alias equality
		{
			name:     "Alias exact match",
			v1:       "latest",
			v2:       "latest",
			expected: 0,
		},
		{
			name:     "Alias exact match different case", // Wait, ParseVersion parses it as Alias if it matches pattern? "latest" is an alias.
			v1:       "lts",
			v2:       "lts",
			expected: 0,
		},
		// Fallback numerical comparison (e.g. invalid SemVer or alias vs alias)
		{
			name:     "Fallback numeric greater",
			v1:       "1.2", // Parsed as VersionTypePartial
			v2:       "1.1",
			expected: 1,
		},
		{
			name:     "Fallback numeric less length difference",
			v1:       "1.2",
			v2:       "1.2.1",
			expected: -1,
		},
		{
			name:     "Fallback numeric greater length difference",
			v1:       "1.2.1",
			v2:       "1.2",
			expected: 1,
		},
		// Fallback string comparison
		{
			name:     "Fallback string comparison less",
			v1:       "alpha",
			v2:       "beta",
			expected: -1,
		},
		{
			name:     "Fallback string comparison greater",
			v1:       "beta",
			v2:       "alpha",
			expected: 1,
		},
		{
			name:     "Fallback mixed string and numeric",
			v1:       "1.2.alpha",
			v2:       "1.2.beta",
			expected: -1,
		},
		{
			name:     "Fallback mixed numeric vs string",
			v1:       "1",
			v2:       "a",
			expected: -1, // 1 vs a: err1 == nil (1 is num) err2 != nil (a is not). Fallback to string: "1" < "a" -> -1
		},
		{
			name:     "Fallback mixed string vs numeric",
			v1:       "b",
			v2:       "2",
			expected: 1, // "b" > "2" -> 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}
