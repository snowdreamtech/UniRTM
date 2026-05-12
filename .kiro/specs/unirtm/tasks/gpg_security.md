# Tasks: UniRTM GPG Security Implementation

## Phase 1: 原子化下载与权限加固 (Atomic Download & Permissions) [DONE]
- [x] 1.1 实现 `.tmp.<rand>` 临时文件下载机制
- [x] 1.2 设置下载临时文件权限为 `0600`
- [x] 1.3 实现下载完成后的原子化 `os.Rename`

## Phase 2: Zip Slip 防护与目录平铺 (Zip Slip & Flattening) [DONE]
- [x] 2.1 修复 Gradle 等原生工具的目录平铺逻辑
- [x] 2.2 实现 `validateInstallDir` 路径安全校验函数
- [x] 2.3 增加对恶意软链接（指向外部敏感文件）的拦截
- [x] 2.4 增加 `http` 协议风险警告

## Phase 3: GPG 校验核心引擎 (GPG Verification Engine)
- [ ] 3.1 集成 GPG 核心库 (Integrate GPG Core Library)
    - [ ] 引入 `golang.org/x/crypto/openpgp` (或系统 `gpg` 封装)
    - [ ] 实现核心校验接口 `gpg.Verify(sig, data, fingerprint)`
- [ ] 3.3 完善 Registry 与指纹发现 (Registry & Discovery)
    - [ ] 更新 `native_provider` 支持从 Registry 获取官方公钥指纹
    - [ ] 实现自动从公钥服务器（如 keys.openpgp.org）拉取未知 Key 的逻辑
- [ ] 3.4 交互式信任流程 (Interactive Trust Flow)
    - [ ] 实现 TTY 模式下的 PTerm 询问窗口
    - [ ] 实现非 TTY (CI/CD) 模式下的自动退避（报错退出）策略

## Phase 4: 锁文件与 CI/CD 增强 (Lockfile & CI/CD)
- [ ] 4.1 锁文件扩展 (Extend Lockfile)
    - [ ] 在 `unirtm.lock` 中持久化记录已验证的公钥指纹
- [ ] 4.2 环境策略控制 (Security Policy)
    - [ ] 增加 `UNIRTM_GPG_VERIFY` 环境变量支持 (`strict`, `warn`)
- [ ] 4.3 最终验收测试 (Final E2E Testing)
    - [ ] 编写模拟恶意投毒、签名缺失等场景的集成测试
