# Product Requirements Document (PRD)

## `opencode-sync` - OpenCode Configuration Sync CLI

**Version**: 0.1.0  
**Status**: Approved  
**Date**: 2026-01-16

---

## 1. Executive Summary

**Product Name:** `opencode-sync`

**One-liner:** A cross-platform CLI tool to sync OpenCode configurations across machines via Git, with optional encryption for secrets.

**Problem:** Developers using OpenCode across multiple machines (work laptop, personal desktop, servers) need to manually copy configuration files or risk inconsistent setups. The existing `opencode-synced` plugin only works inside OpenCode, creating a chicken-egg problem on fresh machines.

**Solution:** A standalone Go binary that syncs OpenCode configs to/from a Git repository, works before OpenCode starts, and optionally encrypts sensitive data.

---

## 2. Goals & Non-Goals

### Goals

- Sync OpenCode configs across machines via Git
- Work standalone (before OpenCode runs)
- Cross-platform: Linux, macOS, Windows
- Single binary, no dependencies (no Node, Python, git CLI required)
- Optional encryption for secrets using `age`
- Support GitHub, GitLab, and self-hosted Git remotes

### Non-Goals

- Sync session data or conversation history (privacy concern)
- Real-time sync (not a daemon/service)
- General-purpose dotfiles manager (OpenCode-specific)
- GUI (CLI only for v1)

---

## 3. User Stories

| As a... | I want to... | So that... |
|---------|--------------|------------|
| Developer | Run `opencode-sync init` on my first machine | My configs are stored in Git |
| Developer | Run `opencode-sync clone` on a new machine | I get my configs before starting OpenCode |
| Developer | Run `opencode-sync push` after changing settings | Changes sync to my repo |
| Developer | Run `opencode-sync pull` before starting work | I have the latest configs |
| Developer | Encrypt my API keys | They're not exposed in my Git repo |
| Developer | Use any Git remote | I'm not locked into GitHub |

---

## 4. Technical Specification

### 4.1 Tech Stack

