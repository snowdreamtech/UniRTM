#!/usr/bin/env python3
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License.

import os
import sys
import pty
import fcntl
import termios
import struct
import select
import json

def set_winsize(fd, rows, cols):
    try:
        fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack("HHHH", rows, cols, 0, 0))
    except Exception:
        pass

def main():
    cols = int(sys.argv[1]) if len(sys.argv) > 1 and sys.argv[1].isdigit() else 120
    rows = int(sys.argv[2]) if len(sys.argv) > 2 and sys.argv[2].isdigit() else 30
    shell = sys.argv[3] if len(sys.argv) > 3 and sys.argv[3] else "/bin/bash"
    cwd = sys.argv[4] if len(sys.argv) > 4 and sys.argv[4] else os.getcwd()

    master, slave = pty.openpty()
    set_winsize(master, rows, cols)

    pid = os.fork()
    if pid == 0:
        os.close(master)
        os.setsid()
        try:
            fcntl.ioctl(slave, termios.TIOCSCTTY, 0)
        except Exception:
            pass
        os.dup2(slave, 0)
        os.dup2(slave, 1)
        os.dup2(slave, 2)
        os.close(slave)
        if os.path.isdir(cwd):
            try:
                os.chdir(cwd)
            except Exception:
                pass
        try:
            os.execvp(shell, [shell, "-l"])
        except Exception:
            os.execvp("/bin/sh", ["/bin/sh"])
    else:
        os.close(slave)

        stdin_fd = sys.stdin.fileno()
        stdout_fd = sys.stdout.fileno()

        # Set master non-blocking
        flags = fcntl.fcntl(master, fcntl.F_GETFL)
        fcntl.fcntl(master, fcntl.F_SETFL, flags | os.O_NONBLOCK)

        stdin_buf = bytearray()

        while True:
            rlist, _, _ = select.select([master, stdin_fd], [], [], 1.0)
            if master in rlist:
                try:
                    data = os.read(master, 4096)
                    if not data:
                        break
                    sys.stdout.buffer.write(data)
                    sys.stdout.buffer.flush()
                except BlockingIOError:
                    pass
                except OSError:
                    break

            if stdin_fd in rlist:
                try:
                    chunk = os.read(stdin_fd, 4096)
                    if not chunk:
                        break
                    
                    # Check if it contains a JSON control message like {"type":"resize",...}
                    if b'{"type":' in chunk:
                        # Split by closing brace or newline to safely parse concatenated frames
                        raw_str = chunk.decode('utf-8', 'ignore')
                        start = 0
                        has_control = False
                        while True:
                            idx = raw_str.find('{"type":', start)
                            if idx == -1:
                                break
                            end_idx = raw_str.find('}', idx)
                            if end_idx == -1:
                                break
                            has_control = True
                            json_slice = raw_str[idx:end_idx+1]
                            try:
                                msg = json.loads(json_slice)
                                if msg.get('type') == 'resize':
                                    r = int(msg.get('rows', rows))
                                    c = int(msg.get('cols', cols))
                                    set_winsize(master, r, c)
                                elif msg.get('type') == 'kill':
                                    return
                            except Exception:
                                pass
                            start = end_idx + 1
                        if has_control:
                            continue
                    
                    os.write(master, chunk)
                except Exception:
                    break

            # Check if child process has exited
            wpid, _ = os.waitpid(pid, os.WNOHANG)
            if wpid != 0:
                break

        try:
            os.kill(pid, 9)
        except Exception:
            pass
        os.close(master)

if __name__ == "__main__":
    main()
