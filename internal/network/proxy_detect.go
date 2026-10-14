// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package network

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DetectedProxy represents proxy details discovered automatically from the host environment.
type DetectedProxy struct {
	Detected    bool   `json:"detected"`
	Type        string `json:"type"` // "sing-box", "clash", "v2ray", "env", "port"
	HTTP        string `json:"http"`
	HTTPS       string `json:"https"`
	SOCKS5      string `json:"socks5"`
	Port        int    `json:"port"`
	Description string `json:"description"`
}

// DetectSystemProxy inspects sing-box, Clash, env vars, and active ports to identify running local proxies.
func DetectSystemProxy() DetectedProxy {
	// 1. Check sing-box configuration files and active ports
	if dp := detectSingBox(); dp.Detected {
		return dp
	}

	// 2. Check Clash / Mihomo configurations
	if dp := detectClash(); dp.Detected {
		return dp
	}

	// 3. Check environment variables
	if dp := detectEnvProxy(); dp.Detected {
		return dp
	}

	// 4. Active probing of common proxy ports
	if dp := probeCommonPorts(); dp.Detected {
		return dp
	}

	return DetectedProxy{Detected: false}
}

func isPortOpen(host string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func detectSingBox() DetectedProxy {
	home, _ := os.UserHomeDir()
	candidatePaths := []string{
		"/etc/sing-box/config.json",
		"/var/lib/sing-box/config.json",
		filepath.Join(home, ".config", "sing-box", "config.json"),
		"/usr/local/etc/sing-box/config.json",
	}

	type Inbound struct {
		Type       string `json:"type"`
		Listen     string `json:"listen"`
		ListenPort int    `json:"listen_port"`
	}
	type SingBoxConfig struct {
		Inbounds []Inbound `json:"inbounds"`
	}

	for _, p := range candidatePaths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}

		var cfg SingBoxConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}

		dp := DetectedProxy{
			Type: "sing-box",
		}

		for _, in := range cfg.Inbounds {
			listenHost := in.Listen
			if listenHost == "" || listenHost == "0.0.0.0" || listenHost == "::" {
				listenHost = "127.0.0.1"
			}
			if in.ListenPort <= 0 {
				continue
			}

			// Verify that the port is actively listening
			if !isPortOpen(listenHost, in.ListenPort, 150*time.Millisecond) {
				continue
			}

			dp.Port = in.ListenPort
			dp.Detected = true

			switch in.Type {
			case "mixed":
				dp.HTTP = fmt.Sprintf("http://%s:%d", listenHost, in.ListenPort)
				dp.HTTPS = fmt.Sprintf("http://%s:%d", listenHost, in.ListenPort)
				dp.SOCKS5 = fmt.Sprintf("socks5://%s:%d", listenHost, in.ListenPort)
				dp.Description = fmt.Sprintf("检测到本机正在运行的 sing-box 混合代理服务 (端口 %d)", in.ListenPort)
				return dp
			case "http":
				dp.HTTP = fmt.Sprintf("http://%s:%d", listenHost, in.ListenPort)
				dp.HTTPS = fmt.Sprintf("http://%s:%d", listenHost, in.ListenPort)
				dp.Description = fmt.Sprintf("检测到本机正在运行的 sing-box HTTP 代理 (端口 %d)", in.ListenPort)
			case "socks":
				dp.SOCKS5 = fmt.Sprintf("socks5://%s:%d", listenHost, in.ListenPort)
				if dp.Description == "" {
					dp.Description = fmt.Sprintf("检测到本机正在运行的 sing-box SOCKS5 代理 (端口 %d)", in.ListenPort)
				}
			}
		}

		if dp.Detected {
			return dp
		}
	}

	return DetectedProxy{Detected: false}
}

