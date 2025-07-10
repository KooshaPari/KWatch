# Changelog

All notable changes to KWatch will be documented in this file.

## [2.0.0] - 2025-07-10

### 🚀 Major Features Added

#### GitHub Actions Integration
- **Real-time CI/CD monitoring** - Monitor GitHub Actions workflows directly in terminal
- **Workflow status tracking** - Shows latest runs, job results, and execution status
- **API integration** - Full GitHub Actions API client with authentication
- **TUI integration** - GitHub Actions column in matrix display and status views

#### Master KWatch Interface  
- **Multi-directory monitoring** - Monitor multiple projects from single unified interface
- **Matrix display format** - Clean tabular view: `DIR CMD1 CMD2 CMD3 GITHUB`
- **Auto-discovery** - Automatically find projects with .kwatch configurations
- **Unified status** - Single view showing all projects and command statuses
- **Real-time updates** - Live status updates across all monitored directories

#### Secure Authentication System
- **AES-256-GCM encryption** - Military-grade encryption for GitHub token storage
- **System-specific keys** - Encryption keys derived from system identifiers (can't be copied between machines)
- **Interactive setup** - `kwatch auth --init` for secure token configuration
- **Token management** - Full lifecycle management with `--status`, `--clear`, `--init` commands
- **Environment fallback** - Supports GITHUB_TOKEN/GH_TOKEN environment variables
- **Secure file permissions** - 600 permissions on encrypted token files

#### Enhanced TUI Interface
- **GitHub Actions column** - Real-time workflow status in matrix view
- **Live command tracking** - Shows "Command started", "Command completed", "Command FAILED"
- **Manual refresh support** - Manual refresh triggered display
- **Enhanced status display** - Comprehensive status with duration and issue counts

### 🛠️ Technical Improvements

#### API Enhancements
- **GitHub API client** - Full featured client with rate limiting and error handling
- **Extended JSON output** - GitHub Actions data in status endpoints
- **Enhanced compact format** - `TSC:✓0 LINT:✗5 TEST:✓0 GITHUB:✓`

#### Security & Reliability
- **Panic recovery** - Safe error handling for GitHub API failures
- **macOS quarantine fix** - Automatic handling of quarantine attributes
- **Graceful degradation** - Works without GitHub token for local-only monitoring

#### Performance Optimizations
- **Parallel execution** - GitHub Actions monitoring runs alongside other commands
- **Intelligent caching** - Efficient GitHub API response caching
- **Optimized file watching** - Enhanced debouncing and monitoring

### 🔧 Configuration Updates

#### Command Structure
- Added `kwatch auth` command group with `--init`, `--status`, `--clear`, `--json` flags
- Added `kwatch master` command for multi-directory monitoring
- Enhanced `kwatch run --command github` for GitHub Actions testing
- All existing commands maintained backward compatibility

#### File Structure
```
~/.kwatch/
├── secure_token.enc  # Encrypted GitHub token (AES-256-GCM)  
├── token.salt        # Random salt for key derivation
└── kwatch.yaml       # Project configuration
```

### 📚 Documentation

#### New Documentation
- **SECURE_AUTH.md** - Comprehensive secure authentication guide
- **Enhanced README.md** - Complete feature documentation with examples
- **Troubleshooting guides** - macOS quarantine fixes and token management

#### Code Documentation
- Comprehensive JSDoc comments for all public APIs
- Detailed inline documentation for security-critical functions
- Architecture documentation for authentication system

### 🐛 Bug Fixes

#### macOS Compatibility
- **Fixed binary auto-close** - Resolved quarantine attribute issues
- **Installation reliability** - Proper handling of extended attributes
- **PATH execution** - Fixed differences between local vs installed binary execution

#### Authentication Reliability  
- **Secure store initialization** - Fixed hanging issues during token store creation
- **Error handling** - Robust error recovery for authentication failures
- **Token validation** - Proper GitHub token format validation

#### TUI Stability
- **Enhanced error handling** - Graceful handling of GitHub API failures
- **Display consistency** - Consistent matrix formatting across all views
- **Memory management** - Optimized memory usage for long-running sessions

### ⚠️ Breaking Changes

None - All changes are backward compatible. Existing workflows continue to work without modification.

### 🔄 Migration Guide

#### From v1.x to v2.0
1. **No code changes required** - All existing functionality preserved
2. **Optional GitHub Actions** - Enable with `kwatch auth --init` 
3. **Enhanced features** - Master view available with `kwatch master`

#### Security Migration
- **Existing environment variables** - Continue to work and take precedence
- **Optional secure storage** - Upgrade to encrypted storage with `kwatch auth --init`
- **Gradual adoption** - Mix of environment variables and secure storage supported

---

## [1.x.x] - Previous Versions

### Core Features (Maintained)
- Real-time TUI monitoring for TSC, Lint, Test commands
- HTTP API for AI agent integration  
- File system watching with automatic command execution
- Command history and performance metrics
- Cross-platform support (Linux, macOS, Windows)
- Global installation and system integration