# Internal Provider Interface

Since UniRTM is an extensible CLI tool without public programmatic API exposures outside its Go package, the "contract" for this feature is the internal `Provider` interface.

```go
// Provider defines the standard interface for all package manager backends in UniRTM.
type Provider interface {
 Name() string
 Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error
 PostInstall(ctx context.Context, tool string, installPath string, version string) error
 GenerateShims(tool string, installPath string, version string) (map[string]string, error)
 GetBinPaths(tool string, installPath string, version string) ([]string, error)
 GetEnvVars(tool string, installPath string, version string) (map[string]string, error)
 Uninstall(ctx context.Context, tool string, installPath string, version string) error
}
```

Any new backend (Composer, LuaRocks, Pub, Cabal) must implement these methods exactly, ensuring cross-platform isolation and accurate binary path resolution.
