package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/GareArc/opencode-sync/internal/config"
	"github.com/GareArc/opencode-sync/internal/crypto"
	"github.com/GareArc/opencode-sync/internal/git"
	"github.com/GareArc/opencode-sync/internal/paths"
	"github.com/GareArc/opencode-sync/internal/sync"
	"github.com/GareArc/opencode-sync/internal/ui"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("opencode-sync %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
	},
}

// syncCmd represents the sync command (pull + push)
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync configurations (pull then push)",
	Long:  `Pull remote changes and push local changes in one command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSync()
	},
}

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push local changes to remote",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPush()
	},
}

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull remote changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPull()
	},
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show differences between local and remote",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDiff()
	},
}

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run the setup wizard",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetupWizard()
	},
}

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose configuration issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDoctor()
	},
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new sync repository",
	Long: `Initialize a new Git repository for syncing OpenCode configurations.

This command will:
1. Create a new Git repository
2. Copy current OpenCode configs to the repository
3. Create an initial commit
4. Set up the remote (if configured)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone [repository-url]",
	Short: "Clone an existing sync repository",
	Long: `Clone an existing sync repository from a remote URL.

This command will:
1. Clone the repository from the remote URL
2. Apply the configurations to your local OpenCode`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var repoURL string
		if len(args) > 0 {
			repoURL = args[0]
		}
		return runClone(repoURL)
	},
}

// linkCmd represents the link command
var linkCmd = &cobra.Command{
	Use:   "link <repository-url>",
	Short: "Link local configs to an existing remote repository",
	Long: `Link your local OpenCode configurations to an existing remote repository.

This command will:
1. Create a local Git repository from your current configs
2. Add the remote repository URL
3. Push your local configs to the remote (force push, overwriting remote)

Use this when you have a remote repository (initialized on another machine)
and want to sync your local configs to it, keeping local as source of truth.

Example:
  opencode-sync link git@github.com:username/opencode-config.git`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLink(args[0])
	},
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage opencode-sync configuration. Use subcommands to view, edit, or modify settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no subcommand provided, show the config
		return runConfigShow()
	},
}

// configShowCmd shows the current configuration
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigShow()
	},
}

// configPathCmd shows the config file path
var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigPath()
	},
}

// configEditCmd opens config in editor
var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration in $EDITOR",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigEdit()
	},
}

// configSetCmd sets a configuration value
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value using dot notation.

Examples:
  opencode-sync config set repo.url git@github.com:user/repo.git
  opencode-sync config set repo.branch main
  opencode-sync config set encryption.enabled true
  opencode-sync config set sync.includeAuth false`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigSet(args[0], args[1])
	},
}

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage encryption keys",
	Long:  `Manage encryption keys for secure syncing of auth tokens.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runKeyExport()
	},
}

var keyExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export private key for backup",
	Long: `Export your private encryption key.

IMPORTANT: Store this key securely (e.g., password manager).
Without it, encrypted data (auth tokens) cannot be recovered.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runKeyExport()
	},
}

var keyImportCmd = &cobra.Command{
	Use:   "import <key>",
	Short: "Import a private key",
	Long: `Import a private key from backup.

Use this when setting up a new machine to decrypt existing auth tokens.

Example:
  opencode-sync key import "AGE-SECRET-KEY-1QQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQ"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runKeyImport(args[0])
	},
}

var keyRegenCmd = &cobra.Command{
	Use:   "regen",
	Short: "Regenerate encryption key",
	Long: `Generate a new encryption key, replacing the existing one.

WARNING: Previously encrypted data will become unrecoverable!
Only use this if you've lost your key and need to start fresh.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runKeyRegen()
	},
}

var rebindCmd = &cobra.Command{
	Use:   "rebind <url>",
	Short: "Change the remote repository URL",
	Long: `Change the remote repository URL without reinitializing.

This updates the remote URL for an existing sync repository.
Useful when migrating to a new git host or changing repo location.

Examples:
  opencode-sync rebind git@github.com:user/new-repo.git
  opencode-sync rebind https://github.com/user/new-repo.git`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRebind(args[0])
	},
}

