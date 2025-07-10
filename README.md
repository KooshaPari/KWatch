# KWatch

A comprehensive project monitoring tool for TypeScript/JavaScript projects with **GitHub Actions integration** and **multi-directory monitoring**. Provides real-time build status through an advanced TUI interface, HTTP API, and secure authentication system.

## 🚀 Features

### Core Monitoring
- **Real-time TUI Panel** - Monitor TSC, Lint, Test, and **GitHub Actions** with live updates
- **GitHub Actions Integration** - Monitor CI/CD workflows directly in your terminal
- **Master KWatch Interface** - Monitor multiple directories from a single unified view
- **Secure Token Management** - AES-256-GCM encrypted GitHub token storage
- **File Watcher** - Automatically runs checks when files change
- **HTTP API** - Fast polling endpoints for AI agents (<100ms response)
- **Command History** - Track all runs with timestamps and results

### Advanced Features
- **Matrix Display** - Clean tabular view of all projects and their status
- **Multi-Directory Monitoring** - Watch multiple projects simultaneously
- **Cross-platform** - Works on Linux, macOS, and Windows
- **Global Installation** - Install like any system command

## 🛠️ Installation

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
git clone https://github.com/KooshaPari/KWatch.git
cd KWatch
make build
./build/kwatch .
```

### macOS Note
If the binary auto-closes after installation, remove quarantine attributes:
```bash
xattr -c ~/bin/kwatch
```

## 🔐 GitHub Authentication Setup

KWatch includes a secure authentication system for GitHub Actions monitoring:

### Quick Setup
```bash
# Interactive secure token setup (recommended)
kwatch auth --init

# Check authentication status
kwatch auth --status

# Test GitHub Actions monitoring
kwatch run --command github
```

### Token Requirements
- Get a token at: https://github.com/settings/tokens
- Required scopes: `repo` + `actions:read`
- Choose between secure storage or environment variables

### Security Features
- **AES-256-GCM encryption** for token storage
- **System-specific key derivation** (tokens can't be copied between machines)
- **Secure file permissions** (600 - user-only access)
- **Environment variable fallback** (`GITHUB_TOKEN` or `GH_TOKEN`)

## 📋 Usage

### Basic Commands

```bash
# Start TUI panel in current directory
kwatch .

# Start TUI panel in specific directory
kwatch /path/to/project

# Master view - monitor multiple directories
kwatch master /path/to/project1 /path/to/project2

# Get current status (JSON)
kwatch status

# Force manual run of all commands
kwatch run

# Start background daemon
kwatch daemon --port 3737
```

### GitHub Actions Commands

```bash
# Check GitHub Actions status
kwatch run --command github

# View authentication status
kwatch auth --status

# Setup secure token storage
kwatch auth --init

# Clear stored token
kwatch auth --clear
```

### Master KWatch Interface

Monitor multiple projects with a unified matrix display:

```bash
# Monitor multiple directories
kwatch master ~/projects/app1 ~/projects/app2 ~/projects/api

# Auto-discover projects with .kwatch configurations
kwatch master ~/projects
```

**Output:**
```
Master KWatch Status - 2025-07-10 14:14:00
Directories: 3 | Commands: 4 | Passed: 8 | Failed: 4

DIRECTORY           TSC         LINT        TEST        GITHUB      
--------------------------------------------------------------------
app1                ✓           ✓           ✗(5)        ✓           
app2                ✗(12)       ✓           ✓           ✗           
api                 ✓           ✗(3)        ✓           ✓           

Legend: ✓ = Passed, ✗ = Failed, ERR = Error, (-) = Not applicable
Numbers in parentheses show issue count
```

### TUI Panel Controls

- **q** - Quit
- **r** - Refresh/Manual run
- **h** - Show help
- **s** - Switch to status view
- **l** - Switch to logs view
- **↑/↓** - Navigate (in history/logs)

## 🌐 HTTP API for AI Agents

When running in daemon mode, the following endpoints are available:

- `GET /quick` - Ultra-fast health check (just "ok")
- `GET /status` - Full JSON status including GitHub Actions
- `GET /status/compact` - Single-line status: `TSC:✓0 LINT:✗5 TEST:✓0 GITHUB:✓`
- `POST /run` - Trigger manual run
- `GET /history` - Command execution history
- `GET /metrics` - Performance metrics

### AI Agent Integration

**Quick Status Check:**
```bash
curl -s http://localhost:3737/status/compact
# Output: TSC:✓0 LINT:✗5 TEST:✓0 GITHUB:✓
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

