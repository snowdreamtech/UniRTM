# TOML Tasks

Tasks defined in `.unirtm.toml` can have sophisticated configurations:

```toml
[tasks.build]
description = "Build the project"
run = "go build"
env = { CGO_ENABLED = "0" }
```
