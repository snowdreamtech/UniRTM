# Troubleshooting

If you encounter issues while using UniRTM, don't panic. This document compiles the most common problems and their quick fixes.

> [!TIP]
> **The first step when facing any issue:** Always try running `unirtm doctor`. This built-in diagnostic tool automatically scans your system environment, shell hook status, and dependency integrity, providing immediate actionable recommendations.

## 1. Tool versions do not change after switching directories

**Symptom:** After entering a project directory with a `.unirtm.toml`, running `node -v` still uses the system's global Node.js instead of the version pinned by the project.

**Cause & Fix:**
This happens because UniRTM's core "dynamic PATH interception" hook is not correctly loaded into your current Shell session.
Please check your `~/.bashrc` or `~/.zshrc` (or the respective Shell profile) and ensure it contains the following line:

```bash
eval "$(unirtm activate zsh)"  # For zsh
# OR
eval "$(unirtm activate bash)" # For bash
```

After adding it, remember to run `source ~/.zshrc` or simply restart your terminal.

## 2. "API Rate Limit Exceeded" Errors

**Symptom:** When executing `unirtm install`, you encounter network errors like "rate limit exceeded" or "403 Forbidden" from GitHub.

**Cause & Fix:**
By default, unauthenticated requests to the GitHub API (e.g., when resolving the latest tool versions) are subject to strict rate limits.
You can generate a [GitHub Personal Access Token (PAT)](https://github.com/settings/tokens) (no special permissions needed, just public read access) and export it as an environment variable:

```bash
export GITHUB_TOKEN="your_PAT_string"
```

UniRTM will automatically detect this token to bypass standard rate limits.

## 3. "Permission Denied" when installing tools

**Symptom:** You are prompted with permission denied errors during `unirtm install`.

**Cause & Fix:**
UniRTM **never requires** `sudo` privileges. All toolchains, caches, and runtimes are strictly isolated and installed within the current user's `~/.local/share/unirtm` directory.
If you encounter permission errors, it is usually because you mistakenly ran `sudo unirtm` earlier, changing the ownership of some directories to `root`.
**Fix:** Restore directory ownership back to your user:

```bash
sudo chown -R $USER:$USER ~/.local/share/unirtm ~/.config/unirtm
```

## 4. The `.unirtm.toml` configuration is completely ignored

**Symptom:** The current directory clearly has a `.unirtm.toml` specifying `node = "20"`, but no matter how many times you restart the terminal, UniRTM refuses to read it.

**Cause & Fix:**
For **absolute security reasons**, UniRTM incorporates a "Trust System". If a configuration file's path is not marked as trusted, UniRTM will refuse to execute any environment modification instructions from it to prevent malicious script injection (e.g., from cloned random repositories).
**Fix:** Explicitly trust the configuration by running the following command in the directory:

```bash
unirtm trust
```

## 5. Reporting an Issue

If none of the above solutions resolve your problem, please head over to our [GitHub Issues](https://github.com/snowdreamtech/UniRTM/issues) page. When opening an issue, make sure to include the complete output logs from running `unirtm doctor` and `unirtm env`.
