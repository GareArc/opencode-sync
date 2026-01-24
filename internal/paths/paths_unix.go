//go:build unix

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

	// XDG Base Directory Specification
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(home, ".config")
	}

	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(home, ".local", "share")
	}

	return &Paths{
		ConfigDir:         filepath.Join(configHome, "opencode-sync"),
		DataDir:           filepath.Join(dataHome, "opencode-sync"),
		OpenCodeConfigDir: filepath.Join(configHome, "opencode"),
		OpenCodeDataDir:   filepath.Join(dataHome, "opencode"),
		ClaudeSkillsDir:   filepath.Join(home, ".claude", "skills"),
	}, nil
}
