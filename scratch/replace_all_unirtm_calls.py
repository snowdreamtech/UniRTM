import os
import re

scripts_dir = "/Users/snowdream/Workspace/snowdreamtech/UniRTM/scripts"
count = 0

pattern = re.compile(r'(?<![A-Za-z0-9_\-\$\{\"\'.])unirtm(?=\s+(install|settings|x|exec|which|lock)\b)')

for root, dirs, files in os.walk(scripts_dir):
    for file in files:
        if file.endswith('.sh'):
            file_path = os.path.join(root, file)
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()

            lines = content.splitlines()
            modified = False
            for idx, line in enumerate(lines):
                # Skip full-line comments
                if line.strip().startswith('#'):
                    continue
                
                # Skip logging and printing lines to prevent nested quotes inside strings
                lower_line = line.lower()
                if any(x in lower_line for x in ['log_info', 'log_debug', 'log_error', 'log_warn', 'log_success', 'log_status', 'echo ', 'printf ']):
                    continue

                # Split line by '#' to avoid replacing in trailing comments
                parts = line.split('#', 1)
                code_part = parts[0]
                comment_part = parts[1] if len(parts) > 1 else ""

                # Let's perform replacement on code_part
                new_code_part, subs = pattern.subn(
                    '"${_G_UNIRTM_BIN:-unirtm}"',
                    code_part
                )
                if subs > 0:
                    lines[idx] = new_code_part + ('#' + comment_part if comment_part else "")
                    modified = True
                    count += subs

            if modified:
                new_content = "\n".join(lines) + "\n"
                with open(file_path, 'w', encoding='utf-8') as f:
                    f.write(new_content)
                print(f"Updated {os.path.relpath(file_path, scripts_dir)}")

print(f"Total replacements made: {count}")
