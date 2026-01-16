package paths

import (
	"os"
	"path/filepath"
)

// Paths holds all relevant paths for opencode-sync
type Paths struct {
	// ConfigDir is where opencode-sync stores its config
	ConfigDir string

	// DataDir is where opencode-sync stores data (sync repo, etc.)
	DataDir string

	// OpenCodeConfigDir is where OpenCode stores its config
	OpenCodeConfigDir string

	// OpenCodeDataDir is where OpenCode stores its data (auth.json, etc.)
	OpenCodeDataDir string
}

// Get returns the paths for the current platform
func Get() (*Paths, error) {
	return getPlatformPaths()
}

// SyncRepoDir returns the path to the sync repository
func (p *Paths) SyncRepoDir() string {
	return filepath.Join(p.DataDir, "repo")
}

// ConfigFile returns the path to the opencode-sync config file
func (p *Paths) ConfigFile() string {
	return filepath.Join(p.ConfigDir, "config.json")
}

// KeyFile returns the path to the age encryption key
func (p *Paths) KeyFile() string {
	return filepath.Join(p.ConfigDir, "age.key")
}

// OpenCodeConfigFile returns the path to the main OpenCode config
func (p *Paths) OpenCodeConfigFile() string {
	// Try .jsonc first, then .json
	jsonc := filepath.Join(p.OpenCodeConfigDir, "opencode.jsonc")
	if _, err := os.Stat(jsonc); err == nil {
		return jsonc
	}
	return filepath.Join(p.OpenCodeConfigDir, "opencode.json")
}

// OpenCodeAuthFile returns the path to OpenCode's auth.json
func (p *Paths) OpenCodeAuthFile() string {
	return filepath.Join(p.OpenCodeDataDir, "auth.json")
}

// OpenCodeMcpAuthFile returns the path to OpenCode's mcp-auth.json
func (p *Paths) OpenCodeMcpAuthFile() string {
	return filepath.Join(p.OpenCodeDataDir, "mcp-auth.json")
}

// EnsureDirs creates all necessary directories
func (p *Paths) EnsureDirs() error {
	dirs := []string{
		p.ConfigDir,
		p.DataDir,
		p.SyncRepoDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// SyncableOpenCodePaths returns paths that should be synced from OpenCode config
func (p *Paths) SyncableOpenCodePaths() []string {
	return []string{
		filepath.Join(p.OpenCodeConfigDir, "opencode.json"),
		filepath.Join(p.OpenCodeConfigDir, "opencode.jsonc"),
		filepath.Join(p.OpenCodeConfigDir, "AGENTS.md"),
		filepath.Join(p.OpenCodeConfigDir, "agent"),
		filepath.Join(p.OpenCodeConfigDir, "command"),
		filepath.Join(p.OpenCodeConfigDir, "skill"),
		filepath.Join(p.OpenCodeConfigDir, "mode"),
		filepath.Join(p.OpenCodeConfigDir, "themes"),
		filepath.Join(p.OpenCodeConfigDir, "plugin"),
	}
}
