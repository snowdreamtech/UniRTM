// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPypiProvider_More(t *testing.T) {
	p := NewPypiProvider()

	dir := t.TempDir()

	binDir := filepath.Join(dir, "bin")
	os.MkdirAll(binDir, 0755)

	p.PostInstall(context.Background(), "tool", dir, "1.0.0")
	isOnlyDigitsAndDots("1.0.0")
	isOnlyDigitsAndDots("1.0.0.abc")
}
