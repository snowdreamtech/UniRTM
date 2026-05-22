{
  "name": "@snowdreamtech/unirtm",
  "version": "{{VERSION}}",
  "description": "UniRTM - Universal Runtime Manager: enterprise-grade foundational toolchain for multi-AI IDE collaboration",
  "license": "MIT",
  "homepage": "https://github.com/snowdreamtech/unirtm",
  "repository": {
    "type": "git",
    "url": "git+https://github.com/snowdreamtech/unirtm.git"
  },
  "bugs": {
    "url": "https://github.com/snowdreamtech/unirtm/issues"
  },
  "keywords": [
    "unirtm",
    "runtime",
    "manager",
    "toolchain",
    "ai",
    "ide"
  ],
  "bin": {
    "unirtm": "install.js"
  },
  "scripts": {
    "postinstall": "node install.js"
  },
  "files": [
    "install.js",
    "LICENSE",
    "README.md",
    "README_zh-CN.md"
  ],
  "optionalDependencies": {
    "@snowdreamtech/unirtm-darwin-arm64": "{{VERSION}}",
    "@snowdreamtech/unirtm-darwin-x64": "{{VERSION}}",
    "@snowdreamtech/unirtm-linux-x64": "{{VERSION}}",
    "@snowdreamtech/unirtm-linux-arm64": "{{VERSION}}",
    "@snowdreamtech/unirtm-linux-ia32": "{{VERSION}}",
    "@snowdreamtech/unirtm-linux-arm": "{{VERSION}}",
    "@snowdreamtech/unirtm-linux-arm-5": "{{VERSION}}",
    "@snowdreamtech/unirtm-linux-arm-6": "{{VERSION}}",
    "@snowdreamtech/unirtm-linux-loong64": "{{VERSION}}",
    "@snowdreamtech/unirtm-linux-ppc64le": "{{VERSION}}",
    "@snowdreamtech/unirtm-linux-riscv64": "{{VERSION}}",
    "@snowdreamtech/unirtm-linux-s390x": "{{VERSION}}",
    "@snowdreamtech/unirtm-windows-x64": "{{VERSION}}",
    "@snowdreamtech/unirtm-windows-arm64": "{{VERSION}}",
    "@snowdreamtech/unirtm-windows-ia32": "{{VERSION}}"
  },
  "engines": {
    "node": ">=18"
  }
}
