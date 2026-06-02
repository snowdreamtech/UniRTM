# Feature Specification: 支持 Hook 脚本数组

**Feature Branch**: `4-hook-arrays`
**Created**: 2026-06-02
**Status**: Draft
**Input**: User description: "支持 Hook 脚本数组"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure multiple pre/post install hooks (Priority: P1)

As a user, I want to define `pre_install` and `post_install` hooks as arrays of strings in my `unirtm.toml` so that I can easily read and maintain sequential commands without complex `&&` concatenation.

**Why this priority**: Core value of the feature, drastically improves config readability.

**Independent Test**: Can be fully tested by creating a config with an array hook and verifying each command executes sequentially during tool installation.

**Acceptance Scenarios**:

1. **Given** a `unirtm.toml` with `pre_install = ["echo 1", "echo 2"]`, **When** the tool is installed, **Then** both commands run in order.
2. **Given** a legacy config with `pre_install = "echo 1"`, **When** the tool is installed, **Then** it still runs correctly (backward compatibility).

---

### User Story 2 - Configure tasks with array of commands (Priority: P2)

As a user, I want to define task `run` commands as an array of strings so that complex tasks are easier to read and format.

**Why this priority**: Natural extension of the array script feature for tasks.

**Independent Test**: Can be fully tested by running a task defined as an array and verifying execution.

**Acceptance Scenarios**:

1. **Given** a task with `run = ["echo start", "echo end"]`, **When** running `unirtm run <task>`, **Then** it executes both commands successfully.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST parse `pre_install` from TOML/YAML as either a string or an array of strings.
- **FR-002**: System MUST parse `post_install` from TOML/YAML as either a string or an array of strings.
- **FR-003**: System MUST parse task `run` definitions from TOML/YAML as either a string or an array of strings.
- **FR-004**: System MUST execute the array elements sequentially.
- **FR-005**: System MUST maintain backward compatibility with existing single-string script definitions.

### Key Entities

- **StringArray**: A custom configuration parser type that safely coerces TOML/YAML strings or arrays into a string slice.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of existing configuration files continue to load and execute without modification.
- **SC-002**: Array-based hooks correctly execute all elements sequentially.
