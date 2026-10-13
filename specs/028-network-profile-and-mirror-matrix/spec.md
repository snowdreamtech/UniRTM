# Feature Specification: Network Profiles, Domestic Mirror Matrix & Cascading Overrides

**Feature Identifier**: `028-network-profile-and-mirror-matrix`  
**Created**: 2026-10-13  
**Status**: Implemented  
**Input**: "Design and implement a complete network profile, domestic mirror matrix, and cascading proxy management subsystem for pre-compilation environment and network preparation. Support multi-ecosystem public domestic mirrors (built-in dropdowns); strictly forbid embedding any third-party proxy servers or GitHub relays (default to official direct); support full profile lifecycle CRUD (create, edit, save-as, reset to factory defaults, delete); support per-language granular overrides (Overrides > Active Profile); maintain full parity across CLI and Web UI with 100% native localization."

---

## 1. Background & Motivation

### 1.1 Development Lifecycle Positioning
Modern developer toolchains operate in distinct phases:
- **VS Code / AI Coding Agents**: Operate during the **coding, syntax analysis, and build invocation phase**, generating and refining application code.
- **UniRTM (Universal Runtime Manager)**: Operates during the **pre-flight environment and network preparation phase**.
- **Pain Point**: Before UniRTM, developers had to manually install language runtimes, battle firewall timeouts during `go get`, `npm install`, or `pip install`, and manually configure environment variables. If pre-flight environments and network connectivity fail, even advanced IDEs and AI agents cannot compile or run projects.

### 1.2 Security Boundaries & Inviolable Principles
1. **Public Mirror Matrices are Built-in and Selectable**: Compliant domestic public mirrors (Aliyun, Tencent Cloud, Huawei Cloud, Tsinghua University, USTC, npmmirror, goproxy.cn, Ruby China) serve as public infrastructure. The system provides pre-configured templates and dropdown selectors.
2. **Never Embed Third-Party Proxy Servers or GitHub Relays**:
   - Proxy servers belong to private local environments (e.g., `127.0.0.1:7890`) or enterprise gateways. Proxy configuration must strictly default to disabled with empty endpoints.
   - Third-party GitHub relay proxies (such as `ghproxy` or unvetted mirrors) are strictly prohibited to prevent Man-in-the-Middle (MITM) hijacking, token leakage, and backdoor supply-chain attacks. GitHub always defaults to official upstream `github.com`.
3. **100% Pure Native Localization**:
   - UI text and CLI outputs must render in clean, pure mother-tongue localization without awkward bilingual parenthetical mixing.

---

## 2. User Stories & Acceptance Criteria

### User Story 1 - Profile Lifecycle Management (Priority: P1) 🎯 MVP

As a developer collaborating across diverse network environments, I want to switch between network profiles with one click and create, edit, save, reset, and delete profiles so that I can quickly align configurations across home networks, office intranets, and CI/CD runners.

**Acceptance Scenarios**:
1. **Given** a fresh workspace or global configuration, **When** running `unirtm network profile list` or opening Web Settings, **Then** the built-in `official` and `cn` profiles are displayed.
2. **Given** current custom settings, **When** invoking `unirtm network profile save-as <name>` or clicking "Save as New Profile" in Web UI, **Then** a new independent profile is created and persisted to TOML.
3. **Given** a modified built-in profile that needs restoring, **When** executing `unirtm network profile reset cn` or clicking "Reset to Factory Defaults", **Then** the profile instantly reverts to the factory blueprint.
4. **Given** a custom profile, **When** running `unirtm network profile delete <name>` or clicking delete in Web UI, **Then** the custom profile is removed and the active profile safely falls back to `official`.

---

### User Story 2 - Granular Ecosystem Overrides (Cascading Resolution) (Priority: P1)

As a developer with specific language registry requirements (e.g., needing an enterprise npm registry while keeping Go and Python on public mirrors), I want to override individual language registries without being forced into an all-or-nothing profile change, and be able to restore inherited behavior at any time.

**Acceptance Scenarios**:
1. **Given** active profile `cn`, **When** the user overrides `npm` with a private registry URL, **Then** the resolution engine applies `Overrides > Active Profile > Default`, keeping npm overridden while Go, Python, Rust, and Ruby follow `cn`.
2. **Given** an overridden ecosystem, **When** viewed in the Web UI or CLI, **Then** an `[Overridden]` badge and a `[Follow Profile]` reset button are displayed.
3. **Given** the user resets the override, **Then** the override key is removed and the ecosystem immediately resumes following the active profile.

