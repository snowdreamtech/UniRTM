// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/http/httpproxy"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

type contextKey string

const routeStrategyContextKey contextKey = "unirtm_route_strategy"

// WithRouteStrategy stores the route strategy in the request context.
func WithRouteStrategy(ctx context.Context, strat RouteStrategy) context.Context {
	return context.WithValue(ctx, routeStrategyContextKey, strat)
}

// RouteStrategyFromContext retrieves the route strategy from the request context.
func RouteStrategyFromContext(ctx context.Context) (RouteStrategy, bool) {
	val, ok := ctx.Value(routeStrategyContextKey).(RouteStrategy)
	return val, ok
}

// HasConfiguredProxy checks if any HTTP/HTTPS/ALL proxy is set in the environment.
func HasConfiguredProxy() bool {
	return env.Get("HTTP_PROXY") != "" || env.Get("HTTPS_PROXY") != "" || env.Get("ALL_PROXY") != ""
}

// DirectTransport creates an http.Transport configured strictly for direct connections (no proxy).
func DirectTransport() *http.Transport {
	trans := DefaultTransport()
	trans.Proxy = nil
	return trans
}

// ProxyOnlyTransport creates an http.Transport configured to use the environment proxy without bypassing domestic mirrors.
func ProxyOnlyTransport() *http.Transport {
	trans := DefaultTransport()
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
		return proxyFunc(req.URL)
	}
	return trans
}

// AdaptiveTransport implements dynamic apex-domain routing with staggered racing,
// automatic fallback, and route memory caching.
type AdaptiveTransport struct {
	directTransport http.RoundTripper
	proxyTransport  http.RoundTripper
	routeCache      *DomainRouteCache
	staggerDelay    time.Duration
}

// NewAdaptiveTransport creates a new AdaptiveTransport instance.
func NewAdaptiveTransport(direct, proxy http.RoundTripper, cache *DomainRouteCache) *AdaptiveTransport {
	if cache == nil {
		cache = DefaultRouteCache()
	}
	return &AdaptiveTransport{
		directTransport: direct,
		proxyTransport:  proxy,
		routeCache:      cache,
		staggerDelay:    200 * time.Millisecond,
	}
}

// isProxyFailure checks if a response code or error indicates proxy failure.
func isProxyFailure(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	if resp != nil {
		switch resp.StatusCode {
		case http.StatusBadGateway, http.StatusGatewayTimeout, http.StatusProxyAuthRequired:
			return true
		}
	}
	return false
}

// isDirectFailure checks if an error indicates direct connection blockage (e.g. timeout or RST).
func isDirectFailure(resp *http.Response, err error) bool {
	return err != nil
}

