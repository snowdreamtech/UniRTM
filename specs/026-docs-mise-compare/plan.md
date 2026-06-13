# Implementation Plan: Documentation Upgrade: Aligning and Surpassing `mise`

**Branch**: `026-docs-mise-compare` | **Date**: 2026-06-13 | **Spec**: [spec.md](./spec.md)

## Summary

This plan outlines a massive overhaul of the UniRTM documentation to completely align with and surpass the depth, breadth, and quality of `mise`'s documentation. Based on user feedback, the previous iterations were too superficial ("假大空") and lacked technical depth. We will introduce detailed architectural deep-dives, exhaustive CLI reference pages, in-depth environment management guides, and correctly position external security tools as integrations rather than core features.

## Proposed Changes

### 1. Architecture & Caching Deep Dive

We will replace the superficial `structure.md` with a comprehensive `architecture.md` (and its Chinese counterpart) that includes:

- **Code-Level Deep Dive**: Detailed breakdown of UniRTM's 100% native Go architecture, eliminating bash/Ruby plugins.
- **Structural Diagrams**: Mermaid flowcharts illustrating the command layer, backend plugin system, tool resolution pipeline, and parallel execution engine.
- **Caching Strategy**: Deep dive into caching mechanisms (e.g., how remote versions are cached, environment diff caching, parallel download caching, and TTL auto-pruning).

### 2. Exhaustive CLI Overview

We will rewrite `docs/cli/overview.md` and `docs/zh/cli/overview.md` from scratch.

- Instead of a single sentence, it will be an exhaustive, collapsible list of ALL UniRTM commands (e.g., `install`, `use`, `run`, `env`, `ls`, `ls-remote`, `exec`, `outdated`, etc.).
- Each command will feature a detailed introduction, usage syntax, and real-world examples.

### 3. In-Depth Environments Guide

We will rewrite `docs/environments/overview.md` and `docs/zh/environments/overview.md`.

- Deep explanation of the `.unirtm.toml` environment variable syntax, dynamic environment evaluation, and cross-platform `.env` file ingestion.
- Explain how UniRTM achieves "zero shell pollution" dynamically upon directory entry/exit.

### 4. Correcting "Secure by Default" Marketing

We will review and update `getting-started.md` and other promotional materials.

- Trivy, Syft, and Gitleaks must be explicitly documented as **external integrations** that UniRTM supports seamlessly, rather than exaggerating them as native built-in features.
- Eliminate any boastful marketing fluff and replace it with objective technical facts.

## Verification Plan

### Automated Tests

- `npm run docs:build` inside the `docs/` directory to ensure VitePress compiles successfully with the new mermaid diagrams and extensive markdown changes.
- Check for broken markdown links across all newly linked documentation pages.

### Manual Verification

- Manually inspect the generated VitePress site locally (`npm run docs:dev`) to ensure diagrams render correctly.
- Ensure the CLI page exhaustively covers all core commands.
- Ensure the tone remains respectful towards `mise` while technically highlighting where UniRTM's native architecture excels.
