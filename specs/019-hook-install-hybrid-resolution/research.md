# Research: Hook Install Hybrid Resolution

## Decisions & Rationale

**Decision**: Use String Manipulation for Block Injection/Replacement.
**Rationale**: We need to parse an existing bash script, locate the Shebang (if any), check if our block (`# --- BEGIN UNIRTM MANAGED BLOCK ---`) exists, and either replace it or insert a new one immediately after the shebang. Regular string splitting by lines and scanning is robust enough for this and doesn't require a full AST parser for Bash.
**Alternatives considered**: Overwriting the entire hook (causes data loss), using Regex substitution exclusively (can be fragile with multi-line matched blocks if the user manually modified the middle of it).

**Decision**: Act purely as an Environment Injector (Route A).
**Rationale**: As decided in the spec, if we act as a Router and execute `unirtm hook run`, we risk executing the Linters/formatters twice if the hook also contains native code like `# husky` or `# pre-commit`. By just configuring `$PATH`, we let the hook natively run the commands while ensuring they have the dependencies they need.

**Decision**: Use `git rev-parse --show-toplevel` for Sandbox Fallback.
**Rationale**: In headless environments, `$PATH` is stripped and `.profile` isn't loaded. By finding the `.git` root directory via `git rev-parse`, we can locate the project-local `unirtm` binary safely regardless of the current working directory from which git is invoked.
