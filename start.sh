#!/bin/bash

# Function to count errors/warnings in output
count_issues() {
    local output="$1"
    local count=0
    
    # Count lines containing "error" or "warning" (case insensitive)
    count=$(echo "$output" | grep -i -E "(error|warning)" | wc -l)
    echo $count
}

# Function to run command in background and save results to temp files
run_command_bg() {
    local cmd_num="$1"
    local cmd="$2"
    local temp_dir="$3"
    
    (
        # Capture both stdout and stderr, and the exit code
        output=$(eval "$cmd" 2>&1)
        exit_code=$?
        
        # Save results to temp files
        echo "$exit_code" > "$temp_dir/exit_$cmd_num"
        echo "$output" > "$temp_dir/output_$cmd_num"
        echo "$cmd" > "$temp_dir/cmd_$cmd_num"
    ) &
    
    echo $!  # Return the background process PID
}

# Function to process results after all commands complete
process_results() {
    local temp_dir="$1"
    
    for i in 1 2 3; do
        if [ -f "$temp_dir/exit_$i" ]; then
            exit_code=$(cat "$temp_dir/exit_$i")
            output=$(cat "$temp_dir/output_$i")
            cmd=$(cat "$temp_dir/cmd_$i")
            
            # Determine if command passed
            if [ $exit_code -eq 0 ]; then
                passed="true"
            else
                passed="false"
            fi
            
            # Count errors/warnings
            issue_count=$(count_issues "$output")
            
            # Output in requested format
            echo "$i: $cmd: $passed $issue_count"
            
            # Show actual output for debugging
            if [ ! -z "$output" ]; then
                echo "Output:"
                echo "$output"
                echo "---"
            fi
        fi
    done
}

echo "=== Running Commands in Parallel ==="

# Create temp directory for results
temp_dir=$(mktemp -d)
trap "rm -rf $temp_dir" EXIT

# Define commands
commands=(
    "npx tsc --noEmit 2>&1"
    "bun run lint" 
    "bun run test"
)

# Start all commands in parallel
pids=()
for i in "${!commands[@]}"; do
    cmd_num=$((i + 1))
    echo "Starting: ${commands[$i]}"
    pid=$(run_command_bg "$cmd_num" "${commands[$i]}" "$temp_dir")
    pids+=($pid)
done

echo "Waiting for all commands to complete..."

# Wait for all background processes to complete
for pid in "${pids[@]}"; do
    wait $pid
done

echo "=== Results ==="

# Process and display results
process_results "$temp_dir"

echo "=== Done ==="
