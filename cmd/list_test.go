package cmd

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestFormatListSize(t *testing.T) {
	assert.Equal(t, "1.0 KB", formatListSize(1024))
	assert.Equal(t, "1.0 MB", formatListSize(1024*1024))
	assert.Equal(t, "1.0 GB", formatListSize(1024*1024*1024))
	assert.Equal(t, "0 B", formatListSize(0))
	assert.Equal(t, "500 B", formatListSize(500))
}
