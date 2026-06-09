# Tasks: GoReleaser Configuration Alignment

**Input**: Design documents from `/specs/016-goreleaser-alignment/`

**Prerequisites**: spec.md

## Format: `[ID] [Status] [Priority] Description`

---

## Phase 1: 低成本高收益改进（建议优先实施）

> 不需要外部账号/Token，纯配置改动即可完成。

### 1.1 Builds 修复

- [x] T001 ⚠️ P1 添加 `builds.ignore` 忽略 `goos: windows` + `goarch: arm`（Go 官方已废弃）
      > 参考 goreleaser: `ignore: - goos: windows goarch: arm`

### 1.2 NFPMs 增强

- [x] T002 ⚠️ P1 添加 `nfpms.dependencies: [git]`（工具链管理器通常需要 git）
- [x] T003 ⚠️ P1 为 deb 包添加 `copyright` 文件（Debian 规范要求）
      > 参考 goreleaser: `contents: - src: LICENSE dst: /usr/share/doc/unirtm/copyright`
- [x] T004 🔹 P2 添加 `deb.suggests`（golang, rustup, zig, deno, bun 等可选依赖）

### 1.3 Homebrew Casks 增强

- [x] T005 ⚠️ P1 添加 `url.verified: "github.com/snowdreamtech/"`（防 Supply Chain 攻击）

### 1.4 Changelog 增强

- [x] T006 ⚠️ P1 补充 `changelog.filters.exclude` 排除规则（参考 goreleaser 的 14 条模式）
      > 当前 7 条，建议补充: go mod tidy, go generate, chore, build(deps), ci, test, docs 等
- [-] T007 🔹 P2 ~~添加 `changelog.format` 含 SHA + 作者信息~~（`{{ .Logins }}` 需 Pro，取消）

### 1.5 Release 增强

- [x] T008 ⚠️ P1 添加 `release.footer`（changelog 链接 + 帮助信息）
      > 参考 goreleaser: `Full Changelog: v0.14.0...v0.15.0`

### 1.6 Snapcrafts

- [ ] T009 ⚠️ P1 将 `snapcrafts.publish` 改为 `true`（已配置但未发布，浪费了）
      > 需先确认 Snapcraft Store 凭据已配置在 GitHub Secrets 中

---

## Phase 2: Windows 分发渠道（建议实施）

> Windows 用户的主要包管理器，覆盖面广。

- [x] T010 ⚠️ P1 添加 `winget` 分发（自动向 microsoft/winget-pkgs 提交 PR）
      > 需配置 GitHub PAT 并 fork microsoft/winget-pkgs
- [ ] T011 🔹 P2 添加 `scoops` 分发（Windows 备选包管理器）
      > 需创建 Scoop Bucket 仓库

---

## Phase 3: Linux 分发渠道（按需实施）

> 扩大 Linux 用户覆盖面。

- [ ] T012 🔹 P2 添加 `aurs` 分发（Arch Linux 用户需要）
      > 需配置 AUR SSH 密钥到 GitHub Secrets
- [ ] T013 🔹 P3 添加 `nix` 分发（NixOS 用户需要）
      > 需创建 NUR 仓库

---

## ~~Phase 4: CI/CD 工作流增强（按需实施）~~ 已取消

> ~~增强发布流程的健壮性和可验证性。~~

- [-] T014 ~~🔹 P2 添加 `changelog.format` 含 SHA + 作者~~（`{{ .Logins }}` 需 Pro，取消）
- [-] T015 ~~🔹 P3 添加 `nightly` 发布支持~~（暂不实施，取消）
- [-] T016 ~~🔹 P3 添加 `announce` 通知~~（暂不实施，取消）
- [-] T017 ~~🔹 P3 添加 `milestones.close: true`~~（暂不实施，取消）

---

## ~~Phase 5: 高级功能（长期目标）~~ 已取消

> ~~需要外部账号、付费服务或额外基础设施。~~

- [-] T018 ~~🔹 P3 macOS `notarize`~~（暂不实施，取消）
- [-] T019 ~~🔹 P3 `dockers_v2` + `docker_signs`~~（暂不实施，取消）
- [-] T020 ~~🔹 P3 `gemfury` 包分发~~（暂不实施，取消）
- [-] T021 ~~❌ P4 `pro: true`~~（暂不实施，取消）

---

## 优先级说明

| 优先级 | 含义 | 预计工作量 |
|---|---|---|
| ⚠️ P1 | 建议尽快实施 | 每项 10-30 分钟 |
| 🔹 P2 | 建议实施 | 每项 30-60 分钟 |
| 🔹 P3 | 按需实施 | 每项 1-2 小时 |
| ❌ P4 | 暂不实施 | 需额外投入 |

## 状态说明

- `[ ]` 未开始
- `[~]` 进行中
- `[x]` 已完成
- `[-]` 已取消
