# VHS tape for KWatch file watching demo
#
# This demo shows real-time file watching and automatic status updates

# Set up the terminal
Output demos/kwatch-file-watching.gif
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Molokai"

# Start clean
Hide
Type "clear"
Enter
Show

# Start kwatch daemon for file watching demo
Type "./build/kwatch daemon --port 3737 &"
Enter
Sleep 3s

# Show initial status
Type "curl -s http://localhost:3737/status/compact"
Enter
Sleep 2s

# Create a test file to modify
Type "echo 'let x = 1;' > test-file.js"
Enter
Sleep 2s

# Check status after file change
Type "curl -s http://localhost:3737/status/compact"
Enter
Sleep 2s

# Modify file to introduce error
Type "echo 'let x = ;' > test-file.js"
Enter
Sleep 2s

# Check status shows error detection
Type "curl -s http://localhost:3737/status/compact"
Enter
Sleep 2s

# Fix the file
Type "echo 'let x = 1;' > test-file.js"
Enter
Sleep 2s

# Final status check shows fix
Type "curl -s http://localhost:3737/status/compact"
Enter
Sleep 2s

# Show full status JSON
Type "curl -s http://localhost:3737/status"
Enter
Sleep 3s

# Clean up
Type "rm test-file.js"
Enter
Type "pkill -f 'kwatch daemon'"
Enter
Sleep 1s