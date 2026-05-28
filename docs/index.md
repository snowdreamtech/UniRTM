---
layout: home

hero:
  name: "UniRTM"
  text: "The universal runtime manager"
  tagline: The fast, simple, cross-platform tool to manage your dev tools, environments, and tasks.
  actions:
    - theme: brand
      text: Get Started
      link: /guide/introduction
    - theme: alt
      text: View on GitHub
      link: https://github.com/snowdreamtech/UniRTM

features:
  - icon: 🛠️
    title: Dev Tools Manager
    details: A polyglot tool manager. Drop-in replacement for asdf, nvm, pyenv, rbenv, and more. Manage all your runtimes in one place.
  - icon: 🌍
    title: Environment Variables
    details: Seamless `.env` and environment management. Automatically load and unload environment variables when switching directories.
  - icon: ⚡
    title: Task Runner
    details: Run tasks easily across multiple languages without relying on complex Makefiles or npm scripts.
---

<br>

<div align="center">
  <h2>The Idea: Everything in its place, before you code.</h2>
  <p style="max-width: 600px; margin: 0 auto; color: var(--vp-c-text-2);">
    Just like a professional chef prepares their station before cooking (<em>mise en place</em>), UniRTM prepares your development environment before you write a single line of code. It installs the right tools, loads the right environment variables, and wires up the right tasks for the commands you run.
  </p>
</div>

<br><br>

## The Menu: One CLI for the whole project setup.

### 🔪 01. Dev Tools
Install project tools, pin versions, and switch automatically as you move between directories. No more guessing which version of Node or Python you need.

```bash
$ unirtm use node@20 python@3.11 go@1.22
✓ wrote .unirtm.toml

$ unirtm install
✓ installed 3 tools
```

### 🫕 02. Environments
Load project-specific environment variables from `.unirtm.toml`, `.env` files, shell commands, and more. Stop cluttering your global bash profile.

```bash
$ cat .env.local
DATABASE_URL=postgres://localhost/orders

$ unirtm env
export DATABASE_URL=postgres://localhost/orders
```

### 🍳 03. Tasks
Define build, test, lint, and deploy commands next to the tools and env vars they need. A modern replacement for complex Makefiles or npm scripts.

```bash
$ unirtm run test
→ lint · typecheck · unit
✓ 3 tasks complete

$ unirtm run deploy
✓ shipped
```

<br>

## Enterprise-Grade by Design

Unlike legacy tools written in Bash or Ruby, UniRTM is engineered for the modern ecosystem.

::: info 🚀 Blazing Fast
Written in pure **Go**, UniRTM executes in milliseconds. No more waiting for slow Ruby shims or complex Bash initializations to load your environment when opening a new terminal.
:::

::: tip 🔒 Secure & Compliant
Native integration with industry-standard security tools like **Trivy**, **Gitleaks**, and **Syft**. Keep your supply chain secure and compliant from day one.
:::

::: warning 💻 Truly Cross-Platform
Designed from the ground up for seamless operation on **macOS**, **Linux**, and **Windows**. Your team gets a consistent experience everywhere, regardless of their OS.
:::

<br>

<div align="center">
  <h2>Ready When You Are</h2>
  <p><em>Allez,</em> prep your station.</p>
  
  ```bash
  curl -sL https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/install.sh | bash
  ```
</div>
