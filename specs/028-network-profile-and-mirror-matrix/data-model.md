# Data Model: Network Profiles & Mirror Matrix Architecture

## 1. TOML Configuration Schema

```toml
[network]
# 当前激活的 Profile ID，如 "cn"、"official" 或用户自定义 ID
active_profile = "cn"

# 1. 代理服务器配置 (全局代理，默认严格关闭，输入框默认完全空白)
[network.proxy]
enabled = false
http = ""
https = ""
socks5 = ""
no_proxy = "localhost,127.0.0.1,::1,*.local,*.internal"

# 2. 单项语言覆盖区 (优先级最高: Overrides > Profile > Official Default)
# 用户在界面或 CLI 单独修改某语言源时记录在此；未设置的语言自动跟随当前 active_profile
[network.overrides]
npm = ""
go = ""
python = ""
rust = ""
ruby = ""

# 3. 预设档案集合 (内置 official、cn，支持用户自定义扩展)
[network.profiles.official]
id = "official"
name = "官方直连"
description = "全量使用各官方注册表与分发源，不经任何代理加速"
builtin = true
registries = { npm = "https://registry.npmjs.org", go = "https://proxy.golang.org,direct", python = "https://pypi.org/simple", rust = "https://crates.io", ruby = "https://rubygems.org" }
downloads = { node = "https://nodejs.org/dist", go = "https://golang.org/dl", python = "https://www.python.org/ftp/python" }

[network.profiles.cn]
id = "cn"
name = "国内镜像加速"
description = "国内顶级公共镜像与大厂高校加速矩阵"
builtin = true
registries = { npm = "https://registry.npmmirror.com", go = "https://goproxy.cn,direct", python = "https://pypi.tuna.tsinghua.edu.cn/simple", rust = "https://rsproxy.cn/crates.io-index", ruby = "https://gems.ruby-china.com" }
downloads = { node = "https://npmmirror.com/mirrors/node", go = "https://golang.google.cn/dl", python = "https://registry.npmmirror.com/binary.html?b=python" }

# 用户自定义档案示例 (通过 CLI create / save-as 或 Web 界面保存生成)
[network.profiles.company-office]
id = "company-office"
name = "公司内网办公"
description = "企业自建私有包服务与专网源"
builtin = false
registries = { npm = "https://npm.corp.internal", go = "https://goproxy.cn,direct", python = "https://pypi.tuna.tsinghua.edu.cn/simple", rust = "https://rsproxy.cn/crates.io-index", ruby = "https://gems.ruby-china.com" }
downloads = { node = "https://npmmirror.com/mirrors/node", go = "https://golang.google.cn/dl", python = "https://registry.npmmirror.com/binary.html?b=python" }
```

---

## 2. Go Core Data Structures

```go
package config

// NetworkConfig represents the root network section in UniRTM config
type NetworkConfig struct {
    ActiveProfile string                    `json:"active_profile" toml:"active_profile"`
    Proxy         ProxyConfig               `json:"proxy" toml:"proxy"`
    Overrides     map[string]string         `json:"overrides" toml:"overrides"`
    Profiles      map[string]NetworkProfile `json:"profiles" toml:"profiles"`
}

// ProxyConfig defines local/corporate network proxy parameters
type ProxyConfig struct {
    Enabled bool   `json:"enabled" toml:"enabled"`
    HTTP    string `json:"http" toml:"http"`
    HTTPS   string `json:"https" toml:"https"`
    SOCKS5  string `json:"socks5" toml:"socks5"`
    NoProxy string `json:"no_proxy" toml:"no_proxy"`
}

// NetworkProfile defines an end-to-end profile contract
type NetworkProfile struct {
    ID          string            `json:"id" toml:"id"`
    Name        string            `json:"name" toml:"name"`
    Description string            `json:"description" toml:"description"`
    Builtin     bool              `json:"builtin" toml:"builtin"`
    Registries  map[string]string `json:"registries" toml:"registries"`
    Downloads   map[string]string `json:"downloads" toml:"downloads"`
}

// ActiveNetworkStatus represents the effective resolved view for CLI and Web
type ActiveNetworkStatus struct {
    ActiveProfileID string                       `json:"active_profile_id"`
    ActiveProfile   NetworkProfile               `json:"active_profile"`
    Proxy           ProxyConfig                  `json:"proxy"`
    Effective       map[string]EffectiveRegistry `json:"effective"`
}

// EffectiveRegistry denotes the resolved URL and origin for a language ecosystem
type EffectiveRegistry struct {
    Ecosystem   string `json:"ecosystem"` // npm, go, python, rust, ruby
    URL         string `json:"url"`
    Source      string `json:"source"`    // "override" | "profile" | "default"
    IsOverridden bool  `json:"is_overridden"`
}
```

---

## 3. Frontend TypeScript Interfaces

```ts
export interface NetworkStatusPayload {
  activeProfileId: string;
  activeProfile: NetworkProfile;
  proxy: ProxyConfig;
  profiles: Record<string, NetworkProfile>;
  effective: Record<string, EffectiveRegistry>;
}

export interface ProxyConfig {
  enabled: boolean;
  http: string;
  https: string;
  socks5: string;
  noProxy: string;
}

export interface NetworkProfile {
  id: string;
  name: string;
  description: string;
  builtin: boolean;
  registries: Record<string, string>;
  downloads: Record<string, string>;
}

export interface EffectiveRegistry {
  ecosystem: string;
  url: string;
  source: "override" | "profile" | "default";
  isOverridden: boolean;
}

export interface ProbeResult {
  endpoint: string;
  name: string;
  rttMs: number;
  status: "fast" | "normal" | "slow" | "timeout";
}
```
