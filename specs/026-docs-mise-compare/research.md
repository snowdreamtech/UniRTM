# Research: UniRTM vs Competitors

## Context

We need to provide detailed architectural comparisons highlighting UniRTM's zero-pollution philosophy, native implementation, and MCP capabilities compared to legacy and modern tools.

## Findings

### 1. nvm / n / fnm (Node.js)

- **Decision**: Highlight `corepack` integration and cross-platform native binary.
- **Rationale**: `nvm` uses bash scripts that heavily slow down shell startup time. `fnm` is faster but still requires shell hooks. UniRTM does not require slow `.zshrc` hooks and integrates naturally without polluting the user's environment. It natively supports `corepack` for `yarn`/`pnpm` out of the box.

### 2. gvm / goenv (Go)

- **Decision**: Highlight project-level scoping without global env pollution.
- **Rationale**: `gvm` compiles Go from source or downloads binaries but manipulates `GOPATH` and `GOROOT` globally in the shell. UniRTM dynamically injects environments at the process level using `.unirtm.toml`, meaning different directories can use different Go versions seamlessly without running any shell alias commands.

### 3. pyenv / pipx (Python)

- **Decision**: Explain UniRTM's native `venv` isolation logic.
- **Rationale**: `pyenv` only manages Python versions. To install CLI tools globally without conflict, users need `pipx`. UniRTM natively creates isolated `venv`s for global CLI tools (acting exactly like `pipx`) natively in Go, meaning you don't need to install `pipx` as a separate tool at all.

### 4. asdf (Any Language)

- **Decision**: Emphasize zero-dependency architecture.
- **Rationale**: `asdf` relies on community-maintained bash plugins. It requires dependencies like `curl`, `git`, `make`, etc., to be installed just to run plugins. UniRTM has built-in providers natively compiled in Go, meaning a single binary does everything with zero external dependencies.

### 5. mise (The Modern Competitor)

- **Decision**: Highlight the 100% native purist approach and MCP.
- **Rationale**: While `mise` (JDX) is written in Rust and is fast, it still falls back to legacy `asdf` plugins, relies on system `pipx` for python tools, and often requires `direnv`. UniRTM re-implements isolation natively (e.g., native `pipx` fallback), guarantees absolutely zero shell hook pollution if desired, and includes a built-in MCP server for AI Agent integration—something `mise` lacks.

### 6. direnv

- **Decision**: Emphasize declarative vs imperative environment variable injection.
- **Rationale**: `direnv` requires `.envrc` and manual `direnv allow` commands. It runs shell scripts on `cd`. UniRTM declaratively defines variables in `.unirtm.toml` and injects them safely without executing arbitrary shell code on directory change, improving security and performance.

### 7. VitePress I18n Auto-Redirect

- **Decision**: Implement a custom client-side router hook or use a lightweight script injected in the `<head>` of the root `index.html` to perform `navigator.language` detection.
- **Rationale**: VitePress natively supports localized routes (e.g., `/` for EN, `/zh/` for ZH), but it does not auto-redirect users visiting `/` based on their browser language by default. Injecting a small script in the configuration (`head` tags) or using a `setup` hook allows us to seamlessly redirect users with Chinese OS/Browser to the `/zh/` route on their first visit, meeting the FR-006 requirement without needing a backend server.

## Unknowns Resolved

- All tool comparison angles are now fully mapped to UniRTM's USPs.
