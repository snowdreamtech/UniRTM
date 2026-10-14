// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package network

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Manager coordinates profile lifecycle, cascading resolution, and sandboxed env export.
type Manager struct {
	config *NetworkConfig
}

// LoadFromPath reads network configuration from a TOML file.
func (m *Manager) LoadFromPath(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			m.EnsureDefaults()
			return nil
		}
		return err
	}
	var root struct {
		Network *NetworkConfig `toml:"network"`
	}
	if err := toml.Unmarshal(data, &root); err != nil {
		return err
	}
	if root.Network != nil {
		m.config = root.Network
	}
	m.EnsureDefaults()
	return nil
}

// SaveToPath updates the network section of a TOML file without modifying other sections.
func (m *Manager) SaveToPath(path string) error {
	m.EnsureDefaults()

	// Serialize only the [network] table
	netMap := map[string]interface{}{
		"network": m.config,
	}
	netBytes, err := toml.Marshal(netMap)
	if err != nil {
		return err
	}
	netString := strings.TrimSpace(string(netBytes))

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) || len(data) == 0 {
		dir := filepath.Dir(path)
		if dir != "" && dir != "." {
			_ = os.MkdirAll(dir, 0755)
		}
		return os.WriteFile(path, []byte(netString+"\n"), 0644)
	}
	if err != nil {
		return err
	}

	// File exists: remove any existing [network] and [network.*] sections,
	// strictly preserving all other comments, tables, and multi-line formatting.
	lines := strings.Split(string(data), "\n")
	var preservedLines []string
	inNetworkSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			secName := strings.Trim(trimmed, "[]")
			secName = strings.TrimSpace(secName)
			if secName == "network" || strings.HasPrefix(secName, "network.") {
				inNetworkSection = true
				continue
			} else {
				inNetworkSection = false
			}
		}
		if !inNetworkSection {
			preservedLines = append(preservedLines, line)
		}
	}

	// Join preserved content and trim trailing empty lines
	preservedContent := strings.TrimRight(strings.Join(preservedLines, "\n"), "\r\n ")

	var finalContent string
	if preservedContent == "" {
		finalContent = netString + "\n"
	} else {
		finalContent = preservedContent + "\n\n" + netString + "\n"
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	return os.WriteFile(path, []byte(finalContent), 0644)
}

// NewManager creates a network manager wrapping the provided NetworkConfig.
func NewManager(cfg *NetworkConfig) *Manager {
	if cfg == nil {
		cfg = &NetworkConfig{}
	}
	m := &Manager{config: cfg}
	m.EnsureDefaults()
	return m
}

// Config returns the underlying NetworkConfig.
func (m *Manager) Config() *NetworkConfig {
	return m.config
}

// EnsureDefaults guarantees that built-in profiles and baseline configs are always seeded.
func (m *Manager) EnsureDefaults() {
	if m.config.ActiveProfile == "" {
		m.config.ActiveProfile = "official"
	}
	if m.config.Profiles == nil {
		m.config.Profiles = make(map[string]NetworkProfile)
	}
	if m.config.Overrides == nil {
		m.config.Overrides = make(map[string]string)
	}

	builtins := DefaultBuiltinProfiles()
	for id, p := range builtins {
		existing, ok := m.config.Profiles[id]
		if !ok {
			m.config.Profiles[id] = p
		} else {
			// Ensure Builtin flag stays true and canonical metadata stays in sync
			existing.Builtin = true
			existing.Name = p.Name
			existing.Description = p.Description
			existing.FallbackProfile = p.FallbackProfile
			if existing.Registries == nil {
				existing.Registries = make(map[string]string)
			}
			// Clean up non-native placeholders from earlier versions of tuna/aliyun so fallback can take effect
			if id == "tuna" {
				if existing.Registries["npm"] == "https://registry.npmmirror.com" {
					delete(existing.Registries, "npm")
				}
				if existing.Registries["go"] == "https://goproxy.cn,direct" {
					delete(existing.Registries, "go")
				}
				if existing.Registries["composer"] == "https://mirrors.aliyun.com/composer/" {
					delete(existing.Registries, "composer")
				}
				if existing.Registries["nuget"] == "https://api.nuget.org/v3/index.json" {
					delete(existing.Registries, "nuget")
				}
				if existing.Registries["docker"] == "https://registry.cn-hangzhou.aliyuncs.com" {
					delete(existing.Registries, "docker")
				}
			} else if id == "aliyun" {
				if existing.Registries["rust"] == "https://rsproxy.cn/crates.io-index" {
					delete(existing.Registries, "rust")
				}
				if existing.Registries["ruby"] == "https://gems.ruby-china.com" {
					delete(existing.Registries, "ruby")
				}
				if existing.Registries["nuget"] == "https://repo.huaweicloud.com/repository/nuget/v3/index.json" {
					delete(existing.Registries, "nuget")
				}
				if existing.Registries["dart"] == "https://mirrors.tuna.tsinghua.edu.cn/dart-pub" {
					delete(existing.Registries, "dart")
				}
				if existing.Registries["cocoapods"] == "https://cdn.cocoapods.org" {
					delete(existing.Registries, "cocoapods")
				}
			}

			for k, v := range p.Registries {
				if _, hasKey := existing.Registries[k]; !hasKey {
					excluded := false
					for _, exc := range m.config.ExcludedEcosystems {
						if exc == k {
							excluded = true
							break
						}
					}
					if !excluded {
						existing.Registries[k] = v
					}
				}
			}
			if existing.Downloads == nil {
				existing.Downloads = p.Downloads
			}
			m.config.Profiles[id] = existing
		}
	}

	if m.config.Proxy.NoProxy == "" {
		m.config.Proxy.NoProxy = DefaultProxyConfig().NoProxy
	}
}

