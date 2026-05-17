import os
import re

scripts_dir = "/Users/snowdream/Workspace/snowdreamtech/UniRTM/scripts"
patterns = [
    # Match raw unirtm invocations: e.g. unirtm exec, run_quiet unirtm, run_with_timeout unirtm, etc.
    # We want to ignore lines containing _G_UNIRTM_BIN, _UNIRTM_TOML, UNIRTM_CONFIG, or comment character '#'
    r'(?<![A-Za-z0-9_-])unirtm(?![A-Za-z0-9_-])'
]

raw_calls = []

for root, dirs, files in os.walk(scripts_dir):
    for file in files:
        if file.endswith('.sh'):
            file_path = os.path.join(root, file)
            with open(file_path, 'r', encoding='utf-8') as f:
                lines = f.readlines()
            for idx, line in enumerate(lines, 1):
                clean_line = line.strip()
                if not clean_line or clean_line.startswith('#'):
                    continue
                # Find unirtm but check if it's raw
                matches = re.finditer(r'(?<![A-Za-z0-9_\-\$\{\"\'])unirtm(?![A-Za-z0-9_])', line)
                for match in matches:
                    # Ignore comments at the end of the line
                    part_before = line[:match.start()]
                    if '#' in part_before:
                        continue
                    # Ignore assignments like local _UNIRTM_VAR or similar, or environment variables like UNIRTM_YES
                    if re.search(r'\b(UNIRTM_[A-Z0-9_]+|_UNIRTM_[A-Z0-9_]+)\b', line):
                        # But make sure the match itself isn't a separate raw call
                        pass
                    # Let's log it
                    raw_calls.append({
                        "file": os.path.relpath(file_path, scripts_dir),
                        "line": idx,
                        "content": clean_line
                    })

print(f"Found {len(raw_calls)} raw unirtm calls:")
for rc in raw_calls:
    print(f"{rc['file']}:{rc['line']}: {rc['content']}")
