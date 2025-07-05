#!/bin/bash

# Test script for KWatch MCP server
# This script sends JSON-RPC 2.0 messages to test the MCP server functionality

echo "🧪 Testing KWatch MCP Server"
echo "=========================="

# Build the kwatch binary first
echo "📦 Building kwatch..."
go build -o kwatch . || {
    echo "❌ Failed to build kwatch"
    exit 1
}

echo "✅ Build successful"
echo ""

# Function to send JSON-RPC message and get response
test_mcp() {
    local test_name="$1"
    local json_message="$2"
    
    echo "🔧 Testing: $test_name"
    echo "📤 Request: $json_message"
    
    # Send message to MCP server and capture response
    response=$(echo "$json_message" | ./kwatch mcp . 2>/dev/null | head -1)
    
    echo "📥 Response: $response"
    echo ""
    
    # Basic validation - check if response contains expected JSON-RPC structure
    if echo "$response" | jq . >/dev/null 2>&1; then
        echo "✅ Valid JSON response"
    else
        echo "❌ Invalid JSON response"
    fi
    echo "---"
}

# Test 1: Initialize
echo "🚀 Starting MCP tests..."
echo ""

initialize_request='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}'

test_mcp "Initialize" "$initialize_request"

# Test 2: List tools
tools_list_request='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'

test_mcp "List Tools" "$tools_list_request"

# Test 3: Get build status (compact)
build_status_compact='{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_build_status","arguments":{"format":"compact"}}}'

test_mcp "Get Build Status (Compact)" "$build_status_compact"

# Test 4: Get build status (detailed)
build_status_detailed='{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_build_status","arguments":{"format":"detailed"}}}'

test_mcp "Get Build Status (Detailed)" "$build_status_detailed"

# Test 5: Run all commands
run_all_commands='{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"run_commands","arguments":{"command":"all"}}}'

test_mcp "Run All Commands" "$run_all_commands"

# Test 6: Get command history
get_history='{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"get_command_history","arguments":{"limit":5}}}'

test_mcp "Get Command History" "$get_history"

echo "🎉 MCP tests completed!"
echo ""
echo "💡 Usage with MCP clients:"
echo "1. Add kwatch to your MCP client config:"
echo "   {\"kwatch\": {\"command\": \"$(pwd)/kwatch\", \"args\": [\"mcp\", \"$(pwd)\"]}}"
echo ""
echo "2. For Claude Desktop, add to ~/.config/claude-desktop/config.json"
echo "3. Restart your MCP client to load the kwatch server"