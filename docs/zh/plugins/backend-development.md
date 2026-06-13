# 后端开发指南 (Backend Development)

UniRTM 的核心哲学是**极致的高性能与全平台的原生兼容**。传统基于 Bash 的工具管理器（如 asdf）在解析、下载、编译阶段往往充满各种兼容性问题且速度缓慢。

如果你需要一个不在官方支持列表中的工具，你可以为 UniRTM 编写自己的**后端插件 (Backend Plugin)**。

## 架构设计

UniRTM 插件并非基于 Bash 脚本，而是**原生编译的模块**。这允许它在 Windows、macOS 和 Linux 上以原生速度运行，并利用多线程并行能力进行下载和解压。

目前，开发后端插件的主要方式是编写 **Go 语言接口**。

## 核心接口 (The Backend Interface)

每个自定义工具后端都需要实现 UniRTM 暴露的 `ToolBackend` 接口：

```go
type ToolBackend interface {
    // 插件名称，例如 "nodejs"
    Name() string

    // 获取所有可用版本列表 (用于 unirtm install node@latest)
    ListRemoteVersions() ([]string, error)

    // 解析并补全下载信息，包括下载 URL 和 checksum
    ResolveVersion(version string) (*ResolvedTool, error)

    // 下载对应的分发包（tar.gz, zip 等）
    Download(resolved *ResolvedTool, destPath string) error

    // 解压并执行安装逻辑
    Install(downloadedPath string, installDir string) error
}
```

### 1. `ListRemoteVersions`

通常，你会在这里通过向 Github API、Node.js Dist API 等发送 HTTP 请求，获取最新的 Release Tag，并对它们进行语义化版本（SemVer）排序。

### 2. `ResolveVersion`

当用户指定了具体版本（如 `20.11.0`）后，你需要在这里根据用户当前的操作系统（GOOS）和架构（GOARCH）拼接出确切的下载地址（URL）。
如果官方支持校验和，你还应该在这里抓取它的 sha256 并一并返回，这将在写入 `unirtm.lock` 时被用到。

### 3. `Download`

UniRTM 提供了一套内置的高性能、支持断点续传且带有进度条的 `Downloader` 模块，大部分情况下你只需要调用 `unirtm.DefaultDownloader.Download(url, destPath)` 即可。

### 4. `Install`

针对下载下来的压缩包进行解压。如果是二进制分发版，解压到指定目录并建立 `bin/` 软链接即可。如果是源码包，你需要在这里调用 `make && make install`。

## 注册你的插件

当你开发完你的后端后，只需在主程序入口将其注册：

```go
package main

import (
    "github.com/snowdreamtech/UniRTM/core"
    "github.com/my-org/unirtm-plugin-rust"
)

func main() {
    core.RegisterBackend(rust.NewRustBackend())
    core.Execute()
}
```

> **注意：** 未来我们会支持基于 WebAssembly (WASM) 的独立插件系统，以允许开发者使用 Rust、Zig 或 TypeScript 编写在沙箱中运行的高性能插件，而无需重新编译 UniRTM 主程序。敬请期待！
