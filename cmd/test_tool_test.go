package cmd

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestTestExecutable(t *testing.T) {
	err := testExecutable("/non/existent/path/to/executable", nil)
	assert.Error(t, err)
}
