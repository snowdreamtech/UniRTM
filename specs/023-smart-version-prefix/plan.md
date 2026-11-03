# Implementation Plan: Smart Version Prefix Normalization

**Branch**: `023-smart-version-prefix` | **Date**: 2026-06-12 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/023-smart-version-prefix/spec.md`

## Summary

Implement a smart version prefix normalizer (`NormalizeVersionPrefix`) that intelligently prepends or strips the `v` prefix from version strings based on the target backend's constraints, without modifying semantic aliases like `latest` or `stable`.

## Technical Context

**Language/Version**: Go 1.26+

**Primary Dependencies**: None (Standard Go string manipulation)

**Storage**: N/A

**Testing**: Go testing (`go test`)

**Target Platform**: Mac, Linux, Windows

**Project Type**: CLI Tool / Backend Module

**Constraints**: Must execute efficiently (<1ms) and not corrupt non-numeric version strings.

**Scale/Scope**: Impacts all backend providers querying tags/versions.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

No Constitution violations detected.

## Project Structure

### Documentation (this feature)

```text
specs/023-smart-version-prefix/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
└── backend/
    ├── common.go
    └── common_test.go
```

**Structure Decision**: Place the `NormalizeVersionPrefix` utility function inside `internal/backend/common.go` where it can be easily imported and reused by various backend providers (e.g. GitHub, GitLab).

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A       | N/A        | N/A                                 |
