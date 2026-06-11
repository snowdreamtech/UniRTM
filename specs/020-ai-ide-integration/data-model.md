# Data Model

This feature is primarily architectural and directory-based, so there are no traditional database entities. However, the core conceptual "entities" of the architecture are mapping concepts:

### 1. IDE Pointer
- **Description**: A file at the repository root specific to an IDE.
- **Fields**: File path (`.cursorrules`, `.windsurfrules`, `.clinerules`)
- **Relationship**: Points directly to `.agent/rules/00-index.md`.

### 2. Workflow SOP (Standard Operating Procedure)
- **Description**: A Markdown file serving as an interface for AI agents.
- **Fields**: File path (`.agent/workflows/*.md`)
- **Relationship**: Wraps and references shell scripts in the `.specify/` engine.

### 3. Execution Script
- **Description**: The actual logic driving Spec Kit.
- **Fields**: File path (`.specify/scripts/bash/*.sh`)
- **Relationship**: Is invoked by the AI as instructed by the Workflow SOP.
