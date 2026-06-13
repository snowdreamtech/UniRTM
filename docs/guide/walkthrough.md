# Walkthrough

Welcome to the full UniRTM walkthrough! This guide will take you step-by-step through a real-world scenario, demonstrating how UniRTM replaces multiple legacy tools (like `asdf`, `direnv`, and `make`) with a single, lightning-fast native binary.

## 1. Setting up a Project

Imagine you're starting a new full-stack project. Your backend requires Go 1.22, and your frontend requires Node 20.

```bash
mkdir my-fullstack-app
cd my-fullstack-app
```

Instead of manually installing these tools or polluting your system paths, let UniRTM manage them:

```bash
unirtm use go@1.22 node@20
```

This command creates a `.unirtm.toml` file in your directory:

```toml
[tools]
go = "1.22"
node = "20"
```

Because UniRTM dynamically adjusts your environment based on this file, you now have the correct versions active the moment you enter this directory!

## 2. Managing Environment Variables

Your app needs some environment variables (e.g., a database connection string). You can specify these directly in your `.unirtm.toml` or load them from a `.env` file.

Let's use the `.unirtm.toml` file to define them contextually:

```toml
[env]
DATABASE_URL = "postgres://user:pass@localhost:5432/mydb"
NODE_ENV = "development"
# Variables can even reference other variables
API_ENDPOINT = "https://api.dev.example.com"
DEBUG_URL = "${API_ENDPOINT}/debug"
```

Whenever you run a command in this directory, or execute a task, these variables are injected safely and automatically—without `direnv` or modifying `.bashrc`.

## 3. The Task Runner

Instead of relying on `Makefile`s or `package.json` scripts that only work for JS, UniRTM features a universal task runner.

Add the following to your `.unirtm.toml`:

```toml
[tasks.build]
description = "Builds the backend and frontend"
run = """
  go build -o server ./cmd/api
  npm run build
"""

[tasks.dev]
description = "Run the development server"
run = "go run ./cmd/api"
env = { PORT = "8080" }
```

You can now list available tasks:

```bash
$ unirtm tasks
build    Builds the backend and frontend
dev      Run the development server
```

And run them effortlessly:

```bash
unirtm run dev
```

UniRTM guarantees that `go` and `npm` will resolve to exactly the versions pinned in your `[tools]` block, and the environment variables from your `[env]` block (plus task-specific overrides like `PORT`) will be strictly scoped to this execution.

## 4. Temporary Execution Isolation

Sometimes you need to run a tool in an isolated environment without affecting your project configuration. For example, you want to test a script on Python 3.9 without permanently installing it or changing your `PATH`.

Use `unirtm exec`:

```bash
unirtm exec python@3.9 -- python test_script.py
```

UniRTM will securely download Python 3.9, evaluate the environment, execute your script, and then cleanly tear down the temporary context. No global state modified.

## Conclusion

You've successfully:

- Pinned multiple tool versions using `unirtm use`.
- Set up isolated, dynamic environment variables.
- Created a cross-language build pipeline using `[tasks]`.
- Executed isolated commands without polluting global state.

UniRTM brings order to your software supply chain. To dig deeper, check out the [CLI Overview](../cli/overview.md) or the [Architecture Deep Dive](../advanced/architecture.md).
