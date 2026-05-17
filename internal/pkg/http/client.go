package http

import (
	"net/http"
	"net/url"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// DefaultTransport returns a customized http.Transport that includes UniRTM's
// intelligent proxy bypass and HTTP/2 fallback configurations.
func DefaultTransport() *http.Transport {
	// Start with a clone of Go's default transport
	trans := http.DefaultTransport.(*http.Transport).Clone()

	// Reinforce network resilience with extended timeouts suitable for both API and large file downloads
	trans.MaxIdleConns = 100
	trans.IdleConnTimeout = 90 * time.Second
	trans.TLSHandshakeTimeout = 30 * time.Second
	trans.ResponseHeaderTimeout = 30 * time.Second
	trans.ExpectContinueTimeout = 5 * time.Second

	// 1. Configure Smart Proxy Bypass
	trans.Proxy = func(req *http.Request) (*url.URL, error) {
		if ShouldBypassProxy(req.URL.Hostname()) {
			return nil, nil // DIRECT connection
		}

		// Support custom UNIRTM_ and MISE_ prefixes for proxies via env.Get
		if req.URL.Scheme == "http" {
			if v := env.Get("HTTP_PROXY"); v != "" {
				return url.Parse(v)
			}
		} else if req.URL.Scheme == "https" {
			if v := env.Get("HTTPS_PROXY"); v != "" {
				return url.Parse(v)
			}
		}
		if v := env.Get("ALL_PROXY"); v != "" {
			return url.Parse(v)
		}

		// Fallback to standard HTTP_PROXY/HTTPS_PROXY
		return http.ProxyFromEnvironment(req)
	}

	// 2. Configure HTTP/2 Fallback
	if env.Get("HTTP2") == "0" {
		DisableHTTP2(trans)
	}

	return trans
}

// NewClient returns an http.Client pre-configured with UniRTM's robust transport.
func NewClient() *http.Client {
	return &http.Client{
		Transport: DefaultTransport(),
	}
}

// NewClientWithTimeout returns an http.Client with a timeout and the robust transport.
func NewClientWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: DefaultTransport(),
	}
}
