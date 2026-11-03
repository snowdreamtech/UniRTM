# Configuration Environments

In complex applications, you often need different configurations depending on whether you are running in `development`, `staging`, or `production`.

UniRTM allows you to define environment-specific sections within your `.unirtm.toml`.

## Defining Environments

Environments are defined by nesting tables under `[env.<environment_name>]`.

```toml
[env]
# Global variables loaded in all environments
APP_NAME = "My Awesome App"

[env.development]
# Only loaded in development
DATABASE_URL = "postgres://localhost:5432/dev_db"
DEBUG = "true"

[env.production]
# Only loaded in production
DATABASE_URL = "postgres://prod-db.internal:5432/prod_db"
DEBUG = "false"
```

## Activating an Environment

By default, UniRTM assumes the `development` environment if nothing is specified. You can change this by setting the `UNIRTM_ENV` variable.

```bash
# Loads [env] and [env.production]
UNIRTM_ENV=production unirtm run start
```

## Tool Environments

Environments are not limited to variables. You can also specify different tool versions based on the environment!

```toml
[tools]
node = "20"

[tools.production]
# Enforce a strict, locked version in production
node = "20.11.1"
```
