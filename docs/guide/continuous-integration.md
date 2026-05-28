# Continuous Integration

You can use UniRTM in your CI/CD pipelines (GitHub Actions, GitLab CI) to ensure your build servers use the exact same tool versions as your local machine.

```yaml
- uses: snowdreamtech/setup-unirtm@v1
- run: unirtm install
```