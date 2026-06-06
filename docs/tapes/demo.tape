# Output destination
Output docs/tapes/demo.gif
Output docs/tapes/demo.mp4

# Setup Terminal Options
Set FontSize 22
Set Width 1200
Set Height 600
Set Padding 40
Set Theme "Catppuccin Macchiato"
Set WindowBar Colorful
Set Framerate 24


# Initial wait before starting
Sleep 1.5s

# 1. Show that tools are not installed
Type "go version"
Sleep 500ms
Enter
Sleep 1.5s

# 2. Run UniRTM to install the tools
# Notice the speed and the built-in vulnerability scan!
Type "unirtm use --global go@1.26.4"
Sleep 1s
Enter

# Wait for UniRTM to download, install, and run security scans (Trivy/Syft).
# Adjust the Sleep time if your actual installation takes longer or shorter.
Sleep 60s

# 3. Verify successful installation
Type "go version"
Sleep 500ms
Enter
Sleep 1s


# Clear screen and end
Type "clear"
Enter
Sleep 1s
