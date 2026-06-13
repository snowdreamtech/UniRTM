# 配置 TOML 任务

UniRTM 允许你在 `.unirtm.toml` 文件中直接内联定义任务。这类似于 `package.json` 中的 scripts 或者 `Makefile` 中的 target，但它拥有更强大、更结构化的配置能力。

## 基础定义

所有的任务都应当定义在 `[tasks]` 节点下：

```toml
[tasks.build]
description = "编译项目二进制文件"
run = "go build -o ./bin/app ./cmd/app"
```

## 核心配置项

### 1. 依赖管理 (`depends_on`)

如果你的任务有前置要求（例如：编译前需要清理目录和生成代码），你可以使用 `depends_on` 构建有向无环图 (DAG)。UniRTM 的任务引擎会**自动并行**执行无关联的依赖。

```toml
[tasks.build]
depends_on = ["clean", "generate"]
run = "go build"

[tasks.clean]
run = "rm -rf ./bin"

[tasks.generate]
run = "go generate ./..."
```

### 2. 环境变量注入 (`env`)

你可以在任务级别注入专属的环境变量。这些变量仅在当前任务（及其子进程）中可见。你还可以使用动态评估来获取密钥。

```toml
[tasks.deploy]
run = "sls deploy --stage prod"
[tasks.deploy.env]
AWS_REGION = "us-east-1"
# 动态提取密码
PROD_SECRET = "{{ exec(command='op read op://Vault/Prod/secret') }}"
```

### 3. 工作目录 (`dir`)

强制任务在指定的相对或绝对路径中执行：

```toml
[tasks.test-frontend]
dir = "./packages/frontend"
run = "npm run test"
```

### 4. 隐藏任务 (`hide`)

有些任务仅仅是为了作为其他任务的依赖（内部工具函数），你不想让它们在使用 `unirtm run` 列出时污染视线，可以将它们隐藏：

```toml
[tasks.pre-hook]
hide = true
run = "echo 'Preparing...'"
```

### 5. 多行脚本与 Shell 指定

对于较长的逻辑，你可以直接写多行字符串。UniRTM 默认使用系统默认 Shell 执行，但你可以覆盖它。

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

## 最佳实践

- **保持简单**：如果你的内联 TOML 任务超过了 10 行脚本，请考虑将其迁移为独立的[文件型任务](./file-tasks.md)。
- **职责单一**：通过 `depends_on` 将大任务拆解为细粒度的小任务，以便最大化并行执行的性能收益。
