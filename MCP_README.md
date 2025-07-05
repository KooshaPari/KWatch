# KWatch MCP Integration

KWatch now supports the **Model Context Protocol (MCP)**, allowing AI assistants like Claude to monitor your project's build status, run commands, and access command history.

## 🚀 Quick Start

1. **Start MCP Server**:
   ```bash
   kwatch mcp /path/to/your/project
   ```

2. **Configure Claude Desktop** (add to `~/.config/claude-desktop/config.json`):
   ```json
   {
     "mcpServers": {
       "kwatch": {
         "command": "kwatch",
         "args": ["mcp", "/path/to/your/project"]
       }
     }
   }
   ```

3. **Restart Claude Desktop** and start using kwatch tools in your conversations!

## 🛠️ Available MCP Tools

### 1. `get_build_status`
Get current build status including TypeScript, linting, and test results.

**Parameters**:
- `format`: `"compact"` (one-line) or `"detailed"` (full JSON)

**Example**:
```
Get the current build status in compact format
```

### 2. `run_commands`
Execute build commands manually.

**Parameters**:
- `command`: `"all"`, `"tsc"`, `"lint"`, or `"test"`

**Example**:
```
Run the linter on this project
```

### 3. `get_command_history`
Get history of previously executed commands.

**Parameters**:
- `limit`: Maximum number of entries (default: 10)
- `filter`: Filter by command type (`"tsc"`, `"lint"`, or `"test"`)

**Example**:
```
Show the last 5 linting command results
```

## ⚙️ Configuration Examples

### Claude Desktop
```json
{
  "mcpServers": {
    "kwatch-dev": {
      "command": "/usr/local/bin/kwatch",
      "args": ["mcp", "/Users/dev/my-project"],
      "env": {}
    }
  }
}
```

### Multiple Projects
```json
{
  "mcpServers": {
    "kwatch-frontend": {
      "command": "kwatch",
      "args": ["--dir", "/var/www/frontend", "mcp"]
    },
    "kwatch-backend": {
      "command": "kwatch", 
      "args": ["--dir", "/var/www/backend", "mcp"]
    }
  }
}
```

## 🧪 Testing

Test the MCP server manually:

```bash
# Run the test script
./test-mcp.sh

# Or test individual requests
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | kwatch mcp .
```

## 📋 Command Reference

### Start MCP Server
```bash
# Current directory
kwatch mcp

# Specific directory
kwatch mcp /path/to/project

# Using global flag
kwatch --dir /path/to/project mcp
```

### Example MCP Usage
Once configured, you can ask Claude:

- *"What's the current build status?"*
- *"Run the TypeScript checker"*
- *"Show me the last 5 lint results"* 
- *"Are there any failing tests?"*
- *"Run all build commands and show me a summary"*

## 🔧 Technical Details

- **Protocol**: JSON-RPC 2.0 over stdio transport
- **Compatibility**: MCP protocol versions 2024-11-05 and 2025-03-26
- **Transport**: stdio (most compatible with MCP clients)
- **Output**: Structured JSON responses with build status, error counts, and timestamps

## 🐛 Troubleshooting

### Server Not Starting
- Ensure `kwatch` binary is in PATH or use absolute path in config
- Check directory exists and is accessible
- Verify JSON syntax in MCP client config

### No Response from Tools
- Check stderr output: `kwatch mcp . 2>&1`
- Verify protocol version compatibility
- Test with manual JSON-RPC messages

### Build Commands Failing
- Ensure project has proper package.json and dependencies
- Check if TypeScript, ESLint, and test commands work manually
- Verify working directory is correct

## 📚 Resources

- [Model Context Protocol Specification](https://modelcontextprotocol.io)
- [Claude Desktop MCP Guide](https://claude.ai/docs/mcp)
- [KWatch Documentation](./README.md)

## 🎯 Use Cases

- **Development Monitoring**: Ask Claude to check build status while coding
- **CI/CD Integration**: Query build health from AI assistants
- **Code Review**: Get quick status updates during reviews
- **Team Collaboration**: Share build status through AI conversations
- **Debugging**: Query specific command history and error details