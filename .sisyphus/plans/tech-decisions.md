# Technical Decisions & Quick Reference

## Quick Reference

### Commands Summary

**Primary (Interactive):**
```bash
opencode-sync              # Interactive menu - START HERE
```

**Direct Commands (for scripting):**

| Command | Description |
|---------|-------------|
| `opencode-sync sync` | Pull then push (most common action) |
| `opencode-sync pull` | Pull remote changes |
| `opencode-sync push` | Push local changes |
| `opencode-sync status` | Show sync status |
| `opencode-sync diff` | Show differences |
| `opencode-sync setup` | Re-run setup wizard |
| `opencode-sync doctor` | Diagnose issues |
| `opencode-sync config` | Open settings menu |
| `opencode-sync version` | Show version |

---

## Tech Stack

```
┌─────────────────────────────────────────────────────────────┐
│                        opencode-sync                        │
├─────────────────────────────────────────────────────────────┤
│  CLI Layer         │  spf13/cobra                           │
├────────────────────┼────────────────────────────────────────┤
│  Interactive UI    │  charmbracelet/huh (forms, menus)      │
├────────────────────┼────────────────────────────────────────┤
│  Config            │  knadh/koanf (JSON/JSONC)              │
├────────────────────┼────────────────────────────────────────┤
│  Git Operations    │  go-git/go-git/v5 (pure Go)            │
├────────────────────┼────────────────────────────────────────┤
│  Encryption        │  filippo.io/age                        │
├────────────────────┼────────────────────────────────────────┤
│  Keychain          │  zalando/go-keyring                    │
├────────────────────┼────────────────────────────────────────┤
│  Paths             │  os.UserConfigDir() + build tags       │
└─────────────────────────────────────────────────────────────┘
```

---

## Key Decisions

### 1. Standalone CLI (not plugin)

**Decision:** Build standalone Go binary, not OpenCode plugin.

**Rationale:**
- Works before OpenCode starts (solves bootstrap problem)
- No Node.js runtime dependency
- Can be used in automation/scripts
- Single binary distribution

### 2. Fail-Fast Conflict Resolution

**Decision:** Fail immediately on conflicts, require manual resolution.

**Rationale:**
- Config files are critical - silent merges dangerous
- User explicitly decides resolution
- Simplest, least surprising behavior
- No complex merge logic needed

### 3. Age Encryption with Key File

**Decision:** Use age encryption with local key file (not passphrase).

**Rationale:**
- No typing passphrase each time
- Key file easily backed up separately
- age is modern, audited, simple API
- Supports SSH keys as alternative

### 4. Pure Go Git (go-git)

**Decision:** Use go-git library, not shell out to git binary.

**Rationale:**
- No external dependency
- Better error handling
- Testable with in-memory repos
- Cross-platform without git installation

**Fallback:** External git CLI as optional fallback for edge cases.

### 5. koanf over Viper

**Decision:** Use koanf for config parsing.

**Rationale:**
- Lighter weight, fewer dependencies
- Case-sensitive keys (important for paths/URLs)
- Modular - only import what we need
- Better design patterns

### 6. Build Tags for Platform Code

**Decision:** Use Go build tags for platform-specific code.

**Rationale:**
- Clean separation (no runtime checks)
- Compiler excludes irrelevant code
- Pattern proven by chezmoi
- Easy to test each platform

### 7. Interactive-First CLI

**Decision:** Default to interactive mode with menus; direct commands for scripting.

**Rationale:**
- Users don't need to memorize commands
- Guided setup reduces errors
- Power users can still script with direct commands
- Better UX for occasional use

**Library:** charmbracelet/huh for forms and menus

### 8. Optional Auth Sync

**Decision:** Allow encrypted sync of OAuth credentials as opt-in.

**Rationale:**
- Refresh tokens work across machines
- Eliminates re-authentication friction
- Must be encrypted (enforced)
- User explicitly opts in with warning

**Safeguards:**
- Requires `encryption.enabled: true`
- Requires `sync.includeAuth: true`
- Warning during setup
- Private repo recommended

---

## File Paths

### Config Locations

