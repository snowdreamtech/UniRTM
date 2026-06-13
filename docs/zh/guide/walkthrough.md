# 核心特性漫游 (Walkthrough)

欢迎来到 UniRTM 的完整实战漫游！本指南将带你一步步通过一个真实的场景，展示 UniRTM 如何用一个极致快速的原生二进制文件，完美替代众多传统的零散工具（例如 `asdf`，`direnv` 和 `make`）。

## 1. 初始化项目与工具链

假设你正在开始一个新的全栈项目。你的后端需要 Go 1.22，而你的前端需要 Node 20。

```bash
mkdir my-fullstack-app
cd my-fullstack-app
```

不要去手动安装这些工具，也不要污染你的系统全局 PATH，让 UniRTM 来管理它们：

```bash
unirtm use go@1.22 node@20
```

这个命令会在你的目录中创建一个 `.unirtm.toml` 配置文件：

```toml
[tools]
go = "1.22"
node = "20"
```

由于 UniRTM 会根据此文件动态调整你的环境，从这一刻起，只要你在这个目录下，这几个精确版本的工具就已经生效了！

## 2. 管理环境变量

你的应用不可避免地需要环境变量（例如数据库连接字符串）。你可以直接在 `.unirtm.toml` 中配置它们，或者从传统的 `.env` 文件中加载。

让我们直接使用 `.unirtm.toml` 来获得更集成的体验：

```toml
[env]
DATABASE_URL = "postgres://user:pass@localhost:5432/mydb"
NODE_ENV = "development"
# 环境变量甚至可以互相引用插值
API_ENDPOINT = "https://api.dev.example.com"
DEBUG_URL = "${API_ENDPOINT}/debug"
```

此后，每当你在该目录中运行命令或执行任务时，这些变量都会被安全、自动地注入——不再需要 `direnv`，也绝对不会污染你全局的 `.bashrc`。

## 3. 终极任务运行器 (Task Runner)

告别陈旧的 `Makefile`，也告别只能用于 JS 的 `package.json` 脚本。UniRTM 内置了一个跨语言、通用型的任务运行器。

在你的 `.unirtm.toml` 中添加以下内容：

```toml
[tasks.build]
description = "构建后端二进制文件与前端静态资源"
run = """
  go build -o server ./cmd/api
  npm run build
"""

[tasks.dev]
description = "启动开发服务器"
run = "go run ./cmd/api"
env = { PORT = "8080" }
```

现在你可以列出所有可用的任务：

```bash
$ unirtm tasks
build    构建后端二进制文件与前端静态资源
dev      启动开发服务器
```

然后一键执行：

```bash
unirtm run dev
```

**UniRTM 提供绝对的执行保证**：在这个任务执行期间，调用的 `go` 和 `npm` 绝对是你 `[tools]` 中固定版本的二进制文件，同时 `[env]` 块（以及特定任务的 `PORT`）中的环境变量会被严格限制在这个进程隔离沙箱内。

## 4. 临时隔离执行

有时候，你需要在一个完全隔离的环境中运行工具，而不希望修改你的项目配置或全局状态。比如，你想用 Python 3.9 测试一个脚本，但不想全局安装它。

使用 `unirtm exec`：

```bash
unirtm exec python@3.9 -- python test_script.py
```

UniRTM 会自动在后台安全地拉取 Python 3.9，计算并构建隔离的虚拟环境，执行你的脚本，随后在执行完毕时无痕销毁上下文。全局系统零污染。

## 总结

到这里，你已经成功掌握了：

- 使用 `unirtm use` 优雅地固定多语言的工具版本。
- 设置安全隔离的动态环境变量。
- 跨语言构建无缝流转的 `[tasks]` 管道。
- 使用 `exec` 进行零污染的临时执行测试。

UniRTM 致力于为你的软件供应链和日常开发带来绝对的秩序与极速。要想深入了解，请继续阅读详尽的 [CLI 命令手册](../cli/overview.md) 或是探索它的 [底层架构剖析](../advanced/architecture.md)。
