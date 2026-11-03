# Phase 0: Outline & Research

## Unknowns Resolved

- **Handling `v` prefixes safely**: How do we ensure we don't accidentally modify tags like `beta-1.0` or `latest`?
  - **Decision**: We use string indexing and `unicode.IsDigit` in Go. We only strip a leading 'v' or 'V' if it's immediately followed by a digit. We only prepend a 'v' if the first character of the string is a digit.
  - **Rationale**: This guarantees we only ever affect SemVer-like strings (e.g., `1.2.3`, `v1.0.0-rc1`) while leaving aliases (`latest`) or non-standard semantic identifiers (`release-1.0`) completely untouched.

- **Backend Configuration**: How does the function know if it should prepend or strip?
  - **Decision**: The function will accept a boolean flag (e.g. `requiresVPrefix bool`) alongside the version string.
  - **Rationale**: The individual backend providers (e.g. `GithubHandler`) inherently know their own format requirements. Passing a simple boolean is stateless, testable, and highly predictable.
