# Phase 0: Research & Abstraction Mapping

This document resolves the core problem of unifying `unirtm hook run [hookname] [stage]` semantics across multiple heterogeneous hook runners.

## Decision: Unified Semantic Abstraction

**Decision**: We adopt a smart abstraction mapping layer in `internal/hook/` implementations to map the `(hookname, stage)` tuple to the correct engine-specific commands.

**Rationale**: `pre-commit`, `husky`, and `lefthook` have fundamentally different architectures for executing hooks. Directly passing CLI arguments via `args []string` results in tightly coupled, easily broken bridge scripts and CLI inputs. A declarative mapping guarantees stable execution.

### Engine Mapping Strategies

| Engine | Without `--stage` (e.g. `hookname args...`) | With `--stage` (e.g. `hookname --stage stage args...`) |
|--------|-----------------------------------------------|---------------------------------------------------------|
| **pre-commit** | `pre-commit run <hookname> -- <args...>` | If hookname != all: `pre-commit run <hookname> --hook-stage <stage> -- <args...>`<br>If hookname == all: `pre-commit run --hook-stage <stage> -- <args...>` |
| **lefthook** | `lefthook run <hookname> <args...>` | If hookname != all: `lefthook run <stage> --commands <hookname> <args...>`<br>If hookname == all: `lefthook run <stage> <args...>` |
| **husky** | `sh .husky/<hookname> <args...>` | `sh .husky/<stage> <args...>` (ignores hookname as husky scripts are monolithic) |
| **native/shell** | `sh .git/hooks/<hookname> <args...>` | `sh .git/hooks/<stage> <args...>` (similar monolithic fallback) |

**Alternatives considered**:

- *Reject execution if engine doesn't support granular hooks:* Rejected because Husky does not support it natively, and failing would break the UX.
- *Mapping stage via env vars (`--hook-stage`):* Rejected because it leaked Python `pre-commit` implementation details into the Go generic interface.
