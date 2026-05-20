# Makefile Commands

All common tasks are unified under `make`. Run `make help` to see all available targets.

## Setup & Installation

All tools and dependencies are managed and installed via UniRTM.

```bash
unirtm install  # Install all configured development tools and dependencies
```

## Quality Gates

```bash
make lint     # Run ALL linting checks (pre-commit hooks)
make format   # Auto-format code across all languages
make test     # Run the test suite
make check    # Run lint + test in sequence
```

## Build & Release

```bash
make build    # Build the project binary/artifacts
make clean    # Remove build artifacts and temporary files
```

## Reference

| Target    | Description                                                |
| --------- | ---------------------------------------------------------- |
| `help`    | Show all available targets and their descriptions          |
| `lint`    | Run all pre-commit hooks against all files                 |
| `format`  | Auto-format all source files                               |
| `test`    | Execute test suite                                         |
| `build`   | Build production artifacts                                 |
| `check`   | Combined lint + test                                       |
| `clean`   | Remove generated files and caches                          |

## Cross-Platform Behavior

The Makefile automatically detects your operating system and uses the appropriate package manager:

| OS                    | Package Manager   |
| --------------------- | ----------------- |
| macOS                 | Homebrew (`brew`) |
| Linux (Debian/Ubuntu) | APT (`apt-get`)   |
| Linux (RedHat/Alpine) | DNF/APK           |
| Windows               | Scoop or Winget   |
