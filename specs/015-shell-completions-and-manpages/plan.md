# Implementation Plan: Shell Completions and Manpages

## 1. Shell Completions Pipeline (Already Implemented)

- Updated `cmd/5.completion.go` to handle `-d` flag and isolated `--install --all` logic.
- Updated `cmd/5.completion_test.go` with extensive tests.
- Modified `.goreleaser.yaml` to trigger `go run ./main.go completion -d ./completions --all`.
- Configured nfpm contents mapping generated completions to `/usr/share/...` directories.
- Appended `completions/` to `.gitignore`.

## 2. Manpages Pipeline (Already Implemented)

- **Add Command**: Modified `cmd/45.generate.go` to add `generateManpageCmd`.
  - Command format: `unirtm generate manpage -d ./manpages`.
  - Functionality: Utilize `github.com/spf13/cobra/doc` to `doc.GenManTree(rootCmd, nil, dir)`.
- **Git Ignore**: Added `manpages/` to `.gitignore`.
- **Goreleaser Updates**:
  - `before.hooks`: Added `go run ./main.go generate manpage -d ./manpages`.
  - `archives.files`: Added `- src: "manpages/*"` with `dst: "manpages"`.
  - `nfpms.contents`: Added `- src: ./manpages/*` pointing to `dst: /usr/share/man/man1/`.

## 3. Homebrew Distribution (Already Implemented)

- **Goreleaser Brews Config**:
  - Ensured repository points to `snowdreamtech/homebrew-tap`.
  - Added a custom `install` block to automate installation of binaries, completions, and manpages:

    ```ruby
    bin.install "unirtm"
    bash_completion.install "completions/unirtm.bash" => "unirtm"
    zsh_completion.install "completions/unirtm.zsh" => "_unirtm"
    fish_completion.install "completions/unirtm.fish"
    man1.install Dir["manpages/*.1"]
    ```
