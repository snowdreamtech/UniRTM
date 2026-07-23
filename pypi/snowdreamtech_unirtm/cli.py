import os
import subprocess
import sys


def main():
    package_dir = os.path.dirname(os.path.abspath(__file__))
    binary_name = "unirtm.exe" if os.name == "nt" else "unirtm"
    binary_path = os.path.join(package_dir, "bin", binary_name)

    if not os.path.exists(binary_path):
        print(f"Error: UniRTM binary not found at {binary_path}", file=sys.stderr)
        sys.exit(1)

    try:
        result = subprocess.run([binary_path] + sys.argv[1:])
        sys.exit(result.returncode)
    except KeyboardInterrupt:
        sys.exit(130)


if __name__ == "__main__":
    main()
