// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"testing"
	"time"
)

func TestNativeDiskCache(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("UNIRTM_CACHE_DIR", tempDir)

	tool := "node"
	key := "https://nodejs.org/dist|musl"
	versions := []VersionInfo{
		{
			Version: "20.10.0",
			IsLTS:   true,
			LTSName: "Iron",
			Assets: []Asset{
				{
					URL:      "https://nodejs.org/dist/v20.10.0/node-v20.10.0-linux-x64.tar.gz",
					Filename: "node-v20.10.0-linux-x64.tar.gz",
					OS:       "linux",
					Arch:     "amd64",
				},
			},
		},
	}

	// 1. Initial read should miss
	var out []VersionInfo
	if readNativeDiskCache(tool, key, &out, 10*time.Minute) {
		t.Fatal("expected cache miss before write")
	}

	// 2. Write to disk cache
	writeNativeDiskCache(tool, key, versions)

	// 3. Read back
	var readBack []VersionInfo
	if !readNativeDiskCache(tool, key, &readBack, 10*time.Minute) {
		t.Fatal("expected cache hit after write")
	}
	if len(readBack) != 1 || readBack[0].Version != "20.10.0" || !readBack[0].IsLTS {
		t.Fatalf("unexpected cache read back: %+v", readBack)
	}

	// 4. Test TTL expiry
	var expired []VersionInfo
	if readNativeDiskCache(tool, key, &expired, -1*time.Second) {
		t.Fatal("expected expired cache to be rejected")
	}
}
