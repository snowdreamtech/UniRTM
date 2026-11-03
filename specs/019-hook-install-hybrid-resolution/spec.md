# Feature Specification: Hook Install Hybrid Resolution

**Feature Branch**: `019-hook-install-hybrid-resolution`

**Created**: 2026-06-11

**Status**: Draft

**Input**: User description: "Refactor unirtm hook install to be non-destructive and robust for AI/CI environments using a hybrid binary resolution strategy."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Non-Destructive Hook Installation (Priority: P1)

As a developer with existing custom Git hooks, I want `unirtm hook install` to safely co-exist with my existing scripts rather than overwriting them, so that I don't lose my manual configurations.

**Why this priority**: Preventing data loss (overwriting user files) is the highest priority for any CLI tool.
**Independent Test**: Can be tested by creating a custom `.git/hooks/pre-commit` file, running `unirtm hook install`, and verifying the original custom logic is still present and runs.

**Acceptance Scenarios**:

1. **Given** a custom `.git/hooks/pre-commit` exists, **When** the user runs `unirtm hook install`, **Then** the UniRTM block is injected safely at the top (after the shebang) and the existing code remains intact.
2. **Given** a hook that already has the UniRTM block, **When** the user runs `unirtm hook install` again, **Then** the hook file is not duplicated (idempotent).

---

### User Story 2 - Headless AI/CI Environment Execution (Priority: P1)

As an automated CI/CD runner or AI Agent operating in a sandboxed shell, I want the Git hooks to automatically find and load the UniRTM environment without relying on user profile scripts (`~/.zshrc`, `~/.profile`), so that automated commits don't fail due to missing dependencies.

**Why this priority**: Essential for the modern AI-assisted development workflow and reliable CI/CD pipelines.
**Independent Test**: Can be tested by running `git commit` in a completely empty environment (e.g., `env -i sh`) and verifying it still successfully triggers the UniRTM hooks.

**Acceptance Scenarios**:

1. **Given** an environment where `unirtm` is not in `$PATH` and `$HOME/.profile` is absent, **When** a git hook triggers, **Then** the hook dynamically locates the `unirtm` binary in the project root and executes successfully.
2. **Given** a standard GUI Git client, **When** a git hook triggers, **Then** it continues to work seamlessly by falling back to user profiles if the local binary isn't found.

---

### User Story 3 - Batch Installation of Hooks (Priority: P2)

As a developer configuring a new repository, I want to install the UniRTM environment block to all active Git hooks at once using a `-a` or `--all` flag, so that I don't have to manually install them one by one.

**Why this priority**: Improves developer experience and onboarding speed, though not strictly required for core functionality.
**Independent Test**: Can be tested by running `unirtm hook install -a` in a repository with multiple hooks and verifying all non-sample hooks are updated.

**Acceptance Scenarios**:

1. **Given** a repository with several standard `.sample` hooks and some custom hooks, **When** the user runs `unirtm hook install -a`, **Then** the UniRTM block is injected into all active (non-sample) hook files, and `.sample` files are skipped.

---

### Edge Cases

- What happens if the existing hook has no shebang (`#!/bin/sh`)? The block should be inserted at the very top of the file.
- What happens if the `unirtm hook run` command fails? The hook must instantly propagate the failure exit code and abort the git operation.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST parse existing hook files and insert the UniRTM environment loader block immediately following the first line if it is a shebang, or at the top if no shebang exists.
- **FR-002**: System MUST wrap the injected code with clear `BEGIN UNIRTM MANAGED BLOCK` and `END UNIRTM MANAGED BLOCK` markers.
- **FR-003**: System MUST implement idempotency. If the `BEGIN UNIRTM MANAGED BLOCK` is detected, the command must update/replace the block rather than duplicating it.
- **FR-004**: System MUST inject a script that attempts to locate `unirtm` via `git rev-parse --show-toplevel` as a fallback for pure sandbox environments.
- **FR-005**: System MUST act strictly as an environment injector (Route A). It MUST NOT inject `unirtm hook run` or attempt to act as a hook router, to avoid duplicate execution when native hook configurations (e.g. `pre-commit`) are already present in the script.
- **FR-006**: System MUST support a `-a/--all` flag for the `install` command that traverses the `.git/hooks/` directory, filters out non-executable or `.sample` files, and applies the injection logic to all valid hook files.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of pre-existing hook logic is preserved when running `unirtm hook install`.
- **SC-002**: Commits triggered from an empty environment (`env -i PATH=/bin:/usr/bin`) succeed if the `unirtm` binary is present in the repository root, as the injected block correctly sets up `$PATH`.
- **SC-003**: Repeated executions of `unirtm hook install` leave the hook file with exactly one UniRTM environment block without corrupting the file.
- **SC-004**: Running `unirtm hook install` on a hook containing native `pre-commit` code does not result in the linters being executed twice.

## Assumptions

- We assume `git rev-parse` is available in all environments where git hooks run.
- We assume developers will not manually modify the contents inside the `BEGIN/END UNIRTM MANAGED BLOCK`.
