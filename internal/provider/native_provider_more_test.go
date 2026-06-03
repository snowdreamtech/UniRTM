// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNativeProvider_isExecutable(t *testing.T) {
	dir := t.TempDir()

	dummyFile := filepath.Join(dir, "dummy.txt")
	os.WriteFile(dummyFile, []byte("test"), 0644)

	info, _ := os.Stat(dummyFile)
	isExecutable(info)
}
