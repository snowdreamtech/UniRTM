# Quickstart: Hook Install Hybrid Resolution

## Purpose

Validate that `unirtm hook install` safely injects the environment bootstrap block without destroying existing hook scripts, and that it successfully bootstraps the environment for isolated CI/CD runners.

## Prerequisites

- UniRTM binary built locally.

## Scenario 1: Non-Destructive Installation

1. Create a mock pre-commit hook that simulates a user's custom script.

   ```bash
   mkdir -p .git/hooks
   echo '#!/bin/sh' > .git/hooks/pre-commit
   echo 'echo "My Custom Logic"' >> .git/hooks/pre-commit
   chmod +x .git/hooks/pre-commit
   ```

2. Run the install command.

   ```bash
   ./unirtm hook install pre-commit
   ```

3. Verify the contents of `.git/hooks/pre-commit`. It should contain the UniRTM block immediately after `#!/bin/sh`, followed by `echo "My Custom Logic"`.

## Scenario 2: Headless Environment Execution

1. Build UniRTM into the root of the project.

   ```bash
   go build -o unirtm ./cmd/unirtm
   ```

2. Clear the environment and execute the hook.

   ```bash
   env -i PATH=/usr/bin:/bin sh -c '.git/hooks/pre-commit'
   ```

3. The hook should execute without "command not found" errors, and should print "My Custom Logic".
