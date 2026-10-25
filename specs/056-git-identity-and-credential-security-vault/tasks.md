# Task Breakdown: Git Identity & Credential Security Vault

**Spec Document**: [spec.md](spec.md)
**Plan Document**: [plan.md](plan.md)

---

## Phase 0: SpecKit Specification & Planning (Documentation)
- [x] 0.1 Author `spec.md` with zero private key touch, local repo isolation, and user stories
- [x] 0.2 Author `plan.md` technical architecture and API contracts
- [x] 0.3 Author `tasks.md` step-by-step atomic execution checklist
- [x] 0.4 **Execute Commit 0**: `docs(spec): add specification and tasks for 056-git-identity-and-credential-security-vault`

---

## Phase 1: Backend Git Profile Isolation & Persistence (Identity Manager)
- [x] 1.1 Implement `GitProfile` and `GitProfileManager` storage in `internal/server/git_profile.go`
- [x] 1.2 Implement `ApplyProfileToWorkspace`: execute local `git config --local` with zero global pollution
- [x] 1.3 Implement read-only system key discovery (`~/.ssh/*.pub` and GPG public key ring)
- [x] 1.4 Expose REST endpoints: `GET/POST/DELETE /api/git/profiles`, `POST /api/git/profile/apply`, `GET /api/git/system-keys`
- [x] 1.5 Write unit tests in `internal/server/git_profile_test.go` verifying CRUD and local configuration isolation
- [x] 1.6 **Execute Commit 1**: `feat(server): add git identity profile management with local repository isolation`

---

## Phase 2: Host Agent Detection & Hardened Nonce IPC (Security & IPC)
- [x] 2.1 Implement native `ssh-agent` and `gpg-agent` detector returning health diagnostics on `GET /api/git/agent-status`
- [x] 2.2 Implement cryptographic single-use nonce issuance in credential coordinator, injecting `UNIRTM_CREDENTIAL_NONCE`
- [x] 2.3 Enforce single-use nonce verification on AskPass endpoint `/api/git/credential-internal-ask` with immediate invalidation
- [x] 2.4 Implement in-memory `CredentialAuditTrail` buffer and expose `GET /api/git/credential-audit`
- [x] 2.5 Write unit tests in `internal/server/credential_test.go` covering nonce generation, burn-on-use, and audit logging
- [x] 2.6 **Execute Commit 2**: `feat(server): enforce one-time nonce token in askpass ipc and audit trail`

---

## Phase 3: Frontend Git Identity State Machine & Types (Composable & Types)
- [x] 3.1 Define `GitProfile`, `SystemKeysSummary`, `NativeAgentStatus`, and `CredentialAuditRecord` in `frontend/src/types/credential.ts`
- [x] 3.2 Implement `frontend/src/composables/useGitIdentity.ts` for profile management, agent status, and audit retrieval
- [x] 3.3 Write unit tests in `frontend/src/composables/useGitIdentity.test.ts`
- [x] 3.4 **Execute Commit 3**: `feat(frontend): implement useGitIdentity composable and types with tests`

---

## Phase 4: Status Bar Switcher & Security Settings Panel (UI Components)
- [x] 4.1 Create `GitProfileStatus.vue` status bar component with active badge and quick-pick switcher
- [x] 4.2 Integrate `GitProfileStatus` into `BottomStatusBar.vue`
- [x] 4.3 Create `GitIdentitySettings.vue` with Profile CRUD, agent status cards, and audit log table in `SettingsView.vue`
- [x] 4.4 Write unit tests in `GitProfileStatus.test.ts`
- [x] 4.5 **Execute Commit 4**: `feat(ui): add GitProfileStatus status bar switcher and identity settings`

---

## Phase 5: Multi-Language Synchronization & Full Quality Gates (i18n & Verification)
- [x] 5.1 Define 32 `git_identity.*` translation keys in `frontend/src/i18n/types.ts`, `en-US.ts`, and `zh-CN.ts`
- [x] 5.2 Synchronize and natively localize keys across all 53 language files, passing `<25` purity threshold
- [x] 5.3 Verify all tests (`go test` and `npm run test`) and production build (`vue-tsc` + `vite build`)
- [x] 5.4 **Execute Commit 5**: `feat(i18n): sync multi-language keys and finalize git identity security vault`
