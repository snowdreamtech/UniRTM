// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"
)

// ProxyBypassDomains contains a list of common domestic mirror domains
// that should bypass any configured HTTP proxies to prevent connection drops.
var ProxyBypassDomains = []string{
	"aliyun.com",
	"npmmirror.com",
	"tencent.com",
	"huaweicloud.com",
	"163.com",
	"ustc.edu.cn",
	"tsinghua.edu.cn",
	"sjtu.edu.cn",
	"bfsu.edu.cn",
	"lzu.edu.cn",
	"nju.edu.cn",
	"cqu.edu.cn",
	"hit.edu.cn",
	"zju.edu.cn",
	"douban.com",
	"rsproxy.cn",
	"r.cnpmjs.org",
	"goproxy.cn",
	"goproxy.io",
	"gems.ruby-china.com",
	"sn0wdr1am.com",
}

// isCompoundTLD checks if the last two domain labels form a recognized compound ccTLD.
func isCompoundTLD(secondLevel, topLevel string) bool {
	switch topLevel {
	case "cn":
		return secondLevel == "edu" || secondLevel == "com" || secondLevel == "org" || secondLevel == "net" || secondLevel == "gov"
	case "uk":
		return secondLevel == "co" || secondLevel == "org" || secondLevel == "ac" || secondLevel == "gov" || secondLevel == "me"
	case "au":
		return secondLevel == "com" || secondLevel == "net" || secondLevel == "org" || secondLevel == "edu" || secondLevel == "gov"
	case "jp":
		return secondLevel == "co" || secondLevel == "ac" || secondLevel == "ne" || secondLevel == "go" || secondLevel == "or"
	case "tw":
		return secondLevel == "com" || secondLevel == "org" || secondLevel == "edu" || secondLevel == "gov" || secondLevel == "idv"
	case "hk":
		return secondLevel == "com" || secondLevel == "org" || secondLevel == "edu" || secondLevel == "gov"
	case "nz":
		return secondLevel == "co" || secondLevel == "org" || secondLevel == "net" || secondLevel == "govt" || secondLevel == "ac"
	default:
		return false
	}
}

// ExtractApexDomain returns the registrable apex domain (e.g. "github.com",
// "githubusercontent.com", "tsinghua.edu.cn").
//
// If the host is an IP address or an unqualified single-label hostname (e.g. "localhost"),
// it returns the original host unchanged.
func ExtractApexDomain(host string) string {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "" {
		return ""
	}
	if strings.Contains(h, ":") {
		if splitHost, _, err := net.SplitHostPort(h); err == nil {
			h = splitHost
		}
	}
	if net.ParseIP(h) != nil {
		return h
	}
	parts := strings.Split(h, ".")
	if len(parts) <= 1 {
		return h
	}

	// 1. Check for compound ccTLDs (e.g. tsinghua.edu.cn, bbc.co.uk)
	if len(parts) >= 3 {
		secondLevel := parts[len(parts)-2]
		topLevel := parts[len(parts)-1]
		if isCompoundTLD(secondLevel, topLevel) {
			return strings.Join(parts[len(parts)-3:], ".")
		}
	}

	// 2. Standard registrable domain (SLD + TLD), e.g. githubusercontent.com, github.com, aliyun.com
	return strings.Join(parts[len(parts)-2:], ".")
}

// ShouldBypassProxy returns true if the given host should bypass the proxy.
func ShouldBypassProxy(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "localhost" || h == "127.0.0.1" || h == "::1" || strings.HasPrefix(h, "127.") {
		return true
	}
	if strings.HasSuffix(h, ".cn") {
		return true
	}
	apex := ExtractApexDomain(h)
	for _, domain := range ProxyBypassDomains {
		if apex == domain || strings.HasSuffix(h, "."+domain) || h == domain {
			return true
		}
	}
	return false
}

// DisableHTTP2 safely disables HTTP/2 on the given transport.
// This prevents ALPN framing errors when proxies intercept traffic.
func DisableHTTP2(trans *http.Transport) {
	if trans == nil {
		return
	}
	trans.ForceAttemptHTTP2 = false
	trans.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)

	// Strip "h2" from ALPN negotiation to prevent the CDN/proxy from sending HTTP/2 frames
	if trans.TLSClientConfig != nil {
		trans.TLSClientConfig = trans.TLSClientConfig.Clone()
		trans.TLSClientConfig.NextProtos = []string{"http/1.1"}
	} else {
		trans.TLSClientConfig = &tls.Config{
			NextProtos: []string{"http/1.1"},
		}
	}
}
