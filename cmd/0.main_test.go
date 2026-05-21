// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"testing"

	"github.com/pterm/pterm"
)

// TestMain is the entry point for all tests in the cmd package.
// It disables pterm styling before any test runs.
//
// Rationale: pterm.SpinnerPrinter.Start() spawns an animation goroutine that
// reads IsActive in a loop. This goroutine is only created when pterm.RawOutput
// is false (i.e. styled output is enabled). When running 'go test -race', the
// goroutine's unsynchronised read of IsActive races with the write from Stop().
// Calling DisableStyling() sets RawOutput = true, which prevents the goroutine
// from being created and eliminates the DATA RACE entirely.
func TestMain(m *testing.M) {
	pterm.DisableStyling()
	os.Exit(m.Run())
}
