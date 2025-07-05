# Process Service

A development process monitoring service with TUI (Text User Interface) and CLI interface that automatically watches your codebase for changes and runs TypeScript checking, linting, testing, and building.

## Features

- 🖥️ **TUI Interface**: Beautiful terminal interface showing real-time status
- 🔌 **CLI Interface**: Command-line tools for status checking and control
- 👀 **File Watching**: Automatically detects code changes and reruns checks
- 📊 **History Tracking**: Keeps track of all runs with timestamps and results
- 🚀 **Parallel Execution**: Runs all checks simultaneously for speed
- 🌐 **HTTP API**: RESTful API for programmatic access
- 💾 **Persistent History**: Saves run history to disk

## Installation

```bash
# Install dependencies
npm install

# Build the project
npm run build

# Or install globally
npm run install:global
```

## Usage

### Starting the Service

```bash
# Start with TypeScript (development)
npm run proc:start

# Or start with built JavaScript
npm start
```

The service will:
- Open a TUI showing current status and history
- Start watching files for changes
- Run initial checks
- Start HTTP server on port 3737

### TUI Interface

The TUI shows:
- **Status Bar**: Current status and server info
- **History Table**: Last runs with timestamps and pass/fail status
- **Log Panel**: Real-time logs of file changes and runs

**Controls:**
- `q`, `Escape`, or `Ctrl+C`: Quit the service

### CLI Interface

```bash
# Show current status
npm run cli status

# Show run history
npm run cli history

# Trigger manual run
npm run cli run

# Watch status in real-time
npm run cli watch

# Show help
npm run cli help
```

### API Endpoints

The service exposes a REST API on port 3737:

```bash
# Get current status
curl http://localhost:3737/status

# Get full history
curl http://localhost:3737/history

# Trigger manual run
curl http://localhost:3737/run
```

## Commands Monitored

The service runs these commands in parallel:

1. **TypeScript Check**: `npx tsc --noEmit`
2. **Linting**: `npm run lint` or `bun run lint`
3. **Testing**: `npm run test` or `bun run test`
4. **Building**: `npm run build` or `bun run build`

## File Watching

Monitors these paths for changes:
- `src/**/*`
- `package.json`
- `tsconfig.json`
- `*.config.js`
- `*.config.ts`
- `tests/**/*`
- `__tests__/**/*`

Ignores:
- `node_modules/`
- `dist/`
- `build/`
- `.git/`
- `**/*.log`

## Configuration

### Timing
- **File Change Debounce**: 2 seconds (prevents excessive runs)
- **Periodic Check**: Every 60 seconds
- **Command Timeout**: 5 minutes per command

### Customization

You can modify the watched files, commands, and timing by editing the `ProcService` class in `proc-service.ts`.

## History

Run history is automatically saved to `.proc-history.json` and includes:
- Timestamp of each run
- Pass/fail status for each command
- Error/warning counts
- Execution duration
- Command output

## Output Format

Each run shows:
- ✅ Pass or ❌ Fail for each command
- Number of errors/warnings in parentheses
- Execution time in milliseconds

## Example Output

```
📊 Process Service Status
══════════════════════════════════════════════════
Running: ⏸️  No

🕒 Last Run: 1/4/2025, 2:30:45 PM (2847ms)
──────────────────────────────────────────────────
✅ TSC    234ms
✅ Lint   187ms (2 issues)
❌ Test   1205ms (3 issues)
✅ Build  891ms

📈 Overall: 3/4 passed
```

## LLM Integration

The CLI and API are designed to be easily consumed by LLMs and other automated tools:

```bash
# Get machine-readable status
curl -s http://localhost:3737/status | jq '.lastRun'

# Check if all tests are passing
proc-cli status | grep "4/4 passed"
```

## Troubleshooting

### Service Won't Start
- Ensure Node.js 16+ is installed
- Check that port 3737 is available
- Verify all dependencies are installed

### Commands Failing
- Ensure your package.json has the correct scripts
- Check that you have the necessary dependencies installed
- Look at the detailed output in the TUI log panel

### File Watching Not Working
- Verify the watched paths exist
- Check file permissions
- Ensure you're not hitting system file watching limits

## Development

```bash
# Start in development mode
npm run dev

# Build for production
npm run build

# Lint the code
npm run lint

# Run tests
npm run test
```

## License

MIT
