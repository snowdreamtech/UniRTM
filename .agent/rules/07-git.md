# Git, Commit & PR Guidelines

> Objective: Unify commit history, branching, PR workflows, and review standards to improve traceability and collaboration efficiency.

## 0. Atomic Commit Discipline (MANDATORY)

> **Strictly enforced when implementing multiple suggestions or changes in sequence.**

- **One commit = one logical change.** Each commit must address a **single, well-defined topic** (e.g., one feature, one bug fix, one refactor, one config change). Never batch unrelated changes into a single commit.
- **Implement → Commit (NO Push) → Next.** When executing multiple suggestions or tasks in sequence, the AI agent **MUST** perform an automatic commit after each logical change, but **MUST NOT** perform an automatic push to the remote repository.
  1. Implement the first item completely.
  2. Run lint/format auto-fixes.
  3. **Commit immediately** (local only) with a precise Conventional Commit message.
  4. Only then proceed to the next item.
- **Auto-Commit is Mandatory (No `--no-verify`).** AI agents must never wait for the end of a session or a batch of changes to commit. Every feature implementation or bug fix must result in an immediate, independent commit. The use of `--no-verify` to bypass pre-commit hooks or commit-msg hooks is strictly prohibited. If a hook fails, you MUST fix the underlying code or configuration.

```bash
# ✅ CORRECT: atomic, auditable
git commit -m "feat(auth): add OAuth2 login"
# ... implement next item ...
git commit -m "feat(auth): add logout endpoint"

# ❌ WRONG: batch commit, hard to audit
git commit -m "feat(auth): add OAuth2 login and logout and token refresh"
```

## 1. Commit Messages

- Use **Conventional Commits** format (strictly following `@commitlint/config-conventional`): `<type>(<scope>): <description>` (e.g., `feat(auth): add OAuth2 login`). Never write vague commit messages like "fix bug", "update code", or "WIP".
- Common types:

  | Type       | Purpose                            | Example                                     |
  | ---------- | ---------------------------------- | ------------------------------------------- |
  | `feat`     | New feature                        | `feat(user): add avatar upload`             |
  | `fix`      | Bug fix                            | `fix(auth): prevent refresh token reuse`    |
  | `docs`     | Documentation only                 | `docs(api): add rate limit description`     |
  | `style`    | Formatting, whitespace             | `style: fix trailing whitespace in models`  |
  | `refactor` | Restructuring (no behavior change) | `refactor(db): extract query builder`       |
  | `test`     | Adding or updating tests           | `test(order): add edge case for empty cart` |
  | `chore`    | Maintenance, tooling               | `chore: upgrade eslint to v9`               |
  | `ci`       | CI/CD pipeline changes             | `ci: add matrix for Node 22`                |
  | `perf`     | Performance optimization           | `perf(cache): use Redis pipelining`         |
  | `build`    | Build system changes               | `build: switch to esbuild`                  |
  | `revert`   | Reverts a previous commit          | `revert: feat(auth): add OAuth2 login`      |

- Commit messages **MUST** be in **English only** (no Chinese characters or punctuation allowed). The **header** (the entire first line: `<type>(<scope>): <description>`) must be concise (**max 120 characters**), written in the imperative mood ("add", not "added"), and **MUST NOT** end with a period (full stop).
- **Detail Offloading**: If you need to provide more context or detailed information beyond the 120-character limit, **MUST** put it in the **body** section after a blank line. The subject line should remain a high-level summary. The body and footer lines **do not have a strict line-length limit** in order to accommodate AI-generated messages, URLs, or error stack traces.
- While the traditional Git limit is 72, we strictly follow an extended `@commitlint/config-conventional` limit of **120 characters** for the header, but disable line length limits for the body and footer.

  ```
  feat(auth): add OAuth2 login with Google provider

  Implements OAuth2 Authorization Code flow using the `passport-google-oauth20`
  strategy. Adds /auth/google and /auth/google/callback routes.

  Users can now log in with Google alongside existing email/password auth.

  Closes: #234
  BREAKING CHANGE: /auth/login response shape changed — see migration guide.
  ```

- Where the repository policy requires it, **sign commits with GPG** (`git commit -S`). Enforce signature verification on protected branches via branch protection rules.
- Configure `.gitattributes` to normalize line endings and designate merge drivers for generated files:

  ```gitattributes
  * text=auto
  *.sh  text eol=lf
  *.ps1 text eol=crlf
  *.png binary
  package-lock.json merge=ours   # prevent merge conflicts in auto-generated lock files
  ```

## 2. Commit Message Troubleshooting (Commitlint)

If your commit fails CI or local hooks, check for these common violations:

