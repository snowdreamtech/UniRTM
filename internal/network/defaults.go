// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package network

// KnownEcosystems enumerates supported package manager ecosystems.
var KnownEcosystems = []string{
	"npm", "go", "python", "rust", "ruby", "composer", "maven", "nuget", "docker",
	"dart", "cocoapods", "conda", "homebrew",
}

// DefaultProxyConfig returns the standard secure proxy baseline.
// Proxy is strictly disabled by default with empty URLs.
func DefaultProxyConfig() ProxyConfig {
	return ProxyConfig{
		Enabled: false,
		HTTP:    "",
		HTTPS:   "",
		SOCKS5:  "",
		NoProxy: "localhost,127.0.0.1,::1,127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,169.254.0.0/16,*.local,*.internal,*.lan,*.cluster.local",
	}
}

// DefaultBuiltinProfiles returns immutable factory blueprints for built-in profiles.
func DefaultBuiltinProfiles() map[string]NetworkProfile {
	return map[string]NetworkProfile{
		"official": {
			ID:          "official",
			Name:        "Official Direct",
			Description: "Directly use official registries and binary CDNs without any proxy or third-party mirrors",
			Builtin:     true,
			Registries: map[string]string{
				"npm":       "https://registry.npmjs.org",
				"go":        "https://proxy.golang.org,direct",
				"python":    "https://pypi.org/simple",
				"rust":      "https://crates.io",
				"ruby":      "https://rubygems.org",
				"composer":  "https://repo.packagist.org",
				"maven":     "https://repo1.maven.org/maven2",
				"nuget":     "https://api.nuget.org/v3/index.json",
				"docker":    "https://registry-1.docker.io",
				"dart":      "https://pub.dev",
				"cocoapods": "https://cdn.cocoapods.org",
				"conda":     "https://repo.anaconda.com/pkgs/main",
				"homebrew":  "https://ghcr.io/v2/homebrew/core",
			},
			Downloads: map[string]string{
				"node":   "https://nodejs.org/dist",
				"go":     "https://golang.org/dl",
				"python": "https://www.python.org/ftp/python",
			},
		},
		"cn": {
			ID:          "cn",
			Name:        "China Mirror Acceleration",
			Description: "Public mirror matrix from top universities and major providers in China for accelerated dependency downloads",
			Builtin:     true,
			Registries: map[string]string{
				"npm":       "https://registry.npmmirror.com",
				"go":        "https://goproxy.cn,direct",
				"python":    "https://pypi.tuna.tsinghua.edu.cn/simple",
				"rust":      "https://rsproxy.cn/crates.io-index",
				"ruby":      "https://mirrors.tuna.tsinghua.edu.cn/rubygems/",
				"composer":  "https://mirrors.aliyun.com/composer/",
				"maven":     "https://maven.aliyun.com/repository/public",
				"nuget":     "https://repo.huaweicloud.com/repository/nuget/v3/index.json",
				"docker":    "https://registry.cn-hangzhou.aliyuncs.com",
				"dart":      "https://mirrors.tuna.tsinghua.edu.cn/dart-pub",
				"cocoapods": "https://mirrors.tuna.tsinghua.edu.cn/git/CocoaPods/Specs.git",
				"conda":     "https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/main",
				"homebrew":  "https://mirrors.tuna.tsinghua.edu.cn/homebrew-bottles",
			},
			Downloads: map[string]string{
				"node":   "https://npmmirror.com/mirrors/node",
				"go":     "https://golang.google.cn/dl",
				"python": "https://npmmirror.com/mirrors/python",
			},
		},
		"tuna": {
			ID:              "tuna",
			Name:            "Tsinghua TUNA Mirror Dedicated",
			Description:     "Prioritize Tsinghua University Open Source Software Mirror (TUNA) backbone mirrors",
			Builtin:         true,
			FallbackProfile: "cn",
			Registries: map[string]string{
				"python":    "https://pypi.tuna.tsinghua.edu.cn/simple",
				"rust":      "https://mirrors.tuna.tsinghua.edu.cn/git/crates.io-index.git",
				"ruby":      "https://mirrors.tuna.tsinghua.edu.cn/rubygems/",
				"maven":     "https://mirrors.tuna.tsinghua.edu.cn/maven/",
				"dart":      "https://mirrors.tuna.tsinghua.edu.cn/dart-pub",
				"cocoapods": "https://mirrors.tuna.tsinghua.edu.cn/git/CocoaPods/Specs.git",
				"conda":     "https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/main",
				"homebrew":  "https://mirrors.tuna.tsinghua.edu.cn/homebrew-bottles",
			},
			Downloads: map[string]string{
				"node":   "https://mirrors.tuna.tsinghua.edu.cn/nodejs-release/",
				"go":     "https://golang.google.cn/dl",
				"python": "https://mirrors.tuna.tsinghua.edu.cn/python/",
			},
		},
		"aliyun": {
			ID:              "aliyun",
			Name:            "Aliyun Mirror Dedicated",
			Description:     "Prioritize Alibaba Cloud open source mirror high-speed CDN acceleration matrix",
			Builtin:         true,
			FallbackProfile: "cn",
			Registries: map[string]string{
				"npm":      "https://registry.npmmirror.com",
				"go":       "https://mirrors.aliyun.com/goproxy/,direct",
				"python":   "https://mirrors.aliyun.com/pypi/simple/",
				"composer": "https://mirrors.aliyun.com/composer/",
				"maven":    "https://maven.aliyun.com/repository/public",
				"docker":   "https://registry.cn-hangzhou.aliyuncs.com",
				"conda":    "https://mirrors.aliyun.com/anaconda/pkgs/main",
				"homebrew": "https://mirrors.aliyun.com/homebrew/homebrew-bottles",
			},
			Downloads: map[string]string{
				"node":   "https://npmmirror.com/mirrors/node",
				"go":     "https://golang.google.cn/dl",
				"python": "https://mirrors.aliyun.com/python/",
			},
		},
		"ustc": {
			ID:              "ustc",
			Name:            "USTC Mirror Dedicated",
			Description:     "Prioritize University of Science and Technology of China (USTC) core backbone mirrors",
			Builtin:         true,
			FallbackProfile: "cn",
			Registries: map[string]string{
				"python":    "https://mirrors.ustc.edu.cn/pypi/simple",
				"rust":      "https://mirrors.ustc.edu.cn/crates.io-index",
				"ruby":      "https://mirrors.ustc.edu.cn/rubygems/",
				"cocoapods": "https://mirrors.ustc.edu.cn/CocoaPods-Specs.git",
				"conda":     "https://mirrors.ustc.edu.cn/anaconda/pkgs/main",
				"homebrew":  "https://mirrors.ustc.edu.cn/homebrew-bottles",
			},
			Downloads: map[string]string{
				"node":   "https://mirrors.ustc.edu.cn/node/",
				"go":     "https://golang.google.cn/dl",
				"python": "https://mirrors.ustc.edu.cn/python/",
				"rust":   "https://mirrors.ustc.edu.cn/rust-static",
				"java":   "https://mirrors.ustc.edu.cn/adoptium/",
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
			{Name: "Official Direct (npmjs.org)", URL: "https://registry.npmjs.org"},
			{Name: "npmmirror (Taobao / Aliyun)", URL: "https://registry.npmmirror.com"},
			{Name: "Tencent Cloud Mirror", URL: "https://mirrors.cloud.tencent.com/npm/"},
			{Name: "Huawei Cloud Mirror", URL: "https://repo.huaweicloud.com/repository/npm/"},
		},
		"go": {
			{Name: "Official Direct (Google)", URL: "https://proxy.golang.org,direct"},
			{Name: "goproxy.cn (Qiniu / Community)", URL: "https://goproxy.cn,direct"},
			{Name: "Aliyun GoProxy", URL: "https://mirrors.aliyun.com/goproxy/,direct"},
			{Name: "goproxy.io", URL: "https://goproxy.io,direct"},
		},
		"python": {
			{Name: "Official Direct (PyPI)", URL: "https://pypi.org/simple"},
			{Name: "Tsinghua TUNA", URL: "https://pypi.tuna.tsinghua.edu.cn/simple"},
			{Name: "USTC Mirror", URL: "https://mirrors.ustc.edu.cn/pypi/simple"},
			{Name: "Aliyun PyPI", URL: "https://mirrors.aliyun.com/pypi/simple/"},
			{Name: "Tencent Cloud PyPI", URL: "https://mirrors.cloud.tencent.com/pypi/simple/"},
			{Name: "Huawei Cloud PyPI", URL: "https://repo.huaweicloud.com/repository/pypi/simple/"},
		},
		"rust": {
			{Name: "Official Direct (crates.io)", URL: "https://crates.io"},
			{Name: "rsproxy.cn (ByteDance / Community)", URL: "https://rsproxy.cn/crates.io-index"},
			{Name: "USTC Mirror", URL: "https://mirrors.ustc.edu.cn/crates.io-index"},
			{Name: "Tsinghua TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/git/crates.io-index.git"},
		},
		"ruby": {
			{Name: "Official Direct (RubyGems)", URL: "https://rubygems.org"},
			{Name: "Tsinghua TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/rubygems/"},
			{Name: "USTC Mirror", URL: "https://mirrors.ustc.edu.cn/rubygems/"},
			{Name: "Tencent Cloud RubyGems", URL: "https://mirrors.cloud.tencent.com/rubygems/"},
		},
		"composer": {
			{Name: "Official Direct (Packagist)", URL: "https://repo.packagist.org"},
			{Name: "Aliyun Composer", URL: "https://mirrors.aliyun.com/composer/"},
			{Name: "Tencent Cloud Composer", URL: "https://mirrors.cloud.tencent.com/composer/"},
			{Name: "Huawei Cloud Composer", URL: "https://repo.huaweicloud.com/repository/php/"},
		},
		"maven": {
			{Name: "Official Direct (Maven Central)", URL: "https://repo1.maven.org/maven2"},
			{Name: "Aliyun Maven", URL: "https://maven.aliyun.com/repository/public"},
			{Name: "Huawei Cloud Maven", URL: "https://repo.huaweicloud.com/repository/maven/"},
			{Name: "Tsinghua TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/maven/"},
		},
		"nuget": {
			{Name: "Official Direct (NuGet.org)", URL: "https://api.nuget.org/v3/index.json"},
			{Name: "Huawei Cloud NuGet", URL: "https://repo.huaweicloud.com/repository/nuget/v3/index.json"},
			{Name: "Tencent Cloud NuGet", URL: "https://mirrors.cloud.tencent.com/nuget/"},
		},
		"docker": {
			{Name: "Official Direct (Docker Hub)", URL: "https://registry-1.docker.io"},
			{Name: "Aliyun Docker Mirror", URL: "https://registry.cn-hangzhou.aliyuncs.com"},
			{Name: "NetEase Docker Mirror", URL: "https://hub-mirror.c.163.com"},
			{Name: "Baidu Docker Mirror", URL: "https://mirror.baidubce.com"},
		},
		"dart": {
			{Name: "Official Direct (pub.dev)", URL: "https://pub.dev"},
			{Name: "Tsinghua TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/dart-pub"},
			{Name: "SJTU SJTUG Mirror", URL: "https://mirrors.sjtug.sjtu.edu.cn/dart-pub"},
		},
		"cocoapods": {
			{Name: "Official Direct (CocoaPods CDN)", URL: "https://cdn.cocoapods.org"},
			{Name: "Tsinghua TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/git/CocoaPods/Specs.git"},
			{Name: "USTC Mirror", URL: "https://mirrors.ustc.edu.cn/CocoaPods-Specs.git"},
		},
		"conda": {
			{Name: "Official Direct (Anaconda)", URL: "https://repo.anaconda.com/pkgs/main"},
			{Name: "Tsinghua TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/main"},
			{Name: "BFSU Mirror", URL: "https://mirrors.bfsu.edu.cn/anaconda/pkgs/main"},
			{Name: "USTC Mirror", URL: "https://mirrors.ustc.edu.cn/anaconda/pkgs/main"},
		},
		"homebrew": {
			{Name: "Official Direct (Homebrew Bottles)", URL: "https://ghcr.io/v2/homebrew/core"},
			{Name: "Tsinghua TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/homebrew-bottles"},
			{Name: "USTC Mirror", URL: "https://mirrors.ustc.edu.cn/homebrew-bottles"},
			{Name: "Aliyun Homebrew", URL: "https://mirrors.aliyun.com/homebrew/homebrew-bottles"},
		},
		"cran": {
			{Name: "Official Direct (CRAN)", URL: "https://cran.r-project.org"},
			{Name: "Tsinghua TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/CRAN/"},
			{Name: "USTC Mirror", URL: "https://mirrors.ustc.edu.cn/CRAN/"},
			{Name: "SUSTech Mirror", URL: "https://mirrors.sustech.edu.cn/CRAN/"},
			{Name: "Aliyun CRAN", URL: "https://mirrors.aliyun.com/CRAN/"},
		},
		"julia": {
			{Name: "Official Direct (Julia Pkg)", URL: "https://pkg.julialang.org"},
			{Name: "Tsinghua TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/julia"},
			{Name: "USTC Mirror", URL: "https://mirrors.ustc.edu.cn/julia"},
			{Name: "BFSU Mirror", URL: "https://mirrors.bfsu.edu.cn/julia"},
		},
		"cpan": {
			{Name: "Official Direct (CPAN)", URL: "https://www.cpan.org"},
			{Name: "Tsinghua TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/CPAN/"},
			{Name: "USTC Mirror", URL: "https://mirrors.ustc.edu.cn/CPAN/"},
			{Name: "Aliyun CPAN", URL: "https://mirrors.aliyun.com/CPAN/"},
		},
		"hackage": {
			{Name: "Official Direct (Hackage)", URL: "https://hackage.haskell.org"},
			{Name: "Tsinghua TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/hackage"},
			{Name: "USTC Mirror", URL: "https://mirrors.ustc.edu.cn/hackage"},
		},
		"clojars": {
			{Name: "Official Direct (Clojars)", URL: "https://repo.clojars.org"},
			{Name: "Tsinghua TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/clojars"},
		},
		"ctan": {
			{Name: "Official Direct (CTAN)", URL: "https://mirror.ctan.org/"},
			{Name: "USTC Mirror", URL: "https://mirrors.ustc.edu.cn/CTAN/"},
			{Name: "Tsinghua TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/CTAN/"},
		},
		"elpa": {
			{Name: "Official Direct (GNU ELPA)", URL: "https://elpa.gnu.org/packages/"},
			{Name: "USTC Mirror", URL: "https://mirrors.ustc.edu.cn/elpa/"},
			{Name: "Tsinghua TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/elpa/"},
		},
		"stackage": {
			{Name: "Official Direct (Stackage)", URL: "https://www.stackage.org"},
			{Name: "USTC Mirror", URL: "https://mirrors.ustc.edu.cn/stackage"},
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
		{Ecosystem: "npm", Name: "npmjs.org (Official)", Endpoint: "https://registry.npmjs.org"},
		{Ecosystem: "go", Name: "goproxy.cn", Endpoint: "https://goproxy.cn"},
		{Ecosystem: "go", Name: "proxy.golang.org (Official)", Endpoint: "https://proxy.golang.org"},
		{Ecosystem: "python", Name: "Tsinghua PyPI", Endpoint: "https://pypi.tuna.tsinghua.edu.cn/simple"},
		{Ecosystem: "python", Name: "pypi.org (Official)", Endpoint: "https://pypi.org/simple"},
		{Ecosystem: "rust", Name: "rsproxy.cn", Endpoint: "https://rsproxy.cn/crates.io-index"},
		{Ecosystem: "rust", Name: "crates.io (Official)", Endpoint: "https://crates.io"},
		{Ecosystem: "ruby", Name: "Tsinghua TUNA", Endpoint: "https://mirrors.tuna.tsinghua.edu.cn/rubygems/"},
		{Ecosystem: "ruby", Name: "rubygems.org (Official)", Endpoint: "https://rubygems.org"},
		{Ecosystem: "composer", Name: "Aliyun Composer", Endpoint: "https://mirrors.aliyun.com/composer/"},
		{Ecosystem: "composer", Name: "Packagist (Official)", Endpoint: "https://repo.packagist.org"},
		{Ecosystem: "maven", Name: "Aliyun Maven", Endpoint: "https://maven.aliyun.com/repository/public"},
		{Ecosystem: "maven", Name: "Maven Central (Official)", Endpoint: "https://repo1.maven.org/maven2"},
		{Ecosystem: "nuget", Name: "Huawei Cloud NuGet", Endpoint: "https://repo.huaweicloud.com/repository/nuget/v3/index.json"},
		{Ecosystem: "nuget", Name: "NuGet.org (Official)", Endpoint: "https://api.nuget.org/v3/index.json"},
		{Ecosystem: "docker", Name: "Aliyun Docker", Endpoint: "https://registry.cn-hangzhou.aliyuncs.com"},
		{Ecosystem: "docker", Name: "Docker Hub (Official)", Endpoint: "https://registry-1.docker.io"},
		{Ecosystem: "dart", Name: "Tsinghua TUNA", Endpoint: "https://mirrors.tuna.tsinghua.edu.cn/dart-pub"},
		{Ecosystem: "dart", Name: "pub.dev (Official)", Endpoint: "https://pub.dev"},
		{Ecosystem: "cocoapods", Name: "CocoaPods CDN (Official)", Endpoint: "https://cdn.cocoapods.org"},
		{Ecosystem: "conda", Name: "Tsinghua Conda", Endpoint: "https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/main"},
		{Ecosystem: "conda", Name: "Anaconda (Official)", Endpoint: "https://repo.anaconda.com/pkgs/main"},
		{Ecosystem: "homebrew", Name: "Tsinghua Homebrew", Endpoint: "https://mirrors.tuna.tsinghua.edu.cn/homebrew-bottles"},
		{Ecosystem: "git", Name: "GitHub (Official Direct)", Endpoint: "https://github.com"},
	}
}
