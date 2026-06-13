# 密钥管理机制 (Secrets Management)

现代应用开发离不开 API Keys、数据库凭证和证书。将这些敏感信息明文写在 `.unirtm.toml` 或代码仓库中是极度危险的行为。

UniRTM 提供了一套灵活且安全的 **Secrets 管理机制**，它不仅支持读取本地的隐蔽文件，更能与成熟的第三方凭证管理器无缝对接。

## 1. 原生支持 `.env` 文件

最基础且常用的方式是 `.env` 文件。当 UniRTM 的环境隔离被触发时，它会自动寻找并加载项目根目录下的 `.env` 文件。

```bash
# .env (务必将其加入 .gitignore)
DATABASE_URL="postgres://user:pass@localhost:5432/db"
STRIPE_SECRET_KEY="sk_test_123456"
```

你也可以在 `.unirtm.toml` 中显式指定要加载的环境文件，从而支持多环境切换：

```toml
[env]
# 根据特定的上下文加载特定的凭证文件
_.file = [".env", ".env.local"]
```

## 2. 外部 Secret Manager 动态求值

这是 UniRTM 最强大的特性之一：**动态评估 (Dynamic Evaluation)**。

我们强烈反对将真实密钥存储在任何本地文件系统中。借助于 `[env]` 区域的命令求值能力，你可以直接在运行时调用 `1Password`、`AWS Secrets Manager` 或 `Vault` 的 CLI 来提取变量。

```toml
[env]
# 调用 1Password CLI 动态读取密钥
GITHUB_TOKEN = "{{ exec(command='op read op://Private/GitHub/credential') }}"

# 调用 AWS CLI 动态提取并解析 JSON
AWS_ACCESS_KEY = "{{ exec(command='aws secretsmanager get-secret-value --secret-id my-key --query SecretString --output text | jq -r .access_key') }}"
```

**为什么这非常安全？**

- 密钥仅在 UniRTM 环境激活的**内存上下文中**存在，绝不会落地成为磁盘文件。
- 如果你使用了 UniRTM 的 Task Runner 执行构建或部署任务，子进程会安全地继承这些内存密钥，并在任务结束后自动销毁。

## 3. 安全环境上下文过滤

有时候你可能希望某些变量**仅在**执行特定任务时暴露，以限制爆炸半径。你可以在 Task 内部单独注入 Secrets：

```toml
[tasks.deploy]
run = "npm run deploy"
description = "部署到生产环境"
[tasks.deploy.env]
# 只有执行 unirtm run deploy 时，这个密钥才可见
PROD_DB_PASSWORD = "{{ exec(command='vault kv get -field=password secret/prod/db') }}"
```

## 4. 泄露扫描与审计

UniRTM 极度重视安全。配合安全检查工具链，如果探测到你正在试图将类似私钥、敏感 Token 明文硬编码到 `[env]` 字典中，我们的推荐规范是立刻移出并改用动态求值模式，防患于未然。
