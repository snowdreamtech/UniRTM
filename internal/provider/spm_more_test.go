// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSpmProvider_copyFile(t *testing.T) {
	p := NewSpmProvider()

	dir := t.TempDir()

	// Create dummy file
	dummyFile := filepath.Join(dir, "dummy.txt")
	os.WriteFile(dummyFile, []byte("test"), 0644)

	dest := filepath.Join(dir, "dest.txt")
	p.copyFile(dummyFile, dest)
}