func detectClash() DetectedProxy {
	home, _ := os.UserHomeDir()
	candidatePaths := []string{
		filepath.Join(home, ".config", "clash", "config.yaml"),
		filepath.Join(home, ".config", "mihomo", "config.yaml"),
		"/etc/clash/config.yaml",
		"/etc/mihomo/config.yaml",
	}

	for _, p := range candidatePaths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		content := string(data)

		findPort := func(key string) int {
			for _, line := range strings.Split(content, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, key+":") {
					val := strings.TrimSpace(strings.TrimPrefix(line, key+":"))
					if p, err := strconv.Atoi(val); err == nil && p > 0 {
						return p
					}
				}
			}
			return 0
		}

		mixedPort := findPort("mixed-port")
		httpPort := findPort("port")
		socksPort := findPort("socks-port")

		if mixedPort > 0 && isPortOpen("127.0.0.1", mixedPort, 150*time.Millisecond) {
			return DetectedProxy{
				Detected:    true,
				Type:        "clash",
				HTTP:        fmt.Sprintf("http://127.0.0.1:%d", mixedPort),
				HTTPS:       fmt.Sprintf("http://127.0.0.1:%d", mixedPort),
				SOCKS5:      fmt.Sprintf("socks5://127.0.0.1:%d", mixedPort),
				Port:        mixedPort,
				Description: fmt.Sprintf("检测到本机正在运行的 Clash/Mihomo 混合代理 (端口 %d)", mixedPort),
			}
		}

		if httpPort > 0 && isPortOpen("127.0.0.1", httpPort, 150*time.Millisecond) {
			dp := DetectedProxy{
				Detected:    true,
				Type:        "clash",
				HTTP:        fmt.Sprintf("http://127.0.0.1:%d", httpPort),
				HTTPS:       fmt.Sprintf("http://127.0.0.1:%d", httpPort),
				Port:        httpPort,
				Description: fmt.Sprintf("检测到本机正在运行的 Clash/Mihomo HTTP 代理 (端口 %d)", httpPort),
			}
			if socksPort > 0 && isPortOpen("127.0.0.1", socksPort, 150*time.Millisecond) {
				dp.SOCKS5 = fmt.Sprintf("socks5://127.0.0.1:%d", socksPort)
			}
			return dp
		}
	}

	return DetectedProxy{Detected: false}
}

func detectEnvProxy() DetectedProxy {
	rawHTTP := os.Getenv("HTTP_PROXY")
	if rawHTTP == "" {
		rawHTTP = os.Getenv("http_proxy")
	}
	rawHTTPS := os.Getenv("HTTPS_PROXY")
	if rawHTTPS == "" {
		rawHTTPS = os.Getenv("https_proxy")
	}
	rawAll := os.Getenv("ALL_PROXY")
	if rawAll == "" {
		rawAll = os.Getenv("all_proxy")
	}

	if rawHTTP == "" && rawHTTPS == "" && rawAll == "" {
		return DetectedProxy{Detected: false}
	}

	dp := DetectedProxy{
		Detected:    true,
		Type:        "env",
		Description: "检测到系统环境变量中配置的网络代理",
	}

	if rawHTTP != "" {
		dp.HTTP = rawHTTP
		if u, err := url.Parse(rawHTTP); err == nil && u.Port() != "" {
			dp.Port, _ = strconv.Atoi(u.Port())
		}
	}
	if rawHTTPS != "" {
		dp.HTTPS = rawHTTPS
	} else if dp.HTTP != "" {
		dp.HTTPS = dp.HTTP
	}

	if rawAll != "" {
		if strings.HasPrefix(rawAll, "socks5://") || strings.HasPrefix(rawAll, "socks5h://") {
			dp.SOCKS5 = rawAll
		}
		if dp.HTTP == "" {
			dp.HTTP = rawAll
			dp.HTTPS = rawAll
		}
	}

	return dp
}

func probeCommonPorts() DetectedProxy {
	// Common proxy ports:
	// 8889: sing-box mixed
	// 7890: clash mixed
	// 10808: v2ray socks
	// 10809: v2ray http
	// 1080: shadowsocks socks
	type portProbe struct {
		port      int
		proxyType string
		isMixed   bool
		isSocks   bool
	}

	probes := []portProbe{
		{port: 8889, proxyType: "sing-box", isMixed: true},
		{port: 7890, proxyType: "clash", isMixed: true},
		{port: 10808, proxyType: "v2ray", isSocks: true},
		{port: 10809, proxyType: "v2ray", isMixed: false},
		{port: 1080, proxyType: "socks5", isSocks: true},
	}

	for _, p := range probes {
		if isPortOpen("127.0.0.1", p.port, 100*time.Millisecond) {
			dp := DetectedProxy{
				Detected: true,
				Type:     p.proxyType,
				Port:     p.port,
			}
			if p.isMixed {
				dp.HTTP = fmt.Sprintf("http://127.0.0.1:%d", p.port)
				dp.HTTPS = fmt.Sprintf("http://127.0.0.1:%d", p.port)
				dp.SOCKS5 = fmt.Sprintf("socks5://127.0.0.1:%d", p.port)
				dp.Description = fmt.Sprintf("检测到本机正在运行的 %s 代理服务 (端口 %d)", p.proxyType, p.port)
			} else if p.isSocks {
				dp.SOCKS5 = fmt.Sprintf("socks5://127.0.0.1:%d", p.port)
				dp.Description = fmt.Sprintf("检测到本机正在运行的 %s SOCKS5 代理 (端口 %d)", p.proxyType, p.port)
			} else {
				dp.HTTP = fmt.Sprintf("http://127.0.0.1:%d", p.port)
				dp.HTTPS = fmt.Sprintf("http://127.0.0.1:%d", p.port)
				dp.Description = fmt.Sprintf("检测到本机正在运行的 %s HTTP 代理 (端口 %d)", p.proxyType, p.port)
			}
			return dp
		}
	}

	return DetectedProxy{Detected: false}
}
