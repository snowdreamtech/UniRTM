# Feature Specification: Smart Version Prefix Normalization

**Feature Branch**: `023-smart-version-prefix`

**Created**: 2026-06-12

**Status**: Draft

**Input**: User description: "023-smart-version-prefix 这个没有留下任何文档？"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Graceful Version Resolution for Git Tags (Priority: P1)

Users often specify versions with or without the `v` prefix indiscriminately (e.g., `1.2.3` or `v1.2.3`), while backend systems or Git repositories strictly require the exact prefix formatting they expect (some require `v1.2.3`, others `1.2.3`).

**Why this priority**: Users shouldn't have to remember whether a specific tool or repository uses `v` prefixes for their versions. The tool should be smart enough to adapt dynamically.

**Independent Test**: Can be fully tested by requesting a version without a `v` for a tool that requires one, and verifying the system successfully resolves and fetches the tool.

**Acceptance Scenarios**:

1. **Given** a user requests version `1.2.3` for a tool, **When** the backend expects versions prefixed with `v`, **Then** the system normalizes the version to `v1.2.3` and successfully queries the backend.
2. **Given** a user requests version `v1.2.3` for a tool, **When** the backend expects versions without a `v`, **Then** the system normalizes the version to `1.2.3` and successfully queries the backend.
3. **Given** a user requests a special alias like `latest` or `stable`, **When** the system normalizes the version, **Then** the alias string remains entirely unchanged.

---

### Edge Cases

- What happens when a version string contains arbitrary letters (e.g. `alpha-1.0`)? (It should not prepend or strip `v` if the logic solely focuses on `v` or `V` prefix stripping for digit-based version formats).
- How does the system handle an already correct prefix? (It should idenpotently return the version string unchanged).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST intelligently prepend a `v` prefix to user-specified version strings if the target backend mandates a `v` prefix, provided the string begins with a digit.
- **FR-002**: System MUST intelligently strip a leading `v` or `V` prefix from user-specified version strings if the target backend mandates no `v` prefix, provided the prefix is immediately followed by a digit.
- **FR-003**: System MUST NOT modify semantic alias tags like `latest` or `stable`.
- **FR-004**: System MUST apply this normalization universally across all supported backend providers that rely on version matching (e.g., GitHub, GitLab, Go, NPM).

### Key Entities

- **VersionString**: The literal version representation provided by the user.
- **Backend**: The package registry or download source which mandates a specific version syntax constraint.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of user requests containing mismatched `v` prefixes are successfully reconciled without failing due to "version not found".
- **SC-002**: Normalization execution takes negligible time (<1ms overhead) per version resolution.
- **SC-003**: All existing unit tests covering various provider versions remain passing or are updated to reflect accurate version normalization constraints.

## Assumptions

- Target backend provider implementations possess the knowledge of whether they inherently require a `v` prefix or not for their respective platforms.
- Versions that genuinely begin with a non-numeric character aside from `v`/`V` (e.g., `release-1.0`) are handled by backend-specific logic or passed through unchanged.
