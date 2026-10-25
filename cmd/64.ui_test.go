// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"
)

func TestUICmdFlagsAndAliases(t *testing.T) {
	if uiCmd.Name() != "ui" {
		t.Fatalf("Expected command name 'ui', got %s", uiCmd.Name())
	}

	// Verify aliases
	aliases := uiCmd.Aliases
	hasDashboard := false
	hasWeb := false
	for _, a := range aliases {
		if a == "dashboard" {
			hasDashboard = true
		}
		if a == "web" {
			hasWeb = true
		}
	}
	if !hasDashboard || !hasWeb {
		t.Errorf("Expected aliases to contain 'dashboard' and 'web', got %v", aliases)
	}

	// Verify flags exist
	portFlag := uiCmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Error("Flag --port not found")
	}

	hostFlag := uiCmd.Flags().Lookup("host")
	if hostFlag == nil {
		t.Error("Flag --host not found")
	}

	noOpenFlag := uiCmd.Flags().Lookup("no-open")
	if noOpenFlag == nil {
		t.Error("Flag --no-open not found")
	}

	browserFlag := uiCmd.Flags().Lookup("browser")
	if browserFlag == nil {
		t.Error("Flag --browser not found")
	}
	if browserFlag.Shorthand != "b" {
		t.Errorf("Expected shorthand 'b' for browser flag, got '%s'", browserFlag.Shorthand)
	}
}
