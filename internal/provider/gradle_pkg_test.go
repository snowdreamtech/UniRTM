// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"strings"
	"testing"
)

func TestGradlePkgProvider_Name(t *testing.T) {
	p := NewGradlePkgProvider()
	if p.Name() != "gradle-pkg" {
		t.Errorf("Expected name 'gradle-pkg', got '%s'", p.Name())
	}
}

// Since GradlePkgProvider delegates to MavenPkgProvider, we just need to verify
// that the delegation properly triggers errors through the MavenPkgProvider logic.
func TestGradlePkgProvider_Install_Delegation(t *testing.T) {
	p := NewGradlePkgProvider()
	ctx := context.Background()
	installPath := t.TempDir()

	err := p.Install(ctx, "invalid-gradle-tool", installPath, "", "1.0.0")
	if err == nil {
		t.Errorf("Expected error for invalid tool name format due to MavenPkgProvider delegation")
	} else if !strings.Contains(err.Error(), "must be in the format") {
		t.Errorf("Expected format error from MavenPkgProvider, got: %v", err)
	}
}
