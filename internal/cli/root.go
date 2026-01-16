package cli

import (
	"fmt"

	"github.com/GareArc/opencode-sync/internal/config"
	"github.com/GareArc/opencode-sync/internal/crypto"
	"github.com/GareArc/opencode-sync/internal/paths"
	"github.com/GareArc/opencode-sync/internal/ui"
	"github.com/spf13/cobra"
)

var (
	// Version info
	version = "dev"
	commit  = "none"
	date    = "unknown"

	// Global flags
	verbose  bool
	dryRun   bool
	noPrompt bool
	cfgFile  string
)

// SetVersionInfo sets version information from main
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "opencode-sync",
	Short: "Sync OpenCode configurations across machines",
	Long: `opencode-sync is a CLI tool to sync your OpenCode configurations 
across multiple machines via Git, with optional encryption for secrets.

Run without arguments for interactive mode, or use subcommands for scripting.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if config exists
		cfg, err := config.Load()
		if err != nil || cfg == nil {
			// No config - run setup wizard
			fmt.Println("Welcome to opencode-sync!")
			fmt.Println()
			return runSetupWizard()
		}

		// Config exists - show main menu
		return runInteractiveMenu(cfg)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be done without making changes")
	rootCmd.PersistentFlags().BoolVar(&noPrompt, "no-prompt", false, "disable interactive prompts (for scripting)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/opencode-sync/config.json)")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(linkCmd)
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(keyCmd)
	rootCmd.AddCommand(rebindCmd)
}

// runSetupWizard runs the first-time setup wizard
func runSetupWizard() error {
	result, err := ui.SetupWizard()
	if err != nil {
		return err
	}

	// Generate encryption keys if encryption is enabled
	if result.Encryption.Enabled {
		if err := generateAndSaveKeys(); err != nil {
			return fmt.Errorf("failed to generate encryption keys: %w", err)
		}
	}

	// Save config
	if err := config.Save(result); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	ui.Success("Setup complete! Your config is ready to sync.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  Run 'opencode-sync' to open the main menu")
	fmt.Println("  Or run 'opencode-sync sync' to sync now")

	return nil
}

// generateAndSaveKeys generates an encryption key pair and saves it
func generateAndSaveKeys() error {
	ui.Info("Generating encryption keys...")

	// Get paths
	p, err := paths.Get()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	// Ensure config directory exists
	if err := p.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Generate keypair
	keyPair, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("failed to generate keypair: %w", err)
	}

	// Save private key
	keyFile := p.KeyFile()
	if err := crypto.SaveKeyToFile(keyPair.PrivateKey, keyFile); err != nil {
		return fmt.Errorf("failed to save key: %w", err)
	}

	ui.Success(fmt.Sprintf("Encryption key saved to: %s", keyFile))
	fmt.Println()
	ui.Warn("IMPORTANT: Back up your private key! Without it, encrypted data is unrecoverable.")
	fmt.Println()
	ui.Info("Run 'opencode-sync key export' to view your key for backup.")
	ui.Info("Store it securely (e.g., password manager).")

	return nil
}

func runKeyMenu() error {
	for {
		choice, err := ui.KeyMenu()
		if err != nil {
			return err
		}

		switch choice {
		case "export":
			if err := runKeyExport(); err != nil {
				ui.Error(err.Error())
			}
		case "import":
			key, err := ui.Input("Paste your private key", "AGE-SECRET-KEY-1...")
			if err != nil {
				ui.Error(err.Error())
				continue
			}
			if key == "" {
				ui.Warn("No key provided, cancelled")
				continue
			}
			if err := runKeyImport(key); err != nil {
				ui.Error(err.Error())
			}
		case "regen":
			if err := runKeyRegen(); err != nil {
				ui.Error(err.Error())
			}
		case "back":
			return nil
		}

		fmt.Println()
	}
}

func runInteractiveMenu(cfg *config.Config) error {
	for {
		choice, err := ui.MainMenu()
		if err != nil {
			return err
		}

		switch choice {
		case "sync":
			if err := runSync(); err != nil {
				ui.Error(err.Error())
			}
		case "pull":
			if err := runPull(); err != nil {
				ui.Error(err.Error())
			}
		case "push":
			if err := runPush(); err != nil {
				ui.Error(err.Error())
			}
		case "status":
			if err := runStatus(); err != nil {
				ui.Error(err.Error())
			}
		case "diff":
			if err := runDiff(); err != nil {
				ui.Error(err.Error())
			}
		case "config":
			if err := runConfigShow(); err != nil {
				ui.Error(err.Error())
			}
		case "init":
			if err := runInit(); err != nil {
				ui.Error(err.Error())
			}
		case "link":
			repoURL, err := ui.Input("Enter repository URL to link", "git@github.com:username/repo.git")
			if err != nil {
				ui.Error(err.Error())
				continue
			}
			if repoURL == "" {
				ui.Warn("No URL provided, cancelled")
				continue
			}
			if err := runLink(repoURL); err != nil {
				ui.Error(err.Error())
			}
		case "clone":
			repoURL, err := ui.Input("Enter repository URL to clone", "git@github.com:username/repo.git")
			if err != nil {
				ui.Error(err.Error())
				continue
			}
			if err := runClone(repoURL); err != nil {
				ui.Error(err.Error())
			}
		case "doctor":
			if err := runDoctor(); err != nil {
				ui.Error(err.Error())
			}
		case "key":
			if err := runKeyMenu(); err != nil {
				ui.Error(err.Error())
			}
		case "rebind":
			newURL, err := ui.Input("Enter new repository URL", "git@github.com:username/repo.git")
			if err != nil {
				ui.Error(err.Error())
				continue
			}
			if newURL == "" {
				ui.Warn("No URL provided, cancelled")
				continue
			}
			if err := runRebind(newURL); err != nil {
				ui.Error(err.Error())
			}
		case "exit":
			return nil
		case "":
			continue
		}

		fmt.Println()
	}
}