- **header-max-length**: First line must be **≤ 120 characters**. If it exceeds this, move extra details to the **body**. (The 120 character limit applies strictly to the header; body and footer have no rigid length limits to allow for long URLs).
- **subject-case**: Description after `<type>(<scope>):` must be **all lowercase** (except acronyms). Never start with a Capital letter.
- **subject-full-stop**: The subject line **must not** end with a period (`.`).
- **subject-empty**: The description part is required.
- **type-enum**: Only use allowed types (see table above).
- **no-chinese**: Commit message must be in English only (no Chinese characters or punctuation allowed).
- **Infinitive Mood**: Use `feat: add` instead of `feat: added` or `feat: adding`.

**Quick Fix for Last Commit:**

```bash
git commit --amend -m "type(scope): correct commit message"
```

## 3. Branching Strategy

- Use a feature branch workflow with clear, consistent branch naming:

  | Branch Pattern                       | Purpose                                   |
  | ------------------------------------ | ----------------------------------------- |
  | `feature/<ticket>-short-description` | New feature development                   |
  | `fix/<ticket>-short-description`     | Bug fixes                                 |
  | `release/x.y.z`                      | Release stabilization                     |
  | `hotfix/<ticket>-short-description`  | Emergency production patches              |
  | `chore/<description>`                | Maintenance (dependency updates, tooling) |

- Branch names must be **lowercase with hyphens**. Include ticket/issue numbers where applicable: `feature/issue-234-oauth-login`.
- Before merging into `main`, the branch MUST:
  - Pass CI completely (all jobs green)
  - Receive approval from at least 1 reviewer (preferably 2 for critical paths)
  - Have no unresolved comments blocking merge
  - Be up-to-date with the base branch (rebased or merged)
- Delete branches after merging — keep the remote repository clean. Enforce with GitHub/GitLab "Delete head branch" on merge settings.
- **Monorepo strategy**: use path-scoped branch names (`feature/auth-<ticket>`) and only run CI for affected packages using Nx, Turborepo, or Bazel affected detection.

## 3. Pull Request Requirements

- PR titles MUST follow the Conventional Commits format for automated changelog generation from PR merge commits.
- PR descriptions MUST include all relevant sections. Use a PR template (`.github/pull_request_template.md`):

  ```markdown
  ## Summary

  <!-- What does this PR change and why? -->

  ## Testing Performed

  <!-- Unit tests, manual test steps, affected environments -->

  ## Related Issues

  Closes: #234

  ## Rollback Plan

  <!-- For risky changes: how to revert if something goes wrong -->

  ## Checklist

  - [ ] Tests added/updated
  - [ ] Documentation updated
  - [ ] No secrets or PII in code/tests
  - [ ] Changelog entry added (if applicable)
  ```

- Use **Draft PRs** for work-in-progress. Convert to "Ready for Review" only when CI passes and the implementation is complete. Do not request reviews on Draft PRs.
- PRs related to **security**, **dependencies**, or **database migrations** MUST:
  - Explicitly list impacts and risks
  - Include tags: `security`, `db-migration`, `breaking-change`
  - Document a rollback procedure
  - Get approval from a second reviewer
- Keep PRs small and focused: target **< 400 lines changed per PR**. Large PRs MUST be broken into smaller, logically independent PRs linked by a parent tracking issue. Gigantic PRs are rarely reviewed well.

## 4. Code Review Standards

- Reviewers MUST check for: correctness, test coverage adequacy, security implications, performance impact, and adherence to project conventions.
- Use **blocking** review comments (`Request changes`) for required changes. Use non-blocking comments (prefixed with `nit:` or `suggestion:`) for optional improvements that should not block merge:

  ```
  # Blocking (must be addressed before merge)
  This function doesn't validate the input — malicious data could cause a SQL injection.

  # Non-blocking (nice to have)
  nit: Consider extracting this into a helper function for reusability.
  suggestion: We could use a Set here instead of an Array for O(1) lookup.
  ```

- **Review SLA**: reviewers MUST respond within **48 business hours** of being requested. If no response within SLA, the PR author may escalate or seek an alternate reviewer.
- Authors MUST respond to all review comments and resolve them before requesting re-review. Do not force-push to a branch under active review without notifying reviewers.
- **AI-assisted review**: treat AI suggestions as a first-pass only. Human reviewers remain accountable for all approved changes. Never merge based solely on an AI tool's approval.

## 5. Workspace Hygiene & Forbidden Files

To maintain a performant repository and a clean development environment, certain file types are strictly prohibited from being staged or committed. These rules are enforced by native pre-commit hooks and must be followed manually if hooks are not present.