func init() {
	// Add config subcommands
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configSetCmd)

	// Add key subcommands
	keyCmd.AddCommand(keyExportCmd)
	keyCmd.AddCommand(keyImportCmd)
	keyCmd.AddCommand(keyRegenCmd)
}

// Command implementations

// initSyncer initializes syncer instance
func initSyncer() (*sync.Syncer, error) {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("no configuration found. Run 'opencode-sync setup' first")
	}

	// Get paths
	p, err := paths.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get paths: %w", err)
	}

	// Initialize git repo
	repo := git.NewBuiltinGit(p.SyncRepoDir())
	if err := repo.Open(); err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Create syncer
	syncer := sync.New(cfg, p, repo)

	// Initialize encryption if enabled
	if cfg.Encryption.Enabled {
		keyFile := p.KeyFile()

		// Check if key file exists
		if _, err := os.Stat(keyFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("encryption key not found at %s. Run 'opencode-sync setup' first", keyFile)
		}

		// Load private key
		privateKey, err := crypto.LoadKeyFromFile(keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load encryption key: %w", err)
		}

		// Initialize encryption
		enc, err := crypto.NewAgeEncryption(privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize encryption: %w", err)
		}

		syncer.SetEncryption(enc)
	}

	return syncer, nil
}

func runSync() error {
	ui.Info("Syncing...")

	// Pull first
	if err := runPull(); err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}

	// Then push
	if err := runPush(); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	ui.Success("Sync complete!")
	return nil
}

