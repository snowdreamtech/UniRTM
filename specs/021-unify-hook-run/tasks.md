# Task List: Unify `unirtm hook run` Arguments

**Feature**: 021-unify-hook-run
**Plan**: [plan.md](./plan.md) | **Spec**: [spec.md](./spec.md)

## Implementation Strategy

- **Foundational**: Update the CLI parser to extract `--stage` and trailing `args...`, and update the `HookRunner` interface signature.
- **Engine Backends**: Implement the updated routing logic across all engine backends (Husky, Lefthook, Pre-commit, Native/Shell) simultaneously, as they all share the same interface contract.
- **Testing**: Manual testing using the scenarios defined in `quickstart.md` to ensure correct native Git argument passthrough and strict syntax enforcement.

## Dependencies

- Phase 2 (Foundational) blocks Phase 3 (Engines).
- Engine implementations in Phase 3 can be executed in parallel.

---

## Phase 1: Setup

*(No new project initialization required, modifying existing architecture)*

---

## Phase 2: Foundational

**Goal**: Update the CLI layer to parse the new `--stage` flag, accept trailing arguments, and define the new `HookRunner` contract.

- [x] T001 Update `HookRunner` interface `Run` method signature to `Run(hookName string, stage string, args []string) error` in `internal/hook/runner.go`
- [x] T002 Update `unirtm hook run` CLI configuration to add persistent `--stage` flag and accept exactly 1 positional argument + trailing arguments in `cmd/62.hook.go`
- [x] T003 Update CLI routing logic to extract trailing arguments `args[1:]` and pass to `runner.Run` in `cmd/62.hook.go`

---

## Phase 3: Engine Implementations `[US1][US2][US3]`

**Goal**: Map the unified command structure down to the heterogeneous hook engines, correctly passing trailing native arguments.

- [x] T004 [P] `[US1][US2][US3]` Update `Run` signature and implementation logic for `pre-commit` backend in `internal/hook/precommit.go`
- [x] T005 [P] `[US1][US2][US3]` Update `Run` signature and implementation logic for `lefthook` backend in `internal/hook/lefthook.go`
- [x] T006 [P] `[US1][US2][US3]` Update `Run` signature and implementation logic for `husky` backend in `internal/hook/husky.go`
- [x] T007 [P] `[US1][US2][US3]` Update `Run` signature and implementation logic for `native` backend in `internal/hook/native.go`
- [x] T008 [P] `[US1][US2][US3]` Update `Run` signature and implementation logic for `shell` backend in `internal/hook/shell.go`

---

## Phase 4: Polish & Cross-Cutting

**Goal**: Verify all execution flows using the quickstart manual testing scenarios.

- [x] T009 Validate argument boundary rejections and parameter passthrough locally as defined in `quickstart.md`
