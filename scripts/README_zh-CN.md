# 自动化脚本工具箱

[English](README.md) | [简体中文](README_zh-CN.md)

> [!NOTE]
> 此目录包含了项目自动化基础设施的实现，使用 POSIX shell 脚本 (`.sh`)，可在所有平台上运行，包括 Windows（通过 Git Bash 或 WSL）。

## 1. 设计与架构

### 概览

该组件提供了一套跨平台的脚本，用于管理开发生命周期，包括环境设置、依赖安装、代码检查、测试和部署。

- **可移植性**: 使用 POSIX 兼容的 shell 编写，确保在 Linux、macOS 和 CI 环境中保持一致。
- **健壮性**: 包含安全守卫、原子操作和标准化的错误处理。
- **Windows 优化**: 从 CMD 和 PowerShell 完全委托给核心逻辑。

### 架构

```text
[ 用户 / 开发者 / CI ]
        |
    [ Makefile ] (便捷入口点)
        |
        v
    [ scripts/*.sh ] (核心 POSIX 逻辑, SSoT)
    /    |      \
  /     |       \
  /      |        \
[lib/common.sh] [lib/langs/*.sh] [POSIX Shell 脚本]
(工具库/SSoT)   (语言模块)            ()
```

### 设计原则

- **POSIX 兼容**: 所有脚本遵循 POSIX shell 标准，确保跨平台兼容性（macOS/Linux/Windows Git Bash/WSL）。
- **可审计性**: 通过详细的日志和退出状态码确保所有决策可追溯。
- **幂等性**: 脚本可以安全地多次运行，无副作用。
- **快速失败**: 出错时立即退出并提供明确的诊断信息。
- **精简性**: 核心逻辑尽可能零依赖（仅依赖 `sh`, `sed`, `awk`）。

### 职责范围

- **环境调配**: 安装运行时工具链与质量工具 (`setup.sh`)。
- **依赖管理**: 标准化多语言栈的依赖安装流程 (`install.sh`)。
- **质量保证**: 编排代码检查、测试套件与安全审计 (`verify.sh`)。
- **生命周期自动化**: 自动化规范化提交、发布与变更日志归档。

## 2. 使用指南

### 前置条件

- **Linux/macOS**: 内置 POSIX shell
- **Windows**: Git Bash（随 [Git for Windows](https://git-scm.com/download/win) 安装）或 WSL
- **Make**: 可选，提供便捷的入口点

> [!TIP]
> Windows 用户：请参阅 [Windows Shell 迁移指南](../docs/migration/windows-shell-migration.md) 了解设置说明。

### 快速开始

```bash
# 1. 设置开发环境
sh scripts/setup.sh

# 2. 安装项目依赖
sh scripts/install.sh

# 3. 验证环境健康状况
sh scripts/verify.sh
```

### 脚本参考

| 脚本                   | 用途               | 关键模块                     |
| :--------------------- | :----------------- | :--------------------------- |
| `setup.sh`             | 安装系统级工具     | node, python, go, rust 等    |
| `install.sh`           | 安装项目依赖       | pnpm, pip, pre-commit        |
| `verify.sh`            | 完整项目验证       | 测试、代码检查、审计         |
| `lint.sh`              | 执行代码静态检查   | pre-commit, 自动修复         |
| `test.sh`              | 执行测试套件       | bats, pytest, vitest         |
| `audit.sh`             | 安全与脆弱性扫描   | gitleaks, osv-scanner        |
| `docs.sh`              | 文档网站管理       | vitepress                    |
| `init-project.sh`      | 修改模板品牌       | 占位符替换、git 初始化       |
| `archive-changelog.sh` | 归档旧版本         | 主版本更迭轮转               |

### 工作流模式

1. **项目初始化**: `setup.sh` → `install.sh` → `verify.sh`。
2. **日常开发**: 编码 → `lint.sh`。
3. **持续集成 (CI)**: `audit.sh`。

### 目录结构

- `scripts/`: 主自动化入口（POSIX shell 脚本）
- `scripts/lib/`: 内部核心库 (`common.sh`)
- `scripts/lib/langs/`: 模块化语言逻辑（由 `common.sh` 自动加载）

## 3. 运维指南

### 预部署检查清单

1. [ ] 运行 `sh scripts/check-env.sh` 确保运行时环境一致。
2. [ ] 运行 `sh scripts/verify.sh` 进行最终 QA 验收。
3. [ ] 运行 `sh scripts/audit.sh` 确保无敏感信息泄漏或严重漏洞。

### 性能考量

- **系统化面板**: 显示版本状态和安装耗时的综合摘要。
- **缓存利用**: Python 虚拟环境 (`.venv`) 与 `node_modules` 在不同运行间复用。

### 故障排除

- **问题**: 权限不足 (Permission Denied)。
  - **解决方案**: 运行 `chmod +x scripts/*.sh`。
- **问题**: Windows 脚本执行策略拦截。
  - **解决方案**: 运行 `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned`。
- **问题**: 下载过程中发生 404 或网络错误。
  - **诊断**: 检查 `GITHUB_PROXY` 是否可达。
  - **解决方案**: 在 `scripts/lib/common.sh` 或环境变量中配置 `GITHUB_PROXY`。

### 维护程序

- **工具更新**: 每周定期运行 `sh scripts/update.sh`。
- **缓存清理**: 运行 `sh scripts/cleanup.sh` 回收磁盘空间。

## 4. 安全考虑

### 安全模型

- **最小权限原则**: 脚本将工具安装至 `.venv` 或用户目录，无需 `sudo`。
- **完整性验证**: 关键安装逻辑包含对外部资源的校验。
- **密钥卫生**: 集成 Gitleaks 静态扫描，防止凭据意外泄漏。

### 最佳实践

| 维度         | 要求                   | 实现方式                       |
| :----------- | :--------------------- | :----------------------------- |
| 文件权限     | 脚本必须具备可执行权限 | `chmod 755 scripts/*.sh`       |
| 凭据完整性   | 禁止硬编码密钥         | 仅从 `.env` 或环境变量读取     |
| 代理处理     | 安全的下载网关         | 通过 TLS 访问 `GITHUB_PROXY`   |
| Windows 安全 | 签名或受限的执行策略   | 使用 `-ExecutionPolicy Bypass` |

## 5. 开发指南

### 代码组织

- 新逻辑必须以新的 `.sh` 文件形式添加至 `scripts/`。
- 共享工具函数必须放入 `scripts/lib/common.sh`。
- 函数内部变量必须使用 `local` 关键字进行私有化。

### 贡献要求

1. 所有新脚本必须包含 `main()` 函数并调用 `parse_common_args`。
2. 所有新脚本必须通过 `shellcheck --shell=sh` 检查。
3.
4. 保持英文与中文 README 版本的同步更新。

### 本地开发环境准备

1. 安装 ShellCheck: `brew install shellcheck`。
2. 安装 PSScriptAnalyzer (Windows 下): `Install-Module -Name PSScriptAnalyzer`。
3. 运行 `make lint` 验证合规性。

### 参考资料

- [POSIX 标准](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html)
- [ShellCheck Wiki](https://github.com/koalaman/shellcheck/wiki)
- [PowerShell 规范](https://docs.microsoft.com/en-us/powershell/scripting/developer/cmdlet/approved-verbs-for-windows-powershell-commands)
