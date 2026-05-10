# CLI 命令对标 mise 补全任务清单

> 参考规划：`.kiro/specs/unirtm/plans/cli_command_parity_plan.md`
> 对比来源：`mise --help` vs `unirtm --help`（2026-05-10 对比）

---

## Phase 1：核心补全（高频刚需）

- `[ ]` **`env` 命令语义升级**
  - `[ ]` 修改 `cmd/9.env.go`，增加 shell 变量导出模式
  - `[ ]` 支持 `eval "$(unirtm env)"` 用法（输出 `export PATH=...`）
  - `[ ]` 保留 `--json` 结构化输出
  - `[ ]` 保留 `-v` 兼容旧版打印行为
  - `[ ]` 编写测试 `cmd/9.env_test.go`
  - `[ ]` 原子化 Commit

- `[ ]` **`outdated` 命令**
  - `[ ]` 实现 `cmd/32.outdated.go`
  - `[ ]` 调用 backend `GetLatestVersion()` 与已安装版本对比
  - `[ ]` 支持表格 / JSON 输出
  - `[ ]` 支持 `unirtm outdated <tool>` 单工具检查
  - `[ ]` 编写测试
  - `[ ]` 原子化 Commit

- `[ ]` **`latest` 命令**
  - `[ ]` 实现 `cmd/33.latest.go`
  - `[ ]` 支持版本前缀过滤 `unirtm latest go 1.22`
  - `[ ]` 支持 `--json` 输出
  - `[ ]` 编写测试
  - `[ ]` 原子化 Commit

- `[ ]` **`list` 命令增强**
  - `[ ]` 修改 `cmd/11.list.go`，增加「激活状态」列
  - `[ ]` 区分 `installed` / `active` / `missing shim` 状态
  - `[ ]` 编写测试
  - `[ ]` 原子化 Commit

- `[ ]` **`set` / `unset` 命令**
  - `[ ]` 实现 `cmd/34.set.go`（含 `set` 和 `unset` 子命令）
  - `[ ]` 支持 `--global` 写入全局配置
  - `[ ]` 读写 unirtm.toml 中的 `[env]` 字段
  - `[ ]` 编写测试
  - `[ ]` 原子化 Commit

- `[ ]` **`tool` 命令**
  - `[ ]` 实现 `cmd/35.tool.go`
  - `[ ]` 显示：tool / backend / installed versions / active version / shim / config source
  - `[ ]` 支持 `--json` 输出
  - `[ ]` 编写测试
  - `[ ]` 原子化 Commit

---

## Phase 2：工具链管理完善

- `[ ]` **`bin-paths` 命令**
  - `[ ]` 实现 `cmd/36.bin-paths.go`
  - `[ ]` 列出所有激活 runtime 的 bin 目录路径
  - `[ ]` 原子化 Commit

- `[ ]` **`backends` 命令**
  - `[ ]` 实现 `cmd/37.backends.go`
  - `[ ]` 子命令：`ls`（列出）、`info <name>`（详情）
  - `[ ]` 显示每个 backend 的名称、状态、支持的工具数
  - `[ ]` 支持 `--json` 输出
  - `[ ]` 原子化 Commit

- `[ ]` **`registry` 命令**
  - `[ ]` 实现 `cmd/38.registry.go`
  - `[ ]` 列出所有注册工具（分页 / 过滤 `--search`）
  - `[ ]` 支持 `--json` 输出
  - `[ ]` 原子化 Commit

- `[ ]` **`tasks` 子命令组**
  - `[ ]` 实现 `cmd/39.tasks.go`
  - `[ ]` 子命令：`list`、`info <task>`、`deps`、`add <name>`、`edit <task>`
  - `[ ]` `tasks list` 显示任务名、来源文件、描述
  - `[ ]` `tasks deps` 显示任务依赖 DAG（文本格式）
  - `[ ]` 编写测试
  - `[ ]` 原子化 Commit

