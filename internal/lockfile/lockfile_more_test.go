// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package lockfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLockFile_More_LoadSave(t *testing.T) {
	// 1. Load malformed file
	malformed := filepath.Join(t.TempDir(), "malformed.toml")
	os.WriteFile(malformed, []byte("bad toml [= "), 0644)
	_, err := Load(malformed)
	assert.Error(t, err)

	// 2. Save unwritable
	unwritableDir := filepath.Join(t.TempDir(), "unwritable")
	os.Mkdir(unwritableDir, 0444)
	lf2 := New(filepath.Join(unwritableDir, "test.toml"))
	err = lf2.Save()
	assert.Error(t, err)
}

func TestLockFile_More_Get(t *testing.T) {
	lf := New("dummy")

	e := lf.GetEntry("notfound", "1.0")
	assert.Nil(t, e)

	p := lf.GetPlatform("notfound", "1.0", "linux/amd64")
	assert.Nil(t, p)

	lf.UpsertEntry("found", &ToolLockEntry{Version: "1.0"})
	p2 := lf.GetPlatform("found", "1.0", "linux/amd64")
	assert.Nil(t, p2)
}

func TestLockFile_More_Upsert(t *testing.T) {
	lf := New("dummy")

	// empty name
	lf.UpsertEntry("", &ToolLockEntry{Version: "1.0"})
	assert.NotNil(t, lf.GetEntry("", "1.0"))

	// missing entry for platform -> inserts the platform to a non-existent entry? Let's just create it first.
	lf.UpsertEntry("missing", &ToolLockEntry{Version: "1.0"})
	lf.UpsertPlatform("missing", "1.0", "linux-amd64", &PlatformEntry{Checksum: "000"})
	assert.NotNil(t, lf.GetPlatform("missing", "1.0", "linux-amd64"))

	// valid upsert twice to trigger update
	lf.UpsertEntry("foo", &ToolLockEntry{Version: "1.0"})
	e := lf.GetEntry("foo", "1.0")
	assert.Equal(t, "1.0", e.Version)

	lf.UpsertPlatform("foo", "1.0", "linux-amd64", &PlatformEntry{Checksum: "123"})
	p := lf.GetPlatform("foo", "1.0", "linux-amd64")
	assert.Equal(t, "123", p.Checksum)
}

func TestPlatformKey_More(t *testing.T) {
	// ParsePlatformKeys edge cases
	keys, err := ParsePlatformKeys("linux-amd64,macos-arm64")
	assert.NoError(t, err)
	assert.ElementsMatch(t, keys, []string{"linux-amd64", "macos-arm64"})

	keys, err = ParsePlatformKeys("")
	assert.NoError(t, err)
	assert.Empty(t, keys)
}
