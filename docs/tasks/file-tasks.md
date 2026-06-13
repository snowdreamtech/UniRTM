# File Tasks

When your automation script logic grows complex (involving extensive if/else conditions, loops, or error handling), cramming it into a `.unirtm.toml` string becomes highly unmaintainable.

To solve this, UniRTM natively supports **File Tasks**.

## How it Works

UniRTM automatically scans the `.unirtm/tasks/` directory (or `.unirtm-tasks/`) located in your project root. It maps every executable script file inside this directory to a UniRTM task.

The basename of the file becomes the task name. For example, a file named `build` can be executed via `unirtm run build`.

### Directory Structure Example

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
> If your file lacks execution permissions (e.g., `chmod +x`), UniRTM may ignore it during scanning or throw a permission denied error at runtime. Ensure your scripts are executable: `chmod +x .unirtm/tasks/build`.

## Writing Scripts & Shebangs

UniRTM places zero restrictions on the language you use for File Tasks! As long as the first line of the script contains a valid `Shebang` (starting with `#!`), UniRTM can execute it.

You can write tasks in Bash, Python, Node.js, or even Go (using `gorun`):

```python
#!/usr/bin/env python3

import os
import sys

print("Deploying...")
if "PROD" in os.environ:
    print("Production deployment")
```

The best part: **If your Shebang references a tool managed by your UniRTM environment (e.g., Python 3.12), the task will automatically and strictly execute using that isolated interpreter, not your system's global one!**

## Inline Metadata (Frontmatter)

Similar to Frontmatter in Markdown files, you can declare metadata at the top of your script files. UniRTM's parser reads specific comments to dynamically generate task configurations.

**Bash Script Example:**

```bash
#!/usr/bin/env bash

# unirtm-description: Deploy system to AWS
# unirtm-depends-on: build, lint
# unirtm-env: AWS_REGION=us-east-1
# unirtm-hide: false

echo "Deploying to AWS in $AWS_REGION..."
```

**Python Script Example:**

```python
#!/usr/bin/env python3

# unirtm-description: Run data cleansing analysis
# unirtm-depends-on: fetch-data

import pandas as pd
print("Data analysis started")
```

With inline metadata, your File Tasks gain the exact same dependency DAG and parallel orchestration capabilities as inline TOML tasks!
