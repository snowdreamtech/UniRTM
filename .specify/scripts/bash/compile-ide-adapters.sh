#!/usr/bin/env bash

# Universal IDE Adapter Compiler for Spec Kit
#
# This script reads the master markdown workflows from .specify/commands/
# and generates the necessary proxy files for various AI IDEs, including:
# - .agent/workflows/*.md (Universal proxy for generic IDEs)
# - .cursor/rules/*.mdc (Cursor specific rules)

set -e

SCRIPT_DIR="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

COMMANDS_DIR="$REPO_ROOT/.specify/commands"
AGENT_WORKFLOWS_DIR="$REPO_ROOT/.agent/workflows"
CURSOR_RULES_DIR="$REPO_ROOT/.cursor/rules"

# 1. Generate Universal Proxy Files in .agent/workflows/
mkdir -p "$AGENT_WORKFLOWS_DIR"

if [[ -d "$COMMANDS_DIR" ]]; then
    for cmd_file in "$COMMANDS_DIR"/*.md; do
        [[ -f "$cmd_file" ]] || continue
        base_name="$(basename "$cmd_file")"
        
        # Write universal markdown proxy
        cat <<EOF > "$AGENT_WORKFLOWS_DIR/$base_name"
---
description: Proxy for $base_name
---

## Execute Command

Please read \`.specify/commands/$base_name\` and execute its instructions exactly.
EOF
    done
fi

# 2. Generate Cursor MDC rules (if .cursor directory exists)
if [[ -d "$REPO_ROOT/.cursor" ]]; then
    mkdir -p "$CURSOR_RULES_DIR"
    
    if [[ -d "$COMMANDS_DIR" ]]; then
        for cmd_file in "$COMMANDS_DIR"/*.md; do
            [[ -f "$cmd_file" ]] || continue
            base_name="$(basename "$cmd_file")"
            cmd_id="${base_name%.md}"
            
            # Write Cursor .mdc proxy rule
            cat <<EOF > "$CURSOR_RULES_DIR/speckit_${cmd_id}.mdc"
---
description: Proxy for the ${cmd_id} workflow
globs: *
---

# Speckit Workflow: ${cmd_id}

When the user asks to run \`/${cmd_id}\`, you MUST read \`.specify/commands/${base_name}\` and follow its instructions exactly.
EOF
        done
    fi
fi
