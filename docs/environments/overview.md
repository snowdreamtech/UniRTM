# Environment Variables Overview

Modern applications rarely operate in a vacuum. They require tokens, database URLs, port bindings, and API keys. UniRTM treats Environment Variable management as a first-class citizen alongside tool versioning.

## The Problem with Traditional `.env`

Traditionally, developers rely on utilities like `dotenv`, `direnv`, or hardcoded exports in `.bashrc` to manage environment variables. These approaches suffer from:

1. **Shell Dependency**: `direnv` requires shell hooks.
2. **Language Dependency**: `dotenv` requires parsing libraries in every language (Node, Python, Go) you use in a monorepo.
3. **Scoping Issues**: Global exports pollute your system and can leak secrets to unrelated applications.

## UniRTM's Solution: Contextual Environments

UniRTM dynamically injects environment variables right before executing a tool or a task. This means the environment variables are strictly scoped to the exact process being spawned.

### Multi-Tier Resolution Strategy

UniRTM evaluates environment variables using a strict hierarchy (from highest to lowest precedence):

1. **CLI Overrides**: `PORT=8080 unirtm run start`
2. **Task Specific Variables**: Variables defined inside a specific `[tasks.start.env]` block in `.unirtm.toml`.
3. **Project Variables (`.unirtm.toml`)**: Global variables defined in the `[env]` block of the project's `.unirtm.toml`.
4. **Local Env Files (`.env`)**: Standard `.env` files located in the project directory.
5. **Global UniRTM Env**: Variables defined in `~/.config/unirtm/config.toml`.
6. **System Environment**: Inherited from the OS.

### Example: The `.unirtm.toml` `[env]` Block

You can define variables natively inside your configuration file:

```toml
[env]
NODE_ENV = "development"
API_ENDPOINT = "https://api.staging.example.com"
# You can even use variable interpolation
DEBUG_URL = "${API_ENDPOINT}/debug"
```

### Supported Formats

UniRTM natively parses and evaluates:

- Standard `key=value` formats.
- Quotes handling: `"value"` and `'value'`.
- Export syntax: `export KEY=value` (the `export` keyword is safely ignored).
- Bash-style parameter expansion: `${VAR:-default}`.

### Integration with Secrets Management

For production secrets or sensitive keys, committing them to `.unirtm.toml` or `.env` is a security risk.
See the [Secrets Management](./secrets.md) guide to learn how UniRTM integrates with tools like 1Password CLI, AWS Secrets Manager, and HashiCorp Vault to securely fetch secrets on-the-fly.
