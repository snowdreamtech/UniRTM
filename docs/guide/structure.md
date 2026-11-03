# Project Structure

A clean overview of the UniRTM repository layout and key components.

## Top-Level Overview

```text
unirtm/
├── cmd/                      # 🚀 CLI Commands (Cobra command hierarchy & subcommands)
├── internal/                 # ⚙️ Core internal implementation
│   ├── config/               # Configuration parsing and schema validation
│   ├── database/             # SQLite state persistence and migrations
│   ├── guard/                # Quality Guard harness & built-in verification rules
│   ├── hook/                 # Git hook integration & hybrid runner
│   ├── provider/             # Multi-language tool providers (Go, Node, Python, etc.)
│   ├── repository/           # Data access repositories
│   └── server/               # Embedded REST / WebSocket API server
├── frontend/                 # 🖥️ Web Workbench GUI (Vue 3, TypeScript, Tailwind)
├── specs/                    # 📋 Specification & architecture deliverables (001–056)
├── docs/                     # 📖 VitePress documentation site
│
├── .github/                  # 🐙 GitHub integration & CI/CD workflows
├── .devcontainer/            # 🐳 DevContainer configuration
├── .vscode/                  # 💻 VS Code workspace settings
│
├── .unirtm.toml              # 🔧 Repository toolchain and quality guard contract
├── unirtm.lock               # 🔒 Pinned tool versions and integrity hashes
├── go.mod                    # 📦 Go module dependencies
├── Makefile                  # 🛠️ Build and development automation
│
├── README.md                 # Project overview and quickstart
├── CONTRIBUTING.md           # Contributor guide
├── CODE_OF_CONDUCT.md        # Community standards
├── SECURITY.md               # Security policy
└── LICENSE                   # MIT License
```

## AI Agent Integration: Zero Repository Pollution

UniRTM employs a **Zero-Footprint Architecture** for AI IDEs and agents:

- **Model Context Protocol (MCP)**: Instead of polluting project roots with 50+ proprietary dot-directories (`.cursor/`, `.windsurf/`, `.cline/`), UniRTM provides a built-in standard MCP server via `unirtm mcp`.
- **Global Rules & Workflows**: AI coding rules and Spec Kit workflows are provided dynamically to AI agents through standard MCP prompts, resources (`unirtm://config`, `unirtm://conventions`), and tools (`quality_check`, `get_rule`).
- **Clean Repositories**: Any AI IDE or agent (Cursor, Windsurf, Claude Desktop, Zed, UniAgent) connects via standard MCP over stdio without committing any tool-specific proxy files.
