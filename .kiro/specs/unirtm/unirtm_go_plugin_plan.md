# UniRTM 插件系统重构计划 (HashiCorp go-plugin)

## 背景与目标

当前 UniRTM 的插件系统（`internal/service/plugin.go`）使用了 Go 原生的 `plugin` 标准库（`buildmode=plugin`）。
这导致了严重的跨平台和兼容性缺陷：
1. **不支持 Windows**：完全打破了 UniRTM 跨平台工具管理器的定位。
2. **严苛的版本锁定**：主程序和插件的 Go 编译器版本、公共依赖版本必须 100% 一致。
3. **CGO 依赖**：破坏了 Go 原生跨平台静态编译的优势。

**目标**：引入 `github.com/hashicorp/go-plugin`，将插件架构从**同进程动态链接库（.so）**重构为**基于 gRPC 的独立子进程隔离模型**。

---

## 用户审查项

> [!IMPORTANT]
> 1. **通信协议选择**：`go-plugin` 支持 `net/rpc` 和 `gRPC` 两种模式。考虑到 `Backend` 和 `Provider` 的接口都使用了 `context.Context`，强烈建议使用 **gRPC**，因为它原生支持 Context 取消（Cancel）和超时传递。是否同意引入 gRPC / Protobuf 依赖？
> 2. **插件命名约定**：采用多进程模型后，插件将是独立的跨平台二进制文件。建议命名规范从 `*.so` 改为 `unirtm-plugin-<name>`（Windows 上是 `unirtm-plugin-<name>.exe`）。是否同意此约定？

---

## 架构对比

### 现有架构（原生 plugin）
- 主程序通过 `dlopen` 加载 `.so`。
- 通过 `Lookup("Plugin")` 寻找暴露的变量。
- 两者运行在同一个内存空间。崩溃会连累主程序。

### 新架构（HashiCorp go-plugin）
- 插件是一个普通的 Go `main` 包编译出的独立二进制文件。
- 主程序（Client）启动插件二进制作为子进程（Server）。
- 主程序和插件之间通过本地 Socket 或 Stdin/Stdout 建立 gRPC 连接。
- 崩溃被完全隔离。

---

## 改造范围与阶段

### 阶段 1：定义 gRPC 接口 (Protobuf)

由于 `Backend` 和 `Provider` 的接口需要跨进程调用，我们需要将它们定义为 `.proto` 文件，并生成 gRPC 代码。

**[NEW] `api/plugin/v1/backend.proto`**
- 对应 `internal/backend.Backend` 接口
- 定义 `ListVersions`, `ResolveVersion`, `GetDownloadInfo` 的 RPC 服务和 Message。

**[NEW] `api/plugin/v1/provider.proto`**
- 对应 `internal/provider.Provider` 接口
- 定义 `Install`, `PostInstall`, `GenerateShims` 等 RPC 服务和 Message。

> 备注：我们会在代码库中引入 `protoc` 生成的 go 代码，或者手写轻量级的 `net/rpc` 封装（如果不愿意引入 protobuf 依赖的话，我们可以先用 `net/rpc` 作为简化版）。

### 阶段 2：实现 HashiCorp Plugin 封装层

在 `internal/plugin` 目录下实现 `plugin.Plugin` 接口（适配器）。

**[NEW] `internal/plugin/backend_grpc.go`**
- `GRPCBackendClient`：实现 `backend.Backend` 接口，内部将方法调用转换为 gRPC 调用。
- `GRPCBackendServer`：将 gRPC 请求反向路由到真实的 Go `backend.Backend` 实现。

**[NEW] `internal/plugin/provider_grpc.go`**
- `GRPCProviderClient`：实现 `provider.Provider` 接口，内部转换 gRPC。
- `GRPCProviderServer`：路由到真实的 `provider.Provider` 实现。

**[NEW] `internal/plugin/shared.go`**
- 定义 `HandshakeConfig`（包含 Magic Cookie 避免非插件进程被误执行）。
- 定义 `PluginMap`。

### 阶段 3：重构 PluginManager

**[MODIFY] `internal/service/plugin.go`**
- 移除对 `plugin.Open()` 的调用。
- 引入 `plugin.NewClient(&plugin.ClientConfig{...})` 启动子进程。
- **生命周期管理**：进程退出时需要调用 `client.Kill()` 释放子进程资源（通常绑定在 context 或 defer 中）。
- **插件发现**：扫描 `~/.local/share/unirtm/plugins/` 下以 `unirtm-plugin-*` 命名的可执行文件。

**[MODIFY] `cmd/24.plugin.go`**
- 适配 `plugin install` 逻辑：拷贝二进制文件并赋予可执行权限，而非 `.so`。

### 阶段 4：改造/提供插件示例

**[NEW] `examples/plugin-go/main.go`**
- 不再使用 `buildmode=plugin`。
- 修改为一个标准的 `main` 包，调用 `plugin.Serve()`：
```go
package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/snowdreamtech/unirtm/internal/plugin"
)

func main() {
    // 实例化具体的 Backend/Provider
	myBackend := &MyCustomBackend{}
	
    // 启动 HashiCorp plugin server
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"backend": &plugin.GRPCBackendPlugin{Impl: myBackend},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
```

---

## 验证计划

1. **跨平台编译验证**：在 macOS 上交叉编译出 Windows 版本的 `unirtm.exe` 和 `unirtm-plugin-dummy.exe`，确保插件能够被正常加载调用。
2. **版本解耦验证**：用 Go 1.22 编译主程序，用 Go 1.20 编译插件，验证能够正常通信。
3. **单元测试**：使用 Mock 进程或同一进程内的内联网络连接，测试 RPC 调用的序列化/反序列化。
4. **集成测试**：编写包含超时中止、异常崩溃恢复的测试用例，确保子进程 Panic 时，主程序能捕获明确的 RPC Error 而不是一起崩溃。
