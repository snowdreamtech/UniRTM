// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestIsUniRTMBinary(t *testing.T) {
	assert.True(t, isUniRTMBinary("unirtm"))
	assert.True(t, isUniRTMBinary("unirtm.exe"))
	assert.True(t, isUniRTMBinary("/path/to/unirtm"))
	assert.True(t, isUniRTMBinary("unirtm-test"))
	assert.True(t, isUniRTMBinary("main"))
	assert.False(t, isUniRTMBinary("go"))
	assert.False(t, isUniRTMBinary("node"))
}

func TestGetBestEditorWithSource(t *testing.T) {
	t.Setenv("UNIRTM_EDITOR", "myeditor")

	editor, source := getBestEditorWithSource(nil)
	assert.Equal(t, "myeditor", editor)
	assert.Equal(t, "$UNIRTM_EDITOR", source)

	t.Setenv("UNIRTM_EDITOR", "")

	cfg := &config.Config{
		Settings: config.Settings{
			Editor: "cfg_editor",
		},
	}
	editor, source = getBestEditorWithSource(cfg)
	assert.Equal(t, "cfg_editor", editor)
	assert.Equal(t, "unirtm settings", source)
}