func runPush() error {
	syncer, err := initSyncer()
	if err != nil {
		return err
	}

	// Copy OpenCode config to repo
	if err := ui.SpinnerWithResult("Copying config files to sync repo", func() error {
		return syncer.CopyToRepo()
	}); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	// Get repo instance
	p, _ := paths.Get()
	repo := git.NewBuiltinGit(p.SyncRepoDir())
	if err := repo.Open(); err != nil {
		return err
	}

	// Check if there are changes
	hasChanges, err := repo.HasChanges()
	if err != nil {
		return fmt.Errorf("failed to check for changes: %w", err)
	}

	if !hasChanges {
		ui.Info("No changes to push")
		return nil
	}

	// Stage all changes
	if err := repo.AddAll(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Commit
	commitMsg := fmt.Sprintf("Sync from %s at %s", getHostname(), time.Now().Format("2006-01-02 15:04:05"))
	if err := repo.Commit(commitMsg); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Push
	if err := ui.SpinnerWithResult("Pushing to remote", func() error {
		return repo.Push()
	}); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

func runPull() error {
	syncer, err := initSyncer()
	if err != nil {
		return err
	}

	// Get repo instance
	p, _ := paths.Get()
	repo := git.NewBuiltinGit(p.SyncRepoDir())
	if err := repo.Open(); err != nil {
		return err
	}

	// Check for local changes before pulling
	hasChanges, err := repo.HasChanges()
	if err != nil {
		return fmt.Errorf("failed to check for changes: %w", err)
	}

	if hasChanges {
		return fmt.Errorf("local changes detected. Commit or discard them before pulling")
	}

	// Pull from remote
	if err := ui.SpinnerWithResult("Fetching from remote", func() error {
		return repo.Pull()
	}); err != nil {
		if conflictErr, ok := err.(*git.ConflictError); ok {
			return fmt.Errorf("merge conflict detected in %d file(s). Please resolve manually", len(conflictErr.Files))
		}
		return fmt.Errorf("failed to pull: %w", err)
	}

	// Copy from repo to OpenCode config
	if err := ui.SpinnerWithResult("Applying changes to OpenCode config", func() error {
		return syncer.CopyFromRepo()
	}); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	return nil
}

func runStatus() error {
	ui.Info("Checking status...")

	syncer, err := initSyncer()
	if err != nil {
		return err
	}

	state, err := syncer.GetState()
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}

	fmt.Println("\nSync Status:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if state.IsClean {
		fmt.Println("✓ Working directory is clean")
	} else {
		fmt.Println("✗ Working directory has changes")
	}

	if state.HasLocalChanges {
		fmt.Printf("\n%d file(s) modified locally\n", len(state.LocalFiles))
	} else {
		fmt.Println("No local changes")
	}

	if len(state.ConflictFiles) > 0 {
		fmt.Printf("\n⚠ %d conflict(s) detected:\n", len(state.ConflictFiles))
		for _, file := range state.ConflictFiles {
			fmt.Printf("  - %s\n", file)
		}
	}

	return nil
}

func runDiff() error {
	ui.Info("Checking differences...")

	p, err := paths.Get()
	if err != nil {
		return err
	}

	repo := git.NewBuiltinGit(p.SyncRepoDir())
	if err := repo.Open(); err != nil {
		return err
	}

	diff, err := repo.Diff()
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}

	if diff == "" {
		fmt.Println("No differences")
		return nil
	}

	fmt.Println("\nDifferences:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(diff)

	return nil
}

func runDoctor() error {
	ui.Info("Running diagnostics...")

	p, err := paths.Get()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	fmt.Println("\nDiagnostics:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	issues := []string{}
	suggestions := []string{}

	// Check OpenCode installation
	fmt.Print("OpenCode config directory... ")
	if _, err := os.Stat(p.OpenCodeConfigDir); err == nil {
		fmt.Println("✓")
	} else {
		fmt.Println("✗ not found")
		issues = append(issues, "OpenCode config directory not found")
		suggestions = append(suggestions, fmt.Sprintf("Install OpenCode or check path: %s", p.OpenCodeConfigDir))
	}

	// Check OpenCode data directory
	fmt.Print("OpenCode data directory... ")
	if _, err := os.Stat(p.OpenCodeDataDir); err == nil {
		fmt.Println("✓")
	} else {
		fmt.Println("✗ not found")
		issues = append(issues, "OpenCode data directory not found")
	}

	// Check sync config
	fmt.Print("opencode-sync config... ")
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		fmt.Println("✗ not found or invalid")
		issues = append(issues, "Configuration not found")
		suggestions = append(suggestions, "Run 'opencode-sync setup' to configure")
	} else {
		fmt.Println("✓")

		// Check encryption key if encryption enabled
		if cfg.Encryption.Enabled {
			fmt.Print("Encryption key... ")
			keyFile := p.KeyFile()
			if _, err := os.Stat(keyFile); err == nil {
				// Try to load the key to verify it's valid
				if privateKey, err := crypto.LoadKeyFromFile(keyFile); err == nil {
					// Try to create encryption instance to verify it works
					if _, err := crypto.NewAgeEncryption(privateKey); err == nil {
						fmt.Println("✓")
					} else {
						fmt.Println("✗ invalid key")
						issues = append(issues, "Encryption key is invalid")
						suggestions = append(suggestions, "Regenerate key or check file corruption")
					}
				} else {
					fmt.Println("✗ failed to load")
					issues = append(issues, "Failed to load encryption key")
					suggestions = append(suggestions, fmt.Sprintf("Check file permissions: %s", keyFile))
				}
			} else {
				fmt.Println("✗ not found")
				issues = append(issues, "Encryption key file not found")
				suggestions = append(suggestions, "Run 'opencode-sync setup' to regenerate key")
			}
		}
	}

	// Check sync repo directory
	fmt.Print("Sync repository directory... ")
	if _, err := os.Stat(p.SyncRepoDir()); err == nil {
		fmt.Println("✓")
	} else {
		fmt.Println("✗ not found")
		issues = append(issues, "Sync repository directory not found")
		suggestions = append(suggestions, "Run 'opencode-sync init' or 'opencode-sync clone' to set up repository")
	}

	// Check git repo
	if cfg != nil {
		fmt.Print("Git repository... ")
		repo := git.NewBuiltinGit(p.SyncRepoDir())
		if err := repo.Open(); err == nil {
			fmt.Println("✓")

			// Check remote
			fmt.Print("Git remote... ")
			remoteURL, err := repo.GetRemoteURL("origin")
			if err == nil {
				fmt.Printf("✓ (%s)\n", remoteURL)

				// Check remote connectivity
				fmt.Print("Remote connectivity... ")
				// Try to fetch to verify connectivity (dry-run)
				if err := repo.Fetch(); err == nil {
					fmt.Println("✓")
				} else {
					fmt.Println("✗ failed to connect")
					issues = append(issues, "Cannot connect to remote")
					suggestions = append(suggestions, "Check network connection and authentication")
				}
			} else {
				fmt.Println("✗ not configured")
				issues = append(issues, "Git remote not configured")
				suggestions = append(suggestions, "Add remote: git remote add origin <url>")
			}

			// Check branch
			fmt.Print("Current branch... ")
			branch, err := repo.GetBranch()
			if err == nil {
				fmt.Println(branch)
			} else {
				fmt.Println("✗ failed to determine")
			}

			// Check for uncommitted changes
			fmt.Print("Working directory... ")
			hasChanges, err := repo.HasChanges()
			if err == nil {
				if !hasChanges {
					fmt.Println("✓ clean")
				} else {
					fmt.Println("⚠ has uncommitted changes")
					suggestions = append(suggestions, "Run 'opencode-sync push' to sync changes")
				}
			} else {
				fmt.Println("✗ failed to check")
			}
		} else {
			fmt.Println("✗ failed to open")
			issues = append(issues, "Git repository is not initialized or corrupted")
			suggestions = append(suggestions, "Run 'opencode-sync init' to reinitialize")
		}
	}

	// Summary
	fmt.Println()
	if len(issues) == 0 {
		ui.Success("All checks passed! Your setup looks good.")
	} else {
		ui.Warn(fmt.Sprintf("Found %d issue(s):", len(issues)))
		for i, issue := range issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}

		if len(suggestions) > 0 {
			fmt.Println()
			ui.Info("Suggested fixes:")
			for i, suggestion := range suggestions {
				fmt.Printf("  %d. %s\n", i+1, suggestion)
			}
		}
	}

	return nil
}

