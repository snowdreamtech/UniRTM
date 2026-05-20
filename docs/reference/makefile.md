# Makefile Commands

All common tasks are unified under `make`.

## Setup & Installation

All tools and dependencies are managed and installed via UniRTM.

```bash
unirtm install  # Install all configured development tools and dependencies
```

## Quality Gates

```bash
make lint     # Run ALL linting checks (pre-commit hooks)
make check    # Run lint + test in sequence
```

## Reference

| Target    | Description                                                |
| --------- | ---------------------------------------------------------- |
| `help`    | Show all available targets and their descriptions          |
| `lint`    | Run all pre-commit hooks against all files                 |
|| `test`    | Execute test suite                                         |
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
