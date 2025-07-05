#!/bin/bash

# Combined script to run all checks
# Returns: N: cmd: bool (pass?) int (# of err/warnings)

echo "Running all checks..."
echo "===================="

# Run TypeScript check
echo "1: tsc: $(./run-tsc.sh | tail -1 | cut -d' ' -f2-)"

# Run lint check  
echo "2: lint: $(./run-lint.sh | tail -1 | cut -d' ' -f2-)"

# Run test check
echo "3: test: $(./run-test.sh | tail -1 | cut -d' ' -f2-)"

echo "===================="