func runConfigShow() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		ui.Warn("No configuration found. Run 'opencode-sync setup' first.")
		return nil
	}

	// Pretty print the config
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println("\nCurrent Configuration:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(string(data))

	return nil
}

func runConfigPath() error {
	p, err := paths.Get()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	fmt.Println(p.ConfigFile())
	return nil
}

func runConfigEdit() error {
	p, err := paths.Get()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	configFile := p.ConfigFile()

	// Check if config exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		ui.Warn("No configuration found. Run 'opencode-sync setup' first.")
		return nil
	}

	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Default editors
		if _, err := os.Stat("/usr/bin/nano"); err == nil {
			editor = "nano"
		} else if _, err := os.Stat("/usr/bin/vim"); err == nil {
			editor = "vim"
		} else if _, err := os.Stat("/usr/bin/vi"); err == nil {
			editor = "vi"
		} else {
			return fmt.Errorf("no editor found. Set $EDITOR or $VISUAL environment variable")
		}
	}

	ui.Info(fmt.Sprintf("Opening %s in %s...", configFile, editor))

	// Execute editor
	cmd := exec.Command(editor, configFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	ui.Success("Configuration updated")
	return nil
}

func runConfigSet(key, value string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		return fmt.Errorf("no configuration found. Run 'opencode-sync setup' first")
	}

	// Parse key and set value
	switch key {
	case "repo.url":
		cfg.Repo.URL = value
	case "repo.branch":
		cfg.Repo.Branch = value
	case "encryption.enabled":
		enabled := value == "true" || value == "yes" || value == "1"
		cfg.Encryption.Enabled = enabled
	case "encryption.keyFile":
		cfg.Encryption.KeyFile = value
	case "sync.includeAuth":
		enabled := value == "true" || value == "yes" || value == "1"
		cfg.Sync.IncludeAuth = enabled
	case "sync.includeMcpAuth":
		enabled := value == "true" || value == "yes" || value == "1"
		cfg.Sync.IncludeMcpAuth = enabled
	default:
		return fmt.Errorf("unknown config key: %s. Valid keys: repo.url, repo.branch, encryption.enabled, encryption.keyFile, sync.includeAuth, sync.includeMcpAuth", key)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Save config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	ui.Success(fmt.Sprintf("Set %s = %s", key, value))
	return nil
}

