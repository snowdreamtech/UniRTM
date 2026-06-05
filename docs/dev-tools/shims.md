# Shims & PATH

Unlike traditional tool managers that rely on slow bash scripts, UniRTM uses native, lightweight symlink shims. By pointing all tools back to the high-performance Go `unirtm` binary, it achieves near-zero overhead routing while completely preventing your `$PATH` from exploding with dozens of tool directories.
