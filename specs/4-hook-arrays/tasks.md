# Tasks: 支持 Hook 脚本数组

- [x] **Task 1:** Update `ToolConfig` to support `StringArray` for `PreInstall` and `PostInstall`.
  - Modify `internal/config/config.go`.
- [x] **Task 2:** Update `Task` to support `StringArray` for `Run`.
  - Modify `internal/config/config.go`.
- [x] **Task 3:** Update Hook execution to use `.Script()`.
  - Modify `internal/service/installation.go`.
- [x] **Task 4:** Update Task execution to use `.Script()`.
  - Modify `internal/task/native.go`. 中的任务执行逻辑，调用 `.Script()`。
- [x] 5. 修正由于类型更改而在其他服务文件（如 `migration.go`）中造成的类型不匹配。
- [x] 6. 修正并补充单元测试（`config_test.go`, `installation_test.go` 等）。
- [x] 7. 手动验证功能：运行带有数组形式脚本的测试配置文件。
