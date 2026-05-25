# Implementation Plan: pip Platform Packaging

**Branch**: `2-pip-packaging` | **Date**: 2026-05-22 | **Spec**: spec.md
**Input**: 参考 npm 打包发布方案，为 UniRTM 建立 PyPI 发布体系

## Summary

为 UniRTM 建立 Python wheel 发布体系。pip 的分发机制与 npm 有本质差异：

- **npm**：一个根包 + N 个平台子包，通过 `optionalDependencies` + `os/cpu` 字段筛选
- **pip**：N 个平台专属 wheel 文件，通过 wheel **文件名中的 platform tag** 自动匹配

pip 安装时自动根据当前系统平台选择正确的 `.whl` 文件，无需 root/wrapper 包。每个 wheel 内含二进制文件 + 轻量 Python 包装层，提供 `unirtm` 命令行入口。

## Technical Context

**Language**: Python 3（构建脚本）+ Go（二进制源）
**Packaging**: Python wheel (PEP 427/660)，wheel tag: `py3-none-{platform_tag}`
**Publishing**: twine → PyPI
**Build Script**: `pip/scripts/build.py`（Python 实现，跨平台）
**Target Platforms**: 15 个平台（与 npm 包一一对应）
**Constraints**: 二进制不提交到 Git；wheel 内的 Python 层仅做 exec 委托

## Core Design

### npm vs pip 对比

| 维度 | npm 方案 | pip 方案 |
|---|---|---|
| 平台选择机制 | `os`/`cpu` 字段 + `optionalDependencies` | wheel 文件名 platform tag |
| 包数量 | 16（1 根 + 15 平台） | 15（每平台 1 个 wheel） |
| 安装命令 | `npm install -g @snowdreamtech/unirtm` | `pip install unirtm` |
| 命令入口 | `bin` 字段 → install.js → 平台二进制 | `entry_points` → `__main__.py` → 平台二进制 |
| 版本文件 | `package.json` | `dist-info/METADATA` |
| 发布工具 | `npm publish` | `twine upload` |

### Wheel 文件命名规范

```
unirtm-{version}-py3-none-{platform_tag}.whl
```

- `py3`：兼容所有 Python 3（无 C 扩展，无 ABI 要求）
- `none`：无 ABI 依赖
- `{platform_tag}`：操作系统 + 架构标识

### Platform Tag 映射表

| GoReleaser dist/ 目录 | wheel platform_tag | 说明 |
|---|---|---|
| `unirtm_darwin_arm64_v8.0` | `macosx_11_0_arm64` | macOS Apple Silicon |
| `unirtm_darwin_amd64_v1` | `macosx_10_9_x86_64` | macOS Intel |
| `unirtm_linux_amd64_v1` | `manylinux_2_17_x86_64` | Linux x64 |
| `unirtm_linux_arm64_v8.0` | `manylinux_2_17_aarch64` | Linux ARM64 |
| `unirtm_linux_386_sse2` | `manylinux_2_17_i686` | Linux x86 32-bit |
| `unirtm_linux_arm_7` | `manylinux_2_17_armv7l` | Linux ARMv7 |
| `unirtm_linux_arm_5` | `linux_armv5l` | Linux ARMv5（嵌入式） |
| `unirtm_linux_arm_6` | `linux_armv6l` | Linux ARMv6（嵌入式） |
| `unirtm_linux_loong64` | `linux_loong64` | Linux LoongArch64 |
| `unirtm_linux_ppc64le_power8` | `manylinux_2_17_ppc64le` | Linux PowerPC LE |
| `unirtm_linux_riscv64_rva20u64` | `linux_riscv64` | Linux RISC-V 64 |
| `unirtm_linux_s390x` | `manylinux_2_17_s390x` | Linux IBM s390x |
| `unirtm_windows_amd64_v1` | `win_amd64` | Windows x64 |
| `unirtm_windows_arm64_v8.0` | `win_arm64` | Windows ARM64 |
| `unirtm_windows_386_sse2` | `win32` | Windows x86 32-bit |

### Wheel 内部结构

```
unirtm-{version}-py3-none-{platform_tag}.whl   (实际上是一个 ZIP 文件)
├── unirtm/
│   ├── __init__.py          # Python 模块入口，包含 main() 函数
│   ├── __main__.py          # 支持 python -m unirtm
│   └── bin/
│       └── unirtm[.exe]     # GoReleaser 构建的平台二进制
└── unirtm-{version}.dist-info/
    ├── METADATA             # 包元数据（名称、版本、描述等）
    ├── WHEEL                # wheel 格式声明
    ├── entry_points.txt     # console_scripts: unirtm = unirtm:main
    └── RECORD               # 文件哈希清单
```

### Python 包装层（`__init__.py`）

