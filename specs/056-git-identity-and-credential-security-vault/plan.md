# Technical Architecture & Implementation Plan: Git Identity & Credential Security Vault

**Feature Branch**: `056-git-identity-and-credential-security-vault`
**Spec Document**: [spec.md](spec.md)
**Status**: Implemented

---

## 1. System Architecture

```text
+---------------------------------------------------------------------------------------+
|                                    UniRTM Frontend                                    |
|  [Status Bar: Active Profile]   [Settings: Git Identity & Vault]   [Passphrase Modal] |
|                                |                                          ^           |
+--------------------------------|------------------------------------------|-----------+
                                 | REST / SSE                               |
+--------------------------------v------------------------------------------v-----------+
|                                  UniRTM Go Backend                                    |
|                                                                                       |
|  +------------------------+  +-------------------------+  +------------------------+  |
|  |   Git Profile Manager  |  |  Native Agent Detector  |  |  Secure Memory Keyring |  |
|  | (Local .git/config)    |  |  (ssh-agent, gpg-agent) |  |  (TTL & Secure Zero)   |  |
|  +------------------------+  +-------------------------+  +------------------------+  |
|               |                                                        |              |
|               v                                                        v              |
|  +---------------------------------------------------------------------------------+  |
|  |             Hardened Credential Coordinator (One-Time Nonce IPC)                |  |
|  |     Injects UNIRTM_CREDENTIAL_NONCE, verifies token, logs to Audit Trail        |  |
|  +---------------------------------------------------------------------------------+  |
+---------------------------------------------------------------------------------------+
        |                                                        |
        | Spawns with AskPass / Loopback                         | Localhost IPC
        v                                                        v
+-------------------------------+              +-------------------------------------+
| git commit -S / git push/fetch| ------------>|  unirtm-askpass / gpg-loopback shim |
+-------------------------------+              +-------------------------------------+
        |
        +-----> Calls OS Native ssh / gpg binaries (Zero Private Key Touch)
```

---

## 2. Data Contracts & Structures

### 2.1 Git Profile Entity
```go
type GitProfile struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    UserName    string `json:"userName"`
    UserEmail   string `json:"userEmail"`
    SSHKeyPath  string `json:"sshKeyPath,omitempty"` // Public or private key path for core.sshCommand
    SigningKey  string `json:"signingKey,omitempty"` // GPG Key ID or fingerprint
    GPGSign     bool   `json:"gpgSign"`              // Automatically sign commits
    IsDefault   bool   `json:"isDefault"`
    UpdatedAt   int64  `json:"updatedAt"`
}
```

### 2.2 Credential Audit Record
```go
type CredentialAuditRecord struct {
    ID        string `json:"id"`
    Timestamp int64  `json:"timestamp"`
    Action    string `json:"action"`    // "askpass_prompt", "ssh_push", "gpg_sign"
    Type      string `json:"type"`      // "ssh", "gpg", "git-auth", "security_violation"
    TargetKey string `json:"targetKey"` // Key path or masked fingerprint
    Result    string `json:"result"`    // "success", "cached", "denied_invalid_nonce", "timeout"
    ClientIP  string `json:"clientIp"`
}
```

### 2.3 Host Native Agent Status
```go
type NativeAgentStatus struct {
    SSHAgentActive bool   `json:"sshAgentActive"`
    SSHAgentSock   string `json:"sshAgentSock,omitempty"` // Masked filename
    SSHAgentKeys   int    `json:"sshAgentKeys"`
    GPGAgentActive bool   `json:"gpgAgentActive"`
}
```

---

## 3. API Contract Specification

1. `GET /api/git/profiles`
   - Returns all configured identity profiles and the active profile ID.
2. `POST /api/git/profiles`
   - Creates or updates an identity profile in `git_profiles.json`.
3. `DELETE /api/git/profiles?id={id}`
   - Removes the specified identity profile.
4. `POST /api/git/profile/apply`
   - Payload: `{ "profileId": "...", "workspaceDir": "..." }`
   - Executes atomic local configurations:
     - `git -C <dir> config --local user.name "<name>"`
     - `git -C <dir> config --local user.email "<email>"`
     - `git -C <dir> config --local user.signingkey "<key>"` (or unset)
     - `git -C <dir> config --local commit.gpgsign true|false`
     - `git -C <dir> config --local core.sshCommand "ssh -i <path> -o IdentitiesOnly=yes"`
5. `GET /api/git/system-keys`
   - Read-only enumeration of `~/.ssh/*.pub` and `gpg --list-public-keys` for UI dropdowns.
6. `GET /api/git/agent-status`
   - Diagnoses active state of host `ssh-agent` and `gpg-agent`.
7. `GET /api/git/credential-audit`
   - Retrieves the last 100 credential prompt access and authorization logs.
8. `POST /api/git/credential-internal-ask`
   - Requires `X-Unirtm-Nonce` header. Verifies against active nonce store and invalidates upon consumption.

---

## 4. Frontend Component Design

1. **`useGitIdentity.ts`**:
   - Central reactive composable managing profiles, system key discovery, agent probes, and audit logs.
2. **`GitProfileStatus.vue`**:
   - Compact status bar widget featuring quick-pick switcher and GPG status indicator.
3. **`GitIdentitySettings.vue`**:
   - Comprehensive settings tab containing Profile CRUD, system agent status cards, and security audit log table.
4. **Internationalization**:
   - 32 translation keys synchronized across all 53 languages in `frontend/src/i18n/locales/`.

---

## 5. Security & Verification Strategy

- **Zero Global Mutation**: Tested via `TestGitProfileManager_ApplyProfileToWorkspace` ensuring `~/.gitconfig` is never touched.
- **Nonce Single-Use Verification**: Tested via `TestCredentialNonceManager` and `TestCredentialAskPass_NonceSecurity`.
- **Quality Gates**: Validated against `npm run test`, `vue-tsc`, `vite build`, and Go test suites.
