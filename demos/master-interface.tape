# VHS tape for Master KWatch interface demo
#
# This demo shows the actual master interface with real multi-directory monitoring

# Set up the terminal
Output demos/kwatch-master-interface.gif
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Molokai"

# Start clean
Hide
Type "clear"
Enter
Show

# Quick setup of demo directories
Type "mkdir -p /tmp/demo-projects/{app1,app2,api}"
Enter
Type "cd /tmp/demo-projects/app1 && echo '{\"name\":\"app1\",\"scripts\":{\"lint\":\"echo ok\"}}' > package.json"
Enter
Type "cd /tmp/demo-projects/app2 && echo '{\"name\":\"app2\",\"scripts\":{\"lint\":\"echo ok\"}}' > package.json"
Enter
Type "cd /tmp/demo-projects/api && echo '{\"name\":\"api\",\"scripts\":{\"lint\":\"echo ok\"}}' > package.json"
Enter
Type "cd /Users/kooshapari/Cwatch"
Enter
Sleep 1s

# Launch master interface with multiple directories - show real matrix
Type "./build/kwatch master /tmp/demo-projects/app1 /tmp/demo-projects/app2 /tmp/demo-projects/api"
Enter

# Wait for master interface to load and show the matrix
Sleep 6s

# Force a refresh to show monitoring in action
Type "r"
Sleep 4s

# Show help to demonstrate controls
Type "h"
Sleep 3s

# Go back to matrix view
Escape
Sleep 2s

# Another refresh cycle to show live updates
Type "r"
Sleep 4s

# Exit the master interface
Type "q"
Sleep 1s

# Clean up demo directories
Type "rm -rf /tmp/demo-projects"
Enter
Sleep 1s