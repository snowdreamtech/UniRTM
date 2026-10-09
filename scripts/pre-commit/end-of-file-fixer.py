#!/usr/bin/env python3

import os
import subprocess
import sys

git_symlinks = set()
try:
    out = subprocess.check_output(["git", "ls-files", "-s"], stderr=subprocess.DEVNULL).decode("utf-8", errors="ignore")
    for line in out.splitlines():
        if line.startswith("120000"):
            parts = line.split(maxsplit=3)
            if len(parts) >= 4:
                git_symlinks.add(os.path.normpath(parts[3]))
except Exception:
    pass

modified = []
for f in sys.argv[1:]:
    norm = os.path.normpath(f)
    if os.path.islink(f) or norm in git_symlinks:
        continue
    if os.path.isfile(f) and os.path.getsize(f) < 2 * 1024 * 1024:
        with open(f, "rb") as fh:
            content = fh.read()
        if not content:
            continue
        stripped = content.rstrip(b"\r\n")
        newline = b"\r\n" if b"\r\n" in content else b"\n"
        new_content = stripped + newline
        if content != new_content:
            with open(f, "wb") as fh:
                fh.write(new_content)
            modified.append(f)
for f in modified:
    print(f"Fixing end of file in: {f}")
sys.exit(1 if modified else 0)
