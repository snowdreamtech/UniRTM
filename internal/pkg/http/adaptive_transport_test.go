// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
	calls         int32
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	atomic.AddInt32(&m.calls, 1)
	return m.roundTripFunc(req)
}

func makeDummyResponse(req *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Request:    req,
		Header:     make(http.Header),
	}
}

func TestAdaptiveTransport_NoProxy_StrictDirect(t *testing.T) {
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("ALL_PROXY", "")

	directRT := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return makeDummyResponse(req, http.StatusOK, "direct-response"), nil
		},
	}
	proxyRT := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			t.Fatal("proxy transport should NEVER be called when no proxy is configured")
			return nil, errors.New("unexpected proxy call")
		},
	}

	cache := NewDomainRouteCache(10 * time.Minute)
	adapt := NewAdaptiveTransport(directRT, proxyRT, cache)

	req, _ := http.NewRequest("GET", "https://github.com/foo/bar", nil)
	resp, err := adapt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if atomic.LoadInt32(&directRT.calls) != 1 {
		t.Errorf("expected 1 direct call, got %d", directRT.calls)
	}
	if atomic.LoadInt32(&proxyRT.calls) != 0 {
		t.Errorf("expected 0 proxy calls, got %d", proxyRT.calls)
	}
}

func TestAdaptiveTransport_DomesticMirror_StrictDirect(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:7890")

	directRT := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return makeDummyResponse(req, http.StatusOK, "mirror-direct"), nil
		},
	}
	proxyRT := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			t.Fatal("proxy transport should NOT be called for domestic mirrors")
			return nil, errors.New("unexpected proxy call")
		},
	}

	cache := NewDomainRouteCache(10 * time.Minute)
	adapt := NewAdaptiveTransport(directRT, proxyRT, cache)

	req, _ := http.NewRequest("GET", "https://mirrors.aliyun.com/foo", nil)
	resp, err := adapt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if atomic.LoadInt32(&directRT.calls) != 1 {
		t.Errorf("expected 1 direct call, got %d", directRT.calls)
	}
	if atomic.LoadInt32(&proxyRT.calls) != 0 {
		t.Errorf("expected 0 proxy calls, got %d", proxyRT.calls)
	}
}

func TestAdaptiveTransport_StaggeredRacing_DirectWins(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:7890")

	directRT := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Responds fast in 20ms
			time.Sleep(20 * time.Millisecond)
			return makeDummyResponse(req, http.StatusOK, "fast-direct"), nil
		},
	}
	proxyRT := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			time.Sleep(100 * time.Millisecond)
			return makeDummyResponse(req, http.StatusOK, "proxy-response"), nil
		},
	}

	cache := NewDomainRouteCache(10 * time.Minute)
	adapt := NewAdaptiveTransport(directRT, proxyRT, cache)
	adapt.staggerDelay = 50 * time.Millisecond

	req, _ := http.NewRequest("GET", "https://example.com/test", nil)
	resp, err := adapt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	// Direct won
	strat, ok := cache.Get("example.com")
	if !ok || strat != RouteDirect {
		t.Errorf("expected RouteDirect cached, got %v (ok=%v)", strat, ok)
	}
}

func TestAdaptiveTransport_StaggeredRacing_ProxyWinsWhenDirectBlocked(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:7890")

	directRT := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Direct is blocked/hung by GFW, will take 500ms
			select {
			case <-time.After(500 * time.Millisecond):
				return nil, errors.New("connection reset by peer")
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		},
	}
	proxyRT := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			time.Sleep(30 * time.Millisecond)
			return makeDummyResponse(req, http.StatusOK, "proxy-success"), nil
		},
	}

	cache := NewDomainRouteCache(10 * time.Minute)
	adapt := NewAdaptiveTransport(directRT, proxyRT, cache)
	adapt.staggerDelay = 50 * time.Millisecond

	req, _ := http.NewRequest("GET", "https://objects.githubusercontent.com/file.tar.gz", nil)
	resp, err := adapt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	// Proxy should win
	strat, ok := cache.Get("objects.githubusercontent.com")
	if !ok || strat != RouteProxy {
		t.Errorf("expected RouteProxy cached, got %v (ok=%v)", strat, ok)
	}

	// Subdomain sharing apex domain should also hit cache
	stratRaw, okRaw := cache.Get("raw.githubusercontent.com")
	if !okRaw || stratRaw != RouteProxy {
		t.Errorf("expected RouteProxy shared for raw.githubusercontent.com, got %v", stratRaw)
	}
}

func TestAdaptiveTransport_CacheHit_ZeroStagger(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:7890")

	var directCalled, proxyCalled int32
	directRT := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&directCalled, 1)
			return nil, errors.New("should not call direct on proxy hit")
		},
	}
	proxyRT := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&proxyCalled, 1)
			return makeDummyResponse(req, http.StatusOK, "cached-proxy-hit"), nil
		},
	}

	cache := NewDomainRouteCache(10 * time.Minute)
	// Pre-populate cache
	cache.Set("githubusercontent.com", RouteProxy)

	adapt := NewAdaptiveTransport(directRT, proxyRT, cache)

	req, _ := http.NewRequest("GET", "https://objects.githubusercontent.com/bar", nil)
	resp, err := adapt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if atomic.LoadInt32(&proxyCalled) != 1 {
		t.Errorf("expected 1 proxy call, got %d", proxyCalled)
	}
	if atomic.LoadInt32(&directCalled) != 0 {
		t.Errorf("expected 0 direct calls on cache hit, got %d", directCalled)
	}
}

func TestAdaptiveTransport_ProxyFailure_FallbackToDirect(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:7890")

	proxyRT := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Proxy gives 502 Bad Gateway
			return makeDummyResponse(req, http.StatusBadGateway, "bad gateway"), nil
		},
	}
	directRT := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return makeDummyResponse(req, http.StatusOK, "fallback-direct-ok"), nil
		},
	}

	cache := NewDomainRouteCache(10 * time.Minute)
	cache.Set("example.org", RouteProxy)

	adapt := NewAdaptiveTransport(directRT, proxyRT, cache)

	req, _ := http.NewRequest("GET", "https://example.org/data", nil)
	resp, err := adapt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Cache should be updated to RouteDirect
	strat, ok := cache.Get("example.org")
	if !ok || strat != RouteDirect {
		t.Errorf("expected cache updated to RouteDirect, got %v", strat)
	}
}

func TestAdaptiveTransport_RedirectContextInheritance(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:7890")

	var directCalls, proxyCalls int32
	directRT := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&directCalls, 1)
			return makeDummyResponse(req, http.StatusOK, "direct"), nil
		},
	}
	proxyRT := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&proxyCalls, 1)
			return makeDummyResponse(req, http.StatusOK, "proxy"), nil
		},
	}

	cache := NewDomainRouteCache(10 * time.Minute)
	adapt := NewAdaptiveTransport(directRT, proxyRT, cache)

	// Simulate inherited RouteProxy context from a 302 redirect
	ctx := WithRouteStrategy(t.Context(), RouteProxy)
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://some-uncached-domain.com/path", nil)

	resp, err := adapt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if atomic.LoadInt32(&proxyCalls) != 1 {
		t.Errorf("expected inherited RouteProxy to be used, got %d proxy calls", proxyCalls)
	}
	if atomic.LoadInt32(&directCalls) != 0 {
		t.Errorf("expected 0 direct calls, got %d", directCalls)
	}
}
