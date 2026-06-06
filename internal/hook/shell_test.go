// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hook

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetShell_ReturnsNonEmpty(t *testing.T) {
	shell := GetShell()
	if shell == "" {
		t.Error("GetShell() returned empty string")
	}
}

func TestGetShell_PlatformSpecific(t *testing.T) {
	shell := GetShell()
	switch runtime.GOOS {
	case "windows":
		// On Windows, should resolve to sh.exe or plain "sh"
		if !strings.HasSuffix(shell, "sh") && !strings.HasSuffix(shell, "sh.exe") {
			t.Errorf("expected sh or sh.exe on Windows, got %q", shell)
		}
	default:
		// On Unix-like systems, must be /bin/sh
		if shell != "/bin/sh" {
			t.Errorf("expected /bin/sh on %s, got %q", runtime.GOOS, shell)
		}
	}
}
