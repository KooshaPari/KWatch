# VHS tape for basic KWatch usage demo
#
# This demo shows basic KWatch functionality including build commands

# Set up the terminal
Output demos/kwatch-basic-usage.gif
Set FontSize 14
Set Width 1200  
Set Height 800
Set Theme "Molokai"

# Title
Type "# KWatch - Project Monitoring Demo"
Enter
Sleep 1s

Type "# KWatch monitors TypeScript, Lint, Test, and GitHub Actions"
Enter
Sleep 1s

Type "# Let's test individual commands directly:"
Enter
Sleep 1s

Type "# TypeScript compilation check"
Enter
Type "npx tsc --noEmit 2>&1"
Enter
Sleep 2s

Type "# ESLint code quality check"  
Enter
Type "npm run lint"
Enter
Sleep 3s

Type "# Test runner (note: this will fail due to missing jest config)"
Enter
Type "npm run test"
Enter
Sleep 2s

Type "# GitHub Actions integration would show CI/CD status here"
Enter
Sleep 1s

Type "# Build the project"
Enter
Type "make build"
Enter
Sleep 3s

Type "# KWatch aggregates all these into a unified monitoring dashboard"
Enter
Sleep 2s