## 📊 Commands Monitored

### Standard Commands
- **TSC** - `npx tsc --noEmit 2>&1`
- **Lint** - `bun run lint` or `npm run lint`
- **Test** - `bun run test` or `npm run test`

### GitHub Actions
- **GitHub** - Latest workflow runs and job status
- **Real-time CI/CD monitoring** - Shows pass/fail status
- **Job-level details** - Individual job results and timing

## 📄 Output Formats

### JSON Status (with GitHub Actions)
```json
{
  "directory": "/path/to/project",
  "timestamp": "2025-07-10T14:14:00Z",
  "commands": {
    "tsc": {
      "passed": true,
      "issue_count": 0,
      "duration": "1.2s"
    },
    "lint": {
      "passed": false,
      "issue_count": 5,
      "duration": "0.8s"
    },
    "test": {
      "passed": true,
      "issue_count": 0,
      "duration": "2.1s"
    },
    "github": {
      "passed": true,
      "issue_count": 0,
      "duration": "0.5s",
      "workflow_name": "CI",
      "workflow_status": "completed",
      "job_results": [
        {
          "name": "build",
          "status": "completed",
          "conclusion": "success"
        }
      ]
    }
  }
}
```

### Compact Status
```
TSC:✓0 LINT:✗5 TEST:✓0 GITHUB:✓
```

- **✓** = Passed
- **✗** = Failed  
- **ERR** = Error
- **Number** = Issue count (errors/warnings/failures)

## 🛠️ Development

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
- Git repository for GitHub Actions monitoring

## 🏗️ Architecture

### Core Components
- **CLI** - Cobra-based command structure with secure authentication
- **TUI** - Enhanced Bubbletea interface with GitHub Actions integration
- **Runner** - Parallel command execution including GitHub API calls
- **Server** - HTTP API with comprehensive status endpoints
- **Watcher** - File system monitoring with intelligent debouncing
- **Auth** - Secure token management with AES-256-GCM encryption

### Security Architecture
- **Encrypted Token Storage** - AES-256-GCM with system-specific keys
- **Secure File Permissions** - 600 permissions on token files
- **Environment Variable Fallback** - GITHUB_TOKEN/GH_TOKEN support
- **Safe Error Handling** - Panic recovery and graceful degradation

## ⚡ Performance

- **TUI Updates** - Real-time, sub-second response
- **HTTP API** - <100ms response times including GitHub Actions
- **File Watching** - Debounced, efficient monitoring
- **Command Execution** - Parallel processing with timeout handling
- **GitHub API** - Cached responses with intelligent rate limiting

## 🚨 Troubleshooting

### macOS Binary Issues
If `kwatch` works locally but fails when installed:
```bash
# Remove quarantine attributes
xattr -c ~/bin/kwatch
```

### GitHub Actions Not Working
```bash
# Check authentication status
kwatch auth --status

# Setup secure token
kwatch auth --init

# Or use environment variable
export GITHUB_TOKEN=your_token_here
```

### Token Management
```bash
# Clear corrupted token
kwatch auth --clear

# Reinitialize
kwatch auth --init
```

## 📚 Documentation

- [Secure Authentication Guide](SECURE_AUTH.md) - Complete GitHub token setup guide
- [API Reference](docs/api.md) - HTTP API documentation
- [Configuration](docs/config.md) - Project configuration options

## 📄 License

MIT License

---

**KWatch** - Comprehensive project monitoring with GitHub Actions integration and secure authentication. Built for developers who need real-time visibility into their entire development pipeline.