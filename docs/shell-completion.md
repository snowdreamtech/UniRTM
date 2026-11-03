# Shell Completion

UniRTM provides shell completion support for Bash, Zsh, Fish, and PowerShell. Shell completion allows you to use tab completion for commands, flags, and arguments, improving the user experience and reducing typing errors.

## Installation

### Bash

#### Linux

To load completions for the current session:

```bash
source <(unirtm completion bash)
```

To load completions for each session, execute once:

```bash
unirtm completion bash > /etc/bash_completion.d/unirtm
```

You may need to use `sudo` for this command if you don't have write permissions to `/etc/bash_completion.d/`.

#### macOS

To load completions for the current session:

```bash
source <(unirtm completion bash)
```

To load completions for each session, execute once:

```bash
unirtm completion bash > $(brew --prefix)/etc/bash_completion.d/unirtm
```

**Note:** You need to have bash-completion installed via Homebrew:

```bash
brew install bash-completion@2
```

Then add the following to your `~/.bash_profile` or `~/.bashrc`:

```bash
export BASH_COMPLETION_COMPAT_DIR="$(brew --prefix)/etc/bash_completion.d"
[[ -r "$(brew --prefix)/etc/profile.d/bash_completion.sh" ]] && . "$(brew --prefix)/etc/profile.d/bash_completion.sh"
```

### Zsh

If shell completion is not already enabled in your environment, you will need to enable it. You can execute the following once:

```zsh
echo "autoload -U compinit; compinit" >> ~/.zshrc
```

To load completions for the current session:

```zsh
source <(unirtm completion zsh)
```

To load completions for each session, execute once:

```zsh
unirtm completion zsh > "${fpath[1]}/_unirtm"
```

**Note:** You will need to start a new shell for this setup to take effect.

#### Alternative Installation (Oh My Zsh)

If you use Oh My Zsh, you can install completions to the custom plugins directory:

```zsh
mkdir -p ~/.oh-my-zsh/custom/plugins/unirtm
unirtm completion zsh > ~/.oh-my-zsh/custom/plugins/unirtm/_unirtm
```

Then add `unirtm` to the plugins array in your `~/.zshrc`:

```zsh
plugins=(... unirtm)
```

### Fish

To load completions for the current session:

```fish
unirtm completion fish | source
```

To load completions for each session, execute once:

```fish
unirtm completion fish > ~/.config/fish/completions/unirtm.fish
```

**Note:** You will need to start a new shell for this setup to take effect.

### PowerShell

To load completions for the current session:

```powershell
unirtm completion powershell | Out-String | Invoke-Expression
```

To load completions for every new session:

1. Generate the completion script:

   ```powershell
   unirtm completion powershell > unirtm.ps1
   ```

2. Find your PowerShell profile location:

   ```powershell
   echo $PROFILE
   ```

3. Add the following line to your PowerShell profile:

   ```powershell
   . /path/to/unirtm.ps1
   ```

   Replace `/path/to/unirtm.ps1` with the actual path where you saved the completion script.

**Note:** You may need to adjust your execution policy to allow running scripts:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

## Usage

Once shell completion is installed, you can use tab completion for:

- **Commands**: Type `unirtm <TAB>` to see available commands
- **Flags**: Type `unirtm --<TAB>` to see available flags
- **Subcommands**: Type `unirtm install <TAB>` to see available options

### Examples

```bash
# Complete commands
$ unirtm <TAB>
completion  config  doctor  install  list  search  uninstall  version

# Complete flags
$ unirtm --<TAB>
--config  --help  --json  --quiet  --verbose  --version

# Complete subcommands
$ unirtm completion <TAB>
bash  fish  powershell  zsh
```

## Troubleshooting

### Bash Completion Not Working

1. Ensure bash-completion is installed:
   - **Linux (Debian/Ubuntu)**: `sudo apt-get install bash-completion`
   - **Linux (RHEL/CentOS)**: `sudo yum install bash-completion`
   - **macOS**: `brew install bash-completion@2`

2. Verify bash-completion is loaded in your shell:

   ```bash
   type _init_completion
   ```

   If this command returns "not found", bash-completion is not loaded.

3. Check that the completion script is in the correct directory:
   - **Linux**: `/etc/bash_completion.d/` or `/usr/share/bash-completion/completions/`
   - **macOS**: `$(brew --prefix)/etc/bash_completion.d/`

### Zsh Completion Not Working

1. Ensure `compinit` is called in your `~/.zshrc`:

   ```zsh
   autoload -U compinit; compinit
   ```

2. Check that the completion script is in a directory in your `$fpath`:

   ```zsh
   echo $fpath
   ```

3. Verify the completion file has the correct name (must start with `_`):

   ```zsh
   ls -la "${fpath[1]}/_unirtm"
   ```

4. Clear the completion cache and restart your shell:

   ```zsh
   rm -f ~/.zcompdump*
   exec zsh
   ```

### Fish Completion Not Working

1. Ensure the completion file is in the correct directory:

   ```fish
   ls -la ~/.config/fish/completions/unirtm.fish
   ```

2. Restart your shell or reload completions:

   ```fish
   source ~/.config/fish/completions/unirtm.fish
   ```

### PowerShell Completion Not Working

1. Check your execution policy:

   ```powershell
   Get-ExecutionPolicy
   ```

   If it's set to `Restricted`, you need to change it:

   ```powershell
   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
   ```

2. Verify the completion script is sourced in your profile:

   ```powershell
   Get-Content $PROFILE
   ```

3. Reload your PowerShell profile:

   ```powershell
   . $PROFILE
   ```

## Uninstallation

To remove shell completions:

### Bash

```bash
# Linux
sudo rm /etc/bash_completion.d/unirtm

# macOS
rm $(brew --prefix)/etc/bash_completion.d/unirtm
```

### Zsh

```zsh
rm "${fpath[1]}/_unirtm"
# or for Oh My Zsh
rm ~/.oh-my-zsh/custom/plugins/unirtm/_unirtm
```

### Fish

```fish
rm ~/.config/fish/completions/unirtm.fish
```

### PowerShell

Remove the line sourcing `unirtm.ps1` from your PowerShell profile, then delete the completion script file.

## Additional Resources

- [Cobra Shell Completions](https://cobra.dev/#shell-completions)
- [Bash Completion Documentation](https://github.com/scop/bash-completion)
- [Zsh Completion System](http://zsh.sourceforge.net/Doc/Release/Completion-System.html)
- [Fish Completion Documentation](https://fishshell.com/docs/current/completions.html)
- [PowerShell Completion Documentation](https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.core/register-argumentcompleter)
