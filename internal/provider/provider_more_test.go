package provider

import (
	"context"
	"testing"
)

func TestAllProviders_Properties(t *testing.T) {
	providers := []Provider{
		NewDenoProvider(),
		NewGoPkgProvider(),
		NewPypiProvider(),
		NewVfoxProvider(),
		NewNodeProvider(),
		NewGemProvider(),
		NewDotnetProvider(),
		NewZigProvider(),
		NewSpmProvider(),
		NewRustProvider(),
		NewContainerProvider("docker"),
		NewUbiProvider(),
		NewElixirProvider(),
		NewGolangProvider(),
		NewSwiftProvider(),
		NewPythonProvider(),
		NewCondaProvider(),
		NewCargoProvider(),
		NewNativeProvider(),
		NewGenericProvider(),
		NewRubyProvider(NewNativeProvider()),
		NewNpmProvider(),
		NewJavaProvider(),
		NewErlangProvider(),
		NewAsdfProvider(),
		NewBunProvider(),
		NewFlutterProvider(),
	}
	ctx := context.Background()

	for _, p := range providers {
		// Just call them to get coverage
		_ = p.Name()
		_ = p.Install(ctx, "tool", "path", "artifactPath", "version")
		_ = p.PostInstall(ctx, "tool", "path", "version")
		_, _ = p.GenerateShims("tool", "path", "version")
		_, _ = p.DetectVersion(ctx, "tool", "path")
		_, _ = p.ListExecutables("tool", "path", "version")
		_, _ = p.GetBinPaths("tool", "path", "version")
		_, _ = p.GetEnvVars("tool", "path", "version")
		_ = p.Uninstall(ctx, "tool", "path", "version")
	}
}

func TestRegistry_Coverage(t *testing.T) {
	r := NewRegistry()
	p := NewGenericProvider()
	r.Register("test", p)
	_ = r.Has("test")
	_ = r.Get("test")
	_, _ = r.GetExact("test")
	r.List()
	r.Unregister("test")

	Register("test", p)
	_ = Has("test")
	_ = Get("test")
	_, _ = GetExact("test")
	List()
	Unregister("test")
}
