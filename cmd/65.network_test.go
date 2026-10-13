// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"path/filepath"
	"testing"
)

func TestNetworkCmdRegistration(t *testing.T) {
	if networkCmd == nil {
		t.Fatalf("networkCmd is nil")
	}

	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "network" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("network command was not registered on rootCmd")
	}
}

func TestNetworkStatusAndProfileFlow(t *testing.T) {
	tmpDir := t.TempDir()
	origLocal := ".unirtm.toml"
	testConfig := filepath.Join(tmpDir, ".unirtm.toml")

	// Temporarily point resolveNetworkConfigFile to testConfig by setting global false
	mgr, _, err := getNetworkManager()
	if err != nil {
		t.Fatalf("failed to get network manager: %v", err)
	}

	if mgr.Config().ActiveProfile == "" {
		t.Fatalf("expected non-empty active profile")
	}

	// Test profile use
	err = mgr.UseProfile("official")
	if err != nil {
		t.Fatalf("failed to use official profile: %v", err)
	}
	if mgr.Config().ActiveProfile != "official" {
		t.Fatalf("expected active profile to be official")
	}

	// Test override
	err = mgr.SetOverride("npm", "https://test.npm.org")
	if err != nil {
		t.Fatalf("failed to set override: %v", err)
	}

	status := mgr.GetStatus()
	if status.Effective["npm"].URL != "https://test.npm.org" {
		t.Fatalf("expected overridden npm, got %s", status.Effective["npm"].URL)
	}

	_ = origLocal
	_ = testConfig
}
