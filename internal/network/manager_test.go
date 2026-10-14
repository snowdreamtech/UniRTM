// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package network

import (
	"path/filepath"
	"testing"
)

func TestDefaultSeedingAndSecurity(t *testing.T) {
	mgr := NewManager(nil)
	cfg := mgr.Config()

	if cfg.ActiveProfile != "official" {
		t.Fatalf("expected active profile to be 'official', got %q", cfg.ActiveProfile)
	}

	if cfg.Proxy.Enabled {
		t.Fatalf("security violation: proxy must be disabled by default")
	}
	if cfg.Proxy.HTTP != "" || cfg.Proxy.SOCKS5 != "" {
		t.Fatalf("security violation: proxy URLs must be empty by default, got http=%q", cfg.Proxy.HTTP)
	}

	if len(cfg.Profiles) < 2 {
		t.Fatalf("expected at least official and cn builtins, got %d", len(cfg.Profiles))
	}

	official, ok := cfg.Profiles["official"]
	if !ok || !official.Builtin {
		t.Fatalf("official profile missing or not marked builtin")
	}

	cn, ok := cfg.Profiles["cn"]
	if !ok || !cn.Builtin {
		t.Fatalf("cn profile missing or not marked builtin")
	}
}

func TestCascadingResolution(t *testing.T) {
	mgr := NewManager(nil)
	status := mgr.GetStatus()

	// 1. By default, follows 'official' profile
	npmEff, ok := status.Effective["npm"]
	if !ok || npmEff.URL != "https://registry.npmjs.org" {
		t.Fatalf("expected default npm to be official npmjs, got %v", npmEff)
	}
	if npmEff.Source != "profile" || npmEff.IsOverridden {
		t.Fatalf("expected source to be 'profile', got %s, isOverridden=%v", npmEff.Source, npmEff.IsOverridden)
	}

	// 2. Set an override for npm only
	err := mgr.SetOverride("npm", "https://npm.corp.internal")
	if err != nil {
		t.Fatalf("unexpected error setting override: %v", err)
	}

	status = mgr.GetStatus()
	npmEff = status.Effective["npm"]
	if npmEff.URL != "https://npm.corp.internal" || npmEff.Source != "override" || !npmEff.IsOverridden {
		t.Fatalf("expected override to take precedence: %+v", npmEff)
	}

	// Go should still follow official
	goEff := status.Effective["go"]
	if goEff.URL != "https://proxy.golang.org,direct" || goEff.Source != "profile" || goEff.IsOverridden {
		t.Fatalf("expected go to still follow profile: %+v", goEff)
	}

	// 3. Reset override
	err = mgr.ResetOverride("npm")
	if err != nil {
		t.Fatalf("unexpected error resetting override: %v", err)
	}
	status = mgr.GetStatus()
	if status.Effective["npm"].URL != "https://registry.npmjs.org" || status.Effective["npm"].IsOverridden {
		t.Fatalf("expected npm to fall back to official profile after reset: %+v", status.Effective["npm"])
	}
}

func TestProfileLifecycle(t *testing.T) {
	mgr := NewManager(nil)

	// Create custom profile
	err := mgr.CreateProfile(NetworkProfile{
		ID:   "corp",
		Name: "Corporate Network",
		Registries: map[string]string{
			"npm": "https://npm.corp.com",
			"go":  "https://goproxy.corp.com",
		},
	})
	if err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}

	err = mgr.UseProfile("corp")
	if err != nil {
		t.Fatalf("failed to switch profile: %v", err)
	}

	status := mgr.GetStatus()
	if status.ActiveProfileID != "corp" {
		t.Fatalf("expected active profile to be 'corp', got %s", status.ActiveProfileID)
	}
	if status.Effective["npm"].URL != "https://npm.corp.com" {
		t.Fatalf("expected npm to resolve to corp, got %s", status.Effective["npm"].URL)
	}

	// Cannot delete built-in profile
	err = mgr.DeleteProfile("cn")
	if err == nil {
		t.Fatalf("expected error deleting built-in profile")
	}

	// Delete custom profile (and fall back to cn)
	err = mgr.DeleteProfile("corp")
	if err != nil {
		t.Fatalf("failed to delete custom profile: %v", err)
	}
	if mgr.Config().ActiveProfile != "official" {
		t.Fatalf("expected active profile to fall back to 'official', got %s", mgr.Config().ActiveProfile)
	}
}

