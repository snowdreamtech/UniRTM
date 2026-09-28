// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShimRecursionGuardClearing(t *testing.T) {
	// Verify that setting and unsetting the recursion guard environment variable works
	os.Setenv("_UNIRTM_SHIM_RECURSION_GUARD", "1")
	assert.Equal(t, "1", os.Getenv("_UNIRTM_SHIM_RECURSION_GUARD"))

	os.Unsetenv("_UNIRTM_SHIM_RECURSION_GUARD")
	assert.Empty(t, os.Getenv("_UNIRTM_SHIM_RECURSION_GUARD"))
}
