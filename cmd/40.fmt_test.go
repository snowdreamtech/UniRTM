package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFmtCommandStructure(t *testing.T) {
	assert.Contains(t, fmtCmd.Use, "fmt")
	assert.NotNil(t, fmtCmd.RunE)
	assert.NotNil(t, fmtCmd.Flags().Lookup("check"))
}

func TestFormatTOML_Order(t *testing.T) {
	input := `
[settings]
lockfile = true

[env]
FOO = "bar"

[tools]
node = "22"
`
	out, err := config.FormatTOML(input)
	require.NoError(t, err)

	content := string(out)
	// [env] should appear before [tools] and [settings]
	envIdx := indexOfSection(content, "[env]")
	toolsIdx := indexOfSection(content, "[tools]")
	settingsIdx := indexOfSection(content, "[settings]")

	assert.Less(t, envIdx, toolsIdx, "[env] should come before [tools]")
	assert.Less(t, toolsIdx, settingsIdx, "[tools] should come before [settings]")
}

func TestFormatTOML_RoundTrip(t *testing.T) {
	input := `
[env]
NODE_ENV = "production"
`
	out, err := config.FormatTOML(input)
	require.NoError(t, err)
	assert.Contains(t, string(out), "NODE_ENV")
}

func TestFmtCmd_NoFile(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))
	defer os.Chdir(orig)

	// No config file → should not error, just warn.
	err := fmtCmd.RunE(fmtCmd, []string{})
	assert.NoError(t, err)
}

func TestFmtCmd_AlreadyFormatted(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))
	defer os.Chdir(orig)

	// Create a simple config.
	require.NoError(t, os.WriteFile(
		filepath.Join(tmp, "unirtm.toml"),
		[]byte("[env]\nFOO = \"bar\"\n"),
		0o644,
	))
	err := fmtCmd.RunE(fmtCmd, []string{})
	assert.NoError(t, err)
}

// indexOfSection returns the byte offset of [section] in s, or -1.
func indexOfSection(s, section string) int {
	idx := 0
	for i := 0; i+len(section) <= len(s); i++ {
		if s[i:i+len(section)] == section {
			return i
		}
		idx = i
	}
	_ = idx
	return -1
}
