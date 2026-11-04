# Data Model: Hook Install Hybrid Resolution

## Constants

The script injection will rely on two primary string markers:

- `managedBlockStart` = `"# --- BEGIN UNIRTM MANAGED BLOCK ---"`
- `managedBlockEnd` = `"# --- END UNIRTM MANAGED BLOCK ---"`

## Injection Payload

```bash
# Auto-load UniRTM environment for Headless (AI/CI) and GUI Git Clients
if ! command -v unirtm >/dev/null 2>&1; then
    UNIRTM_BIN="$(git rev-parse --show-toplevel 2>/dev/null)/unirtm"
    if [ -x "$UNIRTM_BIN" ]; then
        eval "$("$UNIRTM_BIN" env)" 2>/dev/null
    else
        if [ -x "$HOME/bin/unirtm" ]; then
            eval "$("$HOME/bin/unirtm" env)" 2>/dev/null
        elif [ -x "$HOME/.local/bin/unirtm" ]; then
            eval "$("$HOME/.local/bin/unirtm" env)" 2>/dev/null
        fi
    fi
else
    eval "$(unirtm env)" 2>/dev/null
fi
```
