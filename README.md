# opencode-sync

Sync your OpenCode configurations across machines via Git, with optional encryption for secrets.

## Features

- **Cross-platform**: Works on Linux, macOS, and Windows
- **Any git host**: GitHub, GitLab, Bitbucket, self-hosted, etc.
- **Any auth method**: SSH keys, 1Password, gh auth, credential helpers
- **Standalone**: Works before OpenCode starts (no chicken-egg problem)
- **Interactive**: Guided setup and menu-driven interface
- **Encrypted secrets**: Optional age encryption for sensitive data

## Installation

### Quick Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/GareArc/opencode-sync/main/install.sh | bash
```

Or with a specific version:

```bash
VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/GareArc/opencode-sync/main/install.sh | bash
```

### From Source

```bash
go install github.com/GareArc/opencode-sync@latest
```

### From Binary

Download the latest release from the [releases page](https://github.com/GareArc/opencode-sync/releases).

### Homebrew (macOS/Linux)

```bash
brew tap GareArc/tap
brew install opencode-sync
```

### Scoop (Windows)

```powershell
scoop bucket add garearc https://github.com/GareArc/scoop-bucket
scoop install opencode-sync
```

## Quick Start

```bash
# Run interactive setup
opencode-sync

# Or use direct commands
opencode-sync setup     # First-time setup wizard
opencode-sync sync      # Pull and push changes
opencode-sync status    # Check sync status
```

## Usage

### Interactive Mode (Recommended)

Just run `opencode-sync` without arguments to get an interactive menu:

```
$ opencode-sync

  What would you like to do?

  > Sync now (pull + push)
    Pull remote changes
    Push local changes
    View status
    View diff
    Settings
    Exit
```

### Direct Commands

For scripting or power users:

| Command | Description |
|---------|-------------|
| `opencode-sync setup` | Run setup wizard |
| `opencode-sync init` | Initialize new sync repository |
| `opencode-sync link <url>` | Link local configs to existing remote (overwrites remote) |
| `opencode-sync clone <url>` | Clone existing remote (overwrites local) |
| `opencode-sync sync` | Pull then push (most common) |
| `opencode-sync pull` | Pull remote changes |
| `opencode-sync push` | Push local changes |
| `opencode-sync status` | Show sync status |
| `opencode-sync diff` | Show differences |
| `opencode-sync rebind <url>` | Change remote repository URL |
| `opencode-sync doctor` | Diagnose issues |
| `opencode-sync config [show\|path\|edit\|set]` | Manage configuration |
| `opencode-sync key [export\|import\|regen]` | Manage encryption keys |
| `opencode-sync gc` | Optimize repository size (garbage collection) |
| `opencode-sync uninstall` | Uninstall opencode-sync |
| `opencode-sync version` | Show version information |

### Config Subcommands

| Command | Description |
|---------|-------------|
| `opencode-sync config show` | Display current configuration (default) |
| `opencode-sync config path` | Show configuration file path |
| `opencode-sync config edit` | Edit configuration in $EDITOR |
| `opencode-sync config set <key> <value>` | Set a configuration value |

**Available config keys for `set`:**
- `repo.url` - Remote repository URL
- `repo.branch` - Branch name (default: `main`)
- `encryption.enabled` - Enable/disable encryption (`true`/`false`)
- `encryption.keyFile` - Path to encryption key file
- `sync.includeAuth` - Sync auth.json (`true`/`false`)
- `sync.includeMcpAuth` - Sync mcp-auth.json (`true`/`false`)

### Key Subcommands

| Command | Description |
|---------|-------------|
| `opencode-sync key export` | Display private key for backup (default) |
| `opencode-sync key import <key>` | Import key from backup |
| `opencode-sync key regen` | Generate new key (⚠️ old encrypted data lost) |

## Uninstalling

```bash
opencode-sync uninstall
```

This will:
- Remove the binary (may require sudo)
- Optionally remove config (`~/.config/opencode-sync/`) and data (`~/.local/share/opencode-sync/`)
- Your OpenCode configurations are **not affected**

## Requirements

- **git** must be installed and available in PATH

## Repository URL Formats

opencode-sync uses your system's git installation, so any URL format and authentication method your git supports will work:

```bash
# SSH (recommended)
git@github.com:username/repo.git
git@gitlab.com:username/repo.git

