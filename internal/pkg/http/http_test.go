// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"net/http"
	"testing"
	"time"
)

func TestShouldBypassProxy(t *testing.T) {
	tests := []struct {
		host     string
		expected bool
	}{
		{"example.com", false},
		{"google.com", false},
		{"aliyun.com", true},
		{"npmmirror.com", true},
		{"foo.cn", true},
		{"registry.npmmirror.com", true},
		{"mirrors.aliyun.com", true},
		{"test.mirror.com", true},
	}

	for _, tc := range tests {
		t.Run(tc.host, func(t *testing.T) {
			result := ShouldBypassProxy(tc.host)
			if result != tc.expected {
				t.Errorf("ShouldBypassProxy(%q) = %v; expected %v", tc.host, result, tc.expected)
			}
		})
	}
}

func TestDisableHTTP2(t *testing.T) {
	trans := http.DefaultTransport.(*http.Transport).Clone()

	DisableHTTP2(trans)

	if trans.ForceAttemptHTTP2 {
		t.Error("expected ForceAttemptHTTP2 to be false")
	}

	if trans.TLSNextProto == nil {
		t.Error("expected TLSNextProto to be initialized")
	}

	if trans.TLSClientConfig == nil {
		t.Error("expected TLSClientConfig to be initialized")
	} else if len(trans.TLSClientConfig.NextProtos) != 1 || trans.TLSClientConfig.NextProtos[0] != "http/1.1" {
		t.Errorf("expected NextProtos to be ['http/1.1'], got %v", trans.TLSClientConfig.NextProtos)
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Transport == nil {
		t.Error("expected non-nil transport")
	}
}

func TestNewClientWithTimeout(t *testing.T) {
	timeout := 10 * time.Second
	client := NewClientWithTimeout(timeout)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, client.Timeout)
	}
}

func TestDefaultTransport(t *testing.T) {
	trans := DefaultTransport()
	if trans == nil {
		t.Fatal("expected non-nil transport")
	}
	if trans.Proxy == nil {
		t.Error("expected non-nil proxy func")
	}
}

func TestDisableHTTP2_Nil(t *testing.T) {
	// Should not panic
	DisableHTTP2(nil)
}

func TestDisableHTTP2_NoTLS(t *testing.T) {
	trans := &http.Transport{}
	DisableHTTP2(trans)
	if trans.TLSClientConfig == nil {
		t.Error("expected TLSClientConfig to be initialized")
	}
}
