# VHS tape for basic KWatch usage demo
#
# This demo shows the actual KWatch TUI interface in action

# Set up the terminal
Output demos/kwatch-basic-usage.gif
Set FontSize 14
Set Width 1200  
Set Height 800
Set Theme "Molokai"

# Start clean
Hide
Type "clear"
Enter
Show

# Launch KWatch TUI directly - no commentary, just the app
Type "./build/kwatch ."
Enter

# Wait for TUI to fully load and show monitoring interface
Sleep 4s

# Show help screen to demonstrate controls
Type "h"
Sleep 3s

# Exit help back to main monitoring view
Escape
Sleep 1s

# Force a refresh to show real-time monitoring
Type "r"
Sleep 4s

# Switch to status view to show different interface modes
Type "s"
Sleep 3s

# Switch to logs view
Type "l"
Sleep 3s

# Navigate through logs to show interactivity
Down
Sleep 1s
Down
Sleep 1s
Up
Sleep 1s

# Return to main monitoring view
Escape
Sleep 2s

# One final refresh to show live monitoring
Type "r"
Sleep 4s

# Exit the TUI
Type "q"
Sleep 1s