func runInit() error {
	ui.Info("Initializing sync repository...")

	// Load config
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		return fmt.Errorf("no configuration found. Run 'opencode-sync setup' first")
	}

	// Get paths
	p, err := paths.Get()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	// Ensure directories exist
	if err := p.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	repoDir := p.SyncRepoDir()

	// Check if repo already exists
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err == nil {
		return fmt.Errorf("repository already initialized at %s", repoDir)
	}

	// Initialize git repository
	repo := git.NewBuiltinGit(repoDir)
	if err := ui.SpinnerWithResult("Creating Git repository", func() error {
		return repo.Init()
	}); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Add remote if configured
	if cfg.Repo.URL != "" {
		if err := ui.SpinnerWithResult(fmt.Sprintf("Adding remote: %s", cfg.Repo.URL), func() error {
			return repo.AddRemote("origin", cfg.Repo.URL)
		}); err != nil {
			return fmt.Errorf("failed to add remote: %w", err)
		}
	}

	// Create syncer and copy OpenCode configs
	ui.Info("Copying OpenCode configurations...")
	syncer := sync.New(cfg, p, repo)

	// Initialize encryption if enabled
	if cfg.Encryption.Enabled {
		keyFile := p.KeyFile()
		if _, err := os.Stat(keyFile); os.IsNotExist(err) {
			return fmt.Errorf("encryption key not found. Run 'opencode-sync setup' first")
		}

		privateKey, err := crypto.LoadKeyFromFile(keyFile)
		if err != nil {
			return fmt.Errorf("failed to load encryption key: %w", err)
		}

		enc, err := crypto.NewAgeEncryption(privateKey)
		if err != nil {
			return fmt.Errorf("failed to initialize encryption: %w", err)
		}

		syncer.SetEncryption(enc)
	}

	if err := ui.SpinnerWithResult("Copying OpenCode configurations", func() error {
		return syncer.CopyToRepo()
	}); err != nil {
		return fmt.Errorf("failed to copy configs: %w", err)
	}

	// Stage all files and create initial commit
	if err := ui.SpinnerWithResult("Creating initial commit", func() error {
		if err := repo.AddAll(); err != nil {
			return err
		}
		commitMsg := fmt.Sprintf("Initial commit from %s", getHostname())
		return repo.Commit(commitMsg)
	}); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	ui.Success("Repository initialized!")

	// Suggest next steps
	fmt.Println()
	if cfg.Repo.URL != "" {
		ui.Info("Next step: Push to remote with 'opencode-sync push'")
	} else {
		ui.Info("Add a remote URL with: opencode-sync config set repo.url <url>")
	}

	return nil
}

