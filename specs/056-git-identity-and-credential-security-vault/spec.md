# Feature Specification: Git Identity & Credential Security Vault

**Feature Branch**: `056-git-identity-and-credential-security-vault`
**Created**: 2026-10-25
**Status**: Implemented
**Input**: Multi-identity Git isolation, Zero Private Key Touch, system agent delegation, hardened one-time nonce IPC, and step-by-step automated commit assurance for AI agents.

---

## 1. Background & Value Proposition

In modern developer platforms where AI agents pair-program and execute atomic commits in background headless environments, Git operations (`git commit -S`, `git push`, `git fetch`) are executed frequently.

When managing corporate, open-source, and personal projects simultaneously, developers and autonomous AI agents face three fundamental security challenges:
1. **Automated Commit Stalling**: Autonomous AI agents operate without interactive TTYs. When encountering locked SSH keys or passphrase-protected GPG keys without secure session caching, processes hang or fail with timeout errors. Passing raw plaintext credentials in prompts or configs introduces severe leak risks.
2. **Cross-Identity Pollution**: Standard Git tools frequently mutate global `~/.gitconfig` or rely on a single default `~/.ssh/id_rsa`, inadvertently exposing corporate email addresses in open-source projects or causing remote authentication rejections.
3. **User Trust & Credential Sniffing Risks**: Developers rightfully demand strict proof that development tools do not scan, exfiltrate, or harvest private keys (`~/.ssh/id_*`, `~/.gnupg/`), and that unprivileged local processes cannot eavesdrop on passphrases.

---

## 2. Core Security & Architectural Principles

1. **Zero Private Key Touch**:
   - UniRTM (frontend, Go backend daemon, and AI agents) **NEVER reads, parses, copies, or stores SSH or GPG private key files**.
   - Private keys remain exclusively in OS directories (`~/.ssh/`, `~/.gnupg/`) under strict `0600` permissions.
   - All cryptographic operations are performed natively by host `ssh` and `gpg` binaries. UniRTM functions solely as an ephemeral AskPass / Loopback pipe.

2. **Zero Global Pollution**:
   - Identity profiles (`user.name`, `user.email`, `user.signingkey`, `core.sshCommand`) are applied **strictly to the local repository `.git/config`** (`git config --local`).
   - The global `~/.gitconfig` is never modified.

3. **Host Agent Delegation First**:
   - Host `ssh-agent` (`$SSH_AUTH_SOCK`) and `gpg-agent` are automatically detected.
   - If keys are pre-loaded in the host agent, Git operations delegate to the OS kernel agent with zero passphrase prompts.

4. **Hardened Single-Use Nonce IPC**:
   - For every Git subprocess execution, the backend generates a cryptographically secure 128-bit one-time token (`UNIRTM_CREDENTIAL_NONCE`).
   - The AskPass helper script sends this nonce in the `X-Unirtm-Nonce` HTTP header.
   - The backend validates the nonce and **burns it immediately upon first use (single-use & expired)**. Unauthorized local processes cannot harvest credentials.

5. **In-Memory Vault & Transparent Audit Trail**:
   - Cached passphrases reside strictly in volatile memory with configurable TTL (default 2 hours) and are securely overwritten on logout or shutdown.
   - All credential access events are recorded in an in-memory audit log accessible via the UI for complete transparency.

---

## 3. User Scenarios & Acceptance Criteria *(mandatory)*

### User Story 1 - Local Repository Identity Isolation (Priority: P1)

As a developer working on both enterprise and open-source projects, I want to switch the active Git Profile for the current repository from the bottom status bar, so that `user.name`, `user.email`, `signingkey`, and isolated `core.sshCommand` are updated only in `.git/config` without altering my global configuration.

**Why this priority**:
P1 MVP requirement. Prevents accidental exposure of corporate credentials in open-source repositories and eliminates key conflicts.

**Independent Test**:
Switch profile in a test repository and inspect `.git/config` versus `~/.gitconfig`. Confirm that only the local file reflects the changes.

