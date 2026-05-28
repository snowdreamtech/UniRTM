package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFSToolName(t *testing.T) {
	assert.Equal(t, "npm-node", getFSToolName("node", "npm"))
	assert.Equal(t, "github-cli", getFSToolName("cli", "github"))
	assert.Equal(t, "go-golang.org-x-vuln-cmd-govulncheck", getFSToolName("golang.org/x/vuln/cmd/govulncheck", "go"))
	assert.Equal(t, "go_install-golang.org-x-vuln-cmd-govulncheck", getFSToolName("golang.org/x/vuln/cmd/govulncheck", "go_install"))
}
