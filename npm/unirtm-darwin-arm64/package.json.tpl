{
  "name": "@snowdreamtech/unirtm-darwin-arm64",
  "version": "{{VERSION}}",
  "description": "unirtm binary for macOS Apple Silicon (arm64)",
  "license": "MIT",
  "homepage": "https://github.com/snowdreamtech/unirtm",
  "repository": {
    "type": "git",
    "url": "git+https://github.com/snowdreamtech/unirtm.git"
  },
  "bugs": {
    "url": "https://github.com/snowdreamtech/unirtm/issues"
  },
  "os": ["darwin"],
  "cpu": ["arm64"],
  "bin": {
    "unirtm": "bin/unirtm"
  },
  "files": [
    "bin/",
    "LICENSE",
    "README.md",
    "README_zh-CN.md"
  ],
  "engines": {
    "node": ">=18"
  }
}