// GetStatus computes the active effective registries considering cascading overrides.
func (m *Manager) GetStatus() *ActiveNetworkStatus {
	m.EnsureDefaults()

	activeID := m.config.ActiveProfile
	activeProfile, ok := m.config.Profiles[activeID]
	if !ok {
		activeProfile = m.config.Profiles["official"]
		activeID = "official"
	}

	officialBlueprint := DefaultBuiltinProfiles()["official"]
	effective := make(map[string]EffectiveRegistry)

	excludedMap := make(map[string]bool)
	for _, exc := range m.config.ExcludedEcosystems {
		excludedMap[strings.ToLower(strings.TrimSpace(exc))] = true
	}

	seen := make(map[string]bool)
	var allEcosystems []string

	for _, eco := range KnownEcosystems {
		eco = strings.ToLower(strings.TrimSpace(eco))
		if !excludedMap[eco] && !seen[eco] {
			seen[eco] = true
			allEcosystems = append(allEcosystems, eco)
		}
	}

	for eco := range activeProfile.Registries {
		eco = strings.ToLower(strings.TrimSpace(eco))
		if !excludedMap[eco] && !seen[eco] {
			seen[eco] = true
			allEcosystems = append(allEcosystems, eco)
		}
	}

	for eco := range m.config.Overrides {
		eco = strings.ToLower(strings.TrimSpace(eco))
		if !excludedMap[eco] && !seen[eco] {
			seen[eco] = true
			allEcosystems = append(allEcosystems, eco)
		}
	}

	for _, eco := range allEcosystems {
		overrideURL, hasOverride := m.config.Overrides[eco]
		if hasOverride && strings.TrimSpace(overrideURL) != "" {
			effective[eco] = EffectiveRegistry{
				Ecosystem:    eco,
				URL:          strings.TrimSpace(overrideURL),
				Source:       "override",
				IsOverridden: true,
			}
			continue
		}

		profileURL, hasProfile := activeProfile.Registries[eco]
		if hasProfile && strings.TrimSpace(profileURL) != "" {
			effective[eco] = EffectiveRegistry{
				Ecosystem:    eco,
				URL:          strings.TrimSpace(profileURL),
				Source:       "profile",
				IsOverridden: false,
			}
			continue
		}

		// Fallback profile resolution (e.g. tuna or aliyun dedicated profiles falling back to cn matrix)
		fallbackID := activeProfile.FallbackProfile
		if fallbackID == "" && (activeID == "tuna" || activeID == "aliyun") {
			fallbackID = "cn"
		}
		if fallbackID != "" && fallbackID != activeID {
			if fbProfile, exists := m.config.Profiles[fallbackID]; exists {
				if fbURL, hasFB := fbProfile.Registries[eco]; hasFB && strings.TrimSpace(fbURL) != "" {
					effective[eco] = EffectiveRegistry{
						Ecosystem:    eco,
						URL:          strings.TrimSpace(fbURL),
						Source:       "fallback_" + fallbackID,
						IsOverridden: false,
					}
					continue
				}
			}
		}

		fallbackURL := officialBlueprint.Registries[eco]
		if strings.TrimSpace(fallbackURL) != "" {
			effective[eco] = EffectiveRegistry{
				Ecosystem:    eco,
				URL:          fallbackURL,
				Source:       "default",
				IsOverridden: false,
			}
		}
	}

	// Copy profiles map
	profilesCopy := make(map[string]NetworkProfile, len(m.config.Profiles))
	for k, v := range m.config.Profiles {
		profilesCopy[k] = v
	}

	return &ActiveNetworkStatus{
		ActiveProfileID: activeID,
		ActiveProfile:   activeProfile,
		Proxy:           m.config.Proxy,
		Effective:       effective,
		Profiles:        profilesCopy,
	}
}

