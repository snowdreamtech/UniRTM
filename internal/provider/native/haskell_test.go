// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"testing"
)

func TestHaskellHandler(t *testing.T) {
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
