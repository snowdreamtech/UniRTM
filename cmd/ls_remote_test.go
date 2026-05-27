package cmd

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetLsRemoteBackendName(t *testing.T) {
	// Empty tool defaults
	lsRemoteBackend = ""
	assert.Equal(t, "github", getLsRemoteBackendName("golang.org/x/vuln/cmd/govulncheck"))
	
	lsRemoteBackend = "npm"
	assert.Equal(t, "npm", getLsRemoteBackendName("node"))
	lsRemoteBackend = ""
}
