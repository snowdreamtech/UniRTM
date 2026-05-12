# Tasks: UniRTM GPG Security Implementation

## Phase 3: GPG 校验核心引擎 (GPG Verification Engine)
- [ ] 3.1 集成 GPG 核心库 (Integrate GPG Core Library)
    - [ ] 选择并引入 Go GPG 库 (如 `golang.org/x/crypto/openpgp`) 或系统 `gpg` 调用封装
    - [ ] 实现 `gpg.Verify(sigPath, dataPath, fingerprint)` 核心函数
- [ ] 3.2 完善 Registry 公钥指纹支持 (Registry Fingerprint Support)
    - [ ] 在 `internal/repository` 的模型中增加 `GPGKeys` 字段
    - [ ] 更新 `native_provider` 和 `generic_provider` 以从 Registry 读取指纹
- [ ] 3.3 实现公钥自动发现与下载 (Key Discovery & Download)
    - [ ] 实现从 `keys.openpgp.org` 或其他 KeyServer 下载公钥的逻辑
    - [ ] 实现本地公钥缓存/信任库管理
- [ ] 3.4 交互式信任流程 (Interactive Trust Flow)
    - [ ] 在 `installation.go` 中增加 TTY 检测
    - [ ] 实现 TTY 模式下的 PTerm 交互询问窗口

## Phase 4: 锁文件与 CI/CD 增强 (Lockfile & CI/CD)
- [ ] 4.1 扩展锁文件 Schema (Extend Lockfile Schema)
    - [ ] 在 `internal/lockfile` 中为每个工具条目增加 `gpg_fingerprint` 字段
- [ ] 4.2 实现指纹锁定逻辑 (Fingerprint Pinning Logic)
    - [ ] 安装成功后将验证通过的指纹回写至 `unirtm.lock`
    - [ ] 安装时优先校验 `unirtm.lock` 中的指纹
- [ ] 4.3 CI/CD 静默模式适配 (CI/CD Non-interactive Mode)
    - [ ] 实现非 TTY 环境下的“匹配即通过，不匹配即报错”逻辑
    - [ ] 增加 `UNIRTM_GPG_VERIFY` 环境变量控制开关
- [ ] 4.4 最终验证与测试 (Final Validation & Testing)
    - [ ] 编写 E2E 测试用例，模拟签名有效/无效/缺失的各种场景
    - [ ] 在 GitHub Actions 中验证静默模式行为