// RoundTrip executes a single HTTP transaction with adaptive racing and route caching.
func (t *AdaptiveTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	host := req.URL.Hostname()

	// 1. Strict zero-proxy rule: if no proxy is configured, or host is explicitly marked domestic, use direct.
	if !HasConfiguredProxy() || ShouldBypassProxy(host) {
		return t.directTransport.RoundTrip(req)
	}

	// 2. Check if a route strategy was inherited via redirect context.
	if strat, ok := RouteStrategyFromContext(req.Context()); ok {
		resp, err := t.roundTripWithStrategy(req, host, strat)
		if err == nil {
			*req = *req.WithContext(WithRouteStrategy(req.Context(), strat))
		}
		return resp, err
	}

	// 3. Check if apex domain route is cached.
	if strat, ok := t.routeCache.Get(host); ok {
		resp, err := t.roundTripWithStrategy(req, host, strat)
		if err == nil {
			*req = *req.WithContext(WithRouteStrategy(req.Context(), strat))
		}
		return resp, err
	}

	// 4. For non-idempotent or non-rewindable requests, fall back to sequential direct -> proxy.
	if req.Method != http.MethodGet && req.Method != http.MethodHead || (req.Body != nil && req.GetBody == nil) {
		resp, err := t.directTransport.RoundTrip(req)
		if err == nil && !isDirectFailure(resp, err) {
			t.routeCache.Set(host, RouteDirect)
			*req = *req.WithContext(WithRouteStrategy(req.Context(), RouteDirect))
			return resp, nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		resp, err = t.proxyTransport.RoundTrip(req)
		if err == nil {
			t.routeCache.Set(host, RouteProxy)
			*req = *req.WithContext(WithRouteStrategy(req.Context(), RouteProxy))
		}
		return resp, err
	}

	// 5. Staggered Racing (Happy Eyeballs): Direct first, with a 200ms stagger window before proxy.
	resp, winner, err := t.race(req, host)
	if err == nil {
		*req = *req.WithContext(WithRouteStrategy(req.Context(), winner))
	}
	return resp, err
}

func (t *AdaptiveTransport) roundTripWithStrategy(req *http.Request, host string, strat RouteStrategy) (*http.Response, error) {
	primary := t.directTransport
	secondary := t.proxyTransport
	fallbackStrat := RouteProxy

	if strat == RouteProxy {
		primary = t.proxyTransport
		secondary = t.directTransport
		fallbackStrat = RouteDirect
	}

	resp, err := primary.RoundTrip(req)
	failed := false
	if strat == RouteProxy {
		failed = isProxyFailure(resp, err)
	} else {
		failed = isDirectFailure(resp, err)
	}

	if !failed {
		return resp, nil
	}

	// Evict failed route from cache and fallback to secondary
	t.routeCache.Invalidate(host)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}

	resp, err = secondary.RoundTrip(req)
	if err == nil {
		t.routeCache.Set(host, fallbackStrat)
	}
	return resp, err
}

type raceResult struct {
	resp     *http.Response
	err      error
	strategy RouteStrategy
}

func (t *AdaptiveTransport) race(req *http.Request, host string) (*http.Response, RouteStrategy, error) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	resultCh := make(chan raceResult, 2)

	// Helper to launch a pathway
	launch := func(trans http.RoundTripper, strat RouteStrategy) {
		clonedReq := req.Clone(ctx)
		if req.GetBody != nil {
			rc, err := req.GetBody()
			if err == nil {
				clonedReq.Body = rc
			}
		}
		resp, err := trans.RoundTrip(clonedReq)
		resultCh <- raceResult{resp: resp, err: err, strategy: strat}
	}

	// 1. Launch Direct channel
	go launch(t.directTransport, RouteDirect)

	// 2. Wait up to staggerDelay before launching Proxy channel
	timer := time.NewTimer(t.staggerDelay)
	defer timer.Stop()

	var proxyLaunched bool
	var firstErr error

	for i := 0; i < 2; i++ {
		select {
		case res := <-resultCh:
			// Check if winner
			var failed bool
			if res.strategy == RouteProxy {
				failed = isProxyFailure(res.resp, res.err)
			} else {
				failed = isDirectFailure(res.resp, res.err)
			}

			if !failed {
				// We have a winner! Cache the winning route.
				t.routeCache.Set(host, res.strategy)
				cancel() // Cancel the slower or pending channel
				return res.resp, res.strategy, nil
			}

			if res.resp != nil && res.resp.Body != nil {
				res.resp.Body.Close()
			}
			if firstErr == nil {
				firstErr = res.err
			}

			// If the first failed and proxy hasn't been launched yet, launch it immediately!
			if !proxyLaunched {
				proxyLaunched = true
				go launch(t.proxyTransport, RouteProxy)
			}

		case <-timer.C:
			// Direct is taking too long (> 200ms), wake up Proxy channel concurrently!
			if !proxyLaunched {
				proxyLaunched = true
				go launch(t.proxyTransport, RouteProxy)
			}
		}
	}

	if firstErr != nil {
		return nil, RouteDirect, firstErr
	}
	return nil, RouteDirect, errors.New("all connection pathways failed")
}
