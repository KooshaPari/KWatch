# VHS tape for GitHub Actions integration demo
#
# This demo shows GitHub Actions monitoring and secure token management

# Set up the terminal
Output demos/kwatch-github-actions.gif
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Molokai"

# Title  
Type "# KWatch - GitHub Actions Integration"
Enter
Sleep 1s

Type "# KWatch includes secure GitHub token management"
Enter
Sleep 1s

Type "# Let's check what tokens/auth we have:"
Enter
Type "echo $GITHUB_TOKEN | cut -c1-10"
Enter
Sleep 1s

Type "# Check if we have a GitHub repository"
Enter
Type "git remote -v"
Enter
Sleep 2s

Type "# KWatch securely stores GitHub tokens with AES-256-GCM encryption"
Enter
Sleep 1s

Type "# Token setup would be done with:"
Enter
Type "# kwatch auth --init"
Enter
Sleep 1s

Type "# This provides CI/CD monitoring directly in the terminal"
Enter
Type "# showing workflow status, job results, and timing"
Enter
Sleep 2s

Type "# Matrix format shows: ✓ Pass, ✗ Fail, numbers = issues"
Enter
Sleep 1s