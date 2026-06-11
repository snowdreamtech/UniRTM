# CLI Contract: `unirtm hook run`

## Command Schema

```bash
unirtm hook run [hookname] [--stage stage] [args...]
```

### Parameters

- **`hookname`** (required): The name of the specific hook rule to execute (e.g., `shellcheck`, `go-fmt`). The reserved keyword `all` signifies the intention to run the entire `stage`.
- **`--stage`** (optional flag): The Git lifecycle stage context (e.g., `pre-commit`, `commit-msg`).
- **`args...`** (optional): Any arbitrary trailing arguments passed by Git (e.g., `$1` file paths).

### Validation Rules

- Must provide exactly 1 positional argument before `--` or trailing arguments (`hookname`).
- Must not fail if trailing arguments are provided.
