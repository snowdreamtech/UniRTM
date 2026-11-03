# Secrets Management

Modern application development inevitably relies on API Keys, database credentials, and certificates. Hardcoding these sensitive details in plaintext within `.unirtm.toml` or your code repository is an extremely dangerous practice.

UniRTM provides a flexible and highly secure **Secrets Management mechanism** that supports local hidden files as well as seamless integration with enterprise-grade credential managers.

## 1. Native `.env` File Support

The most fundamental approach is using `.env` files. When UniRTM activates an environment, it automatically looks for and loads the `.env` file located in the project's root.

```bash
# .env (Make sure this is in your .gitignore!)
DATABASE_URL="postgres://user:pass@localhost:5432/db"
STRIPE_SECRET_KEY="sk_test_123456"
```

You can also explicitly define which environment files to load in your `.unirtm.toml`, unlocking multi-environment configurations:

```toml
[env]
# Load specific credential files based on context
_.file = [".env", ".env.local"]
```

## 2. Dynamic Evaluation with External Secret Managers

This is one of UniRTM's most powerful features: **Dynamic Evaluation**.

We strongly advise against storing real production keys on any local filesystem. Leveraging the command evaluation capabilities in the `[env]` block, you can dynamically fetch variables at runtime using CLIs from tools like `1Password`, `AWS Secrets Manager`, or `HashiCorp Vault`.

```toml
[env]
# Dynamically fetch credentials using the 1Password CLI
GITHUB_TOKEN = "{{ exec(command='op read op://Private/GitHub/credential') }}"

# Extract and parse JSON dynamically via the AWS CLI
AWS_ACCESS_KEY = "{{ exec(command='aws secretsmanager get-secret-value --secret-id my-key --query SecretString --output text | jq -r .access_key') }}"
```

**Why is this exceptionally secure?**

- The credentials exist **exclusively in memory** during the active UniRTM context and never touch the disk.
- When running tasks via UniRTM's Task Runner, child processes securely inherit these in-memory credentials, which are immediately destroyed once the task completes.

## 3. Scoped Contextual Injection

Sometimes you want certain variables to be exposed **only** during specific operations to minimize the blast radius. You can inject Secrets solely within the scope of a specific Task:

```toml
[tasks.deploy]
run = "npm run deploy"
description = "Deploy application to production"
[tasks.deploy.env]
# This key is ONLY visible when you execute `unirtm run deploy`
PROD_DB_PASSWORD = "{{ exec(command='vault kv get -field=password secret/prod/db') }}"
```

## 4. Leakage Scanning & Auditing

UniRTM takes security seriously. If you attempt to hardcode apparent private keys or sensitive tokens directly in the plaintext `[env]` block, the best practice and strong recommendation is to immediately migrate them to the dynamic evaluation pattern to prevent accidental repository commits.
