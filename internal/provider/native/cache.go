// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

type nativeCacheWrapper struct {
	SavedAt      time.Time     `json:"saved_at"`
	ETag         string        `json:"etag,omitempty"`
	LastModified string        `json:"last_modified,omitempty"`
	Versions     []VersionInfo `json:"versions"`
}

func getNativeDiskCachePath(tool, key string) string {
	safeTool := strings.NewReplacer("/", "_", ":", "_", "@", "_").Replace(tool)
	safeKey := strings.NewReplacer("/", "_", ":", "_", "@", "_", "|", "_").Replace(key)
	return filepath.Join(env.GetCacheDir(), "native", safeTool, safeKey+".json")
}

// readNativeDiskCache reads cached versions if not expired by ttl.
func readNativeDiskCache(tool, key string, versions *[]VersionInfo, ttl time.Duration) bool {
	hit, _, _ := readNativeDiskCacheMeta(tool, key, versions, ttl)
	return hit
}

// readNativeDiskCacheMeta returns hit status, ETag and any existing cached versions (even if expired).
func readNativeDiskCacheMeta(tool, key string, versions *[]VersionInfo, ttl time.Duration) (hit bool, etag string, cached []VersionInfo) {
	if pkgHttp.MockTransport != nil {
		return false, "", nil
	}
	p := getNativeDiskCachePath(tool, key)
	data, err := os.ReadFile(p)
	if err != nil {
		return false, "", nil
	}

	var wrapper nativeCacheWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return false, "", nil
	}

	etag = wrapper.ETag
	cached = wrapper.Versions

	if ttl > 0 && time.Since(wrapper.SavedAt) > ttl {
		return false, etag, cached
	}

	if versions != nil {
		*versions = wrapper.Versions
	}
	return true, etag, cached
}

// readNativeDiskCacheStale returns cached versions even if TTL has expired.
// Useful for fallback on network timeout, offline, or rate-limited responses.
func readNativeDiskCacheStale(tool, key string, versions *[]VersionInfo) bool {
	if pkgHttp.MockTransport != nil {
		return false
	}
	p := getNativeDiskCachePath(tool, key)
	data, err := os.ReadFile(p)
	if err != nil {
		return false
	}

	var wrapper nativeCacheWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return false
	}

	if len(wrapper.Versions) == 0 {
		return false
	}

	if versions != nil {
		*versions = wrapper.Versions
	}
	return true
}

// touchNativeDiskCache updates the SavedAt timestamp to avoid re-fetching on 304 Not Modified.
func touchNativeDiskCache(tool, key string) {
	if pkgHttp.MockTransport != nil {
		return
	}
	p := getNativeDiskCachePath(tool, key)
	data, err := os.ReadFile(p)
	if err != nil {
		return
	}

	var wrapper nativeCacheWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return
	}

	wrapper.SavedAt = time.Now()
	bytes, err := json.Marshal(wrapper)
	if err != nil {
		return
	}

	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, bytes, 0644); err == nil {
		_ = os.Rename(tmp, p)
	}
}

// writeNativeDiskCache writes versions with default/no etag.
func writeNativeDiskCache(tool, key string, versions []VersionInfo) {
	writeNativeDiskCacheWithETag(tool, key, versions, "", "")
}

// writeNativeDiskCacheWithETag writes versions with etag and lastModified metadata.
func writeNativeDiskCacheWithETag(tool, key string, versions []VersionInfo, etag, lastModified string) {
	if pkgHttp.MockTransport != nil || len(versions) == 0 {
		return
	}
	p := getNativeDiskCachePath(tool, key)
	wrapper := nativeCacheWrapper{
		SavedAt:      time.Now(),
		ETag:         etag,
		LastModified: lastModified,
		Versions:     versions,
	}

	bytes, err := json.Marshal(wrapper)
	if err != nil {
		return
	}

	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return
	}

	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, bytes, 0644); err == nil {
		_ = os.Rename(tmp, p)
	}
}