**Acceptance Scenarios**:
1. **Given** an open repository, **When** switching to profile "Corporate Work", **Then** `.git/config` reflects the specified `user.name`, `user.email`, and `core.sshCommand = "ssh -i <key> -o IdentitiesOnly=yes"`.
2. **Given** global `~/.gitconfig`, **When** profiles are applied to workspaces, **Then** `~/.gitconfig` remains completely unchanged.

---

### User Story 2 - Headless AI Agent Step-by-Step Commit Support (Priority: P2)

As a developer relying on AI agents for automated coding, I want consecutive commits and push actions to proceed without blocking timeouts by leveraging secure session memory caching and host agent bridging.

**Why this priority**:
Enables unattended, multi-step agent workflows while keeping GPG-verified signatures intact.

**Independent Test**:
Execute automated GPG-signed commits in sequence; verify that the initial prompt caches the passphrase in memory and subsequent commits succeed instantaneously.

**Acceptance Scenarios**:
1. **Given** a GPG key with passphrase, **When** the user checks "Remember for this session", **Then** subsequent commits in the session sign cleanly without further prompts.
2. **Given** active `ssh-agent` with loaded keys, **When** a push is executed, **Then** it authenticates through the kernel agent with zero UI interruption.

---

### User Story 3 - Hardened Nonce IPC Defense & Audit Trail (Priority: P3)

As a security-conscious engineer, I want AskPass requests to be validated with ephemeral nonces and logged in an audit trail, so that unauthorized local processes cannot extract my credentials.

**Why this priority**:
Prevents local privilege escalation and side-channel sniffing on multi-user or shared development machines.

**Independent Test**:
Send a forged POST request to `/api/git/credential-internal-ask` without a valid nonce; verify that HTTP 403 Forbidden is returned and logged in the audit trail.

**Acceptance Scenarios**:
1. **Given** a forged request with an invalid nonce, **When** submitted to the askpass endpoint, **Then** it is rejected with 403 and logged as `denied_invalid_nonce`.
2. **Given** a valid single-use nonce, **When** used once, **Then** any immediate reuse of the same token is rejected.

---

## 4. Functional Requirements Matrix

### 4.1 Backend Services
- **FR-1.1**: Provide `GitProfileManager` for CRUD operations on profiles stored at `~/.config/unirtm/git_profiles.json`.
- **FR-1.2**: Implement `ApplyProfileToWorkspace(workspaceDir, profileId)` targeting strictly `.git/config`.
- **FR-1.3**: Read-only discovery of `~/.ssh/*.pub` and GPG public key rings (`gpg --list-public-keys`).
- **FR-1.4**: Host agent detection (`$SSH_AUTH_SOCK`, `ssh-add -l`, and `gpg-connect-agent /bye`).
- **FR-1.5**: Cryptographic nonce manager with single-use verification and immediate invalidation.
- **FR-1.6**: Ephemeral `CredentialAuditTrail` buffer retaining the last 100 access events.

### 4.2 Frontend Interfaces
- **FR-2.1**: `GitProfileStatus.vue` status bar widget displaying active profile name and GPG signing indicator with quick-pick switcher.
- **FR-2.2**: `GitIdentitySettings.vue` panel offering Profile CRUD, discovered key dropdowns, agent health status cards, and audit log drawer.
- **FR-2.3**: `useGitIdentity.ts` reactive composable managing state synchronization across components.
- **FR-2.4**: Full native localization across 53 supported languages with strict translation purity standards.

---

## 5. Verification & Quality Gates

1. **Unit Test Coverage**:
   - `internal/server/git_profile_test.go`: Profile CRUD, local isolation, and HTTP endpoints.
   - `internal/server/credential_test.go`: Nonce issuance, single-use invalidation, audit logging, and agent status.
   - `frontend/src/composables/useGitIdentity.test.ts`: State machine, API interaction, and active profile computation.
   - `frontend/src/components/common/GitProfileStatus.test.ts`: Formatting and tooltip presentation.
2. **Quality Guard Rules**:
   - Passes all 34 automated rules (no hardcoded secrets, no unformatted code, strict type checking via `vue-tsc`, zero translation leakage).