---

### User Story 3 - Secure Proxy Decoupling & Child Process Sandbox (Priority: P1)

As a security-conscious developer, I want the runtime to maintain strict proxy neutrality without embedding unvetted third-party relays, injecting proxy environment variables exclusively into executed child processes without mutating host global files.

**Acceptance Scenarios**:
1. **Given** a default setup, **When** inspecting proxy settings, **Then** `enabled` is `false` and URLs are empty.
2. **Given** the user configures a local proxy (e.g., `http://127.0.0.1:7890`), **When** running tasks or toolchain downloads, **Then** child processes receive `HTTP_PROXY`, `HTTPS_PROXY`, `ALL_PROXY`, and `NO_PROXY`.
3. **Given** the task completes, **When** inspecting shell startup files, **Then** the host environment remains untouched.

---

### User Story 4 - Real-Time Latency Probing Matrix (Priority: P2)

As a developer, I want to inspect live round-trip latency (RTT) across official registries and mirror nodes to make informed network decisions.

**Acceptance Scenarios**:
1. **Given** the user initiates a network test, **When** probing endpoints concurrently, **Then** RTT latency results and status ratings (Fast <100ms, Normal <300ms, Slow, Timeout) are returned within 2.5 seconds.
2. **Given** an offline environment, **When** probes fail, **Then** endpoints gracefully report `TIMEOUT` without crashing.

---

### User Story 5 - Full CLI and Web UI Parity (Priority: P1)

As a developer using both command-line workflows and web dashboards, I want 1:1 functional parity across CLI commands and Web UI controls.

**Acceptance Scenarios**:
1. Operations executed in Web UI immediately reflect in CLI state, and vice versa.
2. Complete i18n support across English, Simplified Chinese, and Traditional Chinese.

---

## 3. Functional Requirements

- **FR-001 (Configuration Schema)**: The system MUST organize network configuration under `[network]` with `proxy`, `profiles`, and `overrides` tables.
- **FR-002 (Proxy Defaults)**: Proxy `enabled` MUST default to `false` and endpoint URLs MUST default to empty.
- **FR-003 (Factory Profiles)**: The system MUST ship with immutable factory blueprints:
  - `official`: Direct connections to official registries.
  - `cn`: Curated domestic mirror matrix (npmmirror, goproxy.cn, Tsinghua PyPI, rsproxy, Ruby China).
- **FR-004 (Profile CRUD)**: The system MUST support creating, modifying, saving as, and deleting custom profiles.
- **FR-005 (Factory Reset)**: The system MUST support resetting built-in profiles to factory blueprints.
- **FR-006 (Resolution Precedence)**: The resolution engine MUST resolve effective registries following:
  `Overrides[ecosystem] > ActiveProfile.Registries[ecosystem] > DefaultBaseline[ecosystem]`.
- **FR-007 (Override Reset)**: The system MUST support clearing granular ecosystem overrides.
- **FR-008 (Environment Export)**: The system MUST provide `ExportEnvironment` to format effective registries into standard environment variables (`NPM_CONFIG_REGISTRY`, `GOPROXY`, `PIP_INDEX_URL`, `CARGO_REGISTRIES_CRATES_IO_INDEX`, etc.).
- **FR-009 (CLI Command Suite)**: The CLI MUST provide `unirtm network status`, `profile`, `registry`, `proxy`, and `test` commands.
- **FR-010 (Web REST API)**: The web service MUST expose `/api/network/status`, `/api/network/profile/*`, `/api/network/registry/*`, `/api/network/proxy`, and `/api/network/ping`.
- **FR-011 (Web Settings UI)**: The Web UI MUST provide dedicated sections for Network Proxy and Mirror Matrix with live inputs and theme-adaptive modal dialogs.
- **FR-012 (Probe Timeout)**: Latency probes MUST execute concurrently with a maximum timeout of 2.5 seconds.
- **FR-013 (Native Localization)**: All user-facing strings MUST be fully localized in i18n dictionaries without bilingual parenthetical mixing.

---

## 4. Edge Cases & Error Handling

1. **Active Profile Deleted**: Automatically falls back to `official` with a warning log, avoiding crashes.
2. **Invalid Proxy URL**: Rejects non-HTTP/SOCKS schemes with actionable error messages.
3. **Total Network Disconnection**: Reports `TIMEOUT` gracefully without deadlocking.
4. **Empty Override Input**: Equivalent to resetting the override, returning to active profile inheritance.
