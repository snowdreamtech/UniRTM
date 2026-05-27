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
		os.Setenv("ALL_PROXY", origAll)
		os.Setenv("HTTP2", origH2)
	}()

	os.Setenv("ALL_PROXY", "socks5://127.0.0.1:1080")
	os.Setenv("HTTP2", "0")

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
