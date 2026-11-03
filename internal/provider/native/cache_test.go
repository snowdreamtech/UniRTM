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

	// 5. Test Stale fallback (should succeed even if expired)
	var stale []VersionInfo
	if !readNativeDiskCacheStale(tool, key, &stale) || len(stale) != 1 {
		t.Fatal("expected stale cache to be successfully read")
	}

	// 6. Test Write with ETag
	etagKey := "github/astral-sh/python"
	writeNativeDiskCacheWithETag(tool, etagKey, versions, "W/\"etag-12345\"", "Wed, 21 Oct 2026 07:28:00 GMT")

	var metaVersions []VersionInfo
	hit, etag, _ := readNativeDiskCacheMeta(tool, etagKey, &metaVersions, 10*time.Minute)
	if !hit || etag != "W/\"etag-12345\"" || len(metaVersions) != 1 {
		t.Fatalf("expected etag cache hit, got hit=%v etag=%s", hit, etag)
	}

	// 7. Test touchNativeDiskCache
	touchNativeDiskCache(tool, etagKey)
	hitAfterTouch, etagAfterTouch, _ := readNativeDiskCacheMeta(tool, etagKey, &metaVersions, 10*time.Minute)
	if !hitAfterTouch || etagAfterTouch != "W/\"etag-12345\"" {
		t.Fatalf("expected cache to remain valid after touch, got hit=%v", hitAfterTouch)
	}
}
