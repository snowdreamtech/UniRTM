# Data Model: Documentation Upgrade

*Note: This feature revolves around static documentation generation (Markdown + VitePress). There are no traditional database or domain data models involved.*

## Entities

### 1. `comparisons.md` (Document Entity)

- **Attributes**: Features one-on-one architectural comparisons between UniRTM and competitors (nvm, gvm, pyenv, asdf, direnv, mise).
- **Constraints**: Tone must be respectful; no trash-talking. Must highlight UniRTM's native architecture.

### 2. `architecture.md` (Document Entity)

- **Attributes**: Deep code-level breakdown of UniRTM's engine. Includes Mermaid diagrams.

### 3. `cli/overview.md` (Document Entity)

- **Attributes**: Exhaustive enumeration of all commands.
- **Constraints**: Every single command must be detailed (can be collapsible).

### 4. `environments/overview.md` (Document Entity)

- **Attributes**: Deep dive into `.unirtm.toml`, `.env`, and context swapping.
