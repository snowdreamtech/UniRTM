# Data Model: Plugin & Extension Configuration Isolation

**Feature**: `027-plugin-config-isolation`  
**Spec**: [spec.md](./spec.md)  

---

## 1. Core Go Structs

```go
package config

// PluginDeclaration represents an enabled plugin and its configuration.
type PluginDeclaration struct {
    // Version specifies the semantic version requirement (e.g. "1.0.0", ">=2.0").
    Version string `toml:"version,omitempty" yaml:"version,omitempty" json:"version,omitempty"`

    // Config holds arbitrary namespaced settings owned exclusively by this plugin.
    Config map[string]interface{} `toml:"config,omitempty" yaml:"config,omitempty" json:"config,omitempty"`

    // Source indicates where the config was resolved from: "inline", "sidecar", or "merged".
    Source string `toml:"-" yaml:"-" json:"source,omitempty"`

    // SidecarPath tracks the path to the external file if loaded via sidecar.
    SidecarPath string `toml:"-" yaml:"-" json:"sidecar_path,omitempty"`
}

// PluginManifest defines the extension metadata and contributed configuration schema.
type PluginManifest struct {
    Name        string                               `json:"name" yaml:"name"`
    Version     string                               `json:"version" yaml:"version"`
    Description string                               `json:"description" yaml:"description"`
    Contributes PluginContributes                    `json:"contributes" yaml:"contributes"`
}

type PluginContributes struct {
    Configuration map[string]ConfigPropertySchema    `json:"configuration" yaml:"configuration"`
}

type ConfigPropertySchema struct {
    Type        string        `json:"type" yaml:"type"`               // "string", "number", "boolean", "array", "object"
    Default     interface{}   `json:"default" yaml:"default"`
    Description string        `json:"description" yaml:"description"`
    Enum        []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
}
```

---

## 2. TOML Representation Schema

### Inline Configuration (`.unirtm.toml`)

```toml
[plugins]
"vendor/name" = "1.0.0"

[plugins."vendor/name".config]
key1 = "value"
key2 = true
key3 = 42
```

### Sidecar Configuration (`.unirtm/plugins/vendor-name.toml`)

```toml
# File: .unirtm/plugins/vendor-name.toml
# Automatically mapped to: [plugins."vendor/name".config]

key1 = "value"
key2 = true
key3 = 42
```