| Platform | opencode-sync Config | OpenCode Config |
|----------|---------------------|-----------------|
| Linux | `~/.config/opencode-sync/` | `~/.config/opencode/` |
| macOS | `~/.config/opencode-sync/` | `~/.config/opencode/` |
| Windows | `%APPDATA%\opencode-sync\` | `%APPDATA%\opencode\` |

### Sync Repo Location

Default: `~/.local/share/opencode-sync/repo/`

### Files to Sync

```
ALWAYS SYNC:
  opencode.json / opencode.jsonc
  AGENTS.md
  agent/
  command/
  skill/
  mode/
  themes/
  plugin/*.json (e.g., oh-my-opencode.json)

OPT-IN ENCRYPTED SYNC:
  auth.json        (sync.includeAuth: true)
  mcp-auth.json    (sync.includeMcpAuth: true)

NEVER SYNC:
  node_modules/
  storage/
  log/
```

---

## Code Patterns

### Cross-Platform Paths

```go
// paths/paths.go
type Paths struct {
    ConfigDir string  // opencode-sync config
    DataDir   string  // sync repo location
    OpenCode  string  // OpenCode config dir
}

func Get() (*Paths, error)
```

```go
// paths/paths_unix.go
//go:build unix

func Get() (*Paths, error) {
    home, _ := os.UserHomeDir()
    return &Paths{
        ConfigDir: filepath.Join(home, ".config", "opencode-sync"),
        DataDir:   filepath.Join(home, ".local", "share", "opencode-sync"),
        OpenCode:  filepath.Join(home, ".config", "opencode"),
    }, nil
}
```

```go
// paths/paths_windows.go
//go:build windows

func Get() (*Paths, error) {
    return &Paths{
        ConfigDir: filepath.Join(os.Getenv("APPDATA"), "opencode-sync"),
        DataDir:   filepath.Join(os.Getenv("LOCALAPPDATA"), "opencode-sync"),
        OpenCode:  filepath.Join(os.Getenv("APPDATA"), "opencode"),
    }, nil
}
```

### Git Interface

```go
// git/git.go
type Repository interface {
    Clone(url string) error
    Init() error
    Status() ([]FileStatus, error)
    Add(paths []string) error
    Commit(message string) error
    Push() error
    Pull() error
    Diff() (string, error)
}
```

### Encryption Interface

```go
// crypto/encryption.go
type Encryption interface {
    Encrypt(plaintext []byte) ([]byte, error)
    Decrypt(ciphertext []byte) ([]byte, error)
    GenerateKey() (publicKey, privateKey string, error)
}
```

### Interactive UI (charmbracelet/huh)

```go
// cli/interactive.go
import "github.com/charmbracelet/huh"

func ShowMainMenu() (string, error) {
    var choice string
    
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("What would you like to do?").
                Options(
                    huh.NewOption("Sync now (pull + push)", "sync"),
                    huh.NewOption("Pull remote changes", "pull"),
                    huh.NewOption("Push local changes", "push"),
                    huh.NewOption("View status", "status"),
                    huh.NewOption("View diff", "diff"),
                    huh.NewOption("Settings", "config"),
                    huh.NewOption("Exit", "exit"),
                ).
                Value(&choice),
        ),
    )
    
    err := form.Run()
    return choice, err
}

func SetupWizard() (*Config, error) {
    var repoURL string
    var enableEncryption bool
    var includeAuth bool
    
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Enter your Git repo URL").
                Placeholder("git@github.com:user/opencode-config.git").
                Value(&repoURL),
        ),
        huh.NewGroup(
            huh.NewConfirm().
                Title("Enable encryption for secrets?").
                Affirmative("Yes (recommended)").
                Negative("No").
                Value(&enableEncryption),
        ),
        huh.NewGroup(
            huh.NewConfirm().
                Title("Sync OAuth credentials (auth.json)?").
                Description("Warning: Requires secure key transfer to new machines").
                Affirmative("Yes (encrypted)").
                Negative("No (re-authenticate each machine)").
                Value(&includeAuth),
        ).WithHideFunc(func() bool { return !enableEncryption }),
    )
    
    if err := form.Run(); err != nil {
        return nil, err
    }
    
    return &Config{
        Repo:       RepoConfig{URL: repoURL},
        Encryption: EncryptionConfig{Enabled: enableEncryption},
        Sync:       SyncConfig{IncludeAuth: includeAuth},
    }, nil
}
```

---

## Error Handling

### User-Facing Errors

```go
// Use consistent error format
fmt.Fprintf(os.Stderr, "error: %s\n", err)
fmt.Fprintf(os.Stderr, "\nRun 'opencode-sync doctor' to diagnose issues.\n")
os.Exit(1)
```

### Conflict Error

```
error: conflict detected in opencode.json
  local changes would be overwritten by remote

Resolution options:
  opencode-sync pull --force   # Discard local, use remote
  opencode-sync push --force   # Discard remote, use local
  opencode-sync diff           # View differences
```

---

## Testing Strategy

### Unit Tests
- All packages have `*_test.go` files
- Use table-driven tests
- Mock filesystem with `afero` if needed
- Mock git with in-memory repos

### Integration Tests
- Test full workflows (init → push → clone → pull)
- Use temp directories
- Test on all platforms in CI

### Test Coverage Target
- Minimum 80% coverage
- Critical paths (sync, encryption) at 90%+

---

## CI/CD

### GitHub Actions Matrix

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
    go: ['1.22']
```

### Release Process

1. Tag version: `git tag v0.1.0`
2. Push tag: `git push origin v0.1.0`
3. GoReleaser builds binaries
4. Publishes to GitHub Releases
5. Updates Homebrew tap

---

## Dependencies (go.mod)

```go
require (
    github.com/spf13/cobra v1.8.0
    github.com/charmbracelet/huh v0.6.0
    github.com/charmbracelet/lipgloss v1.0.0
    github.com/go-git/go-git/v5 v5.12.0
    github.com/knadh/koanf/v2 v2.1.0
    filippo.io/age v1.2.0
    github.com/zalando/go-keyring v0.2.5
)
```

---

## Interactive Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     $ opencode-sync                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │ Config exists?   │
                    └──────────────────┘
                      │           │
                     No          Yes
                      │           │
                      ▼           ▼
            ┌──────────────┐  ┌──────────────┐
            │ Setup Wizard │  │  Main Menu   │
            └──────────────┘  └──────────────┘
                  │                   │
                  ▼                   ▼
            ┌──────────────┐  ┌──────────────────────────┐
            │ • Repo URL   │  │ • Sync now               │
            │ • Encryption │  │ • Pull / Push            │
            │ • Auth sync  │  │ • Status / Diff          │
            └──────────────┘  │ • Settings               │
                  │           │ • Exit                   │
                  ▼           └──────────────────────────┘
            ┌──────────────┐
            │ Main Menu    │
            └──────────────┘
```
