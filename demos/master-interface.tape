# VHS tape for Master KWatch interface demo
#
# This demo shows the actual master interface with real multi-directory monitoring

# Set up the terminal
Output demos/kwatch-master-interface.gif
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Molokai"

# Clear screen and start demo
Hide
Type "clear"
Enter
Show

# Title
Type "# KWatch - Master Interface"
Enter
Sleep 1s

Type "# Master KWatch monitors multiple projects from a unified interface"
Enter
Sleep 1s

# Create some example directories for demo
Type "mkdir -p /tmp/demo-projects/{app1,app2,api}"
Enter
Sleep 0.5s

Type "# Let's create basic project structures for our demo"
Enter

# Set up app1 with package.json
Type "cd /tmp/demo-projects/app1"
Enter
Type "echo '{\"name\":\"app1\",\"scripts\":{\"lint\":\"echo 'Linting app1...'\"}}' > package.json"
Enter

# Set up app2 with package.json  
Type "cd /tmp/demo-projects/app2"
Enter
Type "echo '{\"name\":\"app2\",\"scripts\":{\"lint\":\"echo 'Linting app2...'\"}}' > package.json"
Enter

# Set up api with package.json
Type "cd /tmp/demo-projects/api"
Enter
Type "echo '{\"name\":\"api\",\"scripts\":{\"lint\":\"echo 'Linting api...'\"}}' > package.json"
Enter

# Go back to our main directory
Type "cd /Users/kooshapari/Cwatch"
Enter
Sleep 1s

Type "# Now let's launch Master KWatch to monitor all three projects"
Enter
Sleep 1s

# Launch master interface with multiple directories
Type "./build/kwatch master /tmp/demo-projects/app1 /tmp/demo-projects/app2 /tmp/demo-projects/api"
Enter

# Wait for master interface to load and show the matrix
Sleep 5s

# Force a refresh to show monitoring in action
Key "r"
Sleep 3s

# Show help to demonstrate controls
Key "h"
Sleep 2s

# Go back to matrix view
Key "Escape"
Sleep 1s

# Another refresh cycle
Key "r"
Sleep 3s

# Exit the master interface
Key "q"
Sleep 1s

Type "# Master KWatch provides:"
Enter
Type "# • Unified matrix view of multiple projects"
Enter
Type "# • Real-time status updates across all directories"
Enter
Type "# • Centralized monitoring dashboard"
Enter
Type "# • Project health at a glance"
Enter
Sleep 2s

# Clean up demo directories
Type "rm -rf /tmp/demo-projects"
Enter