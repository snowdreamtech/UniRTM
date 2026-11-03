// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion_EdgeCases(t *testing.T) {
	// ParseVersion errors
	_, err := ParseVersion("^99999999999999999999999.0.0")
	assert.Error(t, err)

	_, err = ParseVersion("~99999999999999999999999.0.0")
	assert.Error(t, err)

	_, err = ParseVersion(">=99999999999999999999999.0.0")
	assert.Error(t, err)

	_, err = ParseVersion("v99999999999999999999999.0.0")
	assert.Error(t, err)

	_, err = ParseSemVer("v1.99999999999999999999999.0")
	assert.Error(t, err)

	_, err = ParseSemVer("v1.0.99999999999999999999999")
	assert.Error(t, err)

	_, err = ParseSemVerPartial("1.0.99999999999999999999999-alpha")
	assert.Error(t, err)

	_, err = ParseSemVerPartial("1.0.99999999999999999999999+build")
	assert.Error(t, err)

	// FormatVersion errors
	_, err = FormatVersion(&Version{
		Type:     VersionTypeRange,
		RangeVer: &SemVer{Major: 1},
		RangeOp:  RangeOperator("unknown"),
	})
	assert.Error(t, err)

	_, err = FormatVersion(&Version{
		Type: VersionType(99),
	})
	assert.Error(t, err)

	// Equal with different types
	v1 := &Version{Type: VersionTypeExact}
	v2 := &Version{Type: VersionTypeAlias}
	assert.False(t, v1.Equal(v2))

	v3 := &Version{Type: VersionType(99)}
	v4 := &Version{Type: VersionType(99)}
	assert.False(t, v3.Equal(v4))
}