- `[ ]` **`fmt` 命令**
  - `[ ]` 实现 `cmd/40.fmt.go`
  - `[ ]` 格式化 unirtm.toml（键排序、对齐、统一缩进）
  - `[ ]` 支持 `--check` 模式（CI 用，仅检查不修改）
  - `[ ]` 原子化 Commit

- `[ ]` **`link` 命令**
  - `[ ]` 实现 `cmd/41.link.go`
  - `[ ]` 软链接已有工具路径进 UniRTM 管理体系
  - `[ ]` 写入安装记录至数据库
  - `[ ]` 原子化 Commit

- `[ ]` **`unuse` / `rm` / `remove` 命令**
  - `[ ]` 实现 `cmd/42.unuse.go`
  - `[ ]` 注册 `rm` 和 `remove` 为别名
  - `[ ]` 从 unirtm.toml 中删除工具条目（不删除已安装文件）
  - `[ ]` 原子化 Commit

---

## Phase 3：高级 / 实验性功能

- `[ ]` **`self-update` 命令**
  - `[ ]` 实现 `cmd/43.self-update.go`（参考 `.kiro/specs/unirtm/unirtm_selfupdate_plan.md`）
  - `[ ]` 支持 `--version <version>` 指定目标版本
  - `[ ]` 原子化 Commit

- `[ ]` **`implode` 命令**
  - `[ ]` 实现 `cmd/44.implode.go`
  - `[ ]` 需要用户二次确认（交互 prompt）
  - `[ ]` 支持 `--yes` / `-y` 跳过确认（脚本模式）
  - `[ ]` 清理：数据目录 + shims + 数据库 + 缓存
  - `[ ]` 原子化 Commit

- `[ ]` **`generate` 命令**
  - `[ ]` 实现 `cmd/45.generate.go`
  - `[ ]` 子命令：`github-action`、`pre-commit`、`shell-alias`
  - `[ ]` 输出生成的文件内容（支持 `--output` 指定路径）
  - `[ ]` 原子化 Commit

- `[ ]` **`en` 命令**
  - `[ ]` 实现 `cmd/46.en.go`
  - `[ ]` 在新 sub-shell 中运行 UniRTM 激活环境
  - `[ ]` 支持 `-- <cmd>` 直接执行命令
  - `[ ]` 原子化 Commit

- `[ ]` **`shell-alias` 命令**
  - `[ ]` 实现 `cmd/47.shell-alias.go`
  - `[ ]` 子命令：`list`、`add <alias>`、`remove <alias>`
  - `[ ]` 原子化 Commit

- `[ ]` **`install-into` 命令**
  - `[ ]` 实现 `cmd/48.install-into.go`
  - `[ ]` 安装工具到指定自定义路径
  - `[ ]` 原子化 Commit

- `[ ]` **`edit` 命令**
  - `[ ]` 实现 `cmd/49.edit.go`
  - `[ ]` 打开 `$EDITOR` / `$VISUAL` 编辑配置文件
  - `[ ]` 支持 `--global` 编辑全局配置
  - `[ ]` 原子化 Commit

- `[ ]` **`token` 命令**
  - `[ ]` 实现 `cmd/50.token.go`
  - `[ ]` 显示当前各 provider 使用的 token（掩码处理）
  - `[ ]` 支持 `unirtm token github` 指定 provider
  - `[ ]` 原子化 Commit

- `[ ]` **`mcp` 命令（实验性）**
  - `[ ]` 实现 `cmd/51.mcp.go`
  - `[ ]` 运行 MCP server（stdio 模式，供 AI 工具调用）
  - `[ ]` 暴露：install / list / outdated / tool 等工具为 MCP tools
  - `[ ]` 原子化 Commit

---

## 统计

| Phase | 命令数 | 状态 |
|-------|--------|------|
| Phase 1 — 核心补全 | 6 | 未开始 |
| Phase 2 — 管理完善 | 8 | 未开始 |
| Phase 3 — 高级功能 | 9 | 未开始 |
| **合计** | **23** | — |
