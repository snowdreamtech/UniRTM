// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package network

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// ProbeAll concurrently pings probe targets and returns their RTT results.
func ProbeAll(targets []ProbeTarget, timeout time.Duration) []ProbeResult {
	if len(targets) == 0 {
		targets = DefaultProbeTargets()
	}
	if timeout <= 0 {
		timeout = 2500 * time.Millisecond
	}

	results := make([]ProbeResult, len(targets))
	var wg sync.WaitGroup

	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	for i, target := range targets {
		wg.Add(1)
		go func(idx int, t ProbeTarget) {
			defer wg.Done()
			results[idx] = ProbeOne(client, t, timeout)
		}(i, target)
	}

	wg.Wait()
	return results
}

// ProbeOne executes a single latency probe against a target URL.
func ProbeOne(client *http.Client, t ProbeTarget, timeout time.Duration) ProbeResult {
	res := ProbeResult{
		Ecosystem: t.Ecosystem,
		Name:      t.Name,
		Endpoint:  t.Endpoint,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, t.Endpoint, nil)
	if err != nil {
		res.Status = "error"
		res.Error = err.Error()
		return res
	}

	req.Header.Set("User-Agent", "UniRTM-Probe/0.1.0")

	start := time.Now()
	resp, err := client.Do(req)
	rtt := time.Since(start).Milliseconds()
	res.RTTMs = rtt

	if err != nil {
		res.Status = "timeout"
		res.Error = "timeout"
		return res
	}
	defer resp.Body.Close()

	if rtt < 100 {
		res.Status = "fast"
	} else if rtt < 300 {
		res.Status = "normal"
	} else {
		res.Status = "slow"
	}

	return res
}
