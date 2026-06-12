// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
	"testing"
)

func TestDartHandler(t *testing.T) {
	h := &DartHandler{}
	if h.Name() != "dart" {
		t.Errorf("expected name 'dart', got '%s'", h.Name())
	}

	versions, err := h.ResolveVersions(context.Background(), "")
	if err != nil {
		t.Fatalf("failed to resolve versions: %v", err)
	}

	if len(versions) == 0 {
		t.Fatal("expected at least one version")
	}

	foundLatest := false
	for _, v := range versions {
		if v.Version == "latest" {
			foundLatest = true
		}
		if len(v.Assets) == 0 {
			t.Errorf("version %s has no assets", v.Version)
		}
	}

	if !foundLatest {
		t.Error("expected 'latest' version to be present")
	}
}
