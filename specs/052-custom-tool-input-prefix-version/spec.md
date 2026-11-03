# 需求规范：自定义工具前缀选择与固定版本号录入 (Spec)

## 1. 业务背景与问题分析
在 UniRTM 的“添加工具”弹窗中，底部的“或输入自定义工具名”原本仅提供单一输入框，且后台自动将其版本硬编码为 `"latest"`。
这一实现存在如下核心问题：
1. **前缀不透明**：UniRTM 支持多种后端 Provider（如 `npm:`, `cargo:`, `pipx:`, `go:`, `github:` 等），但普通开发者缺乏前缀认知，不知该如何格式化；
2. **缺乏构建确定性**：`latest` 是移动浮动指针，在不同时间拉取的二进制或依赖不同，严重破坏了工作区环境的可重现性（Reproducibility）。UniRTM 主张确定性构建，必须使用**固定的确切版本号（Fixed Explicit Version）**；
3. **输入流割裂**：用户从外部文档或命令复制完整规范（如 `npm:@biomejs/biome@1.9.4`）时，无法智能拆解。

---

## 2. 核心功能规范与交互原则

### 2.1 复合分段输入栏 (Segmented Triple-Field Input Bar)
采用一体式三段复合控件组：
- **常用前缀下拉选择 (Prefix Selector)**：
  - `(默认 / 核心)`：无前缀，用于 `zig`, `elixir`, `php`, `julia` 等语言；
  - `npm:`：Node.js 生态 CLI 工具（支持带 `@scope` 组织包）；
  - `cargo:`：Rust 生态工具；
  - `pipx:`：Python 生态 CLI 工具；
  - `go:`：Go 模块二进制工具；
  - `github:`：GitHub Releases 归档独立二进制；
  - `aqua:`：Aqua 声明式工具包生态；
  - `asdf:`：asdf 社区插件生态。
- **具体工具名输入框 (Tool Name Input)**：
  - 针对所选前缀动态调整 `placeholder` 示例；
  - 支持直接输入包名，如 `@biomejs/biome`, `ripgrep`, `ruff` 等。
- **固定版本号输入框 (Explicit Fixed Version Input)**：
  - **严禁推荐或默认 `latest`**；
  - 提示明确固定版本号（Placeholder：`如: 1.0.0`）；
  - 必须输入有效版本号，未输入时禁用提交。
- **添加动作按钮 (Add Button)**：
  - 校验“工具名”和“版本号”均不为空，且符合命名格式；
  - 提交后写入当前作用域（工作区 `.unirtm.toml` 或用户配置）并触发下载。

### 2.2 智能粘贴自动拆解 (Smart Paste Parsing)
当用户将完整格式文本粘贴到工具输入框时（例如 `npm:@biomejs/biome@1.9.4`、`cargo:ripgrep@14.1.0` 或 `zig@0.13.0`）：
- 自动识别已知前缀并切换前缀下拉选项；
- 提取工具名填入工具名框；
- 提取 `@` 后的版本号填入固定版本号框。

### 2.3 确定性配置实时预览 (Deterministic Config Live Preview)
在输入框下方实时呈现最终生成的 `.unirtm.toml` 格式：
`📝 即将写入 .unirtm.toml: [tools] "npm:@biomejs/biome" = "1.9.4"`

---

## 3. 验收标准
1. 下拉菜单覆盖原生/默认、`npm:`、`cargo:`、`pipx:`、`go:`、`github:`、`aqua:`、`asdf:` 等常用生态；
2. 彻底移除 `latest` 默认值与提示，必须由用户输入明确版本；
3. 支持智能粘贴完整 Spec 并自动拆解三段；
4. 实时展示确定性配置键值预览；
5. 53 种语言包包含国际化字段；
6. 单元测试与类型检查 0 报错通过。
