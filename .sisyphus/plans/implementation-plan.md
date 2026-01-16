# Implementation Plan

## `opencode-sync` - Phased Development Plan

**Target Duration**: 4-5 weeks  
**Start Date**: 2026-01-16

---

## Phase 1: Project Foundation (Week 1)

### Goals
- Set up Go project structure
- Implement cross-platform path resolution
- Create basic CLI scaffold with Cobra
- Implement `init` and `clone` commands

### Tasks

#### 1.1 Project Scaffold
- [ ] Initialize Go module (`go mod init github.com/username/opencode-sync`)
- [ ] Create directory structure per PRD
- [ ] Add `.gitignore`, `Makefile`, `README.md`
- [ ] Set up basic CI with GitHub Actions

#### 1.2 Cross-Platform Paths (`internal/paths/`)
- [ ] Define `Paths` struct with ConfigDir, DataDir, CacheDir
- [ ] Implement `paths_unix.go` with build tag `//go:build unix`
- [ ] Implement `paths_windows.go` with build tag `//go:build windows`
- [ ] Add `GetOpenCodeConfigDir()` and `GetSyncConfigDir()` functions
- [ ] Write tests for all platforms

#### 1.3 CLI Foundation (`internal/cli/`)
- [ ] Set up Cobra root command with version, help
- [ ] Add global flags: `--verbose`, `--dry-run`, `--config`
- [ ] Implement `version` command
- [ ] Add shell completion generation (bash, zsh, fish, powershell)

#### 1.4 Interactive UI (`internal/ui/`)
- [ ] Add charmbracelet/huh dependency
- [ ] Implement `MainMenu()` - shows action selection
- [ ] Implement `SetupWizard()` - guides first-time setup
- [ ] Implement `ConfirmAction()` - yes/no prompts
- [ ] Implement `ShowError()` / `ShowSuccess()` styled output
- [ ] Root command with no args → shows interactive menu

#### 1.5 Config Management (`internal/config/`)
- [ ] Define `Config` struct matching PRD spec
- [ ] Implement config loading with koanf
- [ ] Support JSON and JSONC formats
- [ ] Add config file creation on first run
- [ ] Write tests

#### 1.6 Git Operations (`internal/git/`)
- [ ] Define `Repository` interface
- [ ] Implement `BuiltinGit` using go-git
  - [ ] `Clone(url, path string) error`
  - [ ] `Init(path string) error`
  - [ ] `AddRemote(name, url string) error`
- [ ] Write tests with in-memory git repos

#### 1.7 Commands: `init` and `clone`
- [ ] `init` command:
  - [ ] Detect if repo already exists
  - [ ] Create new repo or link existing
  - [ ] Copy current OpenCode config to repo
  - [ ] Initial commit and push
- [ ] `clone` command:
  - [ ] Clone repo to sync directory
  - [ ] Copy config files to OpenCode config dir
  - [ ] Handle existing config (prompt user)

### Deliverables
- Working `opencode-sync init` command
- Working `opencode-sync clone` command
- Cross-platform binary builds

---

## Phase 2: Core Sync Operations (Week 2)

### Goals
- Implement push/pull/status/diff commands
- Add file change detection
- Handle basic conflict scenarios

### Tasks

#### 2.1 Sync Core (`internal/sync/`)
- [ ] Implement file comparison (hash-based)
- [ ] Track which files should be synced (include/exclude patterns)
- [ ] Create `SyncState` struct to track sync status
- [ ] Implement `GetChanges()` to detect local changes

#### 2.2 Git Operations (continued)
- [ ] Add to `Repository` interface:
  - [ ] `Status() ([]FileStatus, error)`
  - [ ] `Add(paths []string) error`
  - [ ] `Commit(message string) error`
  - [ ] `Push() error`
  - [ ] `Pull() error`
  - [ ] `Diff() (string, error)`
- [ ] Handle authentication (SSH keys, HTTPS)

#### 2.3 Commands: `push`
- [ ] Detect local changes
- [ ] Stage changed files
- [ ] Create commit with message (auto-generate or `-m` flag)
- [ ] Push to remote
- [ ] Handle push failures (conflict, auth)

#### 2.4 Commands: `pull`
- [ ] Fetch remote changes
- [ ] Detect conflicts (local changes + remote changes)
- [ ] **Fail fast on conflict** (per design decision)
- [ ] Apply changes to OpenCode config dir
- [ ] Report what changed

#### 2.5 Commands: `status`
- [ ] Show sync state (in sync, local changes, remote changes)
- [ ] List changed files
- [ ] Show last sync time

#### 2.6 Commands: `diff`
- [ ] Show diff between local and remote
- [ ] Color-coded output
- [ ] Support `--stat` for summary

