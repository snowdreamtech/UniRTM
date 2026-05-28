package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractTOMLFromBootstrap(t *testing.T) {
	content := `#!/bin/bash
# UNIRTM_TOOL_STUB:
# key = "value"
# another_key = 123
# :UNIRTM_TOOL_STUB
echo "hello"
`
	expected := "key = \"value\"\nanother_key = 123"
	assert.Equal(t, expected, extractTOMLFromBootstrap(content))

	content2 := `#!/bin/bash
# MISE_TOOL_STUB:
# key = "value"
# :MISE_TOOL_STUB
`
	expected2 := "key = \"value\""
	assert.Equal(t, expected2, extractTOMLFromBootstrap(content2))

	assert.Equal(t, "no markers here", extractTOMLFromBootstrap("no markers here"))
}
