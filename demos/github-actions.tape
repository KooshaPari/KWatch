# VHS tape for GitHub Actions integration demo
#
# This demo shows actual GitHub Actions monitoring and secure authentication

# Set up the terminal
Output demos/kwatch-github-actions.gif
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Molokai"

# Start clean
Hide
Type "clear"
Enter
Show

# Check current auth status - actual command execution
Type "./build/kwatch auth --status"
Enter
Sleep 3s

# Run GitHub Actions check - show real execution
Type "./build/kwatch run --command github"
Enter
Sleep 5s

# Get full JSON status including GitHub Actions data
Type "./build/kwatch status"
Enter
Sleep 4s

# Start daemon and show API endpoints
Type "./build/kwatch daemon --port 3737 &"
Enter
Sleep 3s

# Show compact status endpoint
Type "curl -s http://localhost:3737/status/compact"
Enter
Sleep 3s

# Show full status endpoint
Type "curl -s http://localhost:3737/status"
Enter
Sleep 3s

# Clean up daemon
Type "pkill -f 'kwatch daemon'"
Enter
Sleep 1s