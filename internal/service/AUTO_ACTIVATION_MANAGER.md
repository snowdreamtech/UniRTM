# Auto-Activation Manager Implementation

## Overview

The Auto-Activation Manager implements automatic environment activation based on directory changes. It detects when the user enters or leaves a project directory and automatically activates or deactivates the project's toolchain.

**Requirements:** 15.6, 15.7

## Architecture

### Core Components

1. **AutoActivationManager**: Main manager that handles directory change events
2. **EnvironmentState**: Represents the current environment state
3. **DirectoryChangeEvent**: Represents a directory change event
4. **ActivationChange**: Represents the changes needed to update the environment

### Key Features

- **Directory-based Detection**: Automatically detects project directories by searching for UniRTM configuration files
- **Automatic Activation**: Activates project toolchains when entering project directories
- **Automatic Deactivation**: Restores the previous environment when leaving project directories
- **Project Switching**: Seamlessly switches between different projects
- **Shell Hook Integration**: Provides shell hooks for automatic activation on directory change

## Implementation Details

### Configuration File Detection

The Auto-Activation Manager searches for the following configuration files (in order):

1. `unirtm.toml`
2. `.unirtm.toml`
3. `mise.toml`
4. `.mise.toml`
5. `.tool-versions`

The search starts from the current directory and walks up the directory tree until a configuration file is found or the root is reached.

### Activation Actions

The manager supports four types of activation actions:

1. **ActionActivate**: Entering a project directory from outside
2. **ActionDeactivate**: Leaving a project directory
3. **ActionSwitch**: Switching from one project to another
4. **ActionNone**: No change needed (staying in the same project or no project)

### Environment State Management

The `EnvironmentState` tracks:

- **ProjectDir**: The current project directory (empty if no project is active)
- **ToolVersions**: Active tool versions for the current project
- **EnvVars**: Environment variables set by the activation
- **PreviousPath**: The PATH before activation (for restoration)

### Shell Hook Integration

The Auto-Activation Manager generates shell-specific hook scripts that:

1. Detect directory changes by comparing `$PWD` with a saved value
2. Call `unirtm hook-env` to get activation changes
3. Evaluate the returned script to update the environment

#### Bash/Zsh Hook

```bash
_unirtm_hook() {
  local old_pwd="${UNIRTM_OLD_PWD:-}"
  local new_pwd="$PWD"

  if [ "$old_pwd" != "$new_pwd" ]; then
    export UNIRTM_OLD_PWD="$new_pwd"
    eval "$(unirtm hook-env --shell bash)"
  fi
}

# Install in PROMPT_COMMAND (bash) or precmd (zsh)
```

#### Fish Hook

```fish
function _unirtm_hook --on-variable PWD
  unirtm hook-env --shell fish | source
end
```

#### PowerShell Hook

```powershell
function Invoke-UnirtmHook {
  $oldPwd = $env:UNIRTM_OLD_PWD
  $newPwd = $PWD.Path

  if ($oldPwd -ne $newPwd) {
    $env:UNIRTM_OLD_PWD = $newPwd
    $script = unirtm hook-env --shell powershell
    if ($script) {
      Invoke-Expression $script
    }
  }
}

# Install in prompt function
```

## Usage

### Basic Usage

```go
// Create managers
activationMgr := NewActivationManager("/path/to/shims", "/path/to/data")
autoMgr := NewAutoActivationManager(activationMgr)

// Handle directory change
event := DirectoryChangeEvent{
    OldDir: "/home/user",
    NewDir: "/home/user/project",
    Shell:  ShellBash,
}

currentState := &EnvironmentState{
    ProjectDir:   "",
    ToolVersions: make(map[string]string),
    EnvVars:      make(map[string]string),
    PreviousPath: os.Getenv("PATH"),
}

change, err := autoMgr.HandleDirectoryChange(ctx, event, currentState)
if err != nil {
    // Handle error
}

// Execute the activation script
if change.Action != ActionNone {
    fmt.Println(change.Script)
    // Update current state
    currentState = change.NewState
}
```

### Generating Shell Hooks

```go
// Generate hook script for bash
hookScript, err := autoMgr.GenerateHookEnvScript(ShellBash)
if err != nil {
    // Handle error
}

// Output the hook script for the user to add to their shell config
fmt.Println(hookScript)
```

## CLI Integration

The Auto-Activation Manager is designed to be used with the following CLI commands:

### `unirtm activate`

Generates the initial activation script and shell hook:

```bash
# Add to ~/.bashrc or ~/.zshrc
eval "$(unirtm activate bash)"

# Or for fish
unirtm activate fish | source
```

### `unirtm hook-env`

Called by the shell hook to get activation changes:

```bash
# Called automatically by the shell hook
unirtm hook-env --shell bash
```

This command:
1. Detects the current directory
2. Compares with the previous directory (from `$UNIRTM_OLD_PWD`)
3. Determines the activation action needed
4. Outputs the activation/deactivation script

## Workflow Examples

### Entering a Project

```
User: cd ~/projects/myapp
Shell Hook: Detects directory change
UniRTM: Finds unirtm.toml in ~/projects/myapp
UniRTM: Generates activation script
Shell: Evaluates script
Result: Project toolchain activated
```

### Leaving a Project

```
User: cd ~
Shell Hook: Detects directory change
UniRTM: No configuration file found in ~
UniRTM: Generates deactivation script
Shell: Evaluates script
Result: Previous environment restored
```

### Switching Projects

```
User: cd ~/projects/another-app
Shell Hook: Detects directory change
UniRTM: Finds unirtm.toml in ~/projects/another-app
UniRTM: Generates switch script (deactivate + activate)
Shell: Evaluates script
Result: Switched to new project's toolchain
```

## Environment Variables

The Auto-Activation Manager uses the following environment variables:

- **`UNIRTM_OLD_PWD`**: Tracks the previous working directory for change detection
- **`UNIRTM_ACTIVATION_SCOPE`**: Indicates the activation scope (global or project)
- **`UNIRTM_PROJECT_DIR`**: The currently active project directory
- **`UNIRTM_<TOOL>_VERSION`**: Active version for each tool (e.g., `UNIRTM_NODE_VERSION`)

## Testing

The implementation includes comprehensive unit tests covering:

- Configuration file detection
- Activation action determination
- Activation script generation
- Deactivation script generation
- Project switching
- Shell hook generation
- Edge cases (nested projects, symlinks, empty states)

Run tests with:

```bash
go test ./internal/service -run TestAutoActivation
```

## Future Enhancements

Potential improvements for future versions:

1. **Configuration Caching**: Cache parsed configuration files to avoid repeated I/O
2. **Incremental Updates**: Only update changed environment variables instead of full deactivation/activation
3. **Performance Monitoring**: Track activation time and optimize slow operations
4. **Custom Hooks**: Allow users to define custom activation/deactivation hooks
5. **Multi-Project Support**: Support multiple active projects simultaneously
6. **Activation History**: Track activation history for debugging and rollback

## References

- **Requirements**: See `requirements.md` section 15 (Environment Activation)
- **Design**: See `design.md` section on Activation Manager
- **Related Code**: `activation.go` (base Activation Manager)
- **Tests**: `auto_activation_test.go`
