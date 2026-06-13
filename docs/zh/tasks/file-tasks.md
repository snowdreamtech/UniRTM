# 文件型任务

当你的自动化脚本逻辑逐渐变得庞大（包含复杂的 if/else、循环或错误处理）时，继续把它们塞在 `.unirtm.toml` 中会导致配置文件变得难以维护。

为此，UniRTM 原生支持**文件型任务 (File Tasks)**。

## 工作机制

UniRTM 会自动扫描项目根目录下的 `.unirtm/tasks/` 文件夹（或者直接名为 `.unirtm-tasks/` 的文件夹）。它会将该目录下的每一个可执行脚本文件直接映射为一个 UniRTM 任务。

文件的主名即为任务名。例如，文件名为 `build`，则你可以通过 `unirtm run build` 执行它。

### 目录结构示例

```text
my-project/
├── .unirtm.toml
├── .unirtm/
│   └── tasks/
│       ├── build         <-- "unirtm run build"
│       ├── test          <-- "unirtm run test"
│       └── deploy.py     <-- "unirtm run deploy.py"
```

> [!TIP]
> 如果你的文件没有执行权限（如 `chmod +x`），UniRTM 会在扫描时忽略它，或者在运行时提示拒绝访问。请确保你的脚本带有执行权限：`chmod +x .unirtm/tasks/build`。

## 编写脚本与 Shebang

UniRTM 对文件使用的语言没有任何限制！只要脚本的第一行包含合法的 `Shebang`（即 `#!` 开头），UniRTM 就可以执行它。

你可以用 Bash、Python、Node.js 甚至 Go (使用 `gorun`) 来编写任务：

```python
#!/usr/bin/env python3

import os
import sys

print("Deploying...")
if "PROD" in os.environ:
    print("Production deployment")
```

最美妙的是：**如果你的 Shebang 使用了受 UniRTM 环境管理的工具（如 Python 3.12），任务执行时将自动且绝对地使用该隔离环境下的解析器，而不是系统全局解析器！**

## 内联元数据 (Frontmatter)

你可以像在 Markdown 文件中写 Frontmatter 一样，在你的脚本文件头部声明元数据。UniRTM 的解析器会读取特定的注释，将其转化为任务配置。

**Bash 脚本示例：**

```bash
#!/usr/bin/env bash

# unirtm-description: 部署系统到 AWS
# unirtm-depends-on: build, lint
# unirtm-env: AWS_REGION=us-east-1
# unirtm-hide: false

echo "Deploying to AWS in $AWS_REGION..."
```

**Python 脚本示例：**

```python
#!/usr/bin/env python3

# unirtm-description: 运行数据清洗分析
# unirtm-depends-on: fetch-data

import pandas as pd
print("Data analysis started")
```

借助内联元数据，你的文件型脚本完全拥有了与 `TOML` 任务同等的依赖图并行编排能力！
