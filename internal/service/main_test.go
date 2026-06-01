// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service_test

import (
	"os"
	"testing"

	"github.com/pterm/pterm"
	"github.com/rs/zerolog"
)

func TestMain(m *testing.M) {
	// Disable logging during tests to prevent log output from interfering with Example test Output checks
	zerolog.SetGlobalLevel(zerolog.Disabled)
	pterm.DisableOutput()
	os.Exit(m.Run())
}
