package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/GareArc/opencode-sync/internal/config"
)

var (
	// Styles
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

// Success prints a success message
func Success(msg string) {
	fmt.Println(successStyle.Render("✓ " + msg))
}

// Error prints an error message
func Error(msg string) {
	fmt.Println(errorStyle.Render("✗ " + msg))
}

// Info prints an info message
func Info(msg string) {
	fmt.Println(infoStyle.Render("→ " + msg))
}

// Warn prints a warning message
func Warn(msg string) {
	fmt.Println(warnStyle.Render("⚠ " + msg))
}

// MainMenu shows the main interactive menu
func MainMenu() (string, error) {
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

// SetupWizard runs the first-time setup wizard
func SetupWizard() (*config.Config, error) {
	var (
		repoURL          string
		enableEncryption bool
		includeAuth      bool
	)

	cfg := config.Default()

	// Step 1: Repository URL
	form1 := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter your Git repository URL").
				Description("This is where your OpenCode config will be synced").
				Placeholder("git@github.com:username/opencode-config.git").
				Value(&repoURL),
		),
	)

	if err := form1.Run(); err != nil {
		return nil, err
	}

	cfg.Repo.URL = repoURL

	// Step 2: Encryption
	form2 := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable encryption for secrets?").
				Description("Recommended: Encrypts sensitive data before syncing").
				Affirmative("Yes (recommended)").
				Negative("No").
				Value(&enableEncryption),
		),
	)

	if err := form2.Run(); err != nil {
		return nil, err
	}

	cfg.Encryption.Enabled = enableEncryption

	// Step 3: Auth sync (only if encryption enabled)
	if enableEncryption {
		form3 := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Sync OAuth credentials (auth.json)?").
					Description("⚠️  Warning: Requires secure key transfer to new machines.\n" +
						"   If enabled, you won't need to re-authenticate on each device.\n" +
						"   If disabled, you'll authenticate separately on each machine.").
					Affirmative("Yes (encrypted)").
					Negative("No (re-authenticate each machine)").
					Value(&includeAuth),
			),
		)

		if err := form3.Run(); err != nil {
			return nil, err
		}

		cfg.Sync.IncludeAuth = includeAuth
	}

	return cfg, nil
}

// Confirm shows a yes/no confirmation prompt
func Confirm(title string, description string) (bool, error) {
	var result bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Description(description).
				Affirmative("Yes").
				Negative("No").
				Value(&result),
		),
	)

	err := form.Run()
	return result, err
}

// Input prompts for text input
func Input(title string, placeholder string) (string, error) {
	var result string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(title).
				Placeholder(placeholder).
				Value(&result),
		),
	)

	err := form.Run()
	return result, err
}

// Spinner runs a function with a spinner animation
func Spinner(message string, fn func() error) error {
	var err error

	action := func() {
		err = fn()
	}

	if err := spinner.New().
		Title(message).
		Action(action).
		Run(); err != nil {
		return err
	}

	return err
}

// SpinnerWithResult runs a function with a spinner and shows success/error
func SpinnerWithResult(message string, fn func() error) error {
	start := time.Now()
	err := Spinner(message, fn)
	duration := time.Since(start)

	if err != nil {
		Error(fmt.Sprintf("%s (failed after %v)", message, duration))
		return err
	}

	Success(fmt.Sprintf("%s (done in %v)", message, duration))
	return nil
}
