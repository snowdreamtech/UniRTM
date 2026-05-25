# Implementation Plan: npm Platform Packaging

**Branch**: `1-npm-packaging` | **Date**: 2026-05-22 | **Spec**: spec.md
**Input**: npm 跨平台二进制包发布需求

## Summary

为 UniRTM 项目建立 npm 发布体系，采用"Root 委托 + 平台子包"模式，将 GoReleaser 构建的 15 个平台二进制文件封装为 npm 可发布格式，通过 `optionalDependencies` + `os`/`cpu` 字段实现安装时自动平台适配。

## Technical Context

**Language/Version**: Go（二进制源）+ Node.js ≥ 18（npm 委托层）
**Primary Dependencies**: GoReleaser（构建）、npm（发布）
**Storage**: N/A
**Testing**: shellcheck、shfmt、npm pack --dry-run
**Target Platform**: darwin (arm64/x64), linux (x64/arm64/ia32/arm/arm-5/arm-6/loong64/ppc64le/riscv64/s390x), windows (x64/arm64/ia32)
**Project Type**: npm 发布包（非 Web 应用）
**Performance Goals**: 安装时间 < 30s，二进制压缩后 ~22MB/平台
**Constraints**: 二进制不提交到 Git，由 CI 动态生成；Node.js 层仅作委托，不引入运行时依赖

## Project Structure

### Documentation (this feature)

```text
specs/1-npm-packaging/
├── plan.md              # 本文件
└── tasks.md             # 任务清单
```

### Source Code (repository root)

```text
npm/                                  # npm 发布目录
├── scripts/
│   └── build.sh                      # 核心构建脚本（从 dist/ 复制二进制）
├── unirtm/                           # 根包 @snowdreamtech/unirtm
│   ├── package.json.tpl              # 版本模板（{{VERSION}} 占位符）
│   ├── install.js                    # 平台检测与委托脚本
│   ├── LICENSE
│   ├── README.md
│   └── README_zh-CN.md
├── unirtm-darwin-arm64/              # 平台子包（共 15 个）
│   └── package.json.tpl
├── unirtm-darwin-x64/
│   └── package.json.tpl
├── unirtm-linux-x64/
│   └── package.json.tpl
├── unirtm-linux-arm64/
│   └── package.json.tpl
├── unirtm-linux-ia32/
│   └── package.json.tpl
├── unirtm-linux-arm/
│   └── package.json.tpl
├── unirtm-linux-arm-5/
│   └── package.json.tpl
├── unirtm-linux-arm-6/
│   └── package.json.tpl
├── unirtm-linux-loong64/
│   └── package.json.tpl
├── unirtm-linux-ppc64le/
│   └── package.json.tpl
├── unirtm-linux-riscv64/
│   └── package.json.tpl
├── unirtm-linux-s390x/
│   └── package.json.tpl
├── unirtm-windows-x64/
│   └── package.json.tpl
├── unirtm-windows-arm64/
│   └── package.json.tpl
└── unirtm-windows-ia32/
    └── package.json.tpl

.goreleaser.yaml                      # 新增 after.hooks 触发 build.sh
.gitignore                            # 新增 npm/*/bin/ 排除规则
```

## Design Decisions

| 决策点 | 选择 | 原因 |
|---|---|---|
| npm scope | `@snowdreamtech` | 与现有 npm 账号一致 |
| ARM 细分 | 保留独立子包（arm-5/arm-6/arm） | 嵌入式场景需精确匹配 |
| 子包内容 | `bin/` + `LICENSE` + `README.md` + `README_zh-CN.md` | 完整文档随包分发 |
| 触发时机 | GoReleaser `after.hooks` | 保证 dist/ 已完整生成 |
| 二进制 Git 管理 | 不提交，CI 时从 dist/ 动态复制 | 避免大文件入仓库 |
| 版本注入 | `{{VERSION}}` 占位符替换 | 单一来源（VERSION 文件）|
| repository.url | `git+https://...git` 格式 | npm 标准格式，避免发布警告 |

## Platform Mapping

| npm 包目录 | dist/ 目录 | 二进制文件 |
|---|---|---|
| `unirtm-darwin-arm64` | `unirtm_darwin_arm64_v8.0` | `unirtm` |
| `unirtm-darwin-x64` | `unirtm_darwin_amd64_v1` | `unirtm` |
| `unirtm-linux-x64` | `unirtm_linux_amd64_v1` | `unirtm` |
| `unirtm-linux-arm64` | `unirtm_linux_arm64_v8.0` | `unirtm` |
| `unirtm-linux-ia32` | `unirtm_linux_386_sse2` | `unirtm` |
| `unirtm-linux-arm` | `unirtm_linux_arm_7` | `unirtm` |
| `unirtm-linux-arm-5` | `unirtm_linux_arm_5` | `unirtm` |
| `unirtm-linux-arm-6` | `unirtm_linux_arm_6` | `unirtm` |
| `unirtm-linux-loong64` | `unirtm_linux_loong64` | `unirtm` |
| `unirtm-linux-ppc64le` | `unirtm_linux_ppc64le_power8` | `unirtm` |
| `unirtm-linux-riscv64` | `unirtm_linux_riscv64_rva20u64` | `unirtm` |
| `unirtm-linux-s390x` | `unirtm_linux_s390x` | `unirtm` |
| `unirtm-windows-x64` | `unirtm_windows_amd64_v1` | `unirtm.exe` |
| `unirtm-windows-arm64` | `unirtm_windows_arm64_v8.0` | `unirtm.exe` |
| `unirtm-windows-ia32` | `unirtm_windows_386_sse2` | `unirtm.exe` |
