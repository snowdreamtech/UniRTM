#!/usr/bin/env bash

# Universal IDE Adapter Compiler for Spec Kit (V2)
#
# Automatically synchronizes .specify/commands/ to various AI IDE environments.
# Features:
# - Idempotent updates (only writes if changed)
# - Orphan cleanup (deletes proxies if source command was removed)
# - Universal injection (auto-appends global pointers for Cline, Roo, Windsurf, Trae, Copilot)

set -e

SCRIPT_DIR="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

COMMANDS_DIR="$REPO_ROOT/.specify/commands"
AGENT_WORKFLOWS_DIR="$REPO_ROOT/.agent/workflows"
CURSOR_RULES_DIR="$REPO_ROOT/.cursor/rules"

# Ensure target directories exist
mkdir -p "$AGENT_WORKFLOWS_DIR"

# Step 1: Collect Active Commands (macOS Bash 3.2 compatible string array)
ACTIVE_COMMANDS_STRING=" "

if [[ -d "$COMMANDS_DIR" ]]; then
    for cmd_file in "$COMMANDS_DIR"/*.md; do
        [[ -f "$cmd_file" ]] || continue
        base_name="$(basename "$cmd_file")"
        ACTIVE_COMMANDS_STRING="${ACTIVE_COMMANDS_STRING}${base_name} "
    done
fi

# Helper function to check if command is active
is_active_command() {
    local cmd_to_check="$1"
    if [[ "$ACTIVE_COMMANDS_STRING" == *" ${cmd_to_check} "* ]]; then
        return 0
    else
        return 1
    fi
}

# Helper function for idempotent writes
write_idempotent() {
    local target_file="$1"
    local content="$2"
    local temp_file="${target_file}.tmp.$$"

    echo "$content" > "$temp_file"
    
    if [[ ! -f "$target_file" ]] || ! cmp -s "$temp_file" "$target_file"; then
        mv "$temp_file" "$target_file"
    else
        rm "$temp_file"
    fi
}

# Helper function to inject global pointers into IDE-specific system prompts
ensure_global_pointer() {
    local target_file="$1"
    local content="
# --- Spec Kit AI IDE Integration ---
# This project uses Spec Kit. 
# CRITICAL: If you need to execute workflows or commands, refer to the files in \`.agent/workflows/\`.
# CRITICAL: For project governance and rules, refer to \`.agent/rules/00-index.md\`.
"
    
    # Check if the directory containing the target file exists (e.g. .github/)
    local dir_path="$(dirname "$target_file")"
    if [[ "$dir_path" != "." ]] && [[ ! -d "$REPO_ROOT/$dir_path" ]]; then
        return 0 # Skip if the IDE directory like .github doesn't exist
    fi

    local full_path="$REPO_ROOT/$target_file"
    
    # If the file doesn't exist, just create it
    if [[ ! -f "$full_path" ]]; then
        echo "$content" > "$full_path"
        return 0
    fi
    
    # If it exists, append only if our specific pointer string isn't present
    if ! grep -q "Spec Kit AI IDE Integration" "$full_path"; then
        echo "$content" >> "$full_path"
    fi
}

# Step 2: Generate Files & Clean Orphans for Generic Agents (.agent/workflows/)
if [[ -d "$COMMANDS_DIR" ]]; then
    for cmd_file in "$COMMANDS_DIR"/*.md; do
        [[ -f "$cmd_file" ]] || continue
        base_name="$(basename "$cmd_file")"
        content="---
description: Proxy for $base_name
---

## Execute Command

Please read \`.specify/commands/$base_name\` and execute its instructions exactly."
        
        write_idempotent "$AGENT_WORKFLOWS_DIR/$base_name" "$content"
    done
fi

# Clean Orphans in .agent/workflows/
for existing_file in "$AGENT_WORKFLOWS_DIR"/*.md; do
    [[ -f "$existing_file" ]] || continue
    base_name="$(basename "$existing_file")"
    if ! is_active_command "$base_name"; then
        rm "$existing_file"
    fi
done

# Step 3: Generate Files & Clean Orphans for Cursor (.cursor/rules/)
if [[ -d "$REPO_ROOT/.cursor" ]]; then
    mkdir -p "$CURSOR_RULES_DIR"
    
    # Generate
    if [[ -d "$COMMANDS_DIR" ]]; then
        for cmd_file in "$COMMANDS_DIR"/*.md; do
            [[ -f "$cmd_file" ]] || continue
            base_name="$(basename "$cmd_file")"
            cmd_id="${base_name%.md}"
            content="---
description: Proxy for the ${cmd_id} workflow
globs: *
---

# Speckit Workflow: ${cmd_id}

When the user asks to run \`/${cmd_id}\`, you MUST read \`.specify/commands/${base_name}\` and follow its instructions exactly."
            
            write_idempotent "$CURSOR_RULES_DIR/speckit_${cmd_id}.mdc" "$content"
        done
    fi
    
    # Clean Orphans in .cursor/rules/
    for existing_file in "$CURSOR_RULES_DIR"/speckit_*.mdc; do
        [[ -f "$existing_file" ]] || continue
        # Extract original base_name from speckit_*.mdc
        filename="$(basename "$existing_file")"
        cmd_id="${filename#speckit_}"
        cmd_id="${cmd_id%.mdc}"
        base_name="${cmd_id}.md"
        
        if ! is_active_command "$base_name"; then
            rm "$existing_file"
        fi
    done
fi

# Step 4: Inject Global Pointers for other mainstream AI IDEs
ensure_global_pointer ".clinerules"
ensure_global_pointer ".roo-rules"
ensure_global_pointer ".windsurfrules"
ensure_global_pointer ".traerules"
ensure_global_pointer ".github/copilot-instructions.md"

exit 0
