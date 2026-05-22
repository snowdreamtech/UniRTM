# Tasks: npm Platform Packaging

**Input**: Design documents from `/specs/1-npm-packaging/`
**Prerequisites**: plan.md ✅
**Status**: ✅ COMPLETED (2026-05-22)

---

## Phase 1: Setup（目录结构初始化）

- [x] T001 创建 `npm/` 顶层目录结构（16 个子目录）
- [x] T002 创建 `npm/scripts/` 脚本目录

---

## Phase 2: Foundational（核心脚本与根包）

**Purpose**: 构建脚本和根包委托脚本是所有平台子包工作的前提

- [x] T003 编写 `npm/scripts/build.sh`：从 dist/ 复制二进制、注入版本、生成 package.json
- [x] T004 编写 `npm/unirtm/install.js`：运行时平台检测与命令透传
- [x] T005 编写 `npm/unirtm/package.json.tpl`：根包模板（含 15 个 optionalDependencies）
- [x] T006 [P] shellcheck 验证 build.sh
- [x] T007 [P] shfmt 格式化 build.sh

**Checkpoint**: 核心脚本就绪，平台子包模板工作可并行开始 ✅

---

## Phase 3: Platform Sub-packages（15 个平台子包模板）

**Goal**: 每个平台子包含正确的 `os`/`cpu` 字段，用于 npm 自动筛选

**Independent Test**: `npm pack --dry-run npm/unirtm-<platform>/` 验证包内容

- [x] T008 [P] `npm/unirtm-darwin-arm64/package.json.tpl`
- [x] T009 [P] `npm/unirtm-darwin-x64/package.json.tpl`
- [x] T010 [P] `npm/unirtm-linux-x64/package.json.tpl`
- [x] T011 [P] `npm/unirtm-linux-arm64/package.json.tpl`
- [x] T012 [P] `npm/unirtm-linux-ia32/package.json.tpl`
- [x] T013 [P] `npm/unirtm-linux-arm/package.json.tpl`（armv7）
- [x] T014 [P] `npm/unirtm-linux-arm-5/package.json.tpl`（armv5）
- [x] T015 [P] `npm/unirtm-linux-arm-6/package.json.tpl`（armv6）
- [x] T016 [P] `npm/unirtm-linux-loong64/package.json.tpl`
- [x] T017 [P] `npm/unirtm-linux-ppc64le/package.json.tpl`
- [x] T018 [P] `npm/unirtm-linux-riscv64/package.json.tpl`
- [x] T019 [P] `npm/unirtm-linux-s390x/package.json.tpl`
- [x] T020 [P] `npm/unirtm-windows-x64/package.json.tpl`
- [x] T021 [P] `npm/unirtm-windows-arm64/package.json.tpl`
- [x] T022 [P] `npm/unirtm-windows-ia32/package.json.tpl`

**Checkpoint**: 15 个平台模板全部就绪 ✅

---

## Phase 4: CI/CD 集成

**Goal**: GoReleaser 完成后自动触发 npm 包构建

- [x] T023 修改 `.goreleaser.yaml`：新增 `after.hooks` 调用 `npm/scripts/build.sh`
- [x] T024 修改 `.gitignore`：排除 `npm/*/bin/` 和生成的 `package.json`

**Checkpoint**: CI 集成完成，GoReleaser 构建后自动生成 npm 包 ✅

---

## Phase 5: 验证与发布

**Goal**: 端到端验证 npm 安装链路可用

- [x] T025 冒烟测试：运行 `sh npm/scripts/build.sh`，验证 15 个平台全部成功
- [x] T026 发布 15 个平台子包到 npm：`npm publish npm/unirtm-*/`
- [x] T027 发布根包到 npm：`npm publish npm/unirtm/`
- [x] T028 验证安装：`npm install -g @snowdreamtech/unirtm` 成功
- [x] T029 验证可用：`unirtm --help` 正确输出

---

## Phase 6: Polish & 修复

- [x] T030 修复 `repository.url` 格式（`git+https://...git`）避免发布警告
- [x] Git 提交：`feat(npm): add npm platform packaging structure`
- [x] Git 提交：`fix(npm): correct repository.url format`

---

## Dependencies & Execution Order

- **Phase 1**: 无依赖，立即开始
- **Phase 2**: 依赖 Phase 1，核心脚本是一切的前提
- **Phase 3**: 依赖 Phase 2（目录存在），各平台模板可并行
- **Phase 4**: 依赖 Phase 2（build.sh 已存在）
- **Phase 5**: 依赖 Phase 2-4 全部完成

## Result

```
npm install -g @snowdreamtech/unirtm  → ✅ added 2 packages
unirtm --help                         → ✅ 正常输出
```

- **16 个 npm 包**（1 个根包 + 15 个平台子包）均已发布至 `registry.npmjs.org`
- **版本**: 0.0.1
- **Commits**: `ee8615f`, `825c08c`
