# Quickstart / Validation Guide: Documentation Upgrade

## Prerequisites

- A local checkout of the UniRTM repository on the `026-docs-mise-compare` branch.
- Node.js installed to run the VitePress dev server.

## Setup

1. Change into the docs directory (if applicable) or root:

   ```bash
   npm install
   ```

## Validation Scenarios

### Scenario 1: Verify the Local Dev Server

Run the documentation site locally:

```bash
npm run docs:dev
```

- Open `http://localhost:5173` in your browser.
- Verify the site loads properly.

### Scenario 2: Verify Structural Alignment in Guide

- Navigate to the **Guide -> Introduction** page.
- Verify the presence of GitHub Alerts / highlighted blocks mentioning the **100% Native Architecture**, **Zero-Pollution Philosophy**, and **MCP Capabilities**.
- Verify the **Getting Started** page accurately reflects the simplified installation process.

### Scenario 3: Verify the New Comparisons Section

- Navigate to the newly created **替代与对比 (Comparisons)** section in the sidebar.
- Verify that each tool (`nvm`, `gvm`, `pyenv/pipx`, `asdf`, `mise`, `direnv`) has its own dedicated subsection.
- Ensure the layout is clear and readable.

### Scenario 4: Verify Bilingual Support

- Switch the language toggle to English/Chinese.
- Ensure the structural changes and the new `comparisons.md` exist and are correctly translated in both `docs/` and `docs/zh/`.
