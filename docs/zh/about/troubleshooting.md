# 故障排查 (Troubleshooting)

如果在使用 UniRTM 的过程中遇到了问题，请不要惊慌，本文档汇总了绝大部分常见问题及其快速修复方案。

> [!TIP]
> 遇到任何问题时的第一步：永远尝试运行 `unirtm doctor`。这是一个内置的诊断工具，会自动为你扫描系统环境、Shell 挂载状态以及依赖完整性，并给出相应的修复建议。

## 1. 切换目录后，工具版本没有改变

**现象：** 进入含有 `.unirtm.toml` 的项目目录后，运行 `node -v` 发现使用的仍然是系统全局的 Node.js，而不是项目锁定的版本。

**原因与修复：**
这是因为 UniRTM 的核心“动态 PATH 拦截”钩子未能正确挂载到你的终端（Shell）会话中。
请检查你的 `~/.bashrc` 或 `~/.zshrc`（或对应 Shell 的配置文件），确保其中包含以下代码：

```bash
eval "$(unirtm activate zsh)"  # 对于 zsh
# 或者
eval "$(unirtm activate bash)" # 对于 bash
```

添加后，别忘了运行 `source ~/.zshrc` 或重启终端。

## 2. 提示 "API Rate Limit Exceeded"

**现象：** 在执行 `unirtm install` 下载工具时，遇到类似 "rate limit exceeded" 或 "403 Forbidden" 的网络错误。

**原因与修复：**
在默认状态下，向 GitHub API 请求最新版本号时，未认证的请求很容易触发流控限制。
你可以生成一个 [GitHub Personal Access Token (PAT)](https://github.com/settings/tokens)（无需任何特殊权限，仅需公共读取），并将其导出为环境变量：

```bash
export GITHUB_TOKEN="你的_PAT_字符串"
```

UniRTM 会自动读取此 Token 以绕过严格的速率限制。

## 3. 安装工具时报 "Permission Denied" (权限拒绝)

**现象：** 使用 `unirtm install` 安装全局包时，系统提示没有权限。

**原因与修复：**
UniRTM **绝不需要** `sudo` 权限。所有的工具链、缓存和运行时都会严格隔离安装在当前用户的 `~/.local/share/unirtm` 目录下。
如果你遇到了权限错误，通常是因为你之前误用了 `sudo unirtm` 导致部分目录的所有者变成了 `root`。
**修复方案：** 修复目录所有权：

```bash
sudo chown -R $USER:$USER ~/.local/share/unirtm ~/.config/unirtm
```

## 4. `.unirtm.toml` 配置似乎被完全忽略了

**现象：** 当前目录明明有 `.unirtm.toml` 且写了 `node = "20"`，但是无论怎么重启终端，UniRTM 都不去读取它。

**原因与修复：**
出于**绝对的安全性考虑**，UniRTM 引入了“信任机制 (Trust System)”。如果一个配置文件的路径没有被标记为可信，UniRTM 将拒绝执行其中的任何环境修改指令，以防止恶意脚本的注入。
**修复方案：** 在该目录下主动运行一次：

```bash
unirtm trust
```

## 5. IDE (Cursor / Windsurf) 无法加载 `.agent` 规则

**现象：** 打开项目后，IDE 中的 AI 助手似乎没有读取到 `UniRTM` 接管的 SpecKit 上下文规则。

**原因与修复：**
UniRTM 的 `.agent` 规则作为项目的 Single Source of Truth，需要通过命令主动注入给 IDE，或者你的 IDE 需要具备动态读取外置规则的能力。
**修复方案：**
尝试在终端主动触发项目的分析和规划动作：

```bash
unirtm run speckit.plan
```

这会强制唤醒后端 Agent 流程，并更新当前目录下的 `.claude` / `.cursorrules` 上下文投影。

## 6. 报告问题

如果以上的排查方案都未能解决你的问题，请带上 `unirtm doctor` 和 `unirtm env` 的输出日志，前往我们的 [GitHub Issues](https://github.com/snowdreamtech/UniRTM/issues) 页面提交一个 Bug 报告。