| Component | Choice | Rationale |
|-----------|--------|-----------|
| **Language** | Go 1.22+ | Single binary, cross-platform, fast |
| **CLI Framework** | [spf13/cobra](https://github.com/spf13/cobra) | Industry standard (kubectl, docker, gh) |
| **Interactive UI** | [charmbracelet/huh](https://github.com/charmbracelet/huh) | Beautiful forms, prompts, menus |
| **Git Operations** | [go-git/go-git/v5](https://github.com/go-git/go-git) | Pure Go, no git binary dependency |
| **Encryption** | [filippo.io/age](https://github.com/FiloSottile/age) | Modern, simple, audited |
| **Config Parsing** | [knadh/koanf](https://github.com/knadh/koanf) | Lightweight, case-sensitive |
| **Keychain** | [zalando/go-keyring](https://github.com/zalando/go-keyring) | Cross-platform keychain access |

### 4.2 OpenCode Config Paths

| Platform | Config Dir | Data Dir |
|----------|------------|----------|
| Linux | `~/.config/opencode/` | `~/.local/share/opencode/` |
| macOS | `~/.config/opencode/` | `~/.local/share/opencode/` |
| Windows | `%APPDATA%\opencode\` | `%LOCALAPPDATA%\opencode\` |

### 4.3 What Gets Synced

**Always Synced:**

```
~/.config/opencode/
├── opencode.json          # Main config
├── opencode.jsonc         # Alternative format
├── AGENTS.md              # Global rules
├── agent/                 # Custom agents
├── command/               # Custom commands
├── skill/                 # Custom skills
├── mode/                  # Custom modes
├── themes/                # Custom themes
└── plugin/                # Plugin configs (e.g., oh-my-opencode.json)
```

**Optionally Synced (Encrypted Only):**

```
~/.local/share/opencode/
├── auth.json              # OAuth tokens - OPT-IN, requires encryption
└── mcp-auth.json          # MCP auth - OPT-IN, requires encryption
```

**Never Synced:**

```
~/.local/share/opencode/
├── storage/               # Sessions - privacy
└── log/                   # Logs - not useful
```

### 4.4 Interactive CLI Design

**Philosophy:** Minimal memorization. Run `opencode-sync` and get guided through options.

#### Primary Usage (Interactive)

```bash
$ opencode-sync
```

Shows interactive menu:
```
┌─────────────────────────────────────────┐
│  opencode-sync v0.1.0                   │
│                                         │
│  What would you like to do?             │
│                                         │
│  > Sync now (pull + push)               │
│    Pull remote changes                  │
│    Push local changes                   │
│    View status                          │
│    View diff                            │
│    Settings                             │
│    Setup wizard                         │
│    Exit                                 │
└─────────────────────────────────────────┘
```

#### First Run (Guided Setup)

```bash
$ opencode-sync

  Welcome to opencode-sync!
  
  Let's set up syncing for your OpenCode config.
  
  ? Do you have an existing sync repo?
    > Yes, I want to clone it
      No, create a new one
  
  ? Enter your repo URL:
    git@github.com:username/opencode-config.git
  
  ? Enable encryption for secrets?
    > Yes (recommended)
      No
  
  ? Sync OAuth credentials (auth.json)?
    > No (re-authenticate on each machine)
      Yes (encrypted, requires key transfer)
  
  ✓ Setup complete! Your config is now synced.
```

#### Direct Commands (for scripting/power users)

```bash
# Quick actions (no prompts)
opencode-sync sync                  # Pull then push (most common)
opencode-sync push                  # Push local changes
opencode-sync pull                  # Pull remote changes
opencode-sync status                # Show sync status

# Setup (interactive by default, flags for scripting)
opencode-sync init [repo-url]       # Create new sync repo
opencode-sync clone <repo-url>      # Clone existing config
opencode-sync setup                 # Re-run setup wizard

# Utilities
opencode-sync diff                  # Show what would change
opencode-sync doctor                # Diagnose issues
opencode-sync config                # Open settings menu
opencode-sync version               # Show version info
```

#### Typical Daily Workflow

```bash
# Option 1: Interactive (recommended)
$ opencode-sync
# Select "Sync now" from menu

# Option 2: One command
$ opencode-sync sync

# Option 3: Explicit
$ opencode-sync pull && opencode-sync push
```

### 4.5 Configuration File

**Location:** `~/.config/opencode-sync/config.json`

```json
{
  "$schema": "https://opencode-sync.dev/config.schema.json",
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

**Configuration Options:**

| Option | Default | Description |
|--------|---------|-------------|
| `repo.url` | - | Git repository URL (SSH or HTTPS) |
| `repo.branch` | `main` | Branch to sync |
| `encryption.enabled` | `false` | Enable age encryption |
| `encryption.keyFile` | `~/.config/opencode-sync/age.key` | Path to age private key |
| `sync.includeAuth` | `false` | Sync OAuth tokens (requires encryption) |
| `sync.includeMcpAuth` | `false` | Sync MCP auth (requires encryption) |
| `sync.exclude` | `[]` | Glob patterns to exclude |

### 4.6 Project Structure

```
opencode-sync/
├── cmd/
│   └── opencode-sync/
│       └── main.go              # Entry point
├── internal/
│   ├── cli/                     # Cobra commands
│   │   ├── root.go
│   │   ├── init.go
│   │   ├── clone.go
│   │   ├── push.go
│   │   ├── pull.go
│   │   ├── status.go
│   │   ├── diff.go
│   │   ├── encrypt.go
│   │   ├── secret.go
│   │   └── doctor.go
│   ├── config/                  # Config parsing
│   │   ├── config.go
│   │   └── opencode.go          # OpenCode config parsing
│   ├── paths/                   # Cross-platform paths
│   │   ├── paths.go             # Common interface
│   │   ├── paths_unix.go        # //go:build unix
│   │   └── paths_windows.go     # //go:build windows
│   ├── git/                     # Dual-mode git
│   │   ├── git.go               # Interface
│   │   ├── builtin.go           # go-git implementation
│   │   └── external.go          # git CLI fallback
│   ├── crypto/                  # Encryption
│   │   ├── encryption.go        # Interface
│   │   ├── age.go               # Age implementation
│   │   └── none.go              # No-op for testing
│   ├── sync/                    # Core sync logic
│   │   ├── sync.go
│   │   ├── diff.go
│   │   └── merge.go
│   └── system/                  # Filesystem abstraction
│       ├── system.go            # Interface
│       ├── real.go              # Real filesystem
│       └── dry.go               # Dry-run mode
├── testdata/                    # Test fixtures
├── .goreleaser.yaml             # Release automation
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 5. Design Decisions

### 5.1 Conflict Resolution

**Decision:** Fail fast, require manual resolution.

**Rationale:**
- Config files are critical; silent merges could break OpenCode
- User should explicitly decide how to resolve conflicts
- Simplest implementation, least surprising behavior

**Behavior:**
```
$ opencode-sync pull
error: conflict detected in opencode.json
  local changes would be overwritten by remote

Resolution options:
  opencode-sync pull --force   # Discard local, use remote
  opencode-sync push --force   # Discard remote, use local
  opencode-sync diff           # View differences
```

### 5.2 Encryption

**Decision:** Key file based encryption using `age`.

**Rationale:**
- No typing passphrase each time
- Key file can be backed up separately (USB, password manager)
- `age` is modern, audited, and simple

**Key Management:**
```
~/.config/opencode-sync/
├── config.json          # Synced (no secrets)
├── age.key              # NEVER synced (local only)
└── age.pub              # Can be synced (public key)
```

**Encryption Flow:**
1. On `init`: Generate age keypair, store in `~/.config/opencode-sync/`
2. On `push`: Encrypt files matching patterns, commit encrypted versions
3. On `pull`: Decrypt files using local key
4. On new machine: User must transfer `age.key` manually (secure channel)

### 5.3 OAuth Credential Sync (Optional)

**Decision:** Allow encrypted sync of `auth.json` as opt-in feature.

**Rationale:**
- Refresh tokens are long-lived and work across machines
- Eliminates need to re-authenticate on each device
- Must be encrypted (age) - never stored in plaintext
- User explicitly opts in understanding the security implications

**auth.json Structure:**
```json
{
  "openai": {
    "type": "oauth",
    "refresh": "rt_xxx...",      // Long-lived, THE valuable part
    "access": "eyJxxx...",       // Short-lived, auto-refreshes
    "expires": 1769241599876
  },
  "anthropic": { ... },
  "google": { ... }
}
```

**Safeguards:**
1. Requires `encryption.enabled: true` (enforced)
2. Requires `sync.includeAuth: true` (explicit opt-in)
3. Warning shown during setup
4. Key file must exist before enabling

**Setup Flow:**
```
? Sync OAuth credentials (auth.json)?

  ⚠️  This will sync your login tokens across machines.
  
  Pros:
  • No need to re-authenticate on each machine
  • Seamless experience across devices
  
  Cons:
  • If your key file is compromised, tokens are exposed
  • You must transfer the key file securely to new machines
  
  > No (re-authenticate on each machine)
    Yes (encrypted, I understand the risks)
```

### 5.4 Git Authentication

**Supported Methods:**
1. SSH keys (recommended) - Uses `~/.ssh/` keys automatically
2. HTTPS with credential helper - Uses system credential store
3. Personal access tokens - Environment variable `OPENCODE_SYNC_TOKEN`

---

## 6. Distribution

### 6.1 Installation Methods

```bash
# Homebrew (macOS/Linux)
brew install opencode-sync

# Go install
go install github.com/username/opencode-sync@latest

# Direct download
curl -fsSL https://opencode-sync.dev/install.sh | sh

# Windows (scoop)
scoop install opencode-sync
```

### 6.2 Release Automation

Using GoReleaser:
- Builds for Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
- Creates GitHub releases with checksums
- Publishes to Homebrew tap

---

## 7. Success Metrics

| Metric | Target |
|--------|--------|
| Binary size | < 15MB |
| Init time | < 2 seconds |
| Push/pull time | < 5 seconds |
| Platforms | Linux, macOS, Windows |
| Test coverage | > 80% |

---

## 8. Alternatives Considered

| Alternative | Why Not Chosen |
|-------------|----------------|
| **opencode-synced plugin** | Can't run before OpenCode starts |
| **chezmoi** | General-purpose, overkill for single tool |
| **Bare git repo** | No encryption, manual process |
| **TypeScript CLI** | Requires Node runtime |

---

## 9. Appendix

### A. Comparison with opencode-synced

| Feature | opencode-synced | opencode-sync |
|---------|-----------------|---------------|
| Works before OpenCode | No | Yes |
| Encrypted secrets | Private repo only | age encryption |
| Standalone CLI | No | Yes |
| Git binary required | Yes (via gh) | No (go-git) |
| Cross-tool support | No | Possible future |

### B. Research Sources

- chezmoi architecture analysis (Go patterns, encryption, git)
- go-git documentation and examples
- age encryption library
- Cobra CLI framework
- koanf configuration library
