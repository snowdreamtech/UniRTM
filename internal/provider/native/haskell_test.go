// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"os"
	"testing"
)

func TestHaskellHandler(t *testing.T) {
	if os.Getenv("TEST_NETWORK") == "" {
		t.Skip("Skipping network test. Set TEST_NETWORK=1 to enable.")
	}
	h := &HaskellHandler{}
	if h.Name() != "haskell" {
		t.Errorf("expected name 'haskell', got '%s'", h.Name())
	}

	versions, err := h.ResolveVersions(context.Background(), "")
	if err != nil {
		t.Fatalf("failed to resolve versions: %v", err)
	}

	if len(versions) == 0 {
		t.Fatal("expected at least one version")
	}

	for _, v := range versions {
		if len(v.Assets) == 0 {
			t.Errorf("version %s has no assets", v.Version)
		}
	}
}
