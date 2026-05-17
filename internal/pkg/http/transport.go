// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"crypto/tls"
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

// ShouldBypassProxy returns true if the given host should bypass the proxy.
func ShouldBypassProxy(host string) bool {
	if strings.Contains(host, "mirror") || strings.HasSuffix(host, ".cn") {
		return true
	}
	for _, domain := range ProxyBypassDomains {
		if strings.Contains(host, domain) {
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
