#!/bin/bash

# Test runner script
# Returns: status and failure count

echo "Running tests..."
output=$(bun run test 2>&1)
exit_code=$?

if [ $exit_code -eq 0 ]; then
    echo "TEST: true 0"
else
    # Count failed tests
    fail_count=$(echo "$output" | grep -c "FAIL\|failed\|Error")
    echo "TEST: false $fail_count"
fi

exit $exit_code