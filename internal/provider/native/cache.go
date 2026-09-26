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
	SavedAt  time.Time     `json:"saved_at"`
	Versions []VersionInfo `json:"versions"`
}

func getNativeDiskCachePath(tool, key string) string {
	safeTool := strings.NewReplacer("/", "_", ":", "_", "@", "_").Replace(tool)
	safeKey := strings.NewReplacer("/", "_", ":", "_", "@", "_", "|", "_").Replace(key)
	return filepath.Join(env.GetCacheDir(), "native", safeTool, safeKey+".json")
}

func readNativeDiskCache(tool, key string, versions *[]VersionInfo, ttl time.Duration) bool {
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

	if time.Since(wrapper.SavedAt) > ttl {
		return false
	}

	*versions = wrapper.Versions
	return true
}

func writeNativeDiskCache(tool, key string, versions []VersionInfo) {
	if pkgHttp.MockTransport != nil || len(versions) == 0 {
		return
	}
	p := getNativeDiskCachePath(tool, key)
	wrapper := nativeCacheWrapper{
		SavedAt:  time.Now(),
		Versions: versions,
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
