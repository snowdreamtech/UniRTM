# 环境隔离配置 (Environments)

在复杂的应用中，你经常需要根据程序是运行在 `development`（开发）、`staging`（测试）还是 `production`（生产）环境来加载不同的配置。

UniRTM 允许你在 `.unirtm.toml` 中定义特定环境的专属配置块。

## 定义专属环境

通过在 `[env.<环境名称>]` 下嵌套表格来定义环境。

```toml
[env]
# 在所有环境下都会加载的全局变量
APP_NAME = "My Awesome App"

[env.development]
# 仅在开发环境中加载
DATABASE_URL = "postgres://localhost:5432/dev_db"
DEBUG = "true"

[env.production]
# 仅在生产环境中加载
DATABASE_URL = "postgres://prod-db.internal:5432/prod_db"
DEBUG = "false"
```

## 激活特定环境

默认情况下，如果没有指定，UniRTM 会默认使用 `development` 开发环境。你可以通过设置 `UNIRTM_ENV` 变量来改变这一行为。

```bash
# 这将会加载 [env] 和 [env.production]
UNIRTM_ENV=production unirtm run start
```

## 开发工具环境隔离

环境隔离不仅局限于变量。你甚至可以根据不同的环境指定不同版本的工具！

```toml
[tools]
node = "20"

[tools.production]
# 在生产环境中强制使用精确锁定的版本
node = "20.11.1"
```
