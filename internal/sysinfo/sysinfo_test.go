// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package sysinfo

import (
	"runtime"
	"testing"
)

func TestIsMusl(t *testing.T) {
	// This test simply verifies that IsMusl runs without panicking.
	// We cannot strictly assert true/false unless we mock the filesystem/exec,
	// but it shouldn't crash.
	result := IsMusl()
	if runtime.GOOS != "linux" && result == true {
		t.Errorf("IsMusl() returned true on non-linux OS (%s)", runtime.GOOS)
	}
	t.Logf("IsMusl() returned: %v on OS: %s", result, runtime.GOOS)
}
