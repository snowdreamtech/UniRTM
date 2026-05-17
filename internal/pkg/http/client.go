package http

import (
	"net/http"
	"net/url"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// DefaultTransport returns UniRTM's standard http.Transport.
//
// It customizes two behaviors that Go's default transport cannot provide:
//
//  1. Smart Proxy Bypass: domestic mirror domains (aliyun.com, npmmirror.com, etc.)
//     are forced to use DIRECT connections, preventing local proxy software from
//     returning "Bad Request" errors when routing Chinese CDN traffic.
//
//  2. UNIRTM_/MISE_ env prefix support: reads HTTP_PROXY/HTTPS_PROXY/ALL_PROXY
//     through env.Get(), which resolves UNIRTM_HTTP_PROXY and MISE_HTTP_PROXY
//     in addition to the standard names that http.ProxyFromEnvironment covers.
//
// All other settings (connection pool, timeouts) are inherited from Go's
// http.DefaultTransport via Clone(), so they stay in sync with upstream defaults.
func DefaultTransport() *http.Transport {
	trans := http.DefaultTransport.(*http.Transport).Clone()

	// 1. Smart proxy bypass + UNIRTM_/MISE_ env prefix support
	trans.Proxy = func(req *http.Request) (*url.URL, error) {
		if ShouldBypassProxy(req.URL.Hostname()) {
			return nil, nil // DIRECT connection for domestic mirrors
		}

		// Resolve UNIRTM_HTTP_PROXY / MISE_HTTP_PROXY / HTTP_PROXY etc.
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

		return http.ProxyFromEnvironment(req)
	}

	// 2. Optional manual HTTP/2 opt-out for environments where proxy software
	//    corrupts HTTP/2 ALPN frames (smart auto-downgrade is handled at call sites).
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

