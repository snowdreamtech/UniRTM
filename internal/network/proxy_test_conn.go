// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package network

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

// ProxyTestResult represents the connectivity and health check response for configured proxy.
type ProxyTestResult struct {
	Success      bool   `json:"success"`
	RTTMs        int64  `json:"rtt_ms"`
	HTTPStatus   string `json:"http_status"`   // "ok", "error", "skipped"
	SOCKS5Status string `json:"socks5_status"` // "ok", "error", "skipped"
	Message      string `json:"message"`
	Error        string `json:"error,omitempty"`
}

// TestProxyConnectivity performs real-world end-to-end connectivity verification on user-supplied proxy settings.
func TestProxyConnectivity(cfg ProxyConfig, timeout time.Duration) ProxyTestResult {
	if timeout <= 0 {
		timeout = 3500 * time.Millisecond
	}

	res := ProxyTestResult{
		HTTPStatus:   "skipped",
		SOCKS5Status: "skipped",
	}

	if cfg.HTTP == "" && cfg.HTTPS == "" && cfg.SOCKS5 == "" {
		res.Success = false
		res.Error = "未配置任何代理地址（HTTP/HTTPS 与 SOCKS5 均为空）"
		res.Message = res.Error
		return res
	}

	testTarget := "http://connectivitycheck.gstatic.com/generate_204"

	var totalRTT int64
	var count int64

	// 1. Test HTTP / HTTPS proxy if configured
	httpURLStr := cfg.HTTP
	if httpURLStr == "" {
		httpURLStr = cfg.HTTPS
	}

	if httpURLStr != "" {
		pURL, err := url.Parse(httpURLStr)
		if err != nil || pURL.Host == "" {
			res.HTTPStatus = "error"
			res.Error = fmt.Sprintf("HTTP 代理地址格式无效: %s", httpURLStr)
			res.Message = res.Error
			return res
		}

		// First, check local/remote TCP port reachability
		conn, err := net.DialTimeout("tcp", pURL.Host, 1500*time.Millisecond)
		if err != nil {
			res.HTTPStatus = "error"
			res.Error = fmt.Sprintf("无法连接到代理服务器端口 %s: %v", pURL.Host, err)
			res.Message = res.Error
			return res
		}
		_ = conn.Close()

		// Second, perform actual HTTP handshake through proxy
		transport := &http.Transport{
			Proxy: http.ProxyURL(pURL),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		}
		client := &http.Client{
			Transport: transport,
			Timeout:   timeout,
		}

		start := time.Now()
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, testTarget, nil)
		if err != nil {
			res.HTTPStatus = "error"
			res.Error = fmt.Sprintf("创建代理测试请求失败: %v", err)
			res.Message = res.Error
			return res
		}
		req.Header.Set("User-Agent", "UniRTM-Proxy-Tester/0.1.0")

		resp, err := client.Do(req)
		rtt := time.Since(start).Milliseconds()
		if err != nil {
			res.HTTPStatus = "error"
			res.Error = fmt.Sprintf("通过 HTTP 代理访问外网失败: %v", err)
			res.Message = res.Error
			return res
		}
		_ = resp.Body.Close()

		res.HTTPStatus = "ok"
		totalRTT += rtt
		count++
	}

	// 2. Test SOCKS5 proxy if configured
	if cfg.SOCKS5 != "" {
		sURL, err := url.Parse(cfg.SOCKS5)
		if err != nil || sURL.Host == "" {
			res.SOCKS5Status = "error"
			if res.Error == "" {
				res.Error = fmt.Sprintf("SOCKS5 代理地址格式无效: %s", cfg.SOCKS5)
				res.Message = res.Error
			}
			return res
		}

		// Check TCP port reachability
		conn, err := net.DialTimeout("tcp", sURL.Host, 1500*time.Millisecond)
		if err != nil {
			res.SOCKS5Status = "error"
			if res.Error == "" {
				res.Error = fmt.Sprintf("无法连接到 SOCKS5 代理服务器端口 %s: %v", sURL.Host, err)
				res.Message = res.Error
			}
			return res
		}
		_ = conn.Close()

		// Test SOCKS5 tunnel
		dialer, err := proxy.FromURL(sURL, proxy.Direct)
		if err != nil {
			res.SOCKS5Status = "error"
			if res.Error == "" {
				res.Error = fmt.Sprintf("创建 SOCKS5 代理拨号器失败: %v", err)
				res.Message = res.Error
			}
			return res
		}

		start := time.Now()
		socksConn, err := dialer.Dial("tcp", "1.1.1.1:53")
		if err != nil {
			socksConn, err = dialer.Dial("tcp", "www.google.com:80")
		}
		rtt := time.Since(start).Milliseconds()

		if err != nil {
			res.SOCKS5Status = "error"
			if res.Error == "" {
				res.Error = fmt.Sprintf("通过 SOCKS5 代理隧道连接外网失败: %v", err)
				res.Message = res.Error
			}
			return res
		}
		_ = socksConn.Close()

		res.SOCKS5Status = "ok"
		totalRTT += rtt
		count++
	}

	if count > 0 {
		avgRTT := totalRTT / count
		if avgRTT <= 0 {
			avgRTT = 1
		}
		res.Success = true
		res.RTTMs = avgRTT
		res.Message = fmt.Sprintf("代理服务器连通正常 (响应延迟 %dms)", avgRTT)
		return res
	}

	res.Success = false
	if res.Error == "" {
		res.Error = "代理连通性检测未通过"
	}
	res.Message = res.Error
	return res
}
