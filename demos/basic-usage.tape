# VHS tape for basic KWatch usage demo
#
# This demo shows the actual KWatch TUI interface in action

# Set up the terminal
Output demos/kwatch-basic-usage.gif
Set FontSize 14
Set Width 1200  
Set Height 800
Set Theme "Molokai"

# Clear screen and start demo
Hide
Type "clear"
Enter
Show

# Show what we're about to do
Type "# KWatch - Real-time Project Monitoring TUI"
Enter
Sleep 1s

Type "# Let's launch the KWatch TUI interface for this project"
Enter
Sleep 1s

Type "./build/kwatch ."
Enter

# Wait for TUI to load and show interface
Sleep 3s

# Show TUI controls
Key "h"
Sleep 2s

# Go back to main view
Key "Escape"
Sleep 1s

# Force a refresh to show monitoring in action
Key "r"
Sleep 3s

# Switch to status view
Key "s"
Sleep 2s

# Switch to logs view
Key "l"
Sleep 2s

# Navigate through logs
Key "Down"
Sleep 0.5s
Key "Down"
Sleep 0.5s
Key "Up"
Sleep 1s

# Go back to main monitoring view
Key "s"
Sleep 2s

# Show one more refresh cycle
Key "r"
Sleep 3s

# Exit the TUI
Key "q"
Sleep 1s

Type "# KWatch provides real-time monitoring of:"
Enter
Type "# • TypeScript compilation (TSC)"
Enter
Type "# • Code linting (LINT)"
Enter
Type "# • Test execution (TEST)"
Enter
Type "# • GitHub Actions (GITHUB)"
Enter
Sleep 2s