func TestSaveAsProfile(t *testing.T) {
	mgr := NewManager(nil)
	_ = mgr.SetOverride("npm", "https://custom.npm.org")

	err := mgr.SaveAsProfile("custom-mix", "My Custom Mix", "Saved from active")
	if err != nil {
		t.Fatalf("failed to save as: %v", err)
	}

	if mgr.Config().ActiveProfile != "custom-mix" {
		t.Fatalf("expected active profile to switch to 'custom-mix'")
	}
	// Overrides should be cleared
	if len(mgr.Config().Overrides) != 0 {
		t.Fatalf("expected overrides to be cleared after save-as, got %v", mgr.Config().Overrides)
	}

	status := mgr.GetStatus()
	if status.Effective["npm"].URL != "https://custom.npm.org" {
		t.Fatalf("expected saved profile to retain custom URL: %s", status.Effective["npm"].URL)
	}
}

func TestResetBuiltinToFactory(t *testing.T) {
	mgr := NewManager(nil)
	_ = mgr.UseProfile("cn")
	p := mgr.Config().Profiles["cn"]
	p.Registries["npm"] = "https://corrupted.url"
	_ = mgr.UpdateProfile(p)

	if mgr.GetStatus().Effective["npm"].URL != "https://corrupted.url" {
		t.Fatalf("profile was not modified")
	}

	err := mgr.ResetProfile("cn")
	if err != nil {
		t.Fatalf("failed to reset builtin profile: %v", err)
	}

	if mgr.GetStatus().Effective["npm"].URL != "https://registry.npmmirror.com" {
		t.Fatalf("failed to restore factory value for cn npmmirror")
	}
}

func TestProxyValidationAndEnvExport(t *testing.T) {
	mgr := NewManager(nil)

	// Invalid scheme
	err := mgr.SetProxy(ProxyConfig{
		Enabled: true,
		HTTP:    "ftp://invalid.proxy:8080",
	})
	if err == nil {
		t.Fatalf("expected error on invalid proxy scheme")
	}

	// Valid http proxy
	err = mgr.SetProxy(ProxyConfig{
		Enabled: true,
		HTTP:    "http://127.0.0.1:7890",
		NoProxy: "localhost,127.0.0.1",
	})
	if err != nil {
		t.Fatalf("unexpected error setting valid proxy: %v", err)
	}

	env := mgr.ExportEnvironment()
	if env["HTTP_PROXY"] != "http://127.0.0.1:7890" {
		t.Fatalf("expected HTTP_PROXY to be exported, got %q", env["HTTP_PROXY"])
	}
	if env["NPM_CONFIG_REGISTRY"] != "https://registry.npmjs.org" {
		t.Fatalf("expected NPM_CONFIG_REGISTRY, got %q", env["NPM_CONFIG_REGISTRY"])
	}
	if env["GOPROXY"] != "https://proxy.golang.org,direct" {
		t.Fatalf("expected GOPROXY, got %q", env["GOPROXY"])
	}
}

func TestFilePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	mgr1 := NewManager(nil)
	_ = mgr1.UseProfile("official")
	_ = mgr1.SetOverride("npm", "https://custom.npm.org")
	_ = mgr1.SetProxy(ProxyConfig{
		Enabled: true,
		HTTP:    "http://127.0.0.1:8888",
		NoProxy: "127.0.0.1",
	})

	err := mgr1.SaveToPath(configPath)
	if err != nil {
		t.Fatalf("failed to save to path: %v", err)
	}

	mgr2 := NewManager(nil)
	err = mgr2.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("failed to load from path: %v", err)
	}

	status := mgr2.GetStatus()
	if status.ActiveProfileID != "official" {
		t.Fatalf("expected active profile to be official, got %s", status.ActiveProfileID)
	}
	if status.Effective["npm"].URL != "https://custom.npm.org" || !status.Effective["npm"].IsOverridden {
		t.Fatalf("expected npm to be overridden, got %+v", status.Effective["npm"])
	}
	if !status.Proxy.Enabled || status.Proxy.HTTP != "http://127.0.0.1:8888" {
		t.Fatalf("expected proxy to be persisted, got %+v", status.Proxy)
	}
}

