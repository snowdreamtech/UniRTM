# 插件生态概览

虽然 UniRTM （通过内置的 Go 核心插件）原生支持 Node、Python 和 Go 等主流语言，但它不可能将世界上所有的开发工具都内置到二进制文件中。

为了支持无限广阔的开发工具生态，UniRTM **100% 向下兼容 `asdf` 的庞大插件生态系统**，并且原生支持直接从 **GitHub Releases** 提取原生二进制文件。

## 插件机制工作原理

当你请求一个非核心插件的工具时（例如 `kubectl`、`terraform`），UniRTM 会将安装逻辑委托给后端的插件处理。

### GitHub 后端（推荐）
如果你需要的工具在 GitHub Releases 上发布了编译好的二进制文件，你根本不需要任何插件。UniRTM 能够根据你的操作系统和 CPU 架构，原生地计算出应该下载哪个文件并自动解压配置。

```bash
# 直接从 github.com/cli/cli 拉取二进制文件
unirtm use github:cli/cli@latest
```

### ASDF 兼容后端
如果该工具需要复杂的源码编译（比如 Erlang）或者特定的构建标志，你可以直接利用成千上万现成的 `asdf` 插件。

```bash
unirtm plugin install erlang https://github.com/asdf-vm/asdf-erlang.git
unirtm use erlang@26.2.2
```

在幕后，UniRTM 会执行由 `asdf` 插件提供的标准 Shell 脚本（如 `bin/download`、`bin/install`、`bin/list-all`），但会接管生成的二进制文件，使用自身极速的 Go 执行引擎来进行版本切换和环境注入。

## 创建自定义插件

如果一个工具既不在 GitHub 上发布，也没有现成的 `asdf` 插件，你可以非常轻松地创建自己的插件。请参阅 [后端开发指南](./backend-development.md) 了解如何编写自定义插件。
