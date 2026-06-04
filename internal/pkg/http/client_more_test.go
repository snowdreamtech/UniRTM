// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func TestDefaultTransport_EnvVars(t *testing.T) {
	// Backup
	origAll := os.Getenv("ALL_PROXY")
	origH2 := os.Getenv("HTTP2")
	defer func() {
		t.Setenv("ALL_PROXY", origAll)
		t.Setenv("HTTP2", origH2)
	}()

	t.Setenv("ALL_PROXY", "socks5://127.0.0.1:1080")
	t.Setenv("HTTP2", "0")

	tr := DefaultTransport()
	if tr == nil {
		t.Fatal("expected transport")
	}

	client := NewClientWithTimeout(5 * time.Second)
	if client.Timeout != 5*time.Second {
		t.Fatalf("expected timeout 5s, got %v", client.Timeout)
	}

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	proxyUrl, err := tr.Proxy(req)
	if err != nil {
		t.Fatalf("unexpected proxy error: %v", err)
	}
	if proxyUrl == nil || proxyUrl.String() != "socks5://127.0.0.1:1080" {
		t.Fatalf("expected proxyUrl to be socks5://127.0.0.1:1080, got %v", proxyUrl)
	}

	reqMirror, _ := http.NewRequest("GET", "https://npmmirror.com", nil)
	proxyUrlMirror, _ := tr.Proxy(reqMirror)
	if proxyUrlMirror != nil {
		t.Fatalf("expected nil proxy for mirror, got %v", proxyUrlMirror)
	}
}

// Test proxy logic without ALL_PROXY
func TestDefaultTransport_EnvVars_NoAllProxy(t *testing.T) {
	origAll := os.Getenv("ALL_PROXY")
	origHttp := os.Getenv("HTTP_PROXY")
	origHttps := os.Getenv("HTTPS_PROXY")
	defer func() {
		t.Setenv("ALL_PROXY", origAll)
		t.Setenv("HTTP_PROXY", origHttp)
		t.Setenv("HTTPS_PROXY", origHttps)
	}()

	t.Setenv("ALL_PROXY", "")
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:8080")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:8080")

	tr := DefaultTransport()
	if tr == nil {
		t.Fatal("expected transport")
	}

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	proxyUrl, _ := tr.Proxy(req)
	if proxyUrl == nil || proxyUrl.String() != "http://127.0.0.1:8080" {
		t.Fatalf("expected proxyUrl to be http://127.0.0.1:8080, got %v", proxyUrl)
	}
}

// Test MockTransport logic
func TestDefaultTransport_MockTransport(t *testing.T) {
	oldMock := MockTransport
	defer func() {
		MockTransport = oldMock
	}()

	// Create a dummy RoundTripper
	mockRt := http.DefaultTransport
	MockTransport = mockRt

	// Should not panic, should register protocols
	tr := DefaultTransport()
	if tr == nil {
		t.Fatal("expected transport")
	}

	client := NewClient()
	if client.Transport != mockRt {
		t.Fatal("expected client.Transport to be mockRt")
	}

	clientTimeout := NewClientWithTimeout(5 * time.Second)
	if clientTimeout.Transport != mockRt {
		t.Fatal("expected clientTimeout.Transport to be mockRt")
	}
}
