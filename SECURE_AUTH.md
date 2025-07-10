# 🔐 Secure GitHub Token Management

KWatch includes a **secure encrypted token storage system** that provides a safer alternative to storing GitHub tokens in environment variables or shell profiles.

## 🎯 **Quick Setup**

```bash
# Setup secure encrypted token storage
kwatch auth --init

# Check authentication status
kwatch auth --status

# Test GitHub Actions monitoring
kwatch run --command github
```

## 🔒 **Security Features**

✅ **AES-256-GCM Encryption**: Military-grade encryption for your tokens  
✅ **System-Specific Keys**: Encryption keys derived from your system's unique identifiers  
✅ **Secure File Permissions**: Files stored with 600 permissions (user-only access)  
✅ **No Plaintext Storage**: No tokens stored in environment variables or config files  
✅ **Automatic Detection**: KWatch automatically finds and uses your encrypted token  

## 📋 **Complete Command Reference**

### **Initial Setup**
```bash
kwatch auth --init                # Interactive secure setup
```
- Prompts for your GitHub token (input is hidden)
- Validates token format
- Encrypts and stores securely
- Shows storage location and security info

### **Status Checking**
```bash
kwatch auth --status              # Human-readable status
kwatch auth --status --json       # JSON format for automation
```

### **Token Management**
```bash
kwatch auth --clear               # Remove encrypted token (with confirmation)
```

### **Default Behavior**
```bash
kwatch auth                       # Shows current status and guidance
```

## 🛡️ **How It Works**

### **Encryption Process**
1. **Key Derivation**: Combines system-specific data (OS, architecture, username, hostname) with a random salt
2. **AES-256-GCM**: Uses authenticated encryption to prevent tampering
3. **Secure Storage**: Saves encrypted token to `~/.kwatch/secure_token.enc` with 600 permissions
4. **Salt Storage**: Random salt stored separately in `~/.kwatch/token.salt`

### **Token Priority Order**
KWatch checks for tokens in this order:
1. `GITHUB_TOKEN` environment variable
2. `GH_TOKEN` environment variable  
3. Encrypted token in secure store
4. No token (GitHub Actions disabled)

### **File Locations**
```
~/.kwatch/
├── secure_token.enc  # Encrypted GitHub token (AES-256-GCM)
├── token.salt        # Random salt for key derivation
└── kwatch.yaml       # Project configuration (unrelated)
```

## 🎫 **Getting Your GitHub Token**

1. **Go to GitHub Settings**: https://github.com/settings/tokens
2. **Generate new token (classic)**
3. **Select scopes**:
   - `repo` (for private repositories)  
   - `actions:read` (for GitHub Actions monitoring)
4. **Copy the token** (starts with `ghp_` or `github_pat_`)

## 🧪 **Testing Your Setup**

```bash
# Check authentication status
kwatch auth --status

# Test GitHub Actions monitoring  
kwatch run --command github

# Use in TUI mode
kwatch .                         # GitHub Actions column appears

# Use in master mode
kwatch master                    # GitHub Actions in matrix view
```

## 🚨 **Troubleshooting**

### **Token Not Working**
```bash
# Check status and validate token
kwatch auth --status

# Clear and reinitialize if corrupted
kwatch auth --clear
kwatch auth --init
```

### **Permission Errors**
```bash
# Check file permissions
ls -la ~/.kwatch/

# Should show:
# -rw------- secure_token.enc  (600 permissions)
# -rw------- token.salt        (600 permissions)
```

### **Multiple Systems**
Each system has unique encryption keys. You'll need to run `kwatch auth --init` on each machine.

### **Migration from Environment Variables**
If you currently use `GITHUB_TOKEN`:
1. Run `kwatch auth --init` to setup secure storage
2. Remove `export GITHUB_TOKEN=...` from shell profile (optional - environment variables take precedence)
3. Restart terminal or run `unset GITHUB_TOKEN`

## 🔄 **Backup & Recovery**

### **Backup Your Token**
```bash
# Get token type and preview
kwatch auth --status --json

# Note: The encrypted files are system-specific and can't be copied between machines
```

### **Recovery Process**
If you lose access:
1. Generate a new GitHub token at https://github.com/settings/tokens
2. Run `kwatch auth --clear` to remove old encrypted token
3. Run `kwatch auth --init` to setup the new token

## ⚡ **Integration Examples**

### **CI/CD Workflows**
```bash
# In GitHub Actions or other CI
export GITHUB_TOKEN="${{ secrets.GITHUB_TOKEN }}"
kwatch run --command github
```

### **Development Workflow**
```bash
# One-time setup
kwatch auth --init

# Daily usage (token automatically detected)
kwatch .                         # TUI with GitHub Actions
kwatch master --watch            # Multi-project monitoring
kwatch status                    # JSON status with GitHub Actions
```

### **Team Usage**
Each developer runs:
```bash
kwatch auth --init              # Individual encrypted token storage
```

## 🎉 **Benefits Over Environment Variables**

✅ **No Shell Profile Pollution**: No need to edit `.zshrc`, `.bashrc`, etc.  
✅ **System-Specific Security**: Tokens can't be copied between machines  
✅ **Automatic Management**: KWatch handles token detection and usage  
✅ **No Accidental Exposure**: No risk of tokens in shell history or config dumps  
✅ **Easy Rotation**: Simple `--clear` and `--init` process for token updates  
✅ **Fallback Support**: Environment variables still work and take precedence

---

**Need help?** Run `kwatch auth --help` for detailed command information!