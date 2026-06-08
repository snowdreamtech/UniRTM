# Implementation Plan: Shell Completions and Manpages

## 1. Shell Completions Pipeline (Already Implemented)
- Updated `cmd/5.completion.go` to handle `-d` flag and isolated `--install --all` logic.
- Updated `cmd/5.completion_test.go` with extensive tests.
- Modified `.goreleaser.yaml` to trigger `go run ./main.go completion -d ./completions --all`.
- Configured nfpm contents and Homebrew `install` block for completions.
- Appended `completions/` to `.gitignore`.

## 2. Manpages Pipeline
- **Add Command**: Modify `cmd/45.generate.go` to add `generateManpageCmd`.
  - Command format: `unirtm generate manpage -d ./manpages`.
  - Functionality: Utilize `github.com/spf13/cobra/doc` to `doc.GenManTree(rootCmd, nil, dir)`.
- **Git Ignore**: Add `manpages/` to `.gitignore`.
- **Goreleaser Updates**:
  - `before.hooks`: Add `go run ./main.go generate manpage -d ./manpages`.
  - `archives.files`: Add `- src: "manpages/*"` with `dst: "manpages"`.
  - `nfpms.contents`: Add `- src: ./manpages/*` pointing to `dst: /usr/share/man/man1/`.
  - `brews.install`: Append `man1.install Dir["manpages/*.1"]` to the install script.
