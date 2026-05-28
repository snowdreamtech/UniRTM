// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestExecutable(t *testing.T) {
	err := testExecutable("/non/existent/path/to/executable", nil)
	assert.Error(t, err)
}
