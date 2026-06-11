# Task Breakdown: Npm Security & Corepack Alignment

- `[x]` **Phase 1: Unit Testing Infrastructure (TDD)**
  - `[x]` Modify `internal/provider/npm_test.go` to include test cases that mock `package.json` parsing and trigger the lifecycle warning.
  - `[x]` Modify `internal/provider/node_test.go` to update `GenerateShims` and `ListExecutables` test cases, asserting that Corepack-related shims are generated correctly.

- `[x]` **Phase 2: npm Supply Chain Security Hardening (Atomic Development)**
  - `[x]` Enforce the `--ignore-scripts=true` parameter for `npm install` in `internal/provider/npm.go`.
  - `[x]` Add `checkAndWarnLifecycleScripts` to dynamically identify POSIX/Windows paths and emit security interception warnings.
  - `[x]` Atomic Commit.

- `[x]` **Phase 3: Corepack Delegation Mechanism Integration (Atomic Development)**
  - `[x]` Inject `corepack`, `yarn`, `pnpm` environments into the whitelist in `ListExecutables()` within `internal/provider/node.go`.
  - `[x]` Implement specialized Shim generators to automatically delegate `yarn` script execution to `corepack yarn`.
  - `[x]` Atomic Commit.
