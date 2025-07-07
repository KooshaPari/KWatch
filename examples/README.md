# KWatch Configuration Examples

This directory contains example configuration files for different environments and use cases.

## Files

- **`kwatch.yaml`** - Complete example with all options documented
- **`development.yaml`** - Development environment optimized for fast feedback
- **`production.yaml`** - Production environment with strict validation

## Usage

1. Copy the appropriate example to your project:
   ```bash
   mkdir -p .kwatch
   cp examples/development.yaml .kwatch/kwatch.yaml
   ```

2. Customize for your project:
   ```bash
   kwatch config edit
   ```

3. Verify configuration:
   ```bash
   kwatch config show
   ```

## Environment-Specific Setup

### Development
- Fast timeouts for quick feedback
- Auto-fix enabled for linting
- Incremental TypeScript compilation
- Prettier formatting enabled

### Production
- Longer timeouts for thorough checking
- Zero warnings policy for linting
- Full test coverage required
- Security audit enabled
- Build verification included

## Custom Commands

You can add any command that returns a meaningful exit code:

```yaml
commands:
  custom-check:
    command: your-custom-script
    args: [--check, --verbose]
    timeout: 30s
    enabled: true
```

## Tips

- Start with the basic `kwatch.yaml` example
- Adjust timeouts based on your project size
- Use `enabled: false` to temporarily disable commands
- Test configuration changes with `kwatch run`
