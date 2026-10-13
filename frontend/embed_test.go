// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package frontend

import (
	"io/fs"
	"testing"
)

func TestGetFS(t *testing.T) {
	sub, err := GetFS()
	if err != nil {
		t.Fatalf("GetFS returned error: %v", err)
	}

	// Verify we can stat files in the embedded FS
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		t.Fatalf("ReadDir on sub FS failed: %v", err)
	}

	if len(entries) == 0 {
		t.Log("Warning: dist is empty or contains only .gitkeep")
	}
}