# HTTPS
https://github.com/username/repo.git
https://gitlab.com/username/repo.git
```

All standard git authentication methods are supported:
- SSH keys (including 1Password SSH agent, ssh-agent)
- HTTPS credentials via `gh auth login`
- Git credential helpers (macOS Keychain, Windows Credential Manager, etc.)
- `.netrc` files

Works with any git host: GitHub, GitLab, Bitbucket, self-hosted, etc.

## Configuration

Config file location:
- Linux/macOS: `~/.config/opencode-sync/config.json`
- Windows: `%APPDATA%\opencode-sync\config.json`

```json
{
  "repo": {
    "url": "git@github.com:username/opencode-config.git",
    "branch": "main"
  },
  "encryption": {
    "enabled": true,
    "keyFile": "~/.config/opencode-sync/age.key"
  },
  "sync": {
    "includeAuth": false,
    "includeMcpAuth": false,
    "exclude": ["node_modules", "*.log"]
  }
}
```

## What Gets Synced

### Always synced:
- `opencode.json` / `opencode.jsonc` - Main config
- `oh-my-opencode.json` - Oh My OpenCode config
- `AGENTS.md` - Global rules
- `agent/`, `command/`, `skills/`, `mode/`, `themes/`, `plugin/` - Custom extensions
- `~/.claude/skills/` - Claude Code skills (many tools use this as their skill directory)

### Optional (encrypted):
- `auth.json` - OAuth tokens (requires `sync.includeAuth: true`)
- `mcp-auth.json` - MCP auth (requires `sync.includeMcpAuth: true`)

### Never synced:
- Session data
- Logs
- `node_modules/`

### Notes:
- The `~/.claude/skills/` directory is **always created** when syncing to local, even if Claude Code is not installed
- This ensures compatibility with multiple Claude-based tools that use this directory for skills

## Encryption

opencode-sync uses [age](https://age-encryption.org/) for encryption of sensitive files (auth tokens).

### How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                        PUSH (Machine A)                         │
├─────────────────────────────────────────────────────────────────┤
│  auth.json ──encrypt──► auth.json.age ──push──► Git Remote     │
│  (local)      (key)       (repo)                                │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                        PULL (Machine B)                         │
├─────────────────────────────────────────────────────────────────┤
│  Git Remote ──pull──► auth.json.age ──decrypt──► auth.json     │
│                         (repo)         (key)       (local)      │
└─────────────────────────────────────────────────────────────────┘
```

**Decryption is automatic** during `pull`/`clone`/`sync` — but only if:
- You have the private key on the machine (`~/.config/opencode-sync/age.key`)
- Encryption is enabled in config (`encryption.enabled: true`)
- Auth sync is enabled (`sync.includeAuth: true`)

### Full Setup Flow

#### First Machine (Initial Setup)

```bash
# 1. Run setup wizard
opencode-sync setup
#    → Enter repo URL
#    → Enable encryption? Yes
#    → Sync auth tokens? Yes (if you want cross-machine auth)

# 2. Key is auto-generated. BACK IT UP NOW:
opencode-sync key export
#    → Copy the AGE-SECRET-KEY-1... to your password manager

# 3. Initialize and push
opencode-sync init
opencode-sync push
```

#### Second Machine (Clone Existing)

```bash
# 1. Import your key FIRST (before clone)
opencode-sync key import "AGE-SECRET-KEY-1..."

# 2. Run setup with same settings
opencode-sync setup
#    → Same repo URL
#    → Enable encryption? Yes
#    → Sync auth tokens? Yes

# 3. Clone - auth tokens are automatically decrypted
opencode-sync clone git@github.com:user/opencode-config.git
#    ✓ Configs applied
#    ✓ auth.json decrypted automatically
#    → You're logged in!
```

#### Without Key Import (New Machine, No Auth Sync)

If you clone without importing the key:
```bash
opencode-sync clone git@github.com:user/opencode-config.git
#    ✓ Configs applied (opencode.json, agents, etc.)
#    ⚠ auth.json.age skipped (no key to decrypt)
#    → You'll need to re-authenticate in OpenCode
```

### Key Management Commands

| Command | Description |
|---------|-------------|
| `opencode-sync key export` | Display private key for backup |
| `opencode-sync key import <key>` | Import key from backup |
| `opencode-sync key regen` | Generate new key (⚠️ old encrypted data lost) |

### Lost Your Key?

If you lose your private key:
- ❌ Encrypted auth tokens are **unrecoverable**
- ✅ Configs, agents, commands, themes are **not encrypted** — still accessible

**Recovery steps:**
```bash
# 1. Generate a new key
opencode-sync key regen

# 2. Re-authenticate in OpenCode (get new auth.json)

# 3. Push with new encryption
opencode-sync push
```

### What Gets Encrypted?

| File | Encrypted | Notes |
|------|-----------|-------|
| `auth.json` | ✅ Yes | OAuth tokens (if `sync.includeAuth: true`) |
| `mcp-auth.json` | ✅ Yes | MCP auth (if `sync.includeMcpAuth: true`) |
| `opencode.json` | ❌ No | Main config |
| `AGENTS.md` | ❌ No | Global rules |
| `agent/`, `command/`, etc. | ❌ No | Custom extensions |

### Security Notes

- Private key stored at: `~/.config/opencode-sync/age.key`
- Key is **never synced** to remote — stays local only
- Encrypted files use `.age` extension in repo
- **Back up your key immediately** after setup to a password manager

## Repository Size Management

opencode-sync uses git to store config history locally at `~/.local/share/opencode-sync/repo/`.

### Space Optimizations

**Automatic optimizations:**
- **Shallow clone**: When cloning, only the latest commit is fetched (saves ~90% space)
- **Auto GC on pull**: Git garbage collection runs automatically after pulling changes

**Manual optimization:**
```bash
opencode-sync gc  # Compress repository (70-90% size reduction)
```

### Projected Storage Usage

| Commits | Without GC | With Auto GC |
|---------|-----------|--------------|
| 100     | ~2 MB     | ~200 KB      |
| 1,000   | ~20 MB    | ~1-2 MB      |
| 10,000  | ~200 MB   | ~5-10 MB     |

**Conclusion**: Even after thousands of syncs, storage usage remains minimal (1-10 MB).

## Development

```bash
# Clone the repo
git clone https://github.com/GareArc/opencode-sync.git
cd opencode-sync

# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Run locally
make run
```

## License

MIT
