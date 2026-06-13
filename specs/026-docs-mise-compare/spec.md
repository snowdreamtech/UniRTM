# Feature Specification: Documentation Upgrade: Aligning and Surpassing `mise`

**Feature Branch**: `026-docs-mise-compare`
**Created**: 2026-06-13
**Status**: Draft
**Input**: User description: "制定一个计划，1. 参考mise的docs，从栏目上，从内容上，对标。2. 超越mise的部分架构和内容必须点出来。3. 新增一个大栏目。用于介绍unirtm和nvm，gvm等等工具一一对比，特性，用法。方便开发者用来作为他们的替代。"

## Clarifications

### Session 2026-06-13

- Q: Should the comparison with `mise` just match features, or explicitly highlight superiority? → A: It must not only match but explicitly surpass `mise`, highlighting areas where UniRTM is superior.
- Q: How should the documentation handle internationalization (i18n)? → A: The site must support both English and Chinese simultaneously, and automatically redirect users to their preferred language based on their browser environment.
- Q: How should we address competitors like `mise` in the documentation? → A: Show respect to competitors and avoid any trash-talking or put-downs (不拉踩).
- Q: What level of detail is expected in the documentation compared to `mise`? → A: Must have depth and breadth, avoiding vague, boastful, or overly marketing-driven language (拒绝假大空). The architecture and caching strategy sections must dive deep into the code and include flowcharts or structural diagrams.
- Q: How should the CLI overview section be structured? → A: All commands must be exhaustively listed and detailed one by one in the CLI overview pages (can be collapsible).
- Q: How detailed should the environments overview be? → A: Must provide a detailed and in-depth explanation, avoiding perfunctory 1-2 sentence summaries.
- Q: How should integrations (like Trivy & Syft) be presented? → A: They must be accurately described as external tool integrations, NOT as native UniRTM features. Avoid exaggerated marketing claims.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Evaluate UniRTM against existing tools (Priority: P1)

As a developer using legacy tools (nvm, gvm, pyenv) or modern tools (mise, asdf), I want to see a clear, side-by-side comparison with UniRTM, so that I can understand its unique benefits and confidently use it as a replacement.

**Why this priority**: Directly addresses user adoption friction by providing clear migration rationale.
**Independent Test**: Can be tested by navigating to the new "Comparisons" section in the docs and verifying clear, structured contrast points for each tool.

**Acceptance Scenarios**:

1. **Given** a user is evaluating UniRTM against nvm, **When** they view the comparison docs, **Then** they see clear advantages (e.g. cross-platform, corepack fallback) and basic usage alternatives.
2. **Given** a user is evaluating UniRTM against mise, **When** they view the docs, **Then** they see UniRTM's architectural superiority highlighted (e.g. native pipx replacement, no external plugins).

---

### User Story 2 - Discover UniRTM's Architectural Superiority (Priority: P1)

As a technical evaluator or architect reading the introduction, I want to clearly understand how UniRTM's architecture surpasses competitors like `mise`, so that I can trust its zero-pollution and zero-dependency guarantees.

**Why this priority**: Essential for establishing the tool's competitive edge.
**Independent Test**: Can be tested by reviewing the Introduction and Getting Started pages for highlighted superiority points.

**Acceptance Scenarios**:

1. **Given** a user reading the introduction, **When** they look for "Why UniRTM", **Then** they find explicit mentions of 100% Native Architecture, Zero-Pollution Philosophy, and built-in MCP server capabilities.

### Edge Cases

- What happens when a user is looking for a comparison with an obscure tool not listed? (They should be directed to the general 'asdf' or generic comparison).
- How does the documentation handle potential future updates to `mise` that might bridge the gap? (Focus on fundamental architectural differences rather than point-in-time features).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The documentation MUST structurally align with modern standard CLI tool docs (like `mise`), ensuring sections like Introduction, Installation, Configuration, and Commands are clear and logical.
- **FR-002**: The `introduction.md` and `getting-started.md` pages MUST explicitly highlight UniRTM's native Go architecture, eliminating the need for bash shims, plugins, or third-party hook managers like `direnv`.
- **FR-003**: The documentation MUST include a new major section titled "Comparisons" (替代与对比).
- **FR-004**: The "Comparisons" section MUST include one-on-one comparisons for `nvm`, `gvm`, `pyenv`/`pipx`, `asdf`, `mise`, and `direnv`.
- **FR-005**: Each tool comparison MUST detail UniRTM's specific feature advantages and usage substitution instructions. In particular, the comparison with `mise` MUST explicitly list areas where UniRTM surpasses it, rather than just matching features. **However**, the tone MUST remain respectful and professional, completely avoiding trash-talking (不拉踩).
- **FR-006**: The documentation site MUST support both English and Chinese, and include a mechanism to automatically detect the user's browser language environment to redirect them to the appropriate locale by default.
- **FR-007**: The documentation MUST have technical depth and breadth, avoiding vague or overly boastful "marketing fluff" (拒绝假大空). Specifically, architecture and caching sections MUST include deep code-level insights and visual diagrams (flowcharts/structural diagrams). External integrations (Trivy, Syft, Gitleaks) MUST be accurately framed as integrations, not core features.
- **FR-008**: The CLI overview pages (`cli/overview.md` and `zh/cli/overview.md`) MUST exhaustively list and detail every single command one by one. Content can be collapsible, but must not be omitted.
- **FR-009**: The environments overview pages (`environments/overview.md` and `zh/environments/overview.md`) MUST provide detailed, non-perfunctory explanations of environment management rather than brief summaries.

### Key Entities

- **Documentation Content**: Markdown files located in `docs/` and `docs/zh/`.
- **Sidebar Configuration**: VitePress configuration controlling the navigation structure.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 6 specified competitor tools (nvm, gvm, pyenv/pipx, asdf, mise, direnv) have dedicated comparison sections in the new document.
- **SC-002**: The documentation site builds successfully without any broken links.
- **SC-003**: At least 3 architectural advantages over `mise` are prominently highlighted using callouts/alerts in the Introduction and the Comparison section.
- **SC-004**: Visiting the root path `/` on a browser with Chinese language preference automatically redirects to `/zh/` (or similar functional behavior).

## Assumptions

- VitePress is the static site generator used for the documentation (based on the presence of `.vitepress`).
- Both English and Chinese documentation need to be updated simultaneously.
