// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/http/httpproxy"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

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
	trans := http.DefaultTransport.(*http.Transport).Clone()

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
	proxyFunc := (&httpproxy.Config{
		HTTPProxy:  httpProxy,
		HTTPSProxy: httpsProxy,
		NoProxy:    env.Get("NO_PROXY"),
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

	return trans
}

// NewClient returns an http.Client pre-configured with UniRTM's robust transport.
func NewClient() *http.Client {
	var tr http.RoundTripper = DefaultTransport()
	if MockTransport != nil {
		tr = MockTransport
	}
	return &http.Client{
		Transport: tr,
	}
}

// NewClientWithTimeout returns an http.Client with a timeout and the robust transport.
func NewClientWithTimeout(timeout time.Duration) *http.Client {
	var tr http.RoundTripper = DefaultTransport()
	if MockTransport != nil {
		tr = MockTransport
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}
}