#### 2.7 Commands: `sync` (pull + push)
- [ ] Implement as convenience wrapper
- [ ] Pull first, then push
- [ ] Stop on conflict (don't push if pull failed)
- [ ] Show summary of what changed

#### 2.8 Interactive Integration
- [ ] Wire up main menu to commands
- [ ] Add progress indicators for long operations
- [ ] Styled success/error messages
- [ ] Return to menu after action (or exit on request)

### Deliverables
- Working `push`, `pull`, `sync`, `status`, `diff` commands
- Interactive menu fully functional
- Proper conflict detection and error messages
- All core sync functionality complete

---

## Phase 3: Encryption (Week 3)

### Goals
- Integrate age encryption
- Implement secret management commands
- Support partial file encryption (JSON paths)

### Tasks

#### 3.1 Encryption Core (`internal/crypto/`)
- [ ] Define `Encryption` interface
- [ ] Implement `AgeEncryption`:
  - [ ] `GenerateKeyPair() (publicKey, privateKey string, error)`
  - [ ] `Encrypt(plaintext []byte, publicKey string) ([]byte, error)`
  - [ ] `Decrypt(ciphertext []byte, privateKey string) ([]byte, error)`
- [ ] Implement `NoOpEncryption` for testing
- [ ] Key file management (load, save, protect permissions)

#### 3.2 Partial File Encryption
- [ ] Parse JSON path patterns (e.g., `provider.*.options.apiKey`)
- [ ] Extract values matching patterns
- [ ] Replace with `{encrypted:base64data}` placeholders
- [ ] Store encrypted values in `secrets.age`
- [ ] Reverse process on decrypt

#### 3.3 Commands: `encrypt` / `decrypt`
- [ ] `encrypt`: Encrypt files before push
- [ ] `decrypt`: Decrypt files after pull
- [ ] Auto-encrypt on push if encryption enabled
- [ ] Auto-decrypt on pull if key available

#### 3.4 Commands: `secret`
- [ ] `secret set <key> <value>`: Add encrypted secret
- [ ] `secret get <key>`: Retrieve and decrypt secret
- [ ] `secret list`: List secret keys (not values)
- [ ] `secret delete <key>`: Remove secret

#### 3.5 Key Setup Flow
- [ ] On `init` with `--encrypt`: Generate keypair
- [ ] Store private key locally (never sync)
- [ ] Optionally store public key in repo
- [ ] On `clone` with encrypted repo: Prompt for key file location

#### 3.6 OAuth Credentials Sync (auth.json)
- [ ] Add `sync.includeAuth` config option (default: false)
- [ ] Add `sync.includeMcpAuth` config option (default: false)
- [ ] Enforce encryption requirement when auth sync enabled
- [ ] Add warning prompt during setup wizard
- [ ] Encrypt entire auth.json file (not partial)
- [ ] Copy from `~/.local/share/opencode/auth.json` to repo
- [ ] Restore to correct location on pull
- [ ] Handle token refresh (access tokens expire, refresh tokens don't)

### Deliverables
- Working encryption/decryption
- Secret management commands
- Secure key file handling
- Optional OAuth credential sync

---

## Phase 4: Polish & Release (Week 4-5)

### Goals
- Add utility commands
- Comprehensive testing
- Documentation
- Release automation

### Tasks

#### 4.1 Commands: `doctor`
- [ ] Check OpenCode installation
- [ ] Verify config paths exist
- [ ] Check git repo health
- [ ] Verify encryption key (if enabled)
- [ ] Check remote connectivity
- [ ] Report issues with suggested fixes

#### 4.2 Commands: `config`
- [ ] `config show`: Display current config
- [ ] `config edit`: Open config in $EDITOR
- [ ] `config set <key> <value>`: Update config value
- [ ] `config path`: Show config file location

#### 4.3 Testing
- [ ] Unit tests for all packages (target >80% coverage)
- [ ] Integration tests with real git operations
- [ ] Cross-platform testing (Linux, macOS, Windows)
- [ ] Test encryption round-trips

#### 4.4 Documentation
- [ ] README with quick start guide
- [ ] `--help` text for all commands
- [ ] Man page generation
- [ ] Example workflows

#### 4.5 Release Automation
- [ ] Set up GoReleaser
- [ ] GitHub Actions for releases
- [ ] Homebrew formula
- [ ] Scoop manifest (Windows)
- [ ] Install script (`curl | sh`)

#### 4.6 Final Polish
- [ ] Consistent error messages
- [ ] Progress indicators for long operations
- [ ] Colored output (with `--no-color` flag)
- [ ] Verbose mode logging

### Deliverables
- Production-ready v0.1.0 release
- Published to Homebrew, Scoop
- Documentation complete

---

## Dependencies Graph

```
Phase 1                 Phase 2              Phase 3           Phase 4
────────                ────────             ────────          ────────
paths ─────────────────────────────────────────────────────────> all
     \
      config ──────────> status ────────────> encrypt ────────> doctor
            \           /      \            /
             git ──────         diff       secret
                  \    \       /
                   init  push/pull ────────> auto-encrypt
                   clone
```

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| go-git limitations | Fallback to external git CLI |
| Windows path issues | Extensive testing, build tags |
| age key loss | Clear documentation, backup prompts |
| Large config files | Test with realistic file sizes |

---

## Definition of Done

### For each command:
- [ ] Implements specified behavior
- [ ] Has unit tests
- [ ] Has integration test
- [ ] `--help` text written
- [ ] Error messages are clear
- [ ] Works on Linux, macOS, Windows

### For release:
- [ ] All commands complete
- [ ] Test coverage >80%
- [ ] Documentation complete
- [ ] Binary size <15MB
- [ ] CI passing on all platforms