func runLink(repoURL string) error {
	ui.Info(fmt.Sprintf("Linking local configs to remote: %s", repoURL))

	// Load config
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		return fmt.Errorf("no configuration found. Run 'opencode-sync setup' first")
	}

	// Get paths
	p, err := paths.Get()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	// Ensure directories exist
	if err := p.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	repoDir := p.SyncRepoDir()

	// Check if repo already exists
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err == nil {
		return fmt.Errorf("repository already exists at %s. Use 'opencode-sync push' to sync, or remove the directory first", repoDir)
	}

	// Initialize git repository
	repo := git.NewBuiltinGit(repoDir)
	if err := ui.SpinnerWithResult("Creating Git repository", func() error {
		return repo.Init()
	}); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Add remote
	if err := ui.SpinnerWithResult(fmt.Sprintf("Adding remote: %s", repoURL), func() error {
		return repo.AddRemote("origin", repoURL)
	}); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	// Update config with remote URL
	cfg.Repo.URL = repoURL
	if err := config.Save(cfg); err != nil {
		ui.Warn("Failed to update config with remote URL, but link will continue")
	}

	// Create syncer and copy OpenCode configs
	syncer := sync.New(cfg, p, repo)

	// Initialize encryption if enabled
	if cfg.Encryption.Enabled {
		keyFile := p.KeyFile()
		if _, err := os.Stat(keyFile); os.IsNotExist(err) {
			return fmt.Errorf("encryption key not found. Run 'opencode-sync setup' first")
		}

		privateKey, err := crypto.LoadKeyFromFile(keyFile)
		if err != nil {
			return fmt.Errorf("failed to load encryption key: %w", err)
		}

		enc, err := crypto.NewAgeEncryption(privateKey)
		if err != nil {
			return fmt.Errorf("failed to initialize encryption: %w", err)
		}

		syncer.SetEncryption(enc)
	}

	if err := ui.SpinnerWithResult("Copying OpenCode configurations", func() error {
		return syncer.CopyToRepo()
	}); err != nil {
		return fmt.Errorf("failed to copy configs: %w", err)
	}

	// Stage all files and create initial commit
	if err := ui.SpinnerWithResult("Creating initial commit", func() error {
		if err := repo.AddAll(); err != nil {
			return err
		}
		commitMsg := fmt.Sprintf("Link from %s at %s", getHostname(), time.Now().Format("2006-01-02 15:04:05"))
		return repo.Commit(commitMsg)
	}); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Force push to overwrite remote
	ui.Warn("This will OVERWRITE the remote repository with your local configs")
	confirmed, err := ui.Confirm("Force push to remote?", "This will replace all remote content")
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !confirmed {
		ui.Info("Link cancelled. Local repository created but not pushed.")
		ui.Info("You can manually push later with: opencode-sync push")
		return nil
	}

	if err := ui.SpinnerWithResult("Force pushing to remote", func() error {
		return repo.ForcePush()
	}); err != nil {
		return fmt.Errorf("failed to force push: %w", err)
	}

	ui.Success("Successfully linked local configs to remote!")
	fmt.Println()
	ui.Info("Your local OpenCode configs are now synced to the remote")
	ui.Info("Use 'opencode-sync sync' to keep them in sync")

	return nil
}

func runClone(repoURL string) error {
	// Load or prompt for repository URL
	if repoURL == "" {
		cfg, err := config.Load()
		if err == nil && cfg != nil && cfg.Repo.URL != "" {
			repoURL = cfg.Repo.URL
		} else {
			return fmt.Errorf("no repository URL provided. Run 'opencode-sync clone <url>' or configure via 'opencode-sync setup'")
		}
	}

	ui.Info(fmt.Sprintf("Cloning repository from %s...", repoURL))

	// Get paths
	p, err := paths.Get()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	// Ensure directories exist
	if err := p.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	repoDir := p.SyncRepoDir()

	// Check if repo already exists
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err == nil {
		return fmt.Errorf("repository already exists at %s. Use 'opencode-sync pull' to update", repoDir)
	}

	// Clone repository
	repo := git.NewBuiltinGit(repoDir)
	if err := ui.SpinnerWithResult(fmt.Sprintf("Cloning repository from %s", repoURL), func() error {
		return repo.Clone(repoURL)
	}); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Load config or create minimal one
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		// Create minimal config
		cfg = config.Default()
		cfg.Repo.URL = repoURL
		if err := config.Save(cfg); err != nil {
			ui.Warn("Failed to save config, but clone succeeded")
		}
	}

	// Create syncer and copy to OpenCode
	ui.Info("Applying configurations to OpenCode...")
	syncer := sync.New(cfg, p, repo)

	// Initialize encryption if enabled
	if cfg.Encryption.Enabled {
		keyFile := p.KeyFile()
		if _, err := os.Stat(keyFile); err == nil {
			privateKey, err := crypto.LoadKeyFromFile(keyFile)
			if err == nil {
				enc, err := crypto.NewAgeEncryption(privateKey)
				if err == nil {
					syncer.SetEncryption(enc)
				}
			}
		} else {
			ui.Warn("Encryption enabled but key file not found. Encrypted files will not be decrypted.")
		}
	}

	if err := ui.SpinnerWithResult("Applying configurations to OpenCode", func() error {
		return syncer.CopyFromRepo()
	}); err != nil {
		return fmt.Errorf("failed to copy configs: %w", err)
	}
	fmt.Println()
	ui.Info("Your OpenCode is now synced. Use 'opencode-sync sync' to keep it up to date.")

	return nil
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func runKeyExport() error {
	p, err := paths.Get()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	keyFile := p.KeyFile()
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return fmt.Errorf("no encryption key found. Run 'opencode-sync setup' with encryption enabled first")
	}

	privateKey, err := crypto.LoadKeyFromFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to load key: %w", err)
	}

	ui.Warn("PRIVATE KEY - Store securely! Anyone with this key can decrypt your auth tokens.")
	fmt.Println()
	fmt.Println(privateKey)
	fmt.Println()
	ui.Info("Copy this key to your password manager or secure storage.")
	ui.Info("Use 'opencode-sync key import <key>' on other machines.")

	return nil
}

