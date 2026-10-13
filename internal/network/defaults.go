// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package network

// KnownEcosystems enumerates supported package manager ecosystems.
var KnownEcosystems = []string{"npm", "go", "python", "rust", "ruby"}

// DefaultProxyConfig returns the standard secure proxy baseline.
// Proxy is strictly disabled by default with empty URLs.
func DefaultProxyConfig() ProxyConfig {
	return ProxyConfig{
		Enabled: false,
		HTTP:    "",
		HTTPS:   "",
		SOCKS5:  "",
		NoProxy: "localhost,127.0.0.1,::1,*.local,*.internal",
	}
}

// DefaultBuiltinProfiles returns immutable factory blueprints for built-in profiles.
func DefaultBuiltinProfiles() map[string]NetworkProfile {
	return map[string]NetworkProfile{
		"official": {
			ID:          "official",
			Name:        "官方直连",
			Description: "全量使用官方源与官方二进制 CDN，不经任何代理或国内镜像",
			Builtin:     true,
			Registries: map[string]string{
				"npm":    "https://registry.npmjs.org",
				"go":     "https://proxy.golang.org,direct",
				"python": "https://pypi.org/simple",
				"rust":   "https://crates.io",
				"ruby":   "https://rubygems.org",
			},
			Downloads: map[string]string{
				"node":   "https://nodejs.org/dist",
				"go":     "https://golang.org/dl",
				"python": "https://www.python.org/ftp/python",
			},
		},
		"cn": {
			ID:          "cn",
			Name:        "国内镜像加速",
			Description: "国内顶级高校与大厂公共镜像加速矩阵，保障编译前依赖极速拉取",
			Builtin:     true,
			Registries: map[string]string{
				"npm":    "https://registry.npmmirror.com",
				"go":     "https://goproxy.cn,direct",
				"python": "https://pypi.tuna.tsinghua.edu.cn/simple",
				"rust":   "https://rsproxy.cn/crates.io-index",
				"ruby":   "https://gems.ruby-china.com",
			},
			Downloads: map[string]string{
				"node":   "https://npmmirror.com/mirrors/node",
				"go":     "https://golang.google.cn/dl",
				"python": "https://registry.npmmirror.com/binary.html?b=python",
			},
		},
	}
}

// MirrorOption defines a recommended public mirror option for UI dropdowns.
type MirrorOption struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// RecommendedMirrors returns vetted, stable public mirrors for each ecosystem.
func RecommendedMirrors() map[string][]MirrorOption {
	return map[string][]MirrorOption{
		"npm": {
			{Name: "官方直连 (npmjs.org)", URL: "https://registry.npmjs.org"},
			{Name: "npmmirror (淘宝/阿里云)", URL: "https://registry.npmmirror.com"},
			{Name: "腾讯云镜像", URL: "https://mirrors.cloud.tencent.com/npm/"},
			{Name: "华为云镜像", URL: "https://repo.huaweicloud.com/repository/npm/"},
		},
		"go": {
			{Name: "官方直连 (Google)", URL: "https://proxy.golang.org,direct"},
			{Name: "goproxy.cn (七牛云/社区)", URL: "https://goproxy.cn,direct"},
			{Name: "阿里云 GoProxy", URL: "https://mirrors.aliyun.com/goproxy/,direct"},
			{Name: "goproxy.io", URL: "https://goproxy.io,direct"},
		},
		"python": {
			{Name: "官方直连 (PyPI)", URL: "https://pypi.org/simple"},
			{Name: "清华大学 TUNA", URL: "https://pypi.tuna.tsinghua.edu.cn/simple"},
			{Name: "阿里云 PyPI", URL: "https://mirrors.aliyun.com/pypi/simple/"},
			{Name: "腾讯云 PyPI", URL: "https://mirrors.cloud.tencent.com/pypi/simple/"},
			{Name: "华为云 PyPI", URL: "https://repo.huaweicloud.com/repository/pypi/simple/"},
		},
		"rust": {
			{Name: "官方直连 (crates.io)", URL: "https://crates.io"},
			{Name: "rsproxy.cn (字节跳动/社区)", URL: "https://rsproxy.cn/crates.io-index"},
			{Name: "中国科学技术大学 USTC", URL: "https://mirrors.ustc.edu.cn/crates.io-index"},
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/git/crates.io-index.git"},
		},
		"ruby": {
			{Name: "官方直连 (RubyGems)", URL: "https://rubygems.org"},
			{Name: "Ruby China", URL: "https://gems.ruby-china.com"},
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/rubygems/"},
		},
	}
}

// ProbeTarget defines an endpoint for latency probing.
type ProbeTarget struct {
	Ecosystem string
	Name      string
	Endpoint  string
}

// DefaultProbeTargets returns the standard matrix of test endpoints.
func DefaultProbeTargets() []ProbeTarget {
	return []ProbeTarget{
		{Ecosystem: "npm", Name: "npmmirror", Endpoint: "https://registry.npmmirror.com"},
		{Ecosystem: "npm", Name: "npmjs.org (官方)", Endpoint: "https://registry.npmjs.org"},
		{Ecosystem: "go", Name: "goproxy.cn", Endpoint: "https://goproxy.cn"},
		{Ecosystem: "go", Name: "proxy.golang.org (官方)", Endpoint: "https://proxy.golang.org"},
		{Ecosystem: "python", Name: "清华大学 PyPI", Endpoint: "https://pypi.tuna.tsinghua.edu.cn/simple"},
		{Ecosystem: "python", Name: "pypi.org (官方)", Endpoint: "https://pypi.org/simple"},
		{Ecosystem: "git", Name: "GitHub (官方直连)", Endpoint: "https://github.com"},
	}
}