// UseProfile sets the active profile ID.
func (m *Manager) UseProfile(profileID string) error {
	m.EnsureDefaults()
	profileID = strings.TrimSpace(profileID)
	if _, ok := m.config.Profiles[profileID]; !ok {
		return fmt.Errorf("profile %q does not exist", profileID)
	}
	m.config.ActiveProfile = profileID
	return nil
}

// CreateProfile registers a new custom profile.
func (m *Manager) CreateProfile(p NetworkProfile) error {
	m.EnsureDefaults()
	id := strings.TrimSpace(p.ID)
	if id == "" {
		return errors.New("profile id cannot be empty")
	}
	if _, exists := m.config.Profiles[id]; exists {
		return fmt.Errorf("profile %q already exists", id)
	}

	p.ID = id
	p.Builtin = false
	if p.Registries == nil {
		p.Registries = make(map[string]string)
	}
	if p.Downloads == nil {
		p.Downloads = make(map[string]string)
	}

	m.config.Profiles[id] = p
	return nil
}

// SaveAsProfile packages the current effective configuration into a new profile.
func (m *Manager) SaveAsProfile(newID, name, description string) error {
	m.EnsureDefaults()
	newID = strings.TrimSpace(newID)
	if newID == "" {
		return errors.New("new profile id cannot be empty")
	}
	if _, exists := m.config.Profiles[newID]; exists {
		return fmt.Errorf("profile %q already exists", newID)
	}

	status := m.GetStatus()
	registries := make(map[string]string)
	for eco, eff := range status.Effective {
		registries[eco] = eff.URL
	}

	downloads := make(map[string]string)
	for k, v := range status.ActiveProfile.Downloads {
		downloads[k] = v
	}

	if name == "" {
		name = newID
	}

	newProfile := NetworkProfile{
		ID:          newID,
		Name:        name,
		Description: description,
		Builtin:     false,
		Registries:  registries,
		Downloads:   downloads,
	}

	m.config.Profiles[newID] = newProfile
	m.config.ActiveProfile = newID
	// Clear overrides since they are now codified into the new profile
	m.config.Overrides = make(map[string]string)
	return nil
}

// UpdateProfile updates an existing profile.
func (m *Manager) UpdateProfile(p NetworkProfile) error {
	m.EnsureDefaults()
	id := strings.TrimSpace(p.ID)
	existing, ok := m.config.Profiles[id]
	if !ok {
		return fmt.Errorf("profile %q not found", id)
	}

	// Preserve builtin status
	p.Builtin = existing.Builtin
	if p.Name == "" {
		p.Name = existing.Name
	}
	if p.Registries == nil {
		p.Registries = existing.Registries
	}
	if p.Downloads == nil {
		p.Downloads = existing.Downloads
	}

	m.config.Profiles[id] = p
	return nil
}

// ResetProfile restores a built-in profile to its factory default state.
func (m *Manager) ResetProfile(profileID string) error {
	m.EnsureDefaults()
	profileID = strings.TrimSpace(profileID)

	builtins := DefaultBuiltinProfiles()
	factory, ok := builtins[profileID]
	if !ok {
		return fmt.Errorf("only built-in profiles can be reset to factory defaults; %q is custom", profileID)
	}

	m.config.Profiles[profileID] = factory
	return nil
}

