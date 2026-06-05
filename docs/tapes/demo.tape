# Output destination
Output docs/tapes/demo.gif

# Setup Terminal Options
Set FontSize 22
Set Width 1200
Set Height 600
Set Padding 40
Set Theme "Catppuccin Macchiato"
Set WindowBar Colorful
Set Framerate 24

# Set up a fake environment or clear screen if needed
Hide
Type 'export PATH=$(echo $PATH | tr ":" "\n" | grep -v "go" | grep -v "node" | grep -v "python" | paste -sd ":" -)'
Enter
Type "clear"
Enter
Show

# Initial wait before starting
Sleep 1.5s

# 1. Show that tools are not installed
Type "go version"
Sleep 500ms
Enter
Sleep 1.5s

Type "node -v"
Sleep 500ms
Enter
Sleep 1.5s

Type "python --version"
Sleep 500ms
Enter
Sleep 1.5s

# 2. Run UniRTM to install the tools
# Notice the speed and the built-in vulnerability scan!
Type "unirtm use --global go@1.26.3 node@22.2.0 python@3.12.3"
Sleep 1s
Enter

# Wait for UniRTM to download, install, and run security scans (Trivy/Syft).
# Adjust the Sleep time if your actual installation takes longer or shorter.
Sleep 30s

# 3. Verify successful installation
Type "go version"
Sleep 500ms
Enter
Sleep 1s

Type "node -v"
Sleep 500ms
Enter
Sleep 1s

Type "python --version"
Sleep 500ms
Enter
Sleep 3s

# Clear screen and end
Type "clear"
Enter
Sleep 1s
