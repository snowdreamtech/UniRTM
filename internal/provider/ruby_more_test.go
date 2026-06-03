// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRubyProvider_More(t *testing.T) {
	p := NewRubyProvider(NewNativeProvider())

	dir := t.TempDir()

	binDir := filepath.Join(dir, "bin")
	os.MkdirAll(binDir, 0755)

	rubyExe := filepath.Join(binDir, "ruby")
	os.WriteFile(rubyExe, []byte("#!/bin/sh\nexit 0\n"), 0755)

	p.Install(context.Background(), "tool", dir, "artifact", "1.0.0")
	p.DetectVersion(context.Background(), "tool", dir)
	p.Uninstall(context.Background(), "tool", dir, "1.0.0")
	p.generateRubyShim("tool", "ruby", rubyExe, dir, "1.0.0")
}
