# 任务运行器概览 (Task Runner)

UniRTM 内置了一个任务运行器，旨在全面替代传统的工具链封装器，如 `make`、`npm scripts` 或者 `just`。

通过将任务运行器直接深度集成到工具管理器中，UniRTM 能够保证所有任务都在你 `.unirtm.toml` 中精确定义的环境和工具链下执行，彻底解决“在我的机器上明明能跑”的经典难题。

## 为什么要使用 UniRTM 的任务运行器？

- **零外部依赖**：在 Windows 上再也不需要安装 `make` 环境，或者在一个纯 Python 项目中被迫安装 `npm` 来跑脚本。
- **跨语言无缝执行**：在同一个项目中轻松混合执行 Go、Python、Bash 和 Node 任务，没有任何割裂感。
- **依赖关系图**：任务可以声明它们所依赖的前置任务。UniRTM 会在执行目标任务前，按照拓扑顺序先执行所有的依赖任务。
- **上下文天然继承**：任务会自动继承你在 `.unirtm.toml` 文件 `[env]` 部分定义的环境变量，以及 `[tools]` 部分劫持的 `$PATH` 环境。

## 快速上手

任务通常定义在 `.unirtm.toml` 的 `[tasks]` 节点下。

```toml
[tasks.lint]
description = "执行代码规范检查"
run = "golangci-lint run ./..."

[tasks.test]
description = "运行单元测试"
run = "go test ./..."
depends = ["lint"]
```

使用 `unirtm run` 命令来执行任务：

```bash
$ unirtm run test
→ running task "lint"
✓ lint completed
→ running task "test"
✓ test completed
```

## 执行多个任务

你可以通过传入多个参数来顺序执行多个任务：

```bash
$ unirtm run lint test build
```

如果在执行链中任何一个任务失败，整个流程会立即停止并返回非零的退出码。

## 探索项目任务

新加入项目的开发者不再需要去阅读晦涩难懂的 `Makefile` 才能搞清楚项目有哪些可用的指令。他们只需要执行：

```bash
$ unirtm run
可用的任务列表:
  build      编译生产环境二进制文件
  lint       执行代码规范检查
  test       运行单元测试
  dev        启动本地开发服务器
```

了解更多高级的任务工作流，请参阅 [TOML 任务配置](./toml-tasks.md) 和 [文件型任务](./file-tasks.md)。