- **Temporary/Debug Files**: Never commit files used for local testing or debugging (e.g., `test.txt`, `tmp.json`, `debug.log`).
- **Build Artifacts & Caches**: Binary outputs (`dist/`, `build/`, `out/`), dependency caches (`node_modules/`, `.venv/`), and compiler artifacts (`__pycache__/`, `.pytest_cache/`) must stay out of Git.
- **Merge Remnants**: Always clean up `*.orig`, `*.rej`, or other remnants from `git merge` or `git rebase` before committing.
- **OS/Editor Noise**: Exclude system-specific metadata like `.DS_Store` (macOS), `Thumbs.db` (Windows), or swap files (`.swp`, `.swo`).
- **Git Markers in Text**: Specifically for large lockfiles (e.g., `package-lock.json`), always ensure no unresolved merge conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`) exist, even if the file size exceeds normal linting limits.
- **Encoding Standards**: All text files must be clean UTF-8. The Byte Order Mark (BOM: `\xef\xbb\xbf`) is considered an illegal character and will be blocked.

```bash
# ✅ CORRECT: Clean workspace, precise staging
git add src/ models/
git status # Review to ensure no "test.txt" or ".DS_Store" is listed

# ❌ WRONG: Staging everything indiscriminately
git add .
git commit -m "update everything" # High risk of committing illegal/temp files
```

## 6. History Hygiene

- Use **`rebase`** (not merge commits) to integrate upstream changes into feature branches to maintain a linear, readable history:

  ```bash
  git fetch origin
  git rebase origin/main   # rebase feature branch onto latest main
  git push --force-with-lease origin feature/my-branch   # force-push rewritten history
  ```

  Use `--force-with-lease` (never bare `--force`) to prevent overwriting others' work.
- **Squash** trivial/fixup commits (`WIP`, `fix typo`, `address review comment`) before merging. Preserve meaningful individual commits that tell the story of a change. Configure **Squash and Merge** as the default merge strategy on GitHub/GitLab for repositories preferring linear history.
- Never rewrite history on shared branches (`main`, `develop`, `release/*`). Force-push is only allowed on personal feature branches with prior team communication.
- Store large binary files (videos, datasets, compiled artifacts > 50 MB) in **Git LFS**:

  ```bash
  git lfs install
  git lfs track "*.bin" "*.mp4" "*.zip" "model-weights/**"
  git add .gitattributes
  ```

  Document LFS usage in the project README and include LFS setup in onboarding instructions.
- Tag every production release with a **signed, annotated tag**:

  ```bash
  git tag -a -s v1.2.3 -m "Release v1.2.3: Add OAuth2 login, fix token refresh race"
  git push origin v1.2.3
  ```

  Enforce tag signing verification via CI: `git tag -v v1.2.3`.

## 7. Case Sensitivity (core.ignorecase)

To prevent rebase conflicts and history fragmentation across different operating systems (especially on macOS and Windows), this project mandates a **case-sensitive** Git configuration.

- **Mandatory Configuration**: You MUST set `core.ignorecase` to `false` in your local project repository:

  ```bash
  git config core.ignorecase false
  ```

- **Renaming Files**: Since the underlying file systems on macOS (APFS) and Windows (NTFS) are case-insensitive, simply renaming a file via your file explorer or IDE may not be detected by Git. You MUST use the `git mv` command for all case-only renames:

  ```bash
  # ✅ CORRECT: Git tracks the case change explicitly
  git mv oldname.txt OldName.txt

  # ❌ WRONG: Git may not see the change on case-insensitive filesystems
  mv oldname.txt OldName.txt
  ```

- **Verification**: Ensure your configuration is active by running `git config core.ignorecase`. It MUST return `false`.

## 8. Cross-Platform File Mode Tracking (core.filemode)

To prevent false file permission modifications on non-POSIX filesystems (especially Windows NTFS/FAT), configure `core.filemode`:

- **Windows Compatibility**: Windows filesystems do not natively support POSIX executable bits (`chmod +x`). When cloning or checking out repositories containing executable scripts (`.sh`, `.py`), Git on Windows may falsely flag them as modified (`100755` -> `100644`).
- **Configuration**: Windows developers SHOULD set `core.filemode` to `false` locally (or globally):

  ```bash
  git config core.filemode false
  ```

- **POSIX Environments (Linux / macOS)**: Keep `core.filemode` set to `true` (default) so that script execution permissions are tracked accurately in Git index.

## 9. Environment-Agnostic Configurations (No Local Absolute Paths)

- **System-Specific Tools**: Never hardcode platform-specific absolute binary paths (such as `gpg.program = /usr/local/bin/gpg` or `C:/Program Files/...`) in repository-level configurations, hooks, or tracked files.
- **Dynamic Resolution**: Always rely on standard system `PATH` resolution or configure system-dependent tools in user-level global configuration (`~/.gitconfig`).