// DeleteProfile safely deletes a custom profile.
func (m *Manager) DeleteProfile(profileID string) error {
	m.EnsureDefaults()
	profileID = strings.TrimSpace(profileID)

	p, ok := m.config.Profiles[profileID]
	if !ok {
		return fmt.Errorf("profile %q not found", profileID)
	}
	if p.Builtin {
		return fmt.Errorf("cannot delete built-in profile %q", profileID)
	}

	delete(m.config.Profiles, profileID)
	if m.config.ActiveProfile == profileID {
		m.config.ActiveProfile = "official"
	}
	return nil
}

// SetOverride registers an ecosystem registry override.
func (m *Manager) SetOverride(ecosystem, rawURL string) error {
	m.EnsureDefaults()
	eco := strings.TrimSpace(strings.ToLower(ecosystem))
	if eco == "" {
		return errors.New("ecosystem cannot be empty")
	}

	cleanURL := strings.TrimSpace(rawURL)
	if cleanURL == "" {
		delete(m.config.Overrides, eco)
		return nil
	}

	m.config.Overrides[eco] = cleanURL
	return nil
}

// ResetOverride clears an ecosystem registry override, falling back to profile.
func (m *Manager) ResetOverride(ecosystem string) error {
	m.EnsureDefaults()
	eco := strings.TrimSpace(strings.ToLower(ecosystem))
	delete(m.config.Overrides, eco)
	return nil
}

// AddEcosystem registers or restores an ecosystem in the active profile or overrides.
func (m *Manager) AddEcosystem(ecosystem, rawURL string) error {
	m.EnsureDefaults()
	eco := strings.ToLower(strings.TrimSpace(ecosystem))
	if eco == "" {
		return errors.New("ecosystem identifier cannot be empty")
	}

	cleanURL := strings.TrimSpace(rawURL)

	// Un-exclude if previously excluded
	var newExcluded []string
	for _, exc := range m.config.ExcludedEcosystems {
		if strings.ToLower(strings.TrimSpace(exc)) != eco {
			newExcluded = append(newExcluded, exc)
		}
	}
	m.config.ExcludedEcosystems = newExcluded

	// If URL is provided, set as override or set into active profile if custom
	activeID := m.config.ActiveProfile
	activeProfile, exists := m.config.Profiles[activeID]
	if exists && !activeProfile.Builtin {
		if activeProfile.Registries == nil {
			activeProfile.Registries = make(map[string]string)
		}
		if cleanURL != "" {
			activeProfile.Registries[eco] = cleanURL
		}
		m.config.Profiles[activeID] = activeProfile
	} else if cleanURL != "" {
		m.config.Overrides[eco] = cleanURL
	}

	return nil
}

// DeleteEcosystem removes or excludes an ecosystem from effective resolution and active profile.
func (m *Manager) DeleteEcosystem(ecosystem string) error {
	m.EnsureDefaults()
	eco := strings.ToLower(strings.TrimSpace(ecosystem))
	if eco == "" {
		return errors.New("ecosystem identifier cannot be empty")
	}

	// Remove from overrides
	delete(m.config.Overrides, eco)

	// Remove from active profile if custom
	activeID := m.config.ActiveProfile
	if activeProfile, exists := m.config.Profiles[activeID]; exists && !activeProfile.Builtin {
		delete(activeProfile.Registries, eco)
		delete(activeProfile.Downloads, eco)
		m.config.Profiles[activeID] = activeProfile
	}

	// Add to ExcludedEcosystems so built-in default resolution suppresses it
	alreadyExcluded := false
	for _, exc := range m.config.ExcludedEcosystems {
		if strings.ToLower(strings.TrimSpace(exc)) == eco {
			alreadyExcluded = true
			break
		}
	}
	if !alreadyExcluded {
		m.config.ExcludedEcosystems = append(m.config.ExcludedEcosystems, eco)
	}

	return nil
}

// ResetDefaultEcosystems clears all exclusions/overrides and restores all standard built-in ecosystems.
func (m *Manager) ResetDefaultEcosystems() error {
	m.EnsureDefaults()
	m.config.ExcludedEcosystems = nil
	m.config.Overrides = make(map[string]string)

	builtins := DefaultBuiltinProfiles()
	for id, bp := range builtins {
		if prof, ok := m.config.Profiles[id]; ok && prof.Builtin {
			prof.Registries = make(map[string]string)
			for k, v := range bp.Registries {
				prof.Registries[k] = v
			}
			m.config.Profiles[id] = prof
		}
	}

	return nil
}

