# VHS tape for Master KWatch interface demo
#
# This demo shows multi-directory monitoring with matrix display

# Set up the terminal
Output demos/kwatch-master-interface.gif
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Molokai"

# Title
Type "# KWatch - Master Interface"
Enter
Sleep 1s

Type "# Master KWatch provides multi-project monitoring"
Enter
Sleep 1s

Type "# Here's what the matrix format looks like:"
Enter
Sleep 1s

Type "echo 'Master KWatch Status - 2025-07-10 14:14:00'"
Enter
Type "echo 'Directories: 3 | Commands: 4 | Passed: 8 | Failed: 4'"
Enter
Type "echo ''"
Enter
Type "echo 'DIRECTORY           TSC         LINT        TEST        GITHUB      '"
Enter
Type "echo '--------------------------------------------------------------------'"
Enter
Type "echo 'app1                ✓           ✓           ✗(5)        ✓           '"
Enter
Type "echo 'app2                ✗(12)       ✓           ✓           ✗           '"
Enter
Type "echo 'api                 ✓           ✗(3)        ✓           ✓           '"
Enter
Sleep 3s

Type "# Legend: ✓ = Passed, ✗ = Failed, numbers = issue count"
Enter
Sleep 1s

Type "# This matrix shows the health of your entire development ecosystem"
Enter
Sleep 1s