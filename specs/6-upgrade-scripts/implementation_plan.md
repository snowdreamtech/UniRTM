# 跨平台升级脚本改进计划（仅 `install.sh` 与 `install.ps1`）

## Goal
- 为 `unirtm self-up` 提供独立、可直接下载执行的 **POSIX 脚本** `install.sh` 与 **PowerShell 脚本** `install.ps1`，二者相互独立。
- 脚本需自行 **下载二进制、校验、安装、回滚**，不依赖项目内部其他脚本或 `install.bat` 包装器。
- 改进版本号获取方式，确保日志中只输出干净的版本号字符串（不受 ASCII‑art 干扰）。
- 满足项目的 **网络操作**（重试、代理） 与 **安全**（校验和） 要求，同步遵循 `.agent/rules/01-general.md` 中的跨平台兼容规范。

## User Review Required
> [!IMPORTANT]
> 请确认以下关键决策是否符合预期：
> - 默认安装目录采用 XDG/AppData 风格（Linux/macOS 为 `~/.local/bin`，Windows 为 `$HOME\bin` 或管理员时 `C:\Program Files\UniRTM`）。
> - 是否需要 `--quiet` 与 `--log-file` 选项的实现细节。
> - 是否接受在 CI 中使用 `bats` 与 `Pester` 进行单元测试的方案。

## Open Questions
> [!WARNING]
> - 是否需要在脚本中支持自定义代理变量名称（如 `UNIRT_GITHUB_PROXY`）？
> - 是否需要在 `install.ps1` 中提供 `-Destination` 参数的别名 `-InstallDir`？

## Proposed Changes
---
### install.sh
- 将默认安装目录从 `~/.unirtm/bin` 改为 `~/.local/bin`（符合 XDG 标准）。
- 将临时目录管理抽离至 `main()`，统一 `trap` 删除，防止子函数提前清理。
- 参数解析保持 `--version|-v`、`--install-dir`、`--no-proxy`、`--help`，并在未知参数时 `warn`。
- 保持已有的 `curl` 重试、代理、checksum 验证、回滚逻辑。
- 添加 `--quiet` 与 `--log-file` 支持（日志写入文件，默认仅错误输出）。

### install.ps1
- 完全独立实现，与 `install.sh` 功能等价。
- 使用 `[System.IO.Path]::GetTempFileName()` 生成临时文件并在脚本结束时清理。
- `Invoke-WebRequest` 带重试 (`-RetryCount`、`-RetryIntervalSec`)，支持 `GITHUB_PROXY` 前缀。
- 使用 `Get-FileHash -Algorithm SHA256` 验证 checksum。
- 实现 `-Quiet`、`-LogFile` 参数，默认将错误写入 `stderr`。
- 安装目录默认 `$HOME\bin`，管理员时使用 `C:\Program Files\UniRTM`。

---
## Verification Plan
### Automated Tests
- **Bats**：在 Linux/macOS CI 运行 `install.sh`，覆盖正常下载、校验成功、回滚路径。
- **Pester**：在 Windows CI 运行 `install.ps1`，验证相同流程。

### Manual Verification
- 在本地 macOS、Linux（Docker）以及 Windows（PowerShell）手动执行脚本，检查二进制是否可运行且版本号输出正确。
- 确认 `PATH` 提示信息在未加入路径时正确显示。

---
*请在审阅后确认是否接受此实现计划，或提供进一步需求（如自定义下载 URL、日志路径等）。*
