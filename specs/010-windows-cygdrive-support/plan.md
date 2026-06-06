# Implementation Plan: Windows Git Bash/MSYS2 Cygdrive Support

## Background

## Goal

Support environment variables like `UNIRTM_CYGDRIVE_PREFIX` (or default to `/c/` for Git Bash and `/cygdrive/c/` for Cygwin) to properly format and convert Windows paths for seamless integration with bash environments on Windows.

## Proposed Changes

# `internal/pkg/envpath/envpath.go`

* Upd  `FormatDirForPosix(dir string)` logic:
    * If `runtime.GOOS == "windows"`, intercept paths starting with `C:\` or `[A-Z]:\`.
    * Convert `C:\foo\bar` into `<prefix>/c/foo/bar`.

  * If `UNIRTM_CYGDRIVE_PREFIX` is empty, use standard forward-slash replacement (as it currently does: `C:/foo/bar`) to prevent breaking non-bash Windows environments, or attempt to auto-detect bash.

### `internal/pkg/envpath/envpath_test.go`

* Add test cases validating the conversion logic when the cygdrive prefix is set.

## Verification

- Run `go test ./internal/pkg/envpath -v`
