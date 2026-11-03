// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLsRemoteBackendName(t *testing.T) {
	// Empty tool defaults
	lsRemoteBackend = ""
	assert.Equal(t, "github", getLsRemoteBackendName("golang.org/x/vuln/cmd/govulncheck"))

	lsRemoteBackend = "npm"
	assert.Equal(t, "npm", getLsRemoteBackendName("node"))
	lsRemoteBackend = ""
}
