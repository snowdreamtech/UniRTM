# 配置 TOML 任务

在 `.unirtm.toml` 中定义的任务可以具有复杂的配置属性：

```toml
[tasks.build]
description = "编译项目"
run = "go build"
env = { CGO_ENABLED = "0" }
```
