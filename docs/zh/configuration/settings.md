# 全局设置 (Settings)

虽然 `.unirtm.toml` 主要负责管理项目级别的配置，但 UniRTM 也支持适用于所有项目的全局设置。这些设置统一配置在 `~/.config/unirtm/config.toml` 文件中。

## 常用设置项

```toml
[settings]
# 启用对传统版本文件的支持，如 .nvmrc, .node-version, .python-version
legacy_version_file = true

# 提取后始终保留下载的压缩包（对调试网络问题很有用）
always_keep_download = false

# 检查 UniRTM 和插件自动更新的频率
plugin_autoupdate_last_check_duration = "7d"

# 配置并行下载和编译的默认并发任务数
jobs = 4

# 输出日志级别 (trace, debug, info, warn, error)
log_level = "info"
```

## 环境变量覆盖

你也可以使用环境变量来临时覆盖全局设置，只需在设置名称前加上 `UNIRTM_` 前缀即可。

例如，临时开启详细追踪日志：

```bash
UNIRTM_LOG_LEVEL=trace unirtm install
```

临时禁用对传统版本文件的解析：

```bash
UNIRTM_LEGACY_VERSION_FILE=false unirtm run build
```
