# Getting Started

Welcome to UniRTM! UniRTM is a universal runtime manager written in pure Go. It serves as a polyglot tool manager, an environment variable loader, and a task runner.

If you have used tools like `asdf`, `mise`, `nvm`, `pyenv`, or `rbenv`, you will feel right at home—but with significantly better performance and built-in security features.

## Why UniRTM?

1. **Lightning Fast**: Written in Go, it avoids the slow startup times of Bash/Ruby-based tools.
2. **Secure by Default**: Integrates with Trivy, Syft, and Gitleaks to scan your tools and environments for vulnerabilities and secrets.
3. **Cross-Platform**: Windows, macOS, and Linux are all treated as first-class citizens.
4. **All-in-One**: Replaces your tool manager, `.env` loaders (like `direnv`), and task runners (like `make`).

## 1. Install UniRTM

The easiest way to install UniRTM is via our installation script:

```bash
curl -sL https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/install.sh | bash
```

*For more installation methods (Homebrew, Cargo, APT), see [Installing UniRTM](./installing-unirtm.md).*

## 2. Hook into your Shell

For UniRTM to magically switch your tools and environments when you `cd` into a directory, it needs to hook into your shell.

```bash
# For Zsh
echo 'eval "$(unirtm env --shell zsh)"' >> ~/.zshrc

# For Bash
echo 'eval "$(unirtm env --shell bash)"' >> ~/.bashrc
```

Restart your terminal for the changes to take effect.

## 3. Install your first tools

Navigate to your project directory and declare the tools you need.

```bash
$ unirtm use node@20 python@3.11 go@1.22
✓ wrote .unirtm.toml

$ unirtm install
✓ installed 3 tools
```

You can verify they are active:

```bash
$ node -v
v20.x.x
$ go version
go version go1.22.x
```

## 4. Setup Environment Variables

Create a `.env` file in your project:

```bash
echo "DATABASE_URL=postgres://localhost:5432/mydb" > .env
```

UniRTM automatically loads these variables when you enter the directory and unloads them when you leave.

## 5. Define Tasks

Open your `.unirtm.toml` and add a task:

```toml
[tasks.test]
description = "Run the test suite"
run = "go test ./..."
```

Run it with:

```bash
unirtm run test
```

## Next Steps

- Take the [Full Walkthrough](./walkthrough.md) to explore advanced features.
- Learn about [The `.unirtm.toml` File](../configuration/unirtm-toml.md).
- Dive into [Tasks](../tasks/overview.md).
