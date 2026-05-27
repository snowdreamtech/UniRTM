package lockfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate_Errors(t *testing.T) {
	lf := &LockFile{
		Tools: map[string][]*ToolLockEntry{
			"empty": {},
			"bad_version": {
				{Version: ""},
				{Version: "1.0"},
				{Version: "1.0"}, // duplicate
			},
			"bad_platform": {
				{
					Version: "2.0",
					Platforms: map[string]*PlatformEntry{
						"invalid-key":   {},
						"linux-amd64":   nil,
						"windows-amd64": {Checksum: "invalid", Size: -1, URL: "ftp://bad"},
					},
				},
			},
		},
	}

	err := lf.Validate()
	assert.Error(t, err)

	ve, ok := err.(*ValidationError)
	assert.True(t, ok)

	errMsg := ve.Error()
	assert.Contains(t, errMsg, "tool \"empty\": empty entry list")
	assert.Contains(t, errMsg, "tool \"bad_version\": entry has empty version")
	assert.Contains(t, errMsg, "tool \"bad_version\": duplicate version \"1.0\"")
	assert.Contains(t, errMsg, "unknown platform key \"invalid-key\"")
	assert.Contains(t, errMsg, "nil entry")
	assert.Contains(t, errMsg, "checksum \"invalid\" must start with 'sha256:' or 'blake3:'")
	assert.Contains(t, errMsg, "size must be ≥ 0, got -1")
	assert.Contains(t, errMsg, "url \"ftp://bad\" does not look like a valid HTTP URL")
}

func TestCheckStrict_Errors(t *testing.T) {
	lf := &LockFile{
		Tools: map[string][]*ToolLockEntry{
			"go": {
				{
					Version: "1.20",
					Backend: "go",
					Platforms: map[string]*PlatformEntry{
						"linux-amd64": {},
					},
				},
			},
			"github:cli/cli": {
				{
					Version: "2.0",
					Backend: "github",
					Platforms: map[string]*PlatformEntry{
						"linux-amd64": {URL: ""},
					},
				},
			},
		},
	}

	// Missing tool
	err := lf.CheckStrict([]LockRequirement{
		{ToolKey: "missing", Version: "1.0"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no locked entry for tool=\"missing\" version=\"1.0\"")

	// Missing platform
	err = lf.CheckStrict([]LockRequirement{
		{ToolKey: "go", Version: "1.20", PlatformKey: "darwin-arm64"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no locked platform \"darwin-arm64\" for tool=\"go\"")

	// Missing URL for backend that requires it
	err = lf.CheckStrict([]LockRequirement{
		{ToolKey: "github:cli/cli", Version: "2.0", PlatformKey: "linux-amd64"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no locked URL for tool=\"github:cli/cli\"")

	// Valid URL for backend that doesn't require it
	err = lf.CheckStrict([]LockRequirement{
		{ToolKey: "go", Version: "1.20", PlatformKey: "linux-amd64"},
	})
	assert.NoError(t, err)
}
