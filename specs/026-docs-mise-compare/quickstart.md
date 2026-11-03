# Quickstart Validation Guide: Documentation Upgrade

This guide explains how to validate the documentation changes end-to-end.

## Prerequisites

- Node.js installed
- VitePress dependencies installed in `docs/`

## Setup Commands

```bash
cd docs
npm install
```

## Run/Validation Commands

### 1. Verify Local Dev Server & Diagram Rendering

```bash
npm run docs:dev
```

- **Expected Outcome**: The dev server starts. Navigate to `http://localhost:5173/UniRTM/guide/architecture.html` and `http://localhost:5173/UniRTM/zh/guide/architecture.html`. Verify that the Mermaid flowcharts and structural diagrams render correctly.

### 2. Verify Content Exhaustiveness

- **Expected Outcome**: Navigate to the CLI overview (`/cli/overview.html`) and verify that all commands are listed exhaustively with detailed descriptions. Navigate to the Environments overview (`/environments/overview.html`) to ensure the content is in-depth and not just 1-2 sentences.

### 3. Verify Static Build Integrity

```bash
npm run docs:build
```

- **Expected Outcome**: The VitePress build completes successfully without any broken link warnings or syntax errors in the markdown files.
