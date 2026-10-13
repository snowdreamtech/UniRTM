// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package network

// NetworkConfig represents the root network section in UniRTM config.
type NetworkConfig struct {
	ActiveProfile string                    `json:"active_profile" toml:"active_profile" yaml:"active_profile"`
	Proxy         ProxyConfig               `json:"proxy" toml:"proxy" yaml:"proxy"`
	Overrides     map[string]string         `json:"overrides,omitempty" toml:"overrides,omitempty" yaml:"overrides,omitempty"`
	Profiles      map[string]NetworkProfile `json:"profiles,omitempty" toml:"profiles,omitempty" yaml:"profiles,omitempty"`
}

// ProxyConfig defines local/corporate network proxy parameters.
// Proxy is strictly disabled by default, and fields default to empty.
type ProxyConfig struct {
	Enabled bool   `json:"enabled" toml:"enabled" yaml:"enabled"`
	HTTP    string `json:"http" toml:"http" yaml:"http"`
	HTTPS   string `json:"https" toml:"https" yaml:"https"`
	SOCKS5  string `json:"socks5" toml:"socks5" yaml:"socks5"`
	NoProxy string `json:"no_proxy" toml:"no_proxy" yaml:"no_proxy"`
}

// NetworkProfile defines an end-to-end profile contract.
type NetworkProfile struct {
	ID          string            `json:"id" toml:"id" yaml:"id"`
	Name        string            `json:"name" toml:"name" yaml:"name"`
	Description string            `json:"description" toml:"description" yaml:"description"`
	Builtin     bool              `json:"builtin" toml:"builtin" yaml:"builtin"`
	Registries  map[string]string `json:"registries" toml:"registries" yaml:"registries"`
	Downloads   map[string]string `json:"downloads" toml:"downloads" yaml:"downloads"`
}

// ActiveNetworkStatus represents the effective resolved view for CLI and Web.
type ActiveNetworkStatus struct {
	ActiveProfileID string                       `json:"active_profile_id"`
	ActiveProfile   NetworkProfile               `json:"active_profile"`
	Proxy           ProxyConfig                  `json:"proxy"`
	Effective       map[string]EffectiveRegistry `json:"effective"`
	Profiles        map[string]NetworkProfile    `json:"profiles"`
}

// EffectiveRegistry denotes the resolved URL and origin for a language ecosystem.
type EffectiveRegistry struct {
	Ecosystem    string `json:"ecosystem"` // npm, go, python, rust, ruby
	URL          string `json:"url"`
	Source       string `json:"source"` // "override" | "profile" | "default"
	IsOverridden bool   `json:"is_overridden"`
}

// ProbeResult represents the latency result of a probe test.
type ProbeResult struct {
	Ecosystem string `json:"ecosystem"`
	Name      string `json:"name"`
	Endpoint  string `json:"endpoint"`
	RTTMs     int64  `json:"rtt_ms"`
	Status    string `json:"status"` // "fast" | "normal" | "slow" | "timeout" | "error"
	Error     string `json:"error,omitempty"`
}
