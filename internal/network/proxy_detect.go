// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package network

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// DetectedProxy represents proxy details discovered automatically from the host environment.
type DetectedProxy struct {
	Detected    bool   `json:"detected"`
	Type        string `json:"type"` // "sing-box", "clash", "v2ray", "v2raya", "surge", "shadowsocks", "privoxy", "os", "env", "port"
	HTTP        string `json:"http"`
	HTTPS       string `json:"https"`
	SOCKS5      string `json:"socks5"`
	Port        int    `json:"port"`
	Description string `json:"description"`
}

// DetectSystemProxy inspects sing-box, Clash, V2Ray/Xray, OS settings, env vars, and active ports.
func DetectSystemProxy() DetectedProxy {
	// 1. Check sing-box configuration files and active ports
	if dp := detectSingBox(); dp.Detected {
		return dp
	}

	// 2. Check Clash / Mihomo configurations
	if dp := detectClash(); dp.Detected {
		return dp
	}

	// 3. Check V2Ray / Xray configurations
	if dp := detectV2Ray(); dp.Detected {
		return dp
	}

	// 4. Check OS / Desktop environment proxy settings
	if dp := detectOSProxy(); dp.Detected {
		return dp
	}

	// 5. Check environment variables
	if dp := detectEnvProxy(); dp.Detected {
		return dp
	}

	// 6. Active probing of common proxy ports
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
		filepath.Join(home, ".config", "clash-verge", "config.yaml"),
		filepath.Join(home, ".config", "clash-nyanpasu", "config.yaml"),
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

func detectV2Ray() DetectedProxy {
	home, _ := os.UserHomeDir()
	candidatePaths := []string{
		"/etc/v2ray/config.json",
		"/usr/local/etc/v2ray/config.json",
		"/etc/xray/config.json",
		"/usr/local/etc/xray/config.json",
		filepath.Join(home, ".config", "v2ray", "config.json"),
		filepath.Join(home, ".config", "xray", "config.json"),
	}

	type V2Inbound struct {
		Port     int    `json:"port"`
		Protocol string `json:"protocol"`
		Listen   string `json:"listen"`
	}
	type V2Config struct {
		Inbounds []V2Inbound `json:"inbounds"`
	}

	for _, p := range candidatePaths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}

		var cfg V2Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}

		var dp DetectedProxy
		dp.Type = "v2ray"

		for _, in := range cfg.Inbounds {
			if in.Port <= 0 {
				continue
			}
			listenHost := in.Listen
			if listenHost == "" || listenHost == "0.0.0.0" {
				listenHost = "127.0.0.1"
			}
			if !isPortOpen(listenHost, in.Port, 150*time.Millisecond) {
				continue
			}

			dp.Detected = true
			dp.Port = in.Port
			if in.Protocol == "http" {
				dp.HTTP = fmt.Sprintf("http://%s:%d", listenHost, in.Port)
				dp.HTTPS = fmt.Sprintf("http://%s:%d", listenHost, in.Port)
			} else if in.Protocol == "socks" {
				dp.SOCKS5 = fmt.Sprintf("socks5://%s:%d", listenHost, in.Port)
			}
		}

		if dp.Detected {
			dp.Description = fmt.Sprintf("检测到本机正在运行的 V2Ray/Xray 代理服务 (端口 %d)", dp.Port)
			return dp
		}
	}

	return DetectedProxy{Detected: false}
}

