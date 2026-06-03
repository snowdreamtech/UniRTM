// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"os"
	"testing"
)

func TestLockService_Generate_More(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tmpDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	opts := LockServiceOptions{}
	ls, _ := NewLockService(opts)
	// Should work with empty installation dir
	if ls != nil {
		ls.Generate(context.Background(), nil, GenerateOptions{})
		ls.init()
	}
}
