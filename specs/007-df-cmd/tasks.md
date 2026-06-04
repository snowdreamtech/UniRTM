# Tasks: df command

**Branch**: `007-df-cmd` | **Date**: 2026-06-04
**Input**: Generated from `spec.md` and `plan.md`

## Dependencies

- No external story dependencies (this is a standalone command).
- Wait for PR merge of `007-df-cmd` for deployment.

## Phase 1: Setup

- [x] T001 Identify and add `github.com/pterm/pterm` dependency to `go.mod` if not already present.

## Phase 2: Foundational

- [x] T002 Implement directory size calculation logic in `internal/utils/disk.go` to sum sizes recursively.
- [x] T003 Implement human-readable byte formatting function in `internal/utils/disk.go` to print sizes in powers of 1024.
- [x] T004 Add unit tests for disk size calculation logic in `internal/utils/disk_test.go`.

## Phase 3: User Story 1 (Show human readable disk usage)

**Goal**: As a user, I want to run `unirtm df` to see the disk usage of various folders.

- [x] T005 [P] [US1] Create the Cobra command definition in `cmd/df.go` with `-h` and `--human-readable` flags.
- [x] T006 [US1] Integrate directory size calculation logic within `cmd/df.go` targeting the `unirtm` data directory.
- [x] T007 [US1] Implement `pterm` formatting logic in `cmd/df.go` to present the size statistics in a structured table or tree format.
- [x] T008 [US1] Register `dfCmd` inside `cmd/root.go` or a unified command registry.

## Phase 4: Polish & Cross-Cutting Concerns

- [x] T009 Run `make test` or `go test ./...` to verify all unit tests pass with `df` logic included.
- [x] T010 Validate the new command using `go run . df` on the CLI to ensure presentation is well-formatted.
- [x] T011 Write documentation or add a mention of `df` in CLI help output.
