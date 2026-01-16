# opencode-sync

Sync your OpenCode configurations across machines via Git, with optional encryption for secrets.

## Features

- **Cross-platform**: Works on Linux, macOS, and Windows
- **Standalone**: Works before OpenCode starts (no chicken-egg problem)
- **Interactive**: Guided setup and menu-driven interface
- **Encrypted secrets**: Optional age encryption for sensitive data
- **OAuth sync**: Optionally sync authentication tokens (encrypted)

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

### Homebrew (coming soon)

```bash
brew install GareArc/tap/opencode-sync
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
| `opencode-sync doctor` | Diagnose issues |
| `opencode-sync config` | Manage configuration |

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
- `AGENTS.md` - Global rules
- `agent/`, `command/`, `skill/`, `mode/`, `themes/` - Custom extensions

### Optional (encrypted):
- `auth.json` - OAuth tokens (requires `sync.includeAuth: true`)
- `mcp-auth.json` - MCP auth (requires `sync.includeMcpAuth: true`)

### Never synced:
- Session data
- Logs
- `node_modules/`

## Encryption

opencode-sync uses [age](https://age-encryption.org/) for encryption.

```bash
# Enable encryption during setup, or manually:
opencode-sync config
```

Your encryption key is stored locally and **never synced**. When setting up a new machine, you'll need to transfer the key file securely.

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