func TestRecommendedMirrorsOrdering(t *testing.T) {
	mgr := NewManager(nil)
	official := mgr.Config().Profiles["official"]
	rec := RecommendedMirrors()

	for _, eco := range KnownEcosystems {
		list, ok := rec[eco]
		if !ok || len(list) == 0 {
			t.Fatalf("ecosystem %s missing recommended mirrors", eco)
		}
		// First entry must be official direct connection
		first := list[0]
		expectedURL := official.Registries[eco]
		if first.URL != expectedURL {
			t.Fatalf("ecosystem %s expected first mirror to be official upstream %q, got %q", eco, expectedURL, first.URL)
		}
	}
}

func TestAddAndDeleteEcosystem(t *testing.T) {
	mgr := NewManager(nil)

	status := mgr.GetStatus()
	if _, ok := status.Effective["composer"]; !ok {
		t.Fatalf("composer expected in expanded default ecosystems")
	}

	// Add custom ecosystem
	err := mgr.AddEcosystem("customlang", "https://customlang.org/repo")
	if err != nil {
		t.Fatalf("failed to add customlang: %v", err)
	}

	status = mgr.GetStatus()
	customEff, ok := status.Effective["customlang"]
	if !ok || customEff.URL != "https://customlang.org/repo" {
		t.Fatalf("expected customlang to be effective: %+v", customEff)
	}

	// Delete custom ecosystem
	err = mgr.DeleteEcosystem("customlang")
	if err != nil {
		t.Fatalf("failed to delete customlang: %v", err)
	}

	status = mgr.GetStatus()
	if _, ok := status.Effective["customlang"]; ok {
		t.Fatalf("expected customlang to be removed from effective")
	}

	// Delete built-in ecosystem (e.g. ruby)
	err = mgr.DeleteEcosystem("ruby")
	if err != nil {
		t.Fatalf("failed to delete ruby: %v", err)
	}

	status = mgr.GetStatus()
	if _, ok := status.Effective["ruby"]; ok {
		t.Fatalf("expected ruby to be excluded from effective")
	}

	// Re-add ruby
	err = mgr.AddEcosystem("ruby", "https://rubygems.org")
	if err != nil {
		t.Fatalf("failed to re-add ruby: %v", err)
	}

	status = mgr.GetStatus()
	if _, ok := status.Effective["ruby"]; !ok {
		t.Fatalf("expected ruby to be restored in effective")
	}

	// Delete dart and homebrew, then reset
	_ = mgr.DeleteEcosystem("dart")
	_ = mgr.DeleteEcosystem("homebrew")
	status = mgr.GetStatus()
	if _, ok := status.Effective["dart"]; ok {
		t.Fatalf("expected dart to be excluded")
	}

	err = mgr.ResetDefaultEcosystems()
	if err != nil {
		t.Fatalf("unexpected error resetting ecosystems: %v", err)
	}

	status = mgr.GetStatus()
	for _, eco := range KnownEcosystems {
		if _, ok := status.Effective[eco]; !ok {
			t.Fatalf("expected default ecosystem %s to be restored", eco)
		}
	}
}

