# kwatch

A fast, lightweight project monitoring tool for TypeScript/JavaScript projects. Provides real-time build status through a TUI panel and HTTP API optimized for AI agent polling.

## Features

- **Real-time TUI Panel** - Monitor TSC, Lint, and Test status with live updates
- **File Watcher** - Automatically runs checks when files change
- **HTTP API** - Fast polling endpoints for AI agents (<100ms response)
- **Command History** - Track all runs with timestamps and results
- **Global Installation** - Install like any system command (`cp`, `ls`, etc.)
- **Cross-platform** - Works on Linux, macOS, and Windows

## Installation

### Quick Install (Global)
```bash
make install
```

### User Install (No sudo required)
```bash
make install-user
# Add ~/bin to your PATH if not already there
export PATH="$HOME/bin:$PATH"
```

### From Source
```bash
git clone <repository>
cd kwatch
make build
./build/kwatch .
```

## Usage

### Basic Commands

```bash
# Start TUI panel in current directory
kwatch .

# Start TUI panel in specific directory
kwatch /path/to/project

# Get current status (JSON)
kwatch . status

# Get compact status (one-line)
kwatch . status --compact

# View command history
kwatch . history

# Force manual run
kwatch . run

# Start background daemon
kwatch . daemon --port 3737
```

### TUI Panel Controls

- **q** - Quit
- **r** - Refresh/Manual run
- **h** - Show help
- **s** - Switch to status view
- **l** - Switch to logs view
- **↑/↓** - Navigate (in history/logs)

### HTTP API for AI Agents

When running in daemon mode, the following endpoints are available:

- `GET /quick` - Ultra-fast health check (just "ok")
- `GET /status` - Full JSON status
- `GET /status/compact` - Single-line status: `TSC:✓0 LINT:✗5 TEST:✓0`
- `POST /run` - Trigger manual run
- `GET /history` - Command execution history
- `GET /metrics` - Performance metrics

### AI Agent Integration

**Quick Status Check:**
```bash
curl -s http://localhost:3737/status/compact
# Output: TSC:✓0 LINT:✗5 TEST:✓0
```

**Before Running Commands:**
```bash
# Check if project is healthy before running
status=$(curl -s http://localhost:3737/quick)
if [ "$status" = "ok" ]; then
    echo "Project is healthy, proceeding..."
else
    echo "Project has issues, checking details..."
    curl -s http://localhost:3737/status | jq .
fi
```

## Commands Monitored

- **TSC** - `npx tsc --noEmit 2>&1`
- **Lint** - `bun run lint`
- **Test** - `bun run test`

## Output Format

### JSON Status
```json
{
  "directory": "/path/to/project",
  "timestamp": "2024-01-01T12:00:00Z",
  "commands": {
    "tsc": {
      "passed": true,
      "issue_count": 0,
      "duration": "1.2s",
      "timestamp": "2024-01-01T12:00:00Z"
    },
    "lint": {
      "passed": false,
      "issue_count": 5,
      "duration": "0.8s",
      "timestamp": "2024-01-01T12:00:00Z"
    },
    "test": {
      "passed": true,
      "issue_count": 0,
      "duration": "2.1s",
      "timestamp": "2024-01-01T12:00:00Z"
    }
  }
}
```

### Compact Status
```
TSC:✓0 LINT:✗5 TEST:✓0
```

- **✓** = Passed
- **✗** = Failed
- **Number** = Issue count (errors/warnings/failures)

## Development

### Build Commands
```bash
make build        # Build binary
make dev          # Build with race detection
make test         # Run tests
make clean        # Clean build artifacts
make build-all    # Cross-compile for all platforms
```

### Requirements
- Go 1.21+
- Node.js/Bun for the monitored commands
- Make for build automation

## Architecture

- **CLI** - Cobra-based command structure
- **TUI** - Bubbletea terminal interface
- **Runner** - Parallel command execution
- **Server** - HTTP API for agent polling
- **Watcher** - File system monitoring

## Performance

- **TUI Updates** - Real-time, sub-second response
- **HTTP API** - <100ms response times
- **File Watching** - Debounced, efficient monitoring
- **Command Execution** - Parallel processing

## License

MIT License