// SetProxy updates proxy configuration with strict validation.
func (m *Manager) SetProxy(p ProxyConfig) error {
	m.EnsureDefaults()

	validateURL := func(raw, label string) error {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return nil
		}
		if !strings.HasPrefix(trimmed, "http://") &&
			!strings.HasPrefix(trimmed, "https://") &&
			!strings.HasPrefix(trimmed, "socks5://") &&
			!strings.HasPrefix(trimmed, "socks5h://") {
			return fmt.Errorf("invalid %s proxy URL %q: must start with http://, https://, or socks5://", label, trimmed)
		}
		_, err := url.Parse(trimmed)
		if err != nil {
			return fmt.Errorf("invalid %s proxy URL %q: %w", label, trimmed, err)
		}
		return nil
	}

	if err := validateURL(p.HTTP, "HTTP"); err != nil {
		return err
	}
	if err := validateURL(p.HTTPS, "HTTPS"); err != nil {
		return err
	}
	if err := validateURL(p.SOCKS5, "SOCKS5"); err != nil {
		return err
	}

	if p.HTTPS == "" && p.HTTP != "" {
		p.HTTPS = p.HTTP
	}

	m.config.Proxy = p
	return nil
}

// ExportEnvironment computes standard environment variables to inject into child processes.
func (m *Manager) ExportEnvironment() map[string]string {
	status := m.GetStatus()
	env := make(map[string]string)

	// 1. Proxy injection
	if status.Proxy.Enabled {
		if status.Proxy.HTTP != "" {
			env["HTTP_PROXY"] = status.Proxy.HTTP
			env["http_proxy"] = status.Proxy.HTTP
		}
		if status.Proxy.HTTPS != "" {
			env["HTTPS_PROXY"] = status.Proxy.HTTPS
			env["https_proxy"] = status.Proxy.HTTPS
		}
		if status.Proxy.SOCKS5 != "" {
			env["ALL_PROXY"] = status.Proxy.SOCKS5
			env["all_proxy"] = status.Proxy.SOCKS5
		}
		if status.Proxy.NoProxy != "" {
			env["NO_PROXY"] = status.Proxy.NoProxy
			env["no_proxy"] = status.Proxy.NoProxy
		}
	}

	// 2. Package registries injection
	if npm, ok := status.Effective["npm"]; ok && npm.URL != "" {
		env["NPM_CONFIG_REGISTRY"] = npm.URL
	}
	if goproxy, ok := status.Effective["go"]; ok && goproxy.URL != "" {
		env["GOPROXY"] = goproxy.URL
	}
	if pip, ok := status.Effective["python"]; ok && pip.URL != "" {
		env["PIP_INDEX_URL"] = pip.URL
	}
	if cargo, ok := status.Effective["rust"]; ok && cargo.URL != "" {
		env["CARGO_REGISTRIES_CRATES_IO_INDEX"] = cargo.URL
		if strings.Contains(cargo.URL, "rsproxy.cn") {
			env["RUSTUP_DIST_SERVER"] = "https://rsproxy.cn/rustup"
			env["RUSTUP_UPDATE_ROOT"] = "https://rsproxy.cn/rustup/rustup"
		} else if strings.Contains(cargo.URL, "tsinghua") {
			env["RUSTUP_DIST_SERVER"] = "https://mirrors.tuna.tsinghua.edu.cn/rustup"
			env["RUSTUP_UPDATE_ROOT"] = "https://mirrors.tuna.tsinghua.edu.cn/rustup/rustup"
		}
	}
	if dart, ok := status.Effective["dart"]; ok && dart.URL != "" {
		env["PUB_HOSTED_URL"] = dart.URL
	}
	if brew, ok := status.Effective["homebrew"]; ok && brew.URL != "" {
		env["HOMEBREW_BOTTLE_DOMAIN"] = brew.URL
	}

	// 3. Downloads mirror
	if nodeMirror, ok := status.ActiveProfile.Downloads["node"]; ok && nodeMirror != "" {
		env["NODEJS_ORG_MIRROR"] = nodeMirror
	}
	if goMirror, ok := status.ActiveProfile.Downloads["go"]; ok && goMirror != "" {
		env["UNIRTM_GO_DOWNLOAD_MIRROR"] = goMirror
	}
	if pyMirror, ok := status.ActiveProfile.Downloads["python"]; ok && pyMirror != "" {
		env["PYTHON_BUILD_MIRROR_URL"] = pyMirror
	}

	return env
}