func TestDedicatedProfilesFallback(t *testing.T) {
	mgr := NewManager(nil)

	// 1. Test TUNA profile: native python/rust/dart, fallback npm/go/composer to cn
	err := mgr.UseProfile("tuna")
	if err != nil {
		t.Fatalf("failed to switch to tuna profile: %v", err)
	}
	status := mgr.GetStatus()
	if status.ActiveProfileID != "tuna" {
		t.Fatalf("expected active profile tuna, got %s", status.ActiveProfileID)
	}

	// Python is natively hosted by TUNA
	pyEff := status.Effective["python"]
	if pyEff.URL != "https://pypi.tuna.tsinghua.edu.cn/simple" || pyEff.Source != "profile" {
		t.Fatalf("expected python to come from tuna profile, got %+v", pyEff)
	}
	// Dart is natively hosted by TUNA
	dartEff := status.Effective["dart"]
	if dartEff.URL != "https://mirrors.tuna.tsinghua.edu.cn/dart-pub" || dartEff.Source != "profile" {
		t.Fatalf("expected dart to come from tuna profile, got %+v", dartEff)
	}
	// NPM is NOT hosted by TUNA -> fallback to cn (npmmirror)
	npmEff := status.Effective["npm"]
	if npmEff.URL != "https://registry.npmmirror.com" || npmEff.Source != "fallback_cn" {
		t.Fatalf("expected npm to fallback to cn npmmirror, got %+v", npmEff)
	}
	// Go is NOT hosted by TUNA -> fallback to cn (goproxy.cn)
	goEff := status.Effective["go"]
	if goEff.URL != "https://goproxy.cn,direct" || goEff.Source != "fallback_cn" {
		t.Fatalf("expected go to fallback to cn goproxy.cn, got %+v", goEff)
	}
	// Composer is NOT hosted by TUNA -> fallback to cn (aliyun composer)
	composerEff := status.Effective["composer"]
	if composerEff.URL != "https://mirrors.aliyun.com/composer/" || composerEff.Source != "fallback_cn" {
		t.Fatalf("expected composer to fallback to cn aliyun composer, got %+v", composerEff)
	}

	// 2. Test Aliyun profile: native npm/go/composer, fallback rust/dart/cocoapods to cn
	err = mgr.UseProfile("aliyun")
	if err != nil {
		t.Fatalf("failed to switch to aliyun profile: %v", err)
	}
	status = mgr.GetStatus()

	// NPM is natively hosted by Aliyun
	aliNpm := status.Effective["npm"]
	if aliNpm.URL != "https://registry.npmmirror.com" || aliNpm.Source != "profile" {
		t.Fatalf("expected npm to come from aliyun profile, got %+v", aliNpm)
	}
	// Go is natively hosted by Aliyun
	aliGo := status.Effective["go"]
	if aliGo.URL != "https://mirrors.aliyun.com/goproxy/,direct" || aliGo.Source != "profile" {
		t.Fatalf("expected go to come from aliyun profile, got %+v", aliGo)
	}
	// Rust is NOT hosted by Aliyun -> fallback to cn (rsproxy.cn)
	aliRust := status.Effective["rust"]
	if aliRust.URL != "https://rsproxy.cn/crates.io-index" || aliRust.Source != "fallback_cn" {
		t.Fatalf("expected rust to fallback to cn rsproxy.cn, got %+v", aliRust)
	}
	// Dart is NOT hosted by Aliyun -> fallback to cn (tuna dart-pub)
	aliDart := status.Effective["dart"]
	if aliDart.URL != "https://mirrors.tuna.tsinghua.edu.cn/dart-pub" || aliDart.Source != "fallback_cn" {
		t.Fatalf("expected dart to fallback to cn tuna dart-pub, got %+v", aliDart)
	}
	// CocoaPods is NOT hosted by Aliyun -> fallback to cn (tuna specs git)
	aliPods := status.Effective["cocoapods"]
	if aliPods.URL != "https://mirrors.tuna.tsinghua.edu.cn/git/CocoaPods/Specs.git" || aliPods.Source != "fallback_cn" {
		t.Fatalf("expected cocoapods to fallback to cn tuna specs git, got %+v", aliPods)
	}

	// 3. Override takes precedence over fallback
	_ = mgr.SetOverride("rust", "https://custom.rust.org")
	status = mgr.GetStatus()
	if status.Effective["rust"].URL != "https://custom.rust.org" || status.Effective["rust"].Source != "override" {
		t.Fatalf("expected override to take precedence over fallback: %+v", status.Effective["rust"])
	}
}



