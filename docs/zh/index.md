---
layout: home

hero:
  name: "UniRTM"
  text: "通用运行时管理器"
  tagline: 快速、简单、跨平台的工具，统一管理您的开发工具、环境变量和任务。
  actions:
    - theme: brand
      text: 快速开始
      link: /zh/guide/introduction
    - theme: alt
      text: 在 GitHub 上查看
      link: https://github.com/snowdreamtech/UniRTM

features:
  - icon: 🛠️
    title: 开发工具管理
    details: 多语言工具管理器。完美替代 asdf、nvm、pyenv、rbenv 等工具，一站式管理所有运行时环境。
  - icon: 🌍
    title: 环境变量
    details: 无缝管理 `.env` 与系统环境变量。在切换项目目录时，自动加载和卸载对应的环境变量。
  - icon: ⚡
    title: 任务运行器
    details: 跨语言轻松执行各种任务，无需再依赖复杂的 Makefile 或 npm scripts。
---

<br>

<div align="center">
  <h2>核心理念：在编写代码前，让一切各就各位。</h2>
  <p style="max-width: 600px; margin: 0 auto; color: var(--vp-c-text-2);">
    就像专业厨师在烹饪前会准备好所有的食材和厨具（<em>mise en place</em>）一样，UniRTM 会在你写下第一行代码前准备好你的开发环境。它会自动为你安装正确的工具，加载正确的环境变量，并为你运行的命令配置好正确的任务。
  </p>
</div>

<br><br>

## 核心菜单：一个命令，搞定整个项目环境

### 🔪 01. 开发工具管理
安装项目工具，锁定版本，并在不同的项目目录之间平滑切换。再也不用去猜当前项目到底需要哪个版本的 Node 或者 Python。

```bash
$ unirtm use node@20 python@3.11 go@1.22
✓ wrote .unirtm.toml

$ unirtm install
✓ installed 3 tools
```

### 🫕 02. 环境变量
从 `.unirtm.toml`、`.env` 文件或者 Shell 脚本中加载项目专属的环境变量。告别混乱的全局 Bash 配置文件。

```bash
$ cat .env.local
DATABASE_URL=postgres://localhost/orders

$ unirtm env
export DATABASE_URL=postgres://localhost/orders
```

### 🍳 03. 任务运行器
在依赖的开发工具和环境变量旁定义构建、测试、代码检查和部署任务。全面替代复杂的 Makefile 和 npm scripts。

```bash
$ unirtm run test
→ lint · typecheck · unit
✓ 3 tasks complete

$ unirtm run deploy
✓ shipped
```

<br>

## 为企业级与极客而生

与那些由 Bash 或 Ruby 编写的传统遗留工具不同，UniRTM 为现代开发生态量身定制。

::: info 🚀 极致的性能表现
使用纯 **Go 语言** 编写，UniRTM 的执行耗时通常在毫秒级别。当你在终端中打开新标签页时，再也不用痛苦地等待缓慢的 Ruby shim 或复杂的 Bash 脚本初始化环境了。
:::

::: tip 🔒 安全与合规
内置集成 **Trivy**、**Gitleaks** 和 **Syft** 等业界标准安全工具。从项目落地的第一天起，就为你的软件供应链安全保驾护航。
:::

::: warning 💻 原生跨平台
从底层设计即考虑了跨平台的一致性。完美支持 **macOS**、**Linux** 和 **Windows**。无论你的团队成员使用什么操作系统，都能获得完全一致的开发体验。
:::

<br>

<div align="center">
  <h2>准备就绪，即刻开始</h2>
  <p><em>Allez,</em> 准备好你的工作台。</p>
  
  ```bash
  curl -sL https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/install.sh | bash
  ```
</div>
