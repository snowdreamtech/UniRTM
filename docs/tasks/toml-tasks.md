# TOML Tasks

UniRTM allows you to define tasks directly inline within your `.unirtm.toml` file. This is similar to scripts in `package.json` or targets in a `Makefile`, but with far more robust and structured configuration capabilities.

## Basic Definition

All tasks should be defined under the `[tasks]` node:

```toml
[tasks.build]
description = "Compile the project binary"
run = "go build -o ./bin/app ./cmd/app"
```

## Core Configuration Options

### 1. Dependency Management (`depends_on`)

If your task has prerequisites (e.g., you need to clean the directory and generate code before compiling), you can use `depends_on` to construct a Directed Acyclic Graph (DAG). UniRTM's task engine will **automatically parallelize** independent dependencies.

```toml
[tasks.build]
depends_on = ["clean", "generate"]
run = "go build"

[tasks.clean]
run = "rm -rf ./bin"

[tasks.generate]
run = "go generate ./..."
```

### 2. Environment Variable Injection (`env`)

You can inject task-specific environment variables. These are only visible within the context of the current task (and its spawned child processes). You can also use dynamic evaluation for secure secrets retrieval.

```toml
[tasks.deploy]
run = "sls deploy --stage prod"
[tasks.deploy.env]
AWS_REGION = "us-east-1"
# Dynamically extract secret
PROD_SECRET = "{{ exec(command='op read op://Vault/Prod/secret') }}"
```

### 3. Working Directory (`dir`)

Force a task to execute within a specific relative or absolute path:

```toml
[tasks.test-frontend]
dir = "./packages/frontend"
run = "npm run test"
```

### 4. Hidden Tasks (`hide`)

Some tasks serve purely as dependencies (internal utility functions). If you don't want them cluttering your UI when running `unirtm run`, you can hide them:

```toml
[tasks.pre-hook]
hide = true
run = "echo 'Preparing...'"
```

### 5. Multi-line Scripts & Shell Customization

For longer logic, you can write multi-line strings directly. UniRTM uses the system's default shell to execute commands, but you can override this.

```toml
[tasks.complex]
shell = "bash -c"
run = """
if [ -d "./tmp" ]; then
  echo "Temp exists"
else
  mkdir ./tmp
fi
"""
```

## Best Practices

- **Keep it Simple**: If your inline TOML task exceeds 10 lines of script logic, strongly consider migrating it to a standalone [File Task](./file-tasks.md).
- **Single Responsibility**: Break down monolithic scripts into fine-grained tasks using `depends_on` to maximize the performance benefits of parallel execution.
