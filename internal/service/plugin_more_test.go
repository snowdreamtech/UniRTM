// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPluginManager_LoadPlugin(t *testing.T) {
	pm := NewPluginManager(t.TempDir(), nil, nil)

	tmpDir := t.TempDir()
	dummyPath := filepath.Join(tmpDir, "dummy_plugin.so")
	os.WriteFile(dummyPath, []byte("dummy"), 0644)

	pm.loadPlugin(context.Background(), dummyPath)
}
