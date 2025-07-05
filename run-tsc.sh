#!/bin/bash

# TypeScript compilation check script
# Returns: status and error count

echo "Running TypeScript compiler..."
output=$(npx tsc --noEmit 2>&1)
exit_code=$?

if [ $exit_code -eq 0 ]; then
    echo "TSC: true 0"
else
    error_count=$(echo "$output" | grep -c "error TS")
    echo "TSC: false $error_count"
fi

exit $exit_code