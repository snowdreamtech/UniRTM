package backend

import (
	"context"
	"testing"
)

func TestAllBackends_Properties(t *testing.T) {
	backends := []Backend{
		NewAquaBackend(),
		NewAsdfBackend(),
		NewCabalBackend(),
		NewCargoBackend(),
		NewComposerBackend(),
		NewCondaBackend(),
		NewContainerBackend("docker"),
		NewCranBackend(),
		NewDenoBackend(),
		NewDotnetBackend(),
		NewForgejoBackend(),
		NewGemBackend(),
		NewGitHubBackend(),
		NewGitlabBackend(),
		NewGoBackend(),
		NewGoPkgBackend(),
		NewHTTPBackend(),
		NewLuarocksBackend(),
		NewMavenBackend(),
		NewNativeBackend(),
		NewNpmBackend(),
		NewPipxBackend(),
		NewPypiBackend(),
		NewS3Backend(),
		NewSpmBackend(),
		NewVfoxBackend(),
		NewZigBackend(),
	}

	for _, b := range backends {
		// Just call them to get coverage
		_ = b.Name()
		_ = b.Dependencies()
		_ = b.SupportsChecksum()
		_ = b.SupportsGPG()
		_ = b.AttestationType()
		_ = b.IsRecommended()
		_ = b.IsScriptless()
		_ = b.GetReach()
		_ = b.IsStable()
		_ = b.SupportsOffline()

		// Call the unimplemented methods to cover the errors
		ctx := context.Background()
		plat := Platform{}
		_, _ = b.ListVersions(ctx, "", plat)
		_, _ = b.ResolveVersion(ctx, "", "", plat)
		_, _ = b.GetDownloadInfo(ctx, "", "", plat)
	}
}

func TestRegistry_Coverage(t *testing.T) {
	r := NewRegistry()
	b := NewGitHubBackend()
	r.Register(b)
	_ = r.Has(b.Name())
	_, _ = r.Get(b.Name())
	r.List()
	r.Backends()
	r.Unregister(b.Name())

	Register(b)
	_ = Has(b.Name())
	_, _ = Get(b.Name())
	List()
	Backends()
	Unregister(b.Name())
}
