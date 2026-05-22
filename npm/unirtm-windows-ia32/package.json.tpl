{
  "name": "@snowdreamtech/unirtm-windows-ia32",
  "version": "{{VERSION}}",
  "description": "unirtm binary for Windows ia32 (i386)",
  "license": "MIT",
  "homepage": "https://github.com/snowdreamtech/unirtm",
  "repository": {
    "type": "git",
    "url": "git+https://github.com/snowdreamtech/unirtm.git"
  },
  "bugs": {
    "url": "https://github.com/snowdreamtech/unirtm/issues"
  },
  "os": ["win32"],
  "cpu": ["ia32"],
  "bin": {
    "unirtm": "bin/unirtm.exe"
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
