# Plugin & Extension Configuration Guide

UniRTM provides a strict three-layer configuration namespace isolation model inspired by industry standards such as **VS Code Configuration Contributions** and **Cargo Metadata Namespaces**.

This architecture guarantees that third-party plugins and extensions can safely expand configuration without collisions, root table pollution, or cross-plugin overrides.

---

## 1. Architecture: Three-Layer Namespace Isolation

```
.unirtm.toml Root Structure
├── 1. Core Reserved Tables       -> [tools], [tasks], [format], [lint], [hooks], [env], [settings]
├── 2. Plugin Declarations        -> [plugins] (version constraints & enabled status)
└── 3. Isolated Plugin Spaces     -> [plugins."<plugin-id>".config] (exclusive entrypoint)
```

### Core Reserved Namespaces
The following top-level tables are reserved for UniRTM core features:
* `[tools]`: Tool version definitions.
* `[tasks]`: Automation task workflows.
* `[format]`: Code formatter rules.
* `[lint]`: Linter specifications.
* `[hooks]`: Lifecycle trigger hooks.
* `[env]`: Environment variables.
* `[settings]`: CLI runtime settings.
* `[plugins]`: Enabled plugin declarations.

Third-party plugins and CLI commands **cannot write new top-level tables** outside this reserved list.

---

## 2. Configuration Patterns

### Pattern A: Inline Configuration (`.unirtm.toml`)

For lightweight configuration, settings can be written directly inside `.unirtm.toml`:

```toml
[plugins]
"snowdreamtech/docker" = "1.0.0"
"community/deploy" = "0.2.1"

[plugins."snowdreamtech/docker".config]
compose_file = "docker-compose.yml"
port = 8080
auto_prune = true

[plugins."community/deploy".config]
cluster = "k8s-staging"
timeout = "45s"
```

### Pattern B: Sidecar Configuration (`.unirtm/plugins/*.toml`)

For complex plugins with dozens of configuration parameters, configuration can live in external sidecar files:

* **Project Sidecar**: `.unirtm/plugins/<plugin-id>.toml` (or nested `vendor/plugin.toml`)
* **Global Sidecar**: `~/.config/unirtm/plugins/<plugin-id>.toml`

Example `.unirtm/plugins/docker.toml`:

```toml
port = 8080
registry = "harbor.internal"
auto_prune = true
```

UniRTM automatically discovers and deep-merges sidecars at startup in `< 5ms`.

---

## 3. Precedence Hierarchy

When a configuration key is defined across multiple locations, UniRTM resolves conflicts according to the following precedence (highest to lowest):

1. **Inline Project Configuration** (`.unirtm.toml` under `[plugins."<id>".config]`) *(Wins conflicts)*
2. **Project Sidecar File** (`.unirtm/plugins/<id>.toml`)
3. **Global Sidecar File** (`~/.config/unirtm/plugins/<id>.toml`)

---

## 4. CLI Operations

### Writing Configuration
```bash
# Set inline plugin setting
unirtm config set plugins."snowdreamtech/docker".config.port 8080

# Set string setting
unirtm config set plugins."community/deploy".config.cluster staging
```

### Reading Configuration
```bash
# Query plugin configuration
unirtm config get plugins."snowdreamtech/docker".config.port

# Inspect all active plugins and configurations
unirtm config
```

---

## 5. Plugin Developer Guide: Manifest Schema Validation

Plugins can provide a manifest schema (`.manifest.json`) declaring types, descriptions, and allowed enums:

```json
{
  "name": "snowdreamtech/docker",
  "version": "1.0.0",
  "contributes": {
    "configuration": {
      "port": {
        "type": "integer",
        "description": "Port to bind container reverse proxy",
        "default": 8080
      },
      "driver": {
        "type": "string",
        "description": "Storage driver",
        "enum": ["overlay2", "btrfs"]
      }
    }
  }
}
```

When a manifest is present:
* `unirtm config set` strictly checks input types and enums before writing.
* Type mismatches or invalid enum values are immediately rejected with clear error messages.

---

## 6. IDE 智能感知与配置 Schema 规范体系

为了在现代 IDE（VS Code、Cursor、Zed、IntelliJ）中为开发者提供**实时语法高亮、参数自动补全、类型错误标红与悬停文档**，UniRTM 建立了双层 JSON Schema 规范体系：

```
UniRTM Schema 体系
├── 1. 核心工程规范 -> schemas/unirtm.schema.json (约束 .unirtm.toml 核心表与三层隔离)
└── 2. 插件元数据规范 -> schemas/plugin-manifest.schema.json (约束第三方插件 manifest)
```

### 1. 核心工程配置绑定 (`.unirtm.toml`)

在项目的 `.unirtm.toml` 头部添加声明：

```toml
#:schema ./schemas/unirtm.schema.json

[tools]
go = "1.24.0"
```

或在 IDE 配置 (`.vscode/settings.json`) 中绑定：

```json
{
  "evenBetterToml.schema.associations": {
    "*(.|)unirtm*.toml": "./schemas/unirtm.schema.json"
  }
}
```

**生效体验**：
* 键入 `[settings]` 时自动补全支持的参数及中文说明。
* 误在项目级定义 `settings.browser` 时，IDE 立即黄色/标红警告该属性属于仅全局偏好。
* 严禁第三方随意创建根表，非法根表会被 IDE 拦截标红。

---

### 2. 插件开发者指南：如何生成插件 Schema

第三方插件作者需要发布一个规范的 `<plugin-id>.manifest.json`，即可自动获得 IDE 级和 CLI 级的双重校验支持。

#### 插件清单格式模板
创建 `<plugin-id>.manifest.json`（可参考 [examples/plugins/docker/docker.manifest.json](../../examples/plugins/docker/docker.manifest.json)）：

```json
{
  "$schema": "https://unirtm.snowdream.tech/schemas/v1/plugin-manifest.schema.json",
  "name": "vendor/plugin-name",
  "version": "1.0.0",
  "description": "插件功能简述",
  "contributes": {
    "configuration": {
      "port": {
        "type": "integer",
        "description": "服务监听端口",
        "default": 8080
      },
      "mode": {
        "type": "string",
        "description": "运行模式",
        "enum": ["standalone", "cluster"],
        "default": "standalone"
      },
      "enable_metrics": {
        "type": "boolean",
        "description": "是否开启监控指标采集",
        "default": true
      }
    }
  }
}
```

#### 存放与发现路径
插件清单可放置在以下任一位置供 UniRTM 自动发现：
1. **项目插件目录**：`.unirtm/plugins/<plugin-id>.manifest.json` 或 `.unirtm/plugins/<slug>.manifest.json`
2. **全局插件目录**：`~/.config/unirtm/plugins/<plugin-id>.manifest.json`

当清单被加载后：
1. 开发者在 CLI 运行 `unirtm config set plugins."<plugin-id>".config.port abc` 时，会自动校验并提示 `"property 'port' must be an integer"`。
2. 结合 VS Code 的 JSON 校验器，IDE 会在编辑插件配置时自动提供智能提示和文档浮窗。
