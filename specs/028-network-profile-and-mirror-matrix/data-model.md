# Data Model: Network Profiles & Mirror Matrix Architecture

**Feature**: Network Profiles, Domestic Mirror Matrix & Cascading Overrides  
**Branch**: `dev`  
**Spec**: [spec.md](./spec.md)  
**Plan**: [plan.md](./plan.md)  
**Tasks**: [tasks.md](./tasks.md)  

---

## 1. TOML Configuration Schema

The network subsystem is stored under the root `[network]` section in `.unirtm.toml` (project workspace scope) or `~/.config/unirtm/unirtm.toml` (global user scope).

```toml
[network]
# Active Profile identifier, e.g., "cn", "official", or a custom user profile ID
active_profile = "cn"

# 1. Proxy Configuration
# Master proxy controls; strictly disabled by default with empty endpoints
[network.proxy]
enabled = false
http = ""
https = ""
socks5 = ""
no_proxy = "localhost,127.0.0.1,::1,*.local,*.internal"

# 2. Per-Ecosystem Overrides (Highest Precedence: Overrides > Profile > Factory Default)
# Records granular user overrides; unset ecosystems dynamically follow active_profile
[network.overrides]
npm = ""
go = ""
python = ""
rust = ""
ruby = ""

# 3. Built-in Factory Profiles
[network.profiles.official]
id = "official"
name = "Official Direct"
description = "Uses upstream official registries and binary CDNs without proxies or third-party mirrors"
builtin = true
registries = { npm = "https://registry.npmjs.org", go = "https://proxy.golang.org,direct", python = "https://pypi.org/simple", rust = "https://crates.io", ruby = "https://rubygems.org" }
downloads = { node = "https://nodejs.org/dist", go = "https://golang.org/dl", python = "https://www.python.org/ftp/python" }

[network.profiles.cn]
id = "cn"
name = "China Fast Mirrors"
description = "Vetted domestic university and cloud provider public mirror matrix for fast pre-flight downloads"
builtin = true
registries = { npm = "https://registry.npmmirror.com", go = "https://goproxy.cn,direct", python = "https://pypi.tuna.tsinghua.edu.cn/simple", rust = "https://rsproxy.cn/crates.io-index", ruby = "https://gems.ruby-china.com" }
downloads = { node = "https://npmmirror.com/mirrors/node", go = "https://golang.google.cn/dl", python = "https://registry.npmmirror.com/binary.html?b=python" }

# 4. User-Defined Custom Profiles (Created via CLI or Web UI)
[network.profiles.company-office]
id = "company-office"
name = "Corporate Office"
description = "Internal enterprise registries and private repository endpoints"
builtin = false
registries = { npm = "https://npm.corp.internal", go = "https://goproxy.cn,direct", python = "https://pypi.tuna.tsinghua.edu.cn/simple", rust = "https://rsproxy.cn/crates.io-index", ruby = "https://gems.ruby-china.com" }
downloads = { node = "https://npmmirror.com/mirrors/node", go = "https://golang.google.cn/dl", python = "https://registry.npmmirror.com/binary.html?b=python" }
```

---

## 2. Go Core Data Structures

Defined in `internal/network/types.go`:

```go
package network

// NetworkConfig represents the root network section in UniRTM configuration.
type NetworkConfig struct {
    ActiveProfile string                    `json:"active_profile" toml:"active_profile"`
    Proxy         ProxyConfig               `json:"proxy" toml:"proxy"`
    Overrides     map[string]string         `json:"overrides" toml:"overrides"`
    Profiles      map[string]NetworkProfile `json:"profiles" toml:"profiles"`
}

// ProxyConfig defines network proxy parameters.
type ProxyConfig struct {
    Enabled bool   `json:"enabled" toml:"enabled"`
    HTTP    string `json:"http" toml:"http"`
    HTTPS   string `json:"https" toml:"https"`
    SOCKS5  string `json:"socks5" toml:"socks5"`
    NoProxy string `json:"no_proxy" toml:"no_proxy"`
}

// NetworkProfile defines an individual mirror profile.
type NetworkProfile struct {
    ID              string            `json:"id" toml:"id"`
    Name            string            `json:"name" toml:"name"`
    Description     string            `json:"description" toml:"description"`
    Builtin         bool              `json:"builtin" toml:"builtin"`
    FallbackProfile string            `json:"fallback_profile,omitempty" toml:"fallback_profile,omitempty"`
    Registries      map[string]string `json:"registries" toml:"registries"`
    Downloads       map[string]string `json:"downloads" toml:"downloads"`
}

// ActiveNetworkStatus represents the resolved state returned to CLI and Web UI.
type ActiveNetworkStatus struct {
    ActiveProfileID string                       `json:"active_profile_id"`
    ActiveProfile   NetworkProfile               `json:"active_profile"`
    Proxy           ProxyConfig                  `json:"proxy"`
    Effective       map[string]EffectiveRegistry `json:"effective"`
    Profiles        map[string]NetworkProfile    `json:"profiles"`
}

// EffectiveRegistry denotes the resolved registry URL and origin for an ecosystem.
type EffectiveRegistry struct {
    Ecosystem    string `json:"ecosystem"` // npm, go, python, rust, ruby, etc.
    URL          string `json:"url"`
    Source       string `json:"source"`    // "override" | "profile" | "fallback_cn" | "default"
    IsOverridden bool   `json:"is_overridden"`
}

// ProbeResult stores latency metrics for endpoint health monitoring.
type ProbeResult struct {
    Ecosystem string `json:"ecosystem"`
    Name      string `json:"name"`
    Endpoint  string `json:"endpoint"`
    RTTMs     int64  `json:"latency_ms"`
    Status    string `json:"status"` // "fast" | "normal" | "slow" | "timeout"
    OK        bool   `json:"ok"`
}
```

---

## 3. Frontend TypeScript Interfaces

Defined in `frontend/src/i18n/types.ts` and view component types:

```ts
export interface ActiveNetworkStatus {
  active_profile_id: string;
  active_profile: NetworkProfile;
  proxy: ProxyConfig;
  effective: Record<string, EffectiveRegistry>;
  profiles: Record<string, NetworkProfile>;
}

export interface ProxyConfig {
  enabled: boolean;
  http: string;
  https: string;
  socks5: string;
  no_proxy: string;
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
  is_overridden: boolean;
}

export interface ProbeResult {
  ecosystem: string;
  name: string;
  endpoint: string;
  latency_ms: number;
  status: "fast" | "normal" | "slow" | "timeout";
  ok: boolean;
}
```
