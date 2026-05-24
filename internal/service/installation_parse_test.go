package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseToolSpec(t *testing.T) {
	im := &InstallationManager{}

	tests := []struct {
		name         string
		spec         string
		wantBackend  string
		wantTool     string
		wantVersion  string
		wantExplicit bool
	}{
		{"npm package", "npm:@commitlint/cli", "npm", "@commitlint/cli", "latest", false},
		{"npm package with version", "npm:@commitlint/cli@20.5.3", "npm", "@commitlint/cli", "20.5.3", true},
		{"go package", "go:github.com/foo/bar@v1.0.0", "go-pkg", "github.com/foo/bar", "v1.0.0", true},
		{"github repo", "github:foo/bar", "github", "foo/bar", "latest", false},
		{"plain github repo auto-detected", "foo/bar", "github", "foo/bar", "latest", false},
		{"plain tool auto-detected", "ripgrep", "asdf", "ripgrep", "latest", false},
		{"native tool auto-detected", "python", "native", "python", "latest", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBackend, gotTool, gotVersion, gotExplicit := im.ParseToolSpec(tt.spec)
			assert.Equal(t, tt.wantBackend, gotBackend)
			assert.Equal(t, tt.wantTool, gotTool)
			assert.Equal(t, tt.wantVersion, gotVersion)
			assert.Equal(t, tt.wantExplicit, gotExplicit)
		})
	}
}
