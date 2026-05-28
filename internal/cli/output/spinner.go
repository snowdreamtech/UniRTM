// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package output

import (
	"os"
	"strings"

	"github.com/pterm/pterm"
)

// IsTesting checks if the current process is a test process.
func IsTesting() bool {
	return strings.HasSuffix(os.Args[0], ".test") ||
		strings.Contains(os.Args[0], "/_test/") ||
		os.Getenv("UNIRTM_TESTING") == "1" ||
		os.Getenv("CI") != "" // Usually CI runs tests, and we don't want spinners messing up CI logs anyway
}

// StartSpinner starts a pterm spinner, or returns a safe dummy if running in tests
// to avoid pterm's internal data race with `go test -race`.
func StartSpinner(text string) (*pterm.SpinnerPrinter, error) {
	sp := pterm.DefaultSpinner
	sp.Text = text

	if IsTesting() {
		// Mock a started spinner to avoid spawning the goroutine
		sp.IsActive = true
		return &sp, nil
	}

	return sp.Start(text)
}