func runKeyImport(key string) error {
	if _, err := crypto.NewAgeEncryption(key); err != nil {
		return fmt.Errorf("invalid key format: %w", err)
	}

	p, err := paths.Get()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	if err := p.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	keyFile := p.KeyFile()
	if _, err := os.Stat(keyFile); err == nil {
		confirmed, err := ui.Confirm("Key already exists. Overwrite?", "This will replace your existing encryption key")
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Info("Import cancelled")
			return nil
		}
	}

	if err := crypto.SaveKeyToFile(key, keyFile); err != nil {
		return fmt.Errorf("failed to save key: %w", err)
	}

	cfg, err := config.Load()
	if err == nil && cfg != nil {
		cfg.Encryption.Enabled = true
		if err := config.Save(cfg); err != nil {
			ui.Warn("Key saved but failed to update config. Run: opencode-sync config set encryption.enabled true")
		}
	}

	ui.Success(fmt.Sprintf("Key imported to: %s", keyFile))
	ui.Info("You can now pull encrypted data from your repo.")

	return nil
}

func runKeyRegen() error {
	ui.Warn("WARNING: Regenerating your key will make previously encrypted data unrecoverable!")
	ui.Warn("Only proceed if you've lost your key and need to start fresh.")
	fmt.Println()

	confirmed, err := ui.Confirm("Regenerate encryption key?", "Previously encrypted auth tokens will be lost")
	if err != nil {
		return err
	}
	if !confirmed {
		ui.Info("Cancelled")
		return nil
	}

	p, err := paths.Get()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	if err := p.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	keyPair, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	keyFile := p.KeyFile()
	if err := crypto.SaveKeyToFile(keyPair.PrivateKey, keyFile); err != nil {
		return fmt.Errorf("failed to save key: %w", err)
	}

	cfg, err := config.Load()
	if err == nil && cfg != nil {
		cfg.Encryption.Enabled = true
		if err := config.Save(cfg); err != nil {
			ui.Warn("Key saved but failed to update config")
		}
	}

	ui.Success(fmt.Sprintf("New encryption key saved to: %s", keyFile))
	fmt.Println()
	ui.Warn("IMPORTANT: Back up your new key!")
	ui.Info("Run 'opencode-sync key export' to view it for backup.")

	return nil
}

func runRebind(newURL string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		return fmt.Errorf("no configuration found. Run 'opencode-sync setup' first")
	}

	p, err := paths.Get()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	repoDir := p.SyncRepoDir()
	gitDir := filepath.Join(repoDir, ".git")

	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("no repository found. Run 'opencode-sync init' or 'opencode-sync clone' first")
	}

	oldURL := cfg.Repo.URL
	if oldURL == newURL {
		ui.Info("URL is already set to: " + newURL)
		return nil
	}

	ui.Info(fmt.Sprintf("Changing remote URL from: %s", oldURL))
	ui.Info(fmt.Sprintf("                     to: %s", newURL))

	if err := runGitCommand(repoDir, "remote", "set-url", "origin", newURL); err != nil {
		return fmt.Errorf("failed to update git remote: %w", err)
	}

	cfg.Repo.URL = newURL
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	ui.Success("Repository URL updated!")
	ui.Info("Run 'opencode-sync sync' to sync with the new remote.")

	return nil
}

func runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
