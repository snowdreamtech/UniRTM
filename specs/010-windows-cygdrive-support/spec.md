# Feature Specification: Windows Git Bash/MSYS2 Cygdrive Support

## 1. Overview
The goal of this feature is to seamlessly support Windows paths running within Bash emulators (such as Git Bash, MSYS2, and Cygwin) by converting standard Windows drive letters (like `C:\`) to Bash-compatible paths (like `/c/` or `/cygdrive/c/`).

## 2. Motivation
Windows developers frequently use Git Bash or similar environments. When UniRTM interacts with or generates paths intended for the shell (for example, setting environment variables or constructing execution paths), passing raw `C:\` paths often breaks the Bash environment. Allowing users to configure an explicit mount prefix (`UNIRTM_CYGDRIVE_PREFIX`) enables flawless interoperability.

## 3. Requirements
*   **Path Translation:** Correctly intercept `C:\` style paths and prepend the provided `CYGDRIVE_PREFIX`.
*   **Opt-in Customization:** Must respect the `UNIRTM_CYGDRIVE_PREFIX` environment variable.
*   **Non-breaking:** Standard Windows Command Prompt (`cmd.exe`) or PowerShell paths must remain unaffected if the user doesn't require Bash translation.

## 4. Proposed Solution
*   Update `FormatDirForPosix` in `internal/pkg/envpath/envpath.go` to evaluate and apply the `UNIRTM_CYGDRIVE_PREFIX` logic when running on `windows`.
