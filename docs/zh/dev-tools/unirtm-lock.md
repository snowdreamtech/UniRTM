# 锁定文件 (unirtm.lock)

在团队协作和 CI/CD 流程中，仅仅声明 `node = "20"` 是远远不够的。为了保证“在我的机器上能跑，在服务器上也能跑”这一终极目标，UniRTM 引入了严格的锁定文件机制：`unirtm.lock`。

> [!IMPORTANT]
> 永远将 `unirtm.lock` 文件提交（Commit）到你的版本控制系统（如 Git）中。

## 为什么我们需要 Lockfile？

如果你在 `.unirtm.toml` 中定义了：

```toml
[tools]
node = "20.x"
golang = "latest"
```

这个配置在今天是 `node 20.11.0`，而在下个月可能就是 `node 20.12.0`。如果底层依赖发生了暗改，会导致不同开发者的环境出现致命的差异。

`unirtm.lock` 的存在就是为了记录解析后的**绝对精确版本**。

## 深入剖析 unirtm.lock

`unirtm.lock` 使用 JSON 格式生成，其内部不仅记录了版本，还记录了防篡改的哈希值。

```json
{
  "version": "1.0",
  "tools": {
    "node": {
      "version": "20.11.0",
      "source": "https://nodejs.org/dist/v20.11.0/node-v20.11.0-darwin-arm64.tar.gz",
      "checksum": "sha256:d8b2...a7f",
      "resolved_at": "2024-01-01T12:00:00Z"
    },
    "golang": {
      "version": "1.22.0",
      "source": "https://go.dev/dl/go1.22.0.darwin-arm64.tar.gz",
      "checksum": "sha256:f4e...b1c",
      "resolved_at": "2024-01-01T12:00:05Z"
    }
  }
}
```

### 安全与防篡改 (Supply Chain Security)

在下载并安装工具时，UniRTM 会强制计算下载包的 `sha256` 校验和，并与 `unirtm.lock` 中的 `checksum` 字段进行比对。
如果发现校验和不匹配，UniRTM 会立刻终止安装并抛出致命错误！这能有效防止中间人攻击（MITM）和官方镜像源被恶意替换等供应链攻击问题。

## 工作流最佳实践

1. **生成锁定文件**：当你运行 `unirtm install` 或修改了 `.unirtm.toml` 时，UniRTM 会自动更新或生成 `unirtm.lock`。
2. **升级依赖**：如果你确实想要将 `node 20.x` 升级到最新版本，可以运行 `unirtm update node`，这会重新解析并覆写锁文件中的具体版本和校验和。
3. **CI/CD 环境**：在 CI 管道中，建议运行带有锁文件验证的命令（或通过统一的 GitHub Actions）。如果发现当前的配置与 Lockfile 不匹配，它会直接报错退出，而不是默默更新，从而保障发布构建的绝对确定性。
