# 🚨 CRITICAL SYSTEM INSTRUCTION 🚨

Before answering ANY prompt or executing ANY code in this repository,
you **MUST** first read and strictly adhere to ALL the rules defined
in the following directory:

## Step 1: Read ALL Core Rule Files

📁 `.agent/rules/` — Read in numerical order:

1. `.agent/rules/01-general.md` — Language, communication, idempotency, cross-platform, network, security & audit
2. `.agent/rules/02-coding-style.md` — Commit messages, code quality, error handling, documentation, naming conventions
3. `.agent/rules/03-architecture.md` — Project structure, AI IDE integration, design principles
4. `.agent/rules/04-security.md` — Credentials, access control, encryption, scanning, incident response
5. `.agent/rules/05-dependencies.md` — Locking, integrity, auditing, release process, changelog
6. `.agent/rules/06-ci-testing.md` — Test types, CI pipeline, test data, fast feedback, quality gates
7. `.agent/rules/07-git.md` — Commits, branching, pull requests, code review, history hygiene
8. `.agent/rules/08-dev-env.md` — Environment consistency, dev container, scripts, pre-commit hooks, debugging
9. `.agent/rules/09-ai-interaction.md` — Safety boundaries, code generation, communication, context handling, quality
10. `.agent/rules/10-ui-ux.md` — Styling, componentization, accessibility, performance, i18n (frontend projects only)
11. `.agent/rules/11-deployment.md` — Containerization, secrets, deployment pipeline, IaC, observability & DR

## Step 2: Read Relevant Language & Framework Rule Files

After reading the core rules, **inspect the project's file structure and configuration files** (e.g., `package.json`, `go.mod`, `Cargo.toml`, `pom.xml`, `pyproject.toml`, `*.csproj`) to identify the languages and frameworks in use.

Then read the corresponding rule files from `.agent/rules/`.

## Why This File Exists

This project uses a **unified rule system** to ensure consistent AI behavior
across all AI-powered IDEs and tools (Cursor, Windsurf, GitHub Copilot,
Cline, Claude, Gemini, Trae, Roo Code, Augment, Amazon Q, Kiro,
Continue, Junie, etc.). This file is a redirect entry point —
the actual rules live in `.agent/rules/` as the Single Source of Truth.

> **Failure to follow the rules inside `.agent/rules/` is completely unacceptable.**

# --- Spec Kit AI IDE Integration ---
# This project uses Spec Kit. 
# CRITICAL: If you need to execute workflows or commands, refer to the files in `.agent/workflows/`.
# CRITICAL: For project governance and rules, refer to `.agent/rules/00-index.md`.

