// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package network

import (
	"testing"
	"time"
)

func TestProxyConnectivity_Empty(t *testing.T) {
	res := TestProxyConnectivity(ProxyConfig{}, 1*time.Second)
	if res.Success {
		t.Errorf("expected failure on empty proxy config, got success")
	}
}

func TestProxyConnectivity_BogusPort(t *testing.T) {
	cfg := ProxyConfig{
		Enabled: true,
		HTTP:    "http://127.0.0.1:59999",
	}
	res := TestProxyConnectivity(cfg, 1*time.Second)
	if res.Success {
		t.Errorf("expected failure on bogus unreachable port, got success")
	}
	t.Logf("Bogus port failure message: %s", res.Error)
}

func TestProxyConnectivity_RealSingBox(t *testing.T) {
	cfg := ProxyConfig{
		Enabled: true,
		HTTP:    "http://127.0.0.1:8889",
	}
	res := TestProxyConnectivity(cfg, 3*time.Second)
	t.Logf("Real sing-box test result: %+v", res)
	if !res.Success {
		t.Logf("Notice: Proxy test failed (external network or proxy might be down): %s", res.Error)
	} else {
		if res.RTTMs <= 0 {
			t.Errorf("expected positive RTT, got %d", res.RTTMs)
		}
	}
}
