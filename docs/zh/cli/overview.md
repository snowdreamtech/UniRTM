# 命令概览 (CLI Overview)

UniRTM 提供了一个丰富、详尽的 CLI 命令行接口。以下是所有可用命令的完整目录。您随时可以在终端中运行 `unirtm help [command]` 来查看特定标志和详细用法。

## 工具管理命令 (Tool Management)

<details open>
<summary><code>unirtm install &lt;tool@version&gt;</code></summary>

下载并安装特定版本的工具。

- **示例**: `unirtm install node@20.14.0`
- **行为**: 如果未指定版本，UniRTM 会读取当前目录的 `.unirtm.toml`，并安装其中定义的所有缺失工具。

</details>

<details open>
<summary><code>unirtm uninstall &lt;tool@version&gt;</code></summary>

从本地缓存中彻底移除已安装的工具版本。

- **示例**: `unirtm uninstall go@1.22.0`

</details>

<details open>
<summary><code>unirtm use &lt;tool@version&gt;</code></summary>

将特定工具版本固定到当前目录的 `.unirtm.toml` 文件中。

- **示例**: `unirtm use python@3.12`
- **行为**: 创建或更新 `.unirtm.toml`。如果该工具尚未在本地缓存中，会自动触发安装流程。

</details>

<details open>
<summary><code>unirtm current</code></summary>

显示当前目录上下文正在激活的工具版本。

- **示例**: `unirtm current`
- **输出**: 列出当前激活的工具、解析后的具体版本，以及强制使用该版本的配置文件路径。

</details>

<details open>
<summary><code>unirtm ls</code></summary>

列出您本地机器上已安装的所有工具及其版本。

- **示例**: `unirtm ls`

</details>

<details open>
<summary><code>unirtm ls-remote &lt;tool&gt;</code></summary>

查询远程注册表，列出特定工具的所有可安装版本。

- **示例**: `unirtm ls-remote java`

</details>

<details open>
<summary><code>unirtm outdated</code></summary>

检查您的 `.unirtm.toml`，并报告远程注册表中是否为您固定的工具发布了较新的版本。
</details>

<details open>
<summary><code>unirtm upgrade &lt;tool&gt;</code></summary>

根据您配置的语义化版本约束 (SemVer constraints)，将 `.unirtm.toml` 中指定的工具升级到最新兼容版本。
</details>

<details open>
<summary><code>unirtm bin-paths</code></summary>

输出所有当前激活工具的 `bin/` 目录的绝对路径。此命令主要由内部逻辑用于动态组装 `PATH`。
</details>

---

## 执行命令 (Execution Commands)

<details open>
<summary><code>unirtm run &lt;task&gt;</code></summary>

执行在 `.unirtm.toml` 文件中定义的命名任务。

- **示例**: `unirtm run build`

</details>

<details open>
<summary><code>unirtm exec &lt;tool&gt; -- &lt;command&gt;</code></summary>

在特定工具的上下文中执行命令，而不修改全局的 `PATH`。

- **示例**: `unirtm exec node@18 -- npm run build`
- **行为**: 这在需要使用特定运行时版本执行一次性脚本，同时又不想影响当前 Shell 环境时极其有用。

</details>

---

## 环境变量命令 (Environment Commands)

<details open>
<summary><code>unirtm env</code></summary>

输出当前目录上下文中求值计算后的环境变量及 `PATH` 修改指令。

- **示例**: `eval "$(unirtm env)"` (这就是 Shell 集成应用环境变量的底层机制)。

</details>

---

## 核心与诊断命令 (Core & Diagnostic Commands)

<details open>
<summary><code>unirtm doctor</code></summary>

对您的环境执行全面的诊断检查。它会验证文件权限、远程注册表的网络连接状态，并确保 `.unirtm.toml` 的语法有效。
</details>

<details open>
<summary><code>unirtm completion &lt;shell&gt;</code></summary>

为 `bash`, `zsh`, `fish` 或 `powershell` 生成 Shell 自动补全脚本。
</details>

<details open>
<summary><code>unirtm config</code></summary>

显示全局的 UniRTM 配置（如：缓存目录位置、最大并发下载限制等）。
</details>

<details open>
<summary><code>unirtm plugin</code></summary>

管理自定义后端插件。支持诸如 `unirtm plugin add`、`unirtm plugin list` 以及 `unirtm plugin remove` 等子命令。
</details>
