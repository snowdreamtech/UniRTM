# Tasks: Shell Completions and Manpages

## 1. Completions Pipeline

- [x] 1. Implement `-d` flag in `unirtm completion` and fix `--all` installation logic.
- [x] 2. Add `completions/` to `.gitignore`.
- [x] 3. Update `.goreleaser.yaml` to generate completions and distribute them inside archives and `nfpm`.
- [x] 4. Commit completion changes to Git.

## 2. Manpages Pipeline

- [x] 5. Add `unirtm generate manpage -d <dir>` command in `cmd/45.generate.go`.
- [x] 6. Add `manpages/` to `.gitignore`.
- [x] 7. Update `.goreleaser.yaml` to generate manpages and distribute them inside archives and `nfpm`.
- [x] 8. Commit manpage changes to Git.

## 3. Homebrew Distribution

- [x] 9. Configure `.goreleaser.yaml` `brews` block to push to `snowdreamtech/homebrew-tap`.
- [x] 10. Write custom Homebrew `install` block using DSL for binary, `bash_completion`, `zsh_completion`, `fish_completion`, and `man1`.
- [x] 11. Commit `.goreleaser.yaml` Homebrew integration changes to Git.
