# Implementation Plan: 支持 Hook 脚本数组

## 1. 概述
为 UniRTM 配置文件增加数组形式的 Hook 和 Task 脚本解析支持。解决单行长字符串维护困难的问题，通过自定义的 `StringArray` 类型无缝兼容原生字符串与数组反序列化。

## 2. 变更范围

### 2.1 引入 `StringArray` 类型
在 `internal/config/config.go` 中：
- 定义 `type StringArray []string`。
- 实现 `UnmarshalTOML(data interface{}) error`、`UnmarshalYAML(*yaml.Node) error`、`UnmarshalJSON([]byte) error` 以兼容单行文本和数组反序列化。
- 提供 `Script() string` 辅助函数：将切片元素使用 `\n` 连接成可直接由 `sh -c` 运行的多行脚本。

### 2.2 更新结构体定义
将以下结构体属性类型从 `string` 变更为 `StringArray`：
- `ToolConfig.PreInstall`
- `ToolConfig.PostInstall`
- `Task.Run`

### 2.3 修改业务执行逻辑
在 `internal/service/installation.go` 提取 Hook 脚本时：
```go
// 从
preInstall = tc.PreInstall
// 变为
preInstall = tc.PreInstall.Script()
```
在 `internal/task/native.go` 提取任务时：
```go
// 从
script := taskDef.Run
// 变为
script := taskDef.Run.Script()
```

### 2.4 测试用例更新
修改 `config_test.go` 和涉及 `PreInstall` 及 `Run` 构造的测试用例（如 `installation_test.go`, `migration.go` 等）。
