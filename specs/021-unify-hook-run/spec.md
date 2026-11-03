# Feature Specification: Unify `unirtm hook run` Arguments

**Feature Branch**: `021-unify-hook-run`

**Created**: 2026-06-11

**Status**: Draft

**Input**: User description: "./unirtm hook run 后面最少一个参数，最多两个参数。如果是一个参数，这就是hookname。如果是两个参数，第一个是hookname，第二个是stage。梳理一下，统一一下. 你可以参考一下precommit代码"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Executing a Specific Hook (Priority: P1)

As a developer or automated script, I want to execute a specific Git hook natively without having to specify its lifecycle stage, so that I can validate a single rule/hook quickly.

**Why this priority**: Executing a specific hook (e.g., `pre-commit run shellcheck`) is the primary way users interact with granular linting rules locally.

**Independent Test**: Can be fully tested by running `./unirtm hook run shellcheck` and verifying that the underlying engine triggers only the `shellcheck` hook.

**Acceptance Scenarios**:

1. **Given** a user is inside a repository with a `pre-commit` setup, **When** they run `unirtm hook run shellcheck`, **Then** it translates to `pre-commit run shellcheck`.
2. **Given** a user is inside a repository with a `lefthook` setup, **When** they run `unirtm hook run pre-commit`, **Then** it translates to `lefthook run pre-commit`.

---

### User Story 2 - Executing a Full Lifecycle Stage (Priority: P1)

As a Git client or automated hook script, I want to execute all hooks associated with a specific git lifecycle stage, so that I can enforce all rules during events like `commit-msg` or `pre-commit`.

**Why this priority**: This is how Git naturally invokes hooks (e.g., `.git/hooks/pre-commit` triggers the entire `pre-commit` stage).

**Independent Test**: Can be fully tested by triggering `./unirtm hook run all pre-commit` from a git hook bridge script.

**Acceptance Scenarios**:

1. **Given** an automated git hook script triggers `unirtm`, **When** it executes `unirtm hook run all pre-commit`, **Then** the underlying engine is invoked to run the entire `pre-commit` stage (e.g., `pre-commit run --hook-stage pre-commit` or `lefthook run pre-commit`).

---

### User Story 3 - Executing a Specific Hook Within a Stage (Priority: P2)

As a developer, I want to execute a specific hook strictly within the context of a lifecycle stage, so that I can isolate issues related to how a rule behaves in a specific git context.

**Why this priority**: Some engines support scoping rules to stages. While less common for manual invocation, it provides full granular control.

**Independent Test**: Can be fully tested by running `./unirtm hook run shellcheck pre-commit`.

**Acceptance Scenarios**:

1. **Given** a developer wants to run `shellcheck` in the `pre-commit` context, **When** they execute `unirtm hook run shellcheck pre-commit`, **Then** the engine is instructed to run only `shellcheck` constrained to the `pre-commit` stage.

### Edge Cases

- What happens when a user provides 0 parameters? The CLI should reject the command with a clear error indicating that at least 1 parameter is required.
- What happens when a user provides 3 or more parameters? The CLI should reject the command with a clear error indicating that at most 2 parameters are allowed.
- What happens when the underlying engine (like Husky) does not natively support running a specific hook by name? The runner should degrade gracefully, either falling back to running the entire stage or printing a warning that granular execution is unsupported by the detected engine.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `unirtm hook run` command MUST accept exactly 1 positional argument representing the `hookname`.
- **FR-002**: The `stage` context MUST be provided via a dedicated `--stage` flag (e.g., `--stage pre-commit`). This avoids positional ambiguity.
- **FR-003**: The command MUST accept and preserve arbitrary trailing arguments (`args...`) to allow Git native arguments (like `$1` for `commit-msg`) to be passed down to the underlying engine.
- **FR-004**: The system MUST implement a reserved keyword `all` for the `hookname` argument. When `hookname` is `all` and a `--stage` is provided, it indicates the intent to run the entire stage.
- **FR-005**: The engine dispatcher MUST map the `hookname`, `--stage`, and `args...` consistently to the respective backend tools (`pre-commit`, `husky`, `lefthook`, `native`, `shell`).

### Key Entities *(include if feature involves data)*

- **HookName**: A string representing the specific rule or hook identifier to execute (e.g., `shellcheck`, `go-fmt`, `all`).
- **Stage**: A string representing the Git lifecycle stage (e.g., `pre-commit`, `commit-msg`, `pre-push`).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: CLI strictly requires exactly 1 positional argument (`hookname`) and accepts `--stage` and trailing arguments.
- **SC-002**: Executing `unirtm hook run [hookname] [args...]` correctly triggers specific rules in engines and passes arguments.
- **SC-003**: Executing `unirtm hook run all --stage [stage] [args...]` correctly triggers the full stage execution with Git native arguments across all supported engines without syntax errors.

## Assumptions

- We assume that the concept of `hookname` and `stage` covers the fundamental execution models of all target engines (Husky, Lefthook, pre-commit).
- We assume that the Git bridge scripts will be updated or currently use a syntax compatible with this new unified model (i.e. `unirtm hook run all $hook_name` or `unirtm hook run $hook_name`).
