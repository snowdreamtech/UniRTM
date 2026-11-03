// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package network

import (
	"os"
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

func TestDetectEnvProxy(t *testing.T) {
	// Set mock environment variables
	origHTTP := os.Getenv("HTTP_PROXY")
	origALL := os.Getenv("ALL_PROXY")
	defer func() {
		_ = os.Setenv("HTTP_PROXY", origHTTP)
		_ = os.Setenv("ALL_PROXY", origALL)
	}()

	_ = os.Setenv("HTTP_PROXY", "http://127.0.0.1:7890")
	_ = os.Setenv("ALL_PROXY", "socks5://127.0.0.1:10808")

	dp := detectEnvProxy()
	if !dp.Detected {
		t.Fatalf("expected env proxy to be detected")
	}
	if dp.HTTP != "http://127.0.0.1:7890" {
		t.Errorf("expected http://127.0.0.1:7890, got %s", dp.HTTP)
	}
	if dp.SOCKS5 != "socks5://127.0.0.1:10808" {
		t.Errorf("expected socks5://127.0.0.1:10808, got %s", dp.SOCKS5)
	}
	if dp.Port != 7890 {
		t.Errorf("expected port 7890, got %d", dp.Port)
	}
}
