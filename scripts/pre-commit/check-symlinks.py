#!/usr/bin/env python3

import os
import sys

errors = 0
for f in sys.argv[1:]:
    if os.path.islink(f):
        target = os.readlink(f)
        if os.path.isabs(target):
            print(f"Absolute symlink forbidden: {f} -> {target}")
            errors += 1
        elif not os.path.exists(f):
            print(f"Broken symlink: {f} -> {target}")
            errors += 1

sys.exit(1 if errors > 0 else 0)
