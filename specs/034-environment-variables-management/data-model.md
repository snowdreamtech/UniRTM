# Data Model: Environment Variables (034-environment-variables-management)

## 1. Environment Variable Entity

```typescript
export interface EnvVarItem {
  name: string;          // 环境变量名称（如 GITHUB_TOKEN、PORT、NODE_ENV）
  value: string;         // 环境变量值
  scope: "workspace" | "user"; // 所属作用域
  isSensitive: boolean;  // 是否被系统判定为敏感字段
  isRevealed?: boolean;  // 前端运行时状态：是否已点击眼睛显示明文
}

export interface UpsertEnvVarPayload {
  key: string;           // 新键名
  original_key?: string; // 旧键名（编辑时若更改了 Key 则传递以清除旧键）
  value: string;         // 键值
  scope: "workspace" | "user";
}
```

## 2. API Contracts

### GET `/api/env?scope={workspace|user}`
- **Query Params**:
  - `scope`: `"workspace"` | `"user"`
- **Response**:
```json
{
  "env": [
    {
      "name": "GITHUB_TOKEN",
      "value": "ghp_xxxxxxxxxxxx",
      "scope": "workspace",
      "is_sensitive": true
    },
    {
      "name": "PORT",
      "value": "8080",
      "scope": "workspace",
      "is_sensitive": false
    }
  ],
  "total": 2,
  "scope": "workspace"
}
```

### POST / PUT `/api/env`
- **Request Body**:
```json
{
  "key": "API_SECRET",
  "original_key": "OLD_SECRET",
  "value": "super-secret-key-123",
  "scope": "workspace"
}
```
- **Response**:
```json
{
  "success": true,
  "env": {
    "name": "API_SECRET",
    "value": "super-secret-key-123",
    "scope": "workspace",
    "is_sensitive": true
  }
}
```

### DELETE `/api/env`
- **Query or Body**:
  - `key`: string
  - `scope`: `"workspace"` | `"user"`
- **Response**:
```json
{
  "success": true,
  "deleted_key": "API_SECRET"
}
```