```python
import os
import sys
import subprocess

def _get_binary_path():
    """Locate the bundled unirtm binary."""
    binary_name = "unirtm.exe" if sys.platform == "win32" else "unirtm"
    binary = os.path.join(os.path.dirname(__file__), "bin", binary_name)
    if not os.path.isfile(binary):
        raise FileNotFoundError(
            f"unirtm binary not found at {binary}. "
            "Your platform may not be supported."
        )
    return binary

def main():
    """Entry point for the 'unirtm' console script."""
    binary = _get_binary_path()
    # Replace current process with the binary (Unix) or subprocess (Windows)
    if sys.platform == "win32":
        result = subprocess.run([binary] + sys.argv[1:])
        sys.exit(result.returncode)
    else:
        os.execv(binary, [binary] + sys.argv[1:])
```

## Project Structure

### Documentation (this feature)

```text
specs/2-pip-packaging/
├── plan.md              # 本文件
└── tasks.md             # 任务清单（待生成）
```

### Source Code (repository root)

```text
pip/                                   # PyPI 发布目录
├── scripts/
│   └── build.py                       # 核心构建脚本（Python）
│       # 读取 dist/ 目录，为每个平台生成 .whl 文件
├── templates/
│   ├── __init__.py.tpl                # Python 模块模板
│   ├── __main__.py.tpl                # __main__ 模板
│   └── METADATA.tpl                   # wheel METADATA 模板
└── dist/                              # 生成的 .whl 文件（不提交 Git）

.gitignore                             # 新增 pip/dist/ 排除规则
.github/workflows/goreleaser.yml       # 新增 pip publish 步骤
```

## Build Script Design (`pip/scripts/build.py`)

```
输入：
  --version 1.2.3          # 来自 dist/metadata.json
  --dist-dir dist/         # GoReleaser 输出目录
  --output-dir pip/dist/   # wheel 输出目录

流程：
  1. 读取 version（必须是有效 semver，无 v 前缀）
  2. 对每个平台：
     a. 确定 dist/ 中的二进制路径
     b. 创建临时 wheel 目录结构
     c. 复制二进制到 unirtm/bin/
     d. 生成 __init__.py、__main__.py（从模板渲染）
     e. 生成 dist-info/{METADATA,WHEEL,entry_points.txt}
     f. 计算所有文件的 SHA-256 哈希，生成 RECORD
     g. 打包为 .whl（ZIP 格式）
  3. 输出到 pip/dist/

标准：遵循 PEP 427（wheel 格式）、PEP 491（wheel 文件名）
```

## CI/CD 集成

在 `goreleaser.yml` 的 npm 发布步骤之后添加：

```yaml
- name: "🐍 Setup Python for PyPI Publishing"
  if: startsWith(github.ref, 'refs/tags/')
  uses: actions/setup-python@...
  with:
    python-version: "3.12"

- name: "🔨 Build pip Platform Wheels"
  if: startsWith(github.ref, 'refs/tags/')
  run: |
    pip install wheel
    VERSION=$(node -p "require('./dist/metadata.json').version")
    python pip/scripts/build.py --version "${VERSION}"

- name: "🚀 Publish pip Wheels to PyPI"
  if: startsWith(github.ref, 'refs/tags/')
  env:
    TWINE_USERNAME: __token__
    TWINE_PASSWORD: ${{ secrets.PYPI_TOKEN }}
  run: |
    pip install twine
    twine upload pip/dist/*.whl \
      --repository-url https://upload.pypi.org/legacy/ \
      --non-interactive
```

## Design Decisions

| 决策点 | 选择 | 原因 |
|---|---|---|
| 包名 | `unirtm`（无 scope）| PyPI 不支持 scope，`unirtm` 简洁直接 |
| wheel tag | `py3-none-{platform}` | 无 C 扩展，无 ABI 要求，兼容所有 Python 3 |
| Linux tag | `manylinux_2_17_*` | glibc 2.17 是主流 Linux 发行版最低兼容基线 |
| 发布工具 | `twine` | PyPI 官方推荐，稳定成熟 |
| 构建脚本语言 | Python | 标准库可处理 ZIP/wheel，无需额外依赖 |
| 二进制放置 | wheel 内 `unirtm/bin/` | 随包安装，路径固定，`__init__.py` 直接引用 |
| Windows 执行 | `subprocess.run` | Windows 不支持 `os.execv` |
| Unix 执行 | `os.execv` | 替换当前进程，信号传递正确，无 zombie 进程 |

## Required Secrets

| Secret 名称 | 用途 | 生成方式 |
|---|---|---|
| `PYPI_TOKEN` | PyPI 发布认证 | [pypi.org → Account Settings → API tokens] |

## Verification Plan

1. `python pip/scripts/build.py --version 0.0.1` → 生成 15 个 `.whl` 文件
2. `pip install pip/dist/unirtm-0.0.1-py3-none-macosx_11_0_arm64.whl`
3. `unirtm --help` → 正常输出
4. PyPI 发布后：`pip install unirtm` → 安装成功
