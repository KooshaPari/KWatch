# VHS tape for GitHub Actions integration demo
#
# This demo shows actual GitHub Actions monitoring and secure authentication

# Set up the terminal
Output demos/kwatch-github-actions.gif
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
Type "# KWatch - GitHub Actions Integration"
Enter
Sleep 1s

Type "# KWatch includes secure GitHub token management with AES-256-GCM encryption"
Enter
Sleep 1s

# Check current auth status
Type "./build/kwatch auth --status"
Enter
Sleep 2s

Type "# Let's test GitHub Actions monitoring"
Enter
Sleep 1s

# Run GitHub Actions check
Type "./build/kwatch run --command github"
Enter
Sleep 4s

Type "# Now let's see the full status including GitHub Actions"
Enter
Sleep 1s

# Get JSON status with GitHub Actions data
Type "./build/kwatch status"
Enter
Sleep 3s

Type "# The compact format shows all statuses at a glance:"
Enter
Sleep 1s

# Start daemon for compact status
Type "./build/kwatch daemon --port 3737 &"
Enter
Sleep 2s

# Show compact status endpoint
Type "curl -s http://localhost:3737/status/compact"
Enter
Sleep 2s

Type "# GitHub integration provides:"
Enter
Type "# • Workflow status monitoring"
Enter
Type "# • Job-level success/failure tracking"
Enter
Type "# • Real-time CI/CD pipeline visibility"
Enter
Type "# • Secure token storage with system-specific encryption"
Enter
Sleep 3s

# Clean up daemon
Type "pkill -f 'kwatch daemon'"
Enter
Sleep 1s