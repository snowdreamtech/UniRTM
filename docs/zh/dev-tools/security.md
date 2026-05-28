# 安全检测 (Trivy & Syft)

UniRTM 在底层原生集成了安全特性。当你下载一个工具时，UniRTM 会自动通过 Syft 生成 SBOM 并使用 Trivy 扫描漏洞。同时，Gitleaks 确保没有任何凭证会通过环境变量意外泄露。