// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEcosystemLabel(t *testing.T) {
	assert.Equal(t, "github-actions", getEcosystemLabel("github-actions"))
	assert.Equal(t, "javascript", getEcosystemLabel("npm"))
	assert.Equal(t, "python", getEcosystemLabel("pip"))
	assert.Equal(t, "go", getEcosystemLabel("gomod"))
	assert.Equal(t, "rust", getEcosystemLabel("cargo"))
	assert.Equal(t, "php", getEcosystemLabel("composer"))
	assert.Equal(t, "ruby", getEcosystemLabel("bundler"))
	assert.Equal(t, "docker", getEcosystemLabel("docker"))
	assert.Equal(t, "other", getEcosystemLabel("unknown"))
}
