# Continuous Integration (CI)

UniRTM provides a first-class GitHub Action, `snowdreamtech/setup-unirtm`, to seamlessly integrate your local development environment directly into your CI/CD pipelines. This ensures **100% parity** between what you run on your laptop and what runs on your build servers.

## Core Workflow

The quickest way to get started is to use the action with its default settings:

```yaml
name: CI
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup UniRTM & Install Tools
        uses: snowdreamtech/setup-unirtm@v1
        with:
          # Optional: Specify UniRTM version
          unirtm-version: latest
          # Optional: Automatically run unirtm install
          install: true
          # Optional: Automatically trust .unirtm.toml
          trust: true
          # Optional: Pass GitHub token to prevent API rate-limiting
          github_token: ${{ secrets.GITHUB_TOKEN }}

      - name: Run Tests
        # All tools defined in .unirtm.toml are now natively available in PATH
        run: npm test
```

## Deep Dive: How `setup-unirtm` Works Under the Hood

The `setup-unirtm` action is not just a simple downloader. It is a highly optimized wrapper designed specifically for ephemeral CI environments.

Here is what the action code does under the hood:

### 1. Zero-Friction Binary Resolution

Instead of building from source or relying on slow package managers, the action queries the GitHub Releases API (using the provided `github_token` to bypass rate limits), detects the runner's OS (Linux, macOS, Windows) and architecture (amd64, arm64), and downloads the exact compiled native binary for `unirtm`. It then appends this to the runner's `$GITHUB_PATH`.

### 2. Intelligent Auto-Caching

One of the slowest parts of CI is downloading and compiling development tools (like Python or Ruby). `setup-unirtm` intercepts this process.

- It hashes your `.unirtm.toml` and `.unirtm.lock` files to generate a unique cache key.
- It attempts to restore `~/.local/share/unirtm/installs` and `~/.local/share/unirtm/downloads` using the `@actions/cache` API.
- If a cache hit occurs, tool installation takes **0 seconds**.
- If a cache miss occurs, the action automatically runs `unirtm install` and then seamlessly saves the newly compiled tools back to the GitHub cache post-run.

### 3. Environment Variable Bridging

UniRTM supports setting project-specific environment variables in `.unirtm.toml` via the `[env]` block.
The action automatically runs `unirtm env` and parses the output, injecting all those variables directly into the runner's `$GITHUB_ENV` context.
This means subsequent steps in your workflow instantly inherit the exact same environment variables as your local machine, without needing to duplicate them in GitHub Secrets/Variables (unless they are sensitive credentials).

### 4. Automatic Tool Shimming

After installing the tools, `setup-unirtm` registers the binary paths of all activated tools (e.g., `node`, `go`, `python`) directly into the runner's `$GITHUB_PATH`.
Therefore, in subsequent steps, you don't need to run `unirtm exec npm -- install`; you can simply run `npm install`. The native runner `PATH` has been correctly poisoned with your pinned toolchain.

## Configuration Inputs

You can customize the behavior of `setup-unirtm` using the following inputs:

| Input | Default | Description |
|---|---|---|
| `unirtm-version` | `''` (latest) | The version of UniRTM to install (e.g. `"0.0.1"`). If empty, installs the latest release. |
| `install` | `false` | If true, runs `unirtm install` after setting up UniRTM. |
| `install_args` | `''` | Arguments to pass to `unirtm install` (e.g. `"node python"`). |
| `trust` | `false` | If true, runs `unirtm trust` after setting up UniRTM. |
| `github_token` | <code v-pre>${{ github.token }}</code> | GitHub token for API authentication to avoid rate limits. |
| `github_proxy` | `''` | Proxy prefix for GitHub download URLs (e.g. `"https://ghproxy.com/"`) to speed up downloads in restricted networks. |

## Other CI Providers (GitLab CI, CircleCI)

If you are not using GitHub Actions, you can easily install UniRTM using the standalone bash script:

```yaml
# GitLab CI Example
test:
  image: ubuntu:latest
  script:
    # 1. Install UniRTM
    - curl -fsSL https://github.com/snowdreamtech/unirtm/raw/main/install.sh | sh
    - export PATH="$HOME/.local/bin:$PATH"

    # 2. Install tools
    - unirtm install

    # 3. Export env vars and paths for this session
    - eval "$(unirtm activate bash)"

    # 4. Run your tasks
    - npm test
```
