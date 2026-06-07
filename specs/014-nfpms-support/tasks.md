# Tasks: nfpms-support

**Input**: Design documents from `/specs/014-nfpms-support/`

**Prerequisites**: spec.md

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Initialize spec.md for nfpms and additional packaging support

---

## Phase 2: User Story 1 - Generate Packages (Priority: P1) 🎯 MVP

**Goal**: Support native package generation without publishing to external registries, attaching all artifacts to GitHub Releases.

### Implementation for User Story 1

- [x] T002 [P] [US1] Add `nfpms` configuration in `.goreleaser.yaml` to support `.deb`, `.rpm`, `.apk`, and `archlinux`
- [x] T003 [P] [US1] Add `snapcrafts` configuration in `.goreleaser.yaml` with `publish: false`
- [x] T004 [P] [US1] Add `flatpak` configuration in `.goreleaser.yaml`
- [x] T005 [P] [US1] Add `universal_binaries` for macOS in `.goreleaser.yaml`

---

## Phase 3: Future Publishing Integrations (Out of Scope for MVP)

**Purpose**: External platform synchronizations that require tokens and accounts.

- [ ] T006 Investigate and setup GitHub PATs for external repositories (Homebrew Tap, Scoop Bucket, Winget)
- [ ] T007 Configure `homebrew_casks`, `scoops`, and `winget` in `.goreleaser.yaml`
- [ ] T008 Setup AUR SSH keys in GitHub Secrets and configure `aurs`
- [ ] T009 Register NPM and Gemfury tokens and configure `npms` and `gemfury`
- [ ] T010 Setup Docker registry authentication and configure `dockers_v2`
- [ ] T011 Investigate Apple Developer account requirements for `notarize`
- [ ] T012 Configure `announce` for community platforms (Discord, Telegram, etc.)
- [ ] T013 Implement `nightly` build CI triggers and GoReleaser configuration
