# Pip Platform Packaging Implementation Plan

## Goal

>
## User Review Required
>
>
> [!IMPORTANT]
> 请确认以下决策：
>
> - 包名注册为 `unirtm` 还是 `snowdreamtech-unirtm`？
>
> - 二进制文件随 pip 包分发（wheel） 还是在安装时通过 `setup.py` 动态下载？（目前建议构建平台相关的 wheel）。
>

## Open Questions
>
> [!WARNING]

>

> - PyPI token 是否已配置至 GitHub Actions 密钥？

> - 有哪些目标 Python 版本需要特殊声明兼容性？

## Proposed Changes

#### [NEW] `setup.py`

- 配置 setuptools 构建流程，实现 wheel 打包配置。

- 现代 Python 打包标准配置。

### CI Integration

- 添加 `twine upload` 任务，自动发布到 PyPI。

## Verification Plan

### Automated Tests

- 配置 GitHub Actions `pip install .` 测试跨平台构建。
- `pytest` 测试二进制能否从 PATH 中成功被 Python 子进程调用。

### Manual Verification

- 下载生成的 `.whl` 文件，在本地执行 `pip install` 并检查 `unirtm version` 是否可运行。
