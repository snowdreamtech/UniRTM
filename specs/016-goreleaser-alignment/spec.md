# Feature Specification: GoReleaser Configuration Alignment

**Feature Branch**: `[016-goreleaser-alignment]`

**Created**: 2026-06-09

**Status**: Draft

**Input**: 与 goreleaser/goreleaser 上游 `.goreleaser.yaml` 的全面对比，识别 UniRTM 可改进项。

## Overview

通过与 goreleaser 上游项目的发布配置进行逐项对比，识别 UniRTM 在 GoReleaser 配置、CI/CD 工作流、分发渠道等方面的差距，并制定分阶段改进计划。

## 对比总览

### 🏗️ 基础配置

| 配置项 | UniRTM | goreleaser | 差距 |
|---|---|---|---|
| GoReleaser 版本 | Free (`distribution: goreleaser`) | Pro (`pro: true`) | ⚠️ Pro 需付费 |
| `report_sizes` | ✅ | ✅ | — |
| `gomod.proxy` | ✅ | ✅ | — |
| `git.ignore_tags` | ✅ | ✅ | — |
| `metadata.mod_timestamp` | ✅ | ✅ | — |
| `env` (GO111MODULE) | ❌ | ✅ | 🔹 可加 |

### 🔨 Builds

| 配置项 | UniRTM | goreleaser | 差距 |
|---|---|---|---|
| `CGO_ENABLED=0` | ✅ | ✅ | — |
| `-trimpath` | ✅ | ✅ | — |
| `ldflags -s -w` | ✅ | ✅ | — |
| `goos` (linux/darwin/windows) | ✅ | ✅ | — |
| `goarch` | 386/amd64/arm/arm64/loong64/ppc64le/riscv64/s390x | 386/amd64/arm/arm64/loong64/ppc64/riscv64 | UniRTM 多 s390x |
| `goarm` | "7" | "7" | — |
| `builds.ignore` (windows/arm) | ❌ | ✅ | ⚠️ 应忽略 |
| `mod_timestamp` | ✅ | ✅ | — |

### 📦 Archives

| 配置项 | UniRTM | goreleaser | 差距 |
|---|---|---|---|
| `name_template` | ✅ (uname 兼容) | ✅ (相同) | — |
| `wrap_in_directory` | ✅ `true` | ❌ | ✅ UniRTM 更好 |
| `builds_info` (root/mtime) | ✅ | ✅ | — |
| `format_overrides` (windows→zip) | ✅ | ✅ | — |
| `files` 内容 | LICENSE/README/CHANGELOG + completions/manpages | README/LICENSE + completions/manpages | UniRTM 多 CHANGELOG |

### 📋 NFPMs (deb/rpm/apk/archlinux)

| 配置项 | UniRTM | goreleaser | 差距 |
|---|---|---|---|
| `formats` | apk/deb/rpm/archlinux | apk/deb/rpm/archlinux | — |
| `bindir` | `/usr/bin` | `/usr/bin` | — |
| `section` | `utils` | `utils` | — |
| completions (bash/zsh/fish) | ✅ | ✅ | — |
| manpages | ✅ | ✅ | — |
| `dependencies` | ❌ | `git` | ⚠️ 应加依赖声明 |
| `deb.suggests` | ❌ | golang/rustup/zig/deno/bun | 🔹 可选 |
| `deb.lintian_overrides` | ✅ | ✅ | — |
| copyright 文件 | ❌ | ✅ (LICENSE→copyright) | ⚠️ Debian 规范要求 |

### 🏠 Homebrew Casks

| 配置项 | UniRTM | goreleaser | 差距 |
|---|---|---|---|
| `repository` | ✅ | ✅ | — |
| `binaries` | ✅ | ❌ | ✅ UniRTM 更好 |
| `manpages` | ✅ | ✅ (`.gz`) | — |
| `completions` | ✅ | ✅ | — |
| `conflicts` | ❌ | ✅ (goreleaser-pro) | 🔹 可选 |
| `url.verified` | ❌ | ✅ | ⚠️ 安全最佳实践 |
| `hooks.post.install` | ✅ (macOS xattr) | ❌ | ✅ UniRTM 更好 |

### 🐧 Snapcrafts & Flatpak

| 配置项 | UniRTM | goreleaser | 差距 |
|---|---|---|---|
| snapcrafts `confinement` | `classic` | `classic` | — |
| snapcrafts `publish` | `false` | `true` | ⚠️ 未发布到 Snap Store |
| flatpak `app_id` | ✅ | ✅ | — |
| flatpak `runtime_version` | "24.08" | "24.08" | — |

### 🔐 签名 & SBOM

| 配置项 | UniRTM | goreleaser | 差距 |
|---|---|---|---|
| cosign (checksum) | ✅ | ✅ | — |
| `docker_signs` | ❌ | ✅ (manifests) | 🔹 UniRTM 无 Docker |
| SBOM `artifacts: archive` | ✅ | ✅ | — |
| SBOM `artifacts: package` | ✅ | ❌ | ✅ UniRTM 更好 |

### 📝 Changelog

| 配置项 | UniRTM | goreleaser | 差距 |
|---|---|---|---|
| `sort` | `asc` | `asc` | — |
| `use` | `github` | `github` | — |
| `format` | 默认 | `{{ .SHA }}: {{ .Message }} ({{ .Logins }})` | 🔹 可加 SHA+作者 |
| `filters.exclude` | 7 条 | 14 条 | ⚠️ 应补充排除规则 |
| `groups` | 7 组 | 4 组 | ✅ UniRTM 更细 |

### 🚀 Release

| 配置项 | UniRTM | goreleaser | 差距 |
|---|---|---|---|
| `name_template` | `{{ .Tag }}` | `{{ .Tag }}` | — |
| `prerelease` | `auto` | `auto` | — |
| `header` | ✅ (项目介绍) | ✅ (发布公告链接) | — |
| `footer` | ❌ | ✅ (changelog 链接 + 帮助信息) | ⚠️ 应加 footer |
| `source` | ✅ (source tarball) | ❌ | ✅ UniRTM 更好 |

### 📦 分发渠道

| 配置项 | UniRTM | goreleaser | 差距 |
|---|---|---|---|
| Homebrew Cask | ✅ | ✅ | — |
| npm | ✅ (独立 pipeline) | ✅ (内置 `npms:`) | 不同实现方式 |
| Snapcrafts | 配置了但 `publish: false` | ✅ | ⚠️ 未发布 |
| Flatpak | ✅ | ✅ | — |
| Nix | ❌ | ✅ | 🔹 可加 |
| Winget | ❌ | ✅ | ⚠️ Windows 用户需要 |
| AUR | ❌ | ✅ | 🔹 Arch 用户需要 |
| Scoop | ❌ | ✅ | 🔹 Windows 用户需要 |
| Gemfury | ❌ | ✅ | 🔹 可选 |

### 🌙 其他

| 配置项 | UniRTM | goreleaser | 差距 |
|---|---|---|---|
| `nightly` | ❌ | ✅ | 🔹 可选 |
| `announce` | ❌ | ✅ (Mastodon/Discord/Telegram) | 🔹 可选 |
| `milestones` | ❌ | ✅ (close: true) | 🔹 可选 |
| macOS `notarize` | ❌ | ✅ | 🔹 需 Apple 开发者证书 |
| `dockers_v2` | ❌ | ✅ | 🔹 UniRTM 暂无 Docker |
