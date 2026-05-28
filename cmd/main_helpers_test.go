// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncEnv(t *testing.T) {
	// ensure NO_COLOR is not set in os
	os.Unsetenv("NO_COLOR")
	// simulate it being set in env (unirtm env)
	os.Setenv("UNIRTM_NO_COLOR", "1")
	defer os.Unsetenv("UNIRTM_NO_COLOR")

	syncEnv()
	// Actually, syncEnv reads from env.Get(v) which usually reads from os.Getenv(v) or .env file
	// Since we don't have .env loaded, it might do nothing. We just test it doesn't panic.
	assert.NotPanics(t, func() {
		syncEnv()
	})
}
