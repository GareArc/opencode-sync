//go:build windows

package paths

import (
	"os"
	"path/filepath"
)

func getPlatformPaths() (*Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	appData := os.Getenv("APPDATA")
	localAppData := os.Getenv("LOCALAPPDATA")

	return &Paths{
		ConfigDir:         filepath.Join(appData, "opencode-sync"),
		DataDir:           filepath.Join(localAppData, "opencode-sync"),
		OpenCodeConfigDir: filepath.Join(appData, "opencode"),
		OpenCodeDataDir:   filepath.Join(localAppData, "opencode"),
		ClaudeSkillsDir:   filepath.Join(home, ".claude", "skills"),
	}, nil
}
