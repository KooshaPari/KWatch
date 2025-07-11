# VHS tape for KWatch file watching demo
#
# This demo shows real-time file watching and automatic status updates

# Set up the terminal
Output demos/kwatch-file-watching.gif
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
Type "# KWatch - Real-time File Watching"
Enter
Sleep 1s

Type "# KWatch automatically detects file changes and runs checks"
Enter
Sleep 1s

# Start kwatch in background mode for this demo
Type "# Let's start KWatch with file watching enabled"
Enter
Type "./build/kwatch daemon --port 3737 &"
Enter
Sleep 2s

# Show initial status
Type "# Check initial status:"
Enter
Type "curl -s http://localhost:3737/status/compact"
Enter
Sleep 2s

Type "# Now let's watch what happens when we modify a file..."
Enter
Sleep 1s

# Create a test file to modify
Type "echo 'let x = 1;' > test-file.js"
Enter
Sleep 1s

# Check status after file change
Type "# Status after file change:"
Enter
Type "curl -s http://localhost:3737/status/compact"
Enter
Sleep 2s

# Modify the file to introduce an error
Type "# Let's introduce a syntax error:"
Enter
Type "echo 'let x = ;' > test-file.js"
Enter
Sleep 1s

# Check status again
Type "curl -s http://localhost:3737/status/compact"
Enter
Sleep 2s

# Fix the file
Type "# Fix the syntax error:"
Enter
Type "echo 'let x = 1;' > test-file.js"
Enter
Sleep 1s

# Final status check
Type "curl -s http://localhost:3737/status/compact"
Enter
Sleep 2s

Type "# File watching provides:"
Enter
Type "# • Automatic triggering of checks on file changes"
Enter
Type "# • Real-time feedback during development"
Enter
Type "# • Continuous integration without manual intervention"
Enter
Sleep 3s

# Clean up
Type "rm test-file.js"
Enter
Type "pkill -f 'kwatch daemon'"
Enter
Sleep 1s