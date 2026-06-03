// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/provider"
)

func TestInstallationManager_More(t *testing.T) {
	br := backend.NewRegistry()
	pr := provider.NewRegistry()
	im := NewInstallationManager(br, pr, nil, nil, nil, nil)
	if im == nil {
		t.Fatal("expected non-nil InstallationManager")
	}
	im.SetDB(nil)
	im.Close()

	// mock empty calls - skip Install to avoid further panics from nil repositories
	im.EnsureInstalledFromSpecs(context.Background(), nil)
}
