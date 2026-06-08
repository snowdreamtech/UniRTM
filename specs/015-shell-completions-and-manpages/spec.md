# Specification: Shell Completions and Manpages Auto-generation

## Overview
The goal is to provide a fully automated generation and distribution pipeline for shell completions (Zsh, Bash, Fish, PowerShell) and Manpages. These artifacts should be generated automatically during the build process and seamlessly integrated into release archives, native Linux packages (nfpm), and Homebrew formulae.

## Requirements

1. **Shell Completions Pipeline (Completed)**
   - Enhance the `unirtm completion` command to support exporting all shell scripts to a specific directory (`-d/--dir`).
   - Refine the `--install --all` logic to always generate the files but only inject into shell configurations if the corresponding shell profile exists.
   - Ignore dynamically generated completions locally (`.gitignore`).
   - Integrate with Goreleaser to dynamically generate completions in `before.hooks` and package them inside archives and nfpm (`/usr/share/...`).

2. **Manpages Pipeline (Completed)**
   - Add a `unirtm generate manpage -d <dir>` command using Cobra's built-in `doc.GenManTree`.
   - Exclude local `manpages/` directory from Git tracking.
   - Update `.goreleaser.yaml` to dynamically generate manpages into a `manpages/` directory.
   - Package manpages into `.tar.gz` and `.zip` archives.
   - Install manpages to standard system paths (`/usr/share/man/man1/`) via `nfpm` for `.deb`/`.rpm`/`.apk`/`.archlinux`.

3. **Homebrew Distribution (Completed)**
   - Configure Homebrew to automatically push generated Ruby Formulae to `snowdreamtech/homebrew-tap`.
   - The Homebrew `install` block must automatically hook up completions and manpages without user intervention.
   - Specifically, use Homebrew DSL:
     - `bash_completion.install "completions/unirtm.bash" => "unirtm"`
     - `zsh_completion.install "completions/unirtm.zsh" => "_unirtm"`
     - `fish_completion.install "completions/unirtm.fish"`
     - `man1.install Dir["manpages/*.1"]`
