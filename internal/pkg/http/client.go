// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/http/httpproxy"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)


// ClientConfig provides HTTP client configurations.
type ClientConfig struct{}

// MockTransport can be set during tests to intercept all HTTP/HTTPS requests
// created by UniRTM's DefaultTransport.
var MockTransport http.RoundTripper

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
	base, ok := http.DefaultTransport.(*http.Transport)
	var trans *http.Transport
	if ok {
		trans = base.Clone()
	} else {
		trans = &http.Transport{}
	}

	// 1. Smart proxy bypass + UNIRTM_/MISE_ env prefix support + NO_PROXY + ALL_PROXY
	//
	// Proxy config is resolved ONCE at transport creation time (not per request).
	// httpproxy.Config is used to correctly enforce NO_PROXY rules alongside
	// UNIRTM_/MISE_ prefixed proxy variables.
	httpProxy := env.Get("HTTP_PROXY")
	httpsProxy := env.Get("HTTPS_PROXY")
	if allProxy := env.Get("ALL_PROXY"); allProxy != "" {
		if httpProxy == "" {
			httpProxy = allProxy
		}
		if httpsProxy == "" {
			httpsProxy = allProxy
		}
	}

	noProxy := env.Get("NO_PROXY")
	if noProxy == "" {
		noProxy = "localhost,127.0.0.1,::1"
	}

	proxyFunc := (&httpproxy.Config{
		HTTPProxy:  httpProxy,
		HTTPSProxy: httpsProxy,
		NoProxy:    noProxy,
	}).ProxyFunc()

	trans.Proxy = func(req *http.Request) (*url.URL, error) {
		if ShouldBypassProxy(req.URL.Hostname()) {
			return nil, nil // DIRECT connection for domestic mirrors
		}
		return proxyFunc(req.URL)
	}

	// 2. Optional manual HTTP/2 opt-out for environments where proxy software
	//    corrupts HTTP/2 ALPN frames (smart auto-downgrade is handled at call sites).
	if env.Get("HTTP2") == "0" {
		DisableHTTP2(trans)
	}

	// 3. Connection pool optimization for high-concurrency downloads:
	// Go's default MaxIdleConnsPerHost is only 2, which causes connection thrashing
	// during parallel downloads to the same host (e.g. github.com / cdn mirrors).
	trans.MaxIdleConns = 200
	trans.MaxIdleConnsPerHost = 50
	trans.IdleConnTimeout = 90 * time.Second
	trans.ForceAttemptHTTP2 = true

	// 4. In-memory DNS cache: resolves hostnames with TTL caching and singleflight
	// deduplication, eliminating redundant DNS lookups across parallel HTTP calls.
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	dnsCache := DefaultDNSCache()
	trans.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return dialer.DialContext(ctx, network, addr)
		}

		ips, err := dnsCache.LookupIP(ctx, host)
		if err != nil || len(ips) == 0 {
			return dialer.DialContext(ctx, network, addr)
		}

		var lastErr error
		for _, ip := range ips {
			target := net.JoinHostPort(ip.String(), port)
			conn, err := dialer.DialContext(ctx, network, target)
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}

		// On failure, evict stale entry and retry once with original addr
		dnsCache.Evict(host)
		conn, err := dialer.DialContext(ctx, network, addr)
		if err == nil {
			return conn, nil
		}
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, err
	}

	return trans
}


var (
	sharedTransport     *http.Transport
	sharedTransportOnce sync.Once
)

// SharedTransport returns the singleton instance of UniRTM's robust http.Transport.
// Reusing this transport allows connection pooling and HTTP/1.1 or HTTP/2 keep-alive
// across different clients and backends, drastically reducing TCP/TLS handshake latency.
func SharedTransport() *http.Transport {
	sharedTransportOnce.Do(func() {
		sharedTransport = DefaultTransport()
	})
	return sharedTransport
}

// ResetSharedTransport resets the shared transport singleton (primarily used in tests).
func ResetSharedTransport() {
	sharedTransportOnce = sync.Once{}
	sharedTransport = nil
}

// NewClient returns an http.Client pre-configured with UniRTM's robust transport.
func NewClient() *http.Client {
	var tr http.RoundTripper
	if MockTransport != nil {
		tr = MockTransport
	} else if _, ok := http.DefaultTransport.(*http.Transport); !ok {
		tr = http.DefaultTransport
	} else {
		tr = SharedTransport()
	}
	return &http.Client{
		Transport: tr,
	}
}

// NewClientWithTimeout returns an http.Client with a timeout and the robust transport.
func NewClientWithTimeout(timeout time.Duration) *http.Client {
	var tr http.RoundTripper
	if MockTransport != nil {
		tr = MockTransport
	} else if _, ok := http.DefaultTransport.(*http.Transport); !ok {
		tr = http.DefaultTransport
	} else {
		tr = SharedTransport()
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}
}
