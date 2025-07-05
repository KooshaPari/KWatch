#!/bin/bash

# ESLint check script
# Returns: status and error/warning count

echo "Running ESLint..."
output=$(bun run lint 2>&1)
exit_code=$?

if [ $exit_code -eq 0 ]; then
    echo "LINT: true 0"
else
    # Count errors and warnings
    error_count=$(echo "$output" | grep -c "error\|warning")
    echo "LINT: false $error_count"
fi

exit $exit_code