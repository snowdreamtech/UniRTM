# Data Model

No new internal data structures or entities are required. The implementation will solely rely on implementing the existing `Provider` interface.

```go
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
