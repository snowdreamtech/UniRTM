// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package network

import (
	"testing"
)

func TestDetectSystemProxy(t *testing.T) {
	detected := DetectSystemProxy()
	t.Logf("Detected proxy: %+v", detected)

	// In this environment, sing-box is actively running on 127.0.0.1:8889
	if detected.Detected {
		if detected.Port <= 0 {
			t.Errorf("expected positive port, got %d", detected.Port)
		}
		if detected.HTTP == "" && detected.SOCKS5 == "" {
			t.Errorf("expected at least http or socks5 proxy URL")
		}
		t.Logf("Successfully detected running proxy: %s (%s)", detected.Type, detected.Description)
	} else {
		t.Log("No system proxy detected")
	}
}
