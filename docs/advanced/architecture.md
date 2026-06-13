# Architecture Deep Dive

UniRTM is engineered from the ground up to be a high-performance, **100% native Go** application. Unlike legacy tool managers that rely heavily on shell scripts, bash hooks, or Ruby plugins, UniRTM executes directly as a compiled binary. This design choice provides exceptional speed, strict memory safety, and cross-platform reliability.

## Core Tenets

1. **Zero Shell Pollution**: UniRTM does not inject complex initialization scripts into your `.bashrc` or `.zshrc`. It dynamically modifies the `PATH` during tool execution and immediately restores it.
2. **Native Performance**: By avoiding subshells (`bash -c`), UniRTM reduces the overhead of process spawning to almost zero.
3. **Concurrency First**: All heavy operations—such as remote version resolution, artifact downloading, and checksum validation—are executed concurrently using Go's lightweight goroutines.

## High-Level Architecture

The UniRTM core is divided into four primary subsystems:

1. **The Command Router (`cli` layer)**
2. **The Resolution Pipeline (`resolver` layer)**
3. **The Execution Engine (`engine` layer)**
4. **The Storage & Caching Layer (`cache` layer)**

```mermaid
graph TD
    User([User CLI / Shell]) --> CLI[CLI Command Router]

    subgraph Core Engine
        CLI --> Config[Configuration Parser .unirtm.toml]
        Config --> Resolver[Tool & Version Resolver]
        Resolver --> Engine[Goroutine Execution Engine]
    end

    subgraph Storage Subsystem
        Resolver -.-> Cache[(Cache & State Layer)]
        Engine --> Cache
    end

    Engine --> Sys[System Process / OS]

    classDef primary fill:#4f46e5,stroke:#312e81,stroke-width:2px,color:#fff;
    classDef secondary fill:#0ea5e9,stroke:#0369a1,stroke-width:2px,color:#fff;
    classDef storage fill:#10b981,stroke:#047857,stroke-width:2px,color:#fff;

    class CLI,Engine primary;
    class Config,Resolver secondary;
    class Cache storage;
```

### 1. The Command Router

When a user invokes a command (e.g., `unirtm use node@20`), the command router:

- Parses flags and arguments using `cobra`.
- Instantiates the application context.
- Validates the requested plugins against the remote registry.

### 2. The Resolution Pipeline

The hardest part of a tool manager is accurately resolving constraints (e.g., `node@^20.1`) against available remote versions.
UniRTM uses a multi-stage resolution pipeline:

```mermaid
sequenceDiagram
    participant User
    participant Resolver
    participant Cache
    participant RemoteRegistry

    User->>Resolver: Request `node@^20`
    Resolver->>Cache: Check Local Cache for Version List
    alt Cache Miss or Expired
        Resolver->>RemoteRegistry: Fetch `node` releases (Goroutine)
        RemoteRegistry-->>Resolver: JSON payload
        Resolver->>Cache: Save Version List (MsgPack)
    end
    Resolver->>Resolver: Apply SemVer Filter (`^20.x`)
    Resolver-->>User: Return Exact Version (`20.14.0`)
```

### 3. Execution Engine

UniRTM leverages Go's `os/exec` package to run tools natively. When you run a tool managed by UniRTM, the engine:

1. Calculates the exact `PATH` string required for the requested tools.
2. Clones the current environment variables.
3. Injects the new `PATH` and any `.unirtm.toml` variables into the clone.
4. Spawns the child process and pipes `stdin`, `stdout`, and `stderr` directly to the user's terminal without a middleman subshell.

This guarantees that signals (`SIGINT`, `SIGTERM`) are propagated correctly and terminal colors/TTY states are preserved.
