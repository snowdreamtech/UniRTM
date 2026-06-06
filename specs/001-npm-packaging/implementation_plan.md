# npm Platform Packaging Implementation Plan

## Goal
>
>
## User Review Required
>
>
> [!IMPORTANT]
> 请确认以下决策：
>
> - 根包命名为 `@snowdreamtech/unirtm`。
> - 支持多个操作系统与架构（darwin, linux, windows 等共 15 个架构）。
>
> - 二进制不提交到 Git，由 CI 从 dist/ 动态生成。
>

## Open Questions
>
> [!WARNING]
>
>
>
> - 是否需要发布测试（beta）版本？
>
> - 根包的 `install.js` 异常情况（找不到对应二进制）处理逻辑。

## Proposed Changes

### [NEW] `package.json.tpl`

- npm 包模板，包含 `{{VERSION}}` 用于 CI 替换。

- 运行时平台检测并启动相应子包二进制。

### [NEW] `build.sh`

- 核心构建脚本，负责将二进制从 dist/ 复制到各自的 npm 子包中。

- 增加 `after.hooks` 触发 npm 的构建与发布流程。

## Verification Plan

### Automated Tests

- 执行 `npm pack --dry-run` 确保子包内包含正确的文件（二进制，README, LICENSE）。
- 使用 `shellcheck` 验证 `build.sh` 的正确性。

### Manual Verification

- 跨平台本地执行 `npm install`，验证委托包和相应架构子包是否正确解析，确保可以调用 `unirtm version`。
