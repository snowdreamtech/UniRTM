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
			Name:        "官方直连",
			Description: "全量使用官方源与官方二进制 CDN，不经任何代理或第三方镜像",
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
			Name:        "中国镜像加速",
			Description: "中国顶级高校与大厂公共镜像加速矩阵，保障编译前依赖极速拉取",
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
			Name:            "清华大学 TUNA 专线",
			Description:     "优先全量采用清华大学开源软件镜像站 (TUNA) 教育网与公网优质骨干源",
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
			Name:            "阿里云镜像专线",
			Description:     "优先全量采用阿里巴巴开源镜像站生态高速 CDN 加速矩阵",
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
			Name:            "中国科学技术大学 USTC 专线",
			Description:     "优先全量采用中国科学技术大学 (USTC) 教育网核心骨干镜像站加速源",
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
			{Name: "中国科学技术大学 USTC", URL: "https://mirrors.ustc.edu.cn/pypi/simple"},
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
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/rubygems/"},
			{Name: "中国科学技术大学 USTC", URL: "https://mirrors.ustc.edu.cn/rubygems/"},
			{Name: "腾讯云 RubyGems", URL: "https://mirrors.cloud.tencent.com/rubygems/"},
		},
		"composer": {
			{Name: "官方直连 (Packagist)", URL: "https://repo.packagist.org"},
			{Name: "阿里云 Composer", URL: "https://mirrors.aliyun.com/composer/"},
			{Name: "腾讯云 Composer", URL: "https://mirrors.cloud.tencent.com/composer/"},
			{Name: "华为云 Composer", URL: "https://repo.huaweicloud.com/repository/php/"},
		},
		"maven": {
			{Name: "官方直连 (Maven Central)", URL: "https://repo1.maven.org/maven2"},
			{Name: "阿里云 Maven", URL: "https://maven.aliyun.com/repository/public"},
			{Name: "华为云 Maven", URL: "https://repo.huaweicloud.com/repository/maven/"},
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/maven/"},
		},
		"nuget": {
			{Name: "官方直连 (NuGet.org)", URL: "https://api.nuget.org/v3/index.json"},
			{Name: "华为云 NuGet", URL: "https://repo.huaweicloud.com/repository/nuget/v3/index.json"},
			{Name: "腾讯云 NuGet", URL: "https://mirrors.cloud.tencent.com/nuget/"},
		},
		"docker": {
			{Name: "官方直连 (Docker Hub)", URL: "https://registry-1.docker.io"},
			{Name: "阿里云 Docker 镜像", URL: "https://registry.cn-hangzhou.aliyuncs.com"},
			{Name: "网易 Docker 镜像", URL: "https://hub-mirror.c.163.com"},
			{Name: "百度 Docker 镜像", URL: "https://mirror.baidubce.com"},
		},
		"dart": {
			{Name: "官方直连 (pub.dev)", URL: "https://pub.dev"},
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/dart-pub"},
			{Name: "上海交通大学 SJTUG", URL: "https://mirrors.sjtug.sjtu.edu.cn/dart-pub"},
		},
		"cocoapods": {
			{Name: "官方直连 (CocoaPods CDN)", URL: "https://cdn.cocoapods.org"},
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/git/CocoaPods/Specs.git"},
			{Name: "中国科学技术大学 USTC", URL: "https://mirrors.ustc.edu.cn/CocoaPods-Specs.git"},
		},
		"conda": {
			{Name: "官方直连 (Anaconda)", URL: "https://repo.anaconda.com/pkgs/main"},
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/main"},
			{Name: "北京外国语大学 BFSU", URL: "https://mirrors.bfsu.edu.cn/anaconda/pkgs/main"},
			{Name: "中国科学技术大学 USTC", URL: "https://mirrors.ustc.edu.cn/anaconda/pkgs/main"},
		},
		"homebrew": {
			{Name: "官方直连 (Homebrew Bottles)", URL: "https://ghcr.io/v2/homebrew/core"},
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/homebrew-bottles"},
			{Name: "中国科学技术大学 USTC", URL: "https://mirrors.ustc.edu.cn/homebrew-bottles"},
			{Name: "阿里云 Homebrew", URL: "https://mirrors.aliyun.com/homebrew/homebrew-bottles"},
		},
		"cran": {
			{Name: "官方直连 (CRAN)", URL: "https://cran.r-project.org"},
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/CRAN/"},
			{Name: "中国科学技术大学 USTC", URL: "https://mirrors.ustc.edu.cn/CRAN/"},
			{Name: "南方科技大学 SUSTech", URL: "https://mirrors.sustech.edu.cn/CRAN/"},
			{Name: "阿里云 CRAN", URL: "https://mirrors.aliyun.com/CRAN/"},
		},
		"julia": {
			{Name: "官方直连 (Julia Pkg)", URL: "https://pkg.julialang.org"},
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/julia"},
			{Name: "中国科学技术大学 USTC", URL: "https://mirrors.ustc.edu.cn/julia"},
			{Name: "北京外国语大学 BFSU", URL: "https://mirrors.bfsu.edu.cn/julia"},
		},
		"cpan": {
			{Name: "官方直连 (CPAN)", URL: "https://www.cpan.org"},
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/CPAN/"},
			{Name: "中国科学技术大学 USTC", URL: "https://mirrors.ustc.edu.cn/CPAN/"},
			{Name: "阿里云 CPAN", URL: "https://mirrors.aliyun.com/CPAN/"},
		},
		"hackage": {
			{Name: "官方直连 (Hackage)", URL: "https://hackage.haskell.org"},
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/hackage"},
			{Name: "中国科学技术大学 USTC", URL: "https://mirrors.ustc.edu.cn/hackage"},
		},
		"clojars": {
			{Name: "官方直连 (Clojars)", URL: "https://repo.clojars.org"},
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/clojars"},
		},
		"ctan": {
			{Name: "官方直连 (CTAN)", URL: "https://mirror.ctan.org/"},
			{Name: "中国科学技术大学 USTC", URL: "https://mirrors.ustc.edu.cn/CTAN/"},
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/CTAN/"},
		},
		"elpa": {
			{Name: "官方直连 (GNU ELPA)", URL: "https://elpa.gnu.org/packages/"},
			{Name: "中国科学技术大学 USTC", URL: "https://mirrors.ustc.edu.cn/elpa/"},
			{Name: "清华大学 TUNA", URL: "https://mirrors.tuna.tsinghua.edu.cn/elpa/"},
		},
		"stackage": {
			{Name: "官方直连 (Stackage)", URL: "https://www.stackage.org"},
			{Name: "中国科学技术大学 USTC", URL: "https://mirrors.ustc.edu.cn/stackage"},
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
		{Ecosystem: "rust", Name: "rsproxy.cn", Endpoint: "https://rsproxy.cn/crates.io-index"},
		{Ecosystem: "rust", Name: "crates.io (官方)", Endpoint: "https://crates.io"},
		{Ecosystem: "ruby", Name: "清华大学 TUNA", Endpoint: "https://mirrors.tuna.tsinghua.edu.cn/rubygems/"},
		{Ecosystem: "ruby", Name: "rubygems.org (官方)", Endpoint: "https://rubygems.org"},
		{Ecosystem: "composer", Name: "阿里云 Composer", Endpoint: "https://mirrors.aliyun.com/composer/"},
		{Ecosystem: "composer", Name: "Packagist (官方)", Endpoint: "https://repo.packagist.org"},
		{Ecosystem: "maven", Name: "阿里云 Maven", Endpoint: "https://maven.aliyun.com/repository/public"},
		{Ecosystem: "maven", Name: "Maven Central (官方)", Endpoint: "https://repo1.maven.org/maven2"},
		{Ecosystem: "nuget", Name: "华为云 NuGet", Endpoint: "https://repo.huaweicloud.com/repository/nuget/v3/index.json"},
		{Ecosystem: "nuget", Name: "NuGet.org (官方)", Endpoint: "https://api.nuget.org/v3/index.json"},
		{Ecosystem: "docker", Name: "阿里云 Docker", Endpoint: "https://registry.cn-hangzhou.aliyuncs.com"},
		{Ecosystem: "docker", Name: "Docker Hub (官方)", Endpoint: "https://registry-1.docker.io"},
		{Ecosystem: "dart", Name: "清华大学 TUNA", Endpoint: "https://mirrors.tuna.tsinghua.edu.cn/dart-pub"},
		{Ecosystem: "dart", Name: "pub.dev (官方)", Endpoint: "https://pub.dev"},
		{Ecosystem: "cocoapods", Name: "CocoaPods CDN (官方)", Endpoint: "https://cdn.cocoapods.org"},
		{Ecosystem: "conda", Name: "清华大学 Conda", Endpoint: "https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/main"},
		{Ecosystem: "conda", Name: "Anaconda (官方)", Endpoint: "https://repo.anaconda.com/pkgs/main"},
		{Ecosystem: "homebrew", Name: "清华大学 Homebrew", Endpoint: "https://mirrors.tuna.tsinghua.edu.cn/homebrew-bottles"},
		{Ecosystem: "git", Name: "GitHub (官方直连)", Endpoint: "https://github.com"},
	}
}
