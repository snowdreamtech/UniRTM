package backend

import (
	"context"
	"testing"
)

func TestBackends_ResolveVersion(t *testing.T) {
	ctx := context.Background()
	platform := Platform{OS: "linux", Arch: "amd64"}

	backends := []Backend{
		NewCabalBackend(),
		NewCargoBackend(),
		NewComposerBackend(),
		NewCondaBackend(),
		NewGemBackend(),
		NewGoBackend(),
		NewMavenBackend(),
		NewNpmBackend(),
		NewPypiBackend(),
		NewSpmBackend(),
		NewVfoxBackend(),
		NewZigBackend(),
		NewPipxBackend(),
		NewS3Backend(),
	}

	for _, b := range backends {
		_, _ = b.ResolveVersion(ctx, "invalid-tool", "1.0.0", platform)
		_, _ = b.ResolveVersion(ctx, "invalid-tool", "latest", platform)
	}
}
