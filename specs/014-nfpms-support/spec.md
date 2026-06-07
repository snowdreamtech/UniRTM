# Feature Specification: nfpms-support

**Feature Branch**: `[014-nfpms-support]`

**Created**: 2026-06-07

**Status**: Draft

**Input**: User description: "实施第一阶段，进行nfpms支持"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Generate Linux Packages (Priority: P1)

As a Linux user or system administrator, I want to install UniRTM using standard Linux package managers (apt, yum/dnf, apk, pacman) so that I can easily deploy and manage it natively without manually downloading and extracting archives.

**Why this priority**: Essential for native Linux distribution support, heavily requested by enterprise users and sysadmins for automated deployments.

**Independent Test**: Can be independently tested by running GoReleaser locally or in CI and verifying that `.deb`, `.rpm`, `.apk`, and Arch Linux packages are generated correctly.

**Acceptance Scenarios**:

1. **Given** a new release is triggered, **When** GoReleaser runs, **Then** it generates `.deb`, `.rpm`, `.apk`, and `archlinux` packages.
2. **Given** a generated `.deb` package, **When** installed via `dpkg -i`, **Then** the `unirtm` binary is placed in `/usr/bin/unirtm`.

### Edge Cases

- What happens when a user installs via package manager and then tries to self-update? (Should be disabled or warn the user).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST configure GoReleaser NFPM integration.
- **FR-002**: System MUST generate `.deb` packages for Debian/Ubuntu environments.
- **FR-003**: System MUST generate `.rpm` packages for RHEL/CentOS/Fedora environments.
- **FR-004**: System MUST generate `.apk` packages for Alpine Linux environments.
- **FR-005**: System MUST generate packages for Arch Linux.
- **FR-006**: Generated packages MUST include proper metadata (description, maintainer, license, homepage).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: CI pipeline successfully generates 4 types of Linux packages (`deb`, `rpm`, `apk`, `archlinux`) upon release.
- **SC-002**: Packages can be successfully installed and uninstalled on their respective target systems without errors.

## Assumptions

- We assume standard GoReleaser NFPM defaults are sufficient for initial support without needing complex pre/post-install scripts.

## Out of Scope (Future Enhancements)

The following capabilities are present in the GoReleaser ecosystem but are currently out of scope because they require external platform synchronization, accounts, or tokens. They are documented here for future tracking:

### 1. External Package Manager Publishing (Repository Sync)
这些功能不会生成直接可供下载的安装包，而是生成配置脚本（如 `.rb`, `JSON`, `PKGBUILD`），并通过 Git/API 自动推送到第三方的仓库中。
- **`homebrew_casks`**: 自动向 macOS 的 Homebrew Tap 仓库推送更新。
- **`winget`**: 自动向微软的 Windows 官方包仓库提交 PR。
- **`scoops`**: 自动向 Windows 的 Scoop Bucket 仓库推送更新。
- **`aurs`**: 自动向 Arch Linux User Repository (AUR) 推送 `PKGBUILD` 脚本。
- **`nix`**: 自动向 Nix 用户的 GitHub NUR 仓库推送配置文件。
- **`npms`**: 自动向 npmjs.com 注册表推送 Node.js 版本的包封装。
- **`gemfury`**: 自动向 Gemfury 托管平台推送包。

### 2. Container Images (容器镜像)
- **`dockers_v2` & `docker_signs`**: 根据编译好的二进制文件打包出对应的 Docker 镜像，并自动 push 到 GitHub Container Registry (ghcr.io) 及 Docker Hub，同时用 Cosign 为镜像签名。

### 3. macOS Official Notarization (苹果官方公证)
- **`notarize`**: 使用苹果开发者证书对 macOS 二进制文件进行签名并送交苹果公证服务器进行安全检查。这需要付费的苹果开发者账号。

### 4. Community and Announcements (社区联动与公告自动分发)
- **`announce`**: 每次发布完成后，自动向 Mastodon, Discord, Telegram 等社交平台发送新版本广播，集成 OpenCollective。
- **`milestones`**: 自动根据 Tag 版本号关闭 GitHub 上对应的 Milestone。

### 5. Testing and Nightly Builds (测试与每日构建)
- **`nightly`**: 针对 Nightly 版本的特殊标签规则，允许通过 CI 定期生成不稳定预览版。
