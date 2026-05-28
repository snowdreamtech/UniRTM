# CI/CD 集成

你可以在 CI/CD 管道（如 GitHub Actions、GitLab CI）中使用 UniRTM，以确保构建服务器使用的工具版本与本地机器完全一致。

```yaml
- uses: snowdreamtech/setup-unirtm@v1
- run: unirtm install
```