func detectOSProxy() DetectedProxy {
	switch runtime.GOOS {
	case "linux":
		// Query GNOME desktop proxy settings
		if _, err := exec.LookPath("gsettings"); err == nil {
			out, err := exec.Command("gsettings", "get", "org.gnome.system.proxy", "mode").Output()
			if err == nil && strings.Contains(string(out), "manual") {
				hOut, _ := exec.Command("gsettings", "get", "org.gnome.system.proxy.http", "host").Output()
				pOut, _ := exec.Command("gsettings", "get", "org.gnome.system.proxy.http", "port").Output()
				host := strings.Trim(strings.TrimSpace(string(hOut)), "'\"")
				portStr := strings.TrimSpace(string(pOut))
				port, _ := strconv.Atoi(portStr)
				if host != "" && port > 0 {
					dp := DetectedProxy{
						Detected:    true,
						Type:        "os",
						HTTP:        fmt.Sprintf("http://%s:%d", host, port),
						HTTPS:       fmt.Sprintf("http://%s:%d", host, port),
						Port:        port,
						Description: fmt.Sprintf("检测到系统桌面环境设置的 HTTP 代理 (%s:%d)", host, port),
					}
					sOut, _ := exec.Command("gsettings", "get", "org.gnome.system.proxy.socks", "host").Output()
					spOut, _ := exec.Command("gsettings", "get", "org.gnome.system.proxy.socks", "port").Output()
					sHost := strings.Trim(strings.TrimSpace(string(sOut)), "'\"")
					sPort, _ := strconv.Atoi(strings.TrimSpace(string(spOut)))
					if sHost != "" && sPort > 0 {
						dp.SOCKS5 = fmt.Sprintf("socks5://%s:%d", sHost, sPort)
					}
					return dp
				}
			}
		}
	case "darwin":
		// Query macOS scutil --proxy
		if out, err := exec.Command("scutil", "--proxy").Output(); err == nil {
			scanner := bufio.NewScanner(bytes.NewReader(out))
			var httpEnable, socksEnable bool
			var httpHost, socksHost string
			var httpPort, socksPort int
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				parts := strings.SplitN(line, ":", 2)
				if len(parts) != 2 {
					continue
				}
				k, v := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
				switch k {
				case "HTTPEnable":
					httpEnable = (v == "1")
				case "HTTPProxy":
					httpHost = v
				case "HTTPPort":
					httpPort, _ = strconv.Atoi(v)
				case "SOCKSEnable":
					socksEnable = (v == "1")
				case "SOCKSProxy":
					socksHost = v
				case "SOCKSPort":
					socksPort, _ = strconv.Atoi(v)
				}
			}
			if (httpEnable && httpHost != "" && httpPort > 0) || (socksEnable && socksHost != "" && socksPort > 0) {
				dp := DetectedProxy{
					Detected:    true,
					Type:        "os",
					Description: "检测到 macOS 系统偏好设置中配置的网络代理",
				}
				if httpEnable && httpHost != "" && httpPort > 0 {
					dp.HTTP = fmt.Sprintf("http://%s:%d", httpHost, httpPort)
					dp.HTTPS = fmt.Sprintf("http://%s:%d", httpHost, httpPort)
					dp.Port = httpPort
				}
				if socksEnable && socksHost != "" && socksPort > 0 {
					dp.SOCKS5 = fmt.Sprintf("socks5://%s:%d", socksHost, socksPort)
					if dp.Port == 0 {
						dp.Port = socksPort
					}
				}
				return dp
			}
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
	// Probe common client port combinations:
	// 1. sing-box mixed: 8889
	// 2. Clash/Mihomo: 7890 (HTTP/mixed), 7891 (SOCKS5)
	// 3. V2Ray/Xray: 10809 (HTTP), 10808 (SOCKS5)
	// 4. v2rayA: 20172 / 20170 (HTTP), 20171 (SOCKS5)
	// 5. Surge: 6152 (HTTP), 6153 (SOCKS5)
	// 6. Shadowsocks: 1080 / 1086 (SOCKS5)
	// 7. Privoxy: 8118 (HTTP)
	// 8. Tinyproxy: 8888 (HTTP)
	// 9. Squid: 3128 (HTTP)

	// Check Clash
	if isPortOpen("127.0.0.1", 7890, 80*time.Millisecond) {
		dp := DetectedProxy{
			Detected:    true,
			Type:        "clash",
			HTTP:        "http://127.0.0.1:7890",
			HTTPS:       "http://127.0.0.1:7890",
			Port:        7890,
			Description: "检测到本机正在运行的 Clash/Mihomo 代理 (端口 7890)",
		}
		if isPortOpen("127.0.0.1", 7891, 50*time.Millisecond) {
			dp.SOCKS5 = "socks5://127.0.0.1:7891"
		} else {
			dp.SOCKS5 = "socks5://127.0.0.1:7890" // Many clash versions use 7890 as mixed-port
		}
		return dp
	}

	// Check V2Ray / Xray
	socks10808 := isPortOpen("127.0.0.1", 10808, 80*time.Millisecond)
	http10809 := isPortOpen("127.0.0.1", 10809, 80*time.Millisecond)
	if socks10808 || http10809 {
		dp := DetectedProxy{
			Detected:    true,
			Type:        "v2ray",
			Port:        10808,
			Description: "检测到本机正在运行的 V2Ray/Xray 代理服务",
		}
		if socks10808 {
			dp.SOCKS5 = "socks5://127.0.0.1:10808"
		}
		if http10809 {
			dp.HTTP = "http://127.0.0.1:10809"
			dp.HTTPS = "http://127.0.0.1:10809"
			dp.Port = 10809
		}
		return dp
	}

	// Check v2rayA
	socks20171 := isPortOpen("127.0.0.1", 20171, 80*time.Millisecond)
	http20172 := isPortOpen("127.0.0.1", 20172, 80*time.Millisecond)
	if socks20171 || http20172 {
		dp := DetectedProxy{
			Detected:    true,
			Type:        "v2raya",
			Port:        20171,
			Description: "检测到本机正在运行的 v2rayA 代理服务",
		}
		if socks20171 {
			dp.SOCKS5 = "socks5://127.0.0.1:20171"
		}
		if http20172 {
			dp.HTTP = "http://127.0.0.1:20172"
			dp.HTTPS = "http://127.0.0.1:20172"
			dp.Port = 20172
		}
		return dp
	}

	// Check Surge
	http6152 := isPortOpen("127.0.0.1", 6152, 80*time.Millisecond)
	socks6153 := isPortOpen("127.0.0.1", 6153, 80*time.Millisecond)
	if http6152 || socks6153 {
		dp := DetectedProxy{
			Detected:    true,
			Type:        "surge",
			Port:        6152,
			Description: "检测到本机正在运行的 Surge 代理服务",
		}
		if http6152 {
			dp.HTTP = "http://127.0.0.1:6152"
			dp.HTTPS = "http://127.0.0.1:6152"
		}
		if socks6153 {
			dp.SOCKS5 = "socks5://127.0.0.1:6153"
			if !http6152 {
				dp.Port = 6153
			}
		}
		return dp
	}

	// Check Shadowsocks / generic SOCKS5
	for _, port := range []int{1080, 1086} {
		if isPortOpen("127.0.0.1", port, 80*time.Millisecond) {
			return DetectedProxy{
				Detected:    true,
				Type:        "shadowsocks",
				SOCKS5:      fmt.Sprintf("socks5://127.0.0.1:%d", port),
				Port:        port,
				Description: fmt.Sprintf("检测到本机正在运行的 SOCKS5 代理服务 (端口 %d)", port),
			}
		}
	}

	// Check Privoxy (8118)
	if isPortOpen("127.0.0.1", 8118, 80*time.Millisecond) {
		return DetectedProxy{
			Detected:    true,
			Type:        "privoxy",
			HTTP:        "http://127.0.0.1:8118",
			HTTPS:       "http://127.0.0.1:8118",
			Port:        8118,
			Description: "检测到本机正在运行的 Privoxy HTTP 代理服务 (端口 8118)",
		}
	}

	// Check Tinyproxy / Fiddler / Charles (8888)
	if isPortOpen("127.0.0.1", 8888, 80*time.Millisecond) {
		return DetectedProxy{
			Detected:    true,
			Type:        "http-proxy",
			HTTP:        "http://127.0.0.1:8888",
			HTTPS:       "http://127.0.0.1:8888",
			Port:        8888,
			Description: "检测到本机正在运行的 HTTP 代理服务 (端口 8888)",
		}
	}

	// Check Squid (3128)
	if isPortOpen("127.0.0.1", 3128, 80*time.Millisecond) {
		return DetectedProxy{
			Detected:    true,
			Type:        "squid",
			HTTP:        "http://127.0.0.1:3128",
			HTTPS:       "http://127.0.0.1:3128",
			Port:        3128,
			Description: "检测到本机正在运行的 Squid 代理服务 (端口 3128)",
		}
	}

	return DetectedProxy{Detected: false}
}
