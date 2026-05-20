# Automation Scripts

[English](README.md) | [简体中文](README_zh-CN.md)

> [!NOTE]
> This directory contains the implementation of the project's automation infrastructure using POSIX shell scripts (`.sh`) that work across all platforms including Windows (via Git Bash or WSL).

## 1. Design & Architecture

### Overview

This component provides a suite of cross-platform scripts to manage the development lifecycle, including environment setup, dependency installation, linting, testing, and deployment.

- **Portable**: Written in POSIX-compliant shell to ensure consistency across Linux, macOS, Windows (Git Bash/WSL), and CI environments.
- **Robust**: Includes safety guards, atomic operations, and standardized error handling.
- **Cross-Platform**: Single source of truth with POSIX shell scripts, no platform-specific wrappers needed.

### Architecture

```text
[ User / Developer / CI ]
        |
    [ Makefile ] (Convenience Entry)
        |
        v
    [ scripts/*.sh ] (POSIX Shell Scripts)
    /      \
  /         \
[lib/common.sh] [lib/langs/*.sh]
(Utils/SSoT)    (Lang Modules)
```

### Design Principles

- **Cross-Platform**: POSIX-compliant shell scripts work on Linux, macOS, and Windows (Git Bash/WSL)
- **Auditable**: Every decision is traceable via detailed logging and exit codes
- **Idempotent**: Scripts can be run multiple times safely without side effects
- **Fail-Fast**: Immediate exit on error with clear diagnostics
- **Lean**: Zero-dependency for core logic where possible (uses `sh`, `sed`, `awk`)

### Responsibilities

- **Environment Provisioning**: Installing runtimes and quality tools (`setup.sh`).
- **Dependency Management**: Standardizing installation across stacks (`install.sh`).
- **Quality Assurance**: Orchestrating linting, testing, and security audits (`verify.sh`).
- **Lifecycle Automation**: Handling commits, releases, and changelog archival.

## 2. Usage Guide

### Prerequisites

- **Linux/macOS**: Built-in POSIX shell
- **Windows**: Git Bash (included with [Git for Windows](https://git-scm.com/download/win)) or WSL
- **Make**: Optional, provides convenient entry points

> [!TIP]
> Windows users: See [Windows Shell Migration Guide](../docs/migration/windows-shell-migration.md) for setup instructions.

### Quick Start

```bash
# 1. Setup development environment
sh scripts/setup.sh

# 2. Install project dependencies
sh scripts/install.sh

# 3. Verify environment health
sh scripts/verify.sh
```

### Script Reference

| Script                 | Purpose                       | Key Modules                   |
| :--------------------- | :---------------------------- | :---------------------------- |
| `setup.sh`             | Install system-level tools    | 20+ languages (via lib/langs) |
| `install.sh`           | Install project dependencies  | pnpm, pip, pre-commit         |
| `verify.sh`            | Full project verification     | lint, test, audit             |
| `lint.sh`              | Run linters and fixers        | pre-commit, auto-fix          |
| `test.sh`              | Execute test suites           | bats, pytest, vitest          |
| `audit.sh`             | Security & vulnerability scan | gitleaks, osv-scanner         |
| `release.sh`           | Standardized tagged release   | git tag (v-prefix), auto-sync |
| `docs.sh`              | Manage documentation site     | vitepress                     |
| `init-project.sh`      | Rebrand template              | placeholders, git init        |
| `archive-changelog.sh` | Archive old versions          | major version rotation        |

### Workflow Patterns

1. **Onboarding**: `setup.sh` → `install.sh` → `verify.sh`.
2. **Daily Development**: Work → `lint.sh` → `test.sh`.
3. **Continuous Integration**: `test.sh` → `audit.sh`.

### Directory Structure

- `scripts/`: Primary automation entry points (POSIX shell scripts)
- `scripts/lib/`: Internal core library (`common.sh`)
- `scripts/lib/langs/`: Modular language-specific logic (sourced by `common.sh`)

## 3. Operations Guide

### Pre-deployment Checklist

1. [ ] Run `sh scripts/check-env.sh` to ensure runtime parity.
2. [ ] Run `sh scripts/verify.sh` for final QA pass.
3. [ ] Run `sh scripts/audit.sh` to ensure no secrets or critical vulnerabilities.
4. [ ] Verify versioning: Run `make release` (Audits local state and syncs manifest version).

### Performance Considerations

- **Unified Dashboard**: Displays a comprehensive summary of version statuses and installation durations.
- **Caching**: Virtualenvs (`.venv`) and `node_modules` are reused across runs.

### Troubleshooting

- **Problem**: Permission Denied
  - **Solution**: Run `chmod +x scripts/*.sh`
- **Problem**: Command not found on Windows
  - **Solution**: Install Git Bash or WSL. See [migration guide](../docs/migration/windows-shell-migration.md)
- **Problem**: 404/Network failure during download
  - **Diagnosis**: Check if `GITHUB_PROXY` is reachable
  - **Solution**: Configure `GITHUB_PROXY` in `scripts/lib/common.sh` or env

### Maintenance Procedures

- **Tool Updates**: Run `sh scripts/update.sh` weekly.
- **Cache Purge**: Run `sh scripts/cleanup.sh` to reclaim disk space.

## 4. Security Considerations

### Security Model

- **Principle of Least Privilege**: Scripts install tools to `.venv` or user directory; no `sudo` required.
- **Validation**: Sub-scripts include checksum verification for external assets.
- **Secret Hygiene**: Gitleaks integration prevents accidental credential leaks.

### Best Practices

| Aspect           | Requirement                | Implementation           |
| :--------------- | :------------------------- | :----------------------- |
| File Permissions | Scripts must be executable | `chmod 755 scripts/*.sh` |
| Secret Integrity | No hardcoded keys          | Source from `.env` only  |
| Proxy Handling   | Secure download gateway    | `GITHUB_PROXY` via TLS   |

## 5. Development Guide

### Code Organization

- New logic MUST be added to a new `.sh` file in `scripts/`.
- Shared utility functions MUST be placed in `scripts/lib/common.sh`.
- Variables MUST be localized using `local` inside functions.

### Contribution Requirements

1. All new scripts MUST include a `main()` function and call `parse_common_args`
2. All new scripts MUST pass `shellcheck --shell=sh`
3. All new scripts MUST be POSIX-compliant for cross-platform compatibility
4. Keep English/Chinese README versions in sync

### Local Development Setup

1. Install ShellCheck: `brew install shellcheck` (macOS) or see [shellcheck.net](https://www.shellcheck.net/)
2. Run `make lint` to verify compliance

### References

- [POSIX Standard](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html)
- [ShellCheck Wiki](https://github.com/koalaman/shellcheck/wiki)
- [Git for Windows](https://git-scm.com/download/win)
- [Windows Shell Migration Guide](../docs/migration/windows-shell-migration.md)
