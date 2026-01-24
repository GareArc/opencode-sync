package sync

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GareArc/opencode-sync/internal/config"
	"github.com/GareArc/opencode-sync/internal/crypto"
	"github.com/GareArc/opencode-sync/internal/git"
	"github.com/GareArc/opencode-sync/internal/paths"
)

// Syncer handles synchronization between OpenCode config and sync repo
type Syncer struct {
	cfg        *config.Config
	paths      *paths.Paths
	repo       git.Repository
	encryption crypto.Encryption
}

// New creates a new Syncer instance
func New(cfg *config.Config, p *paths.Paths, repo git.Repository) *Syncer {
	return &Syncer{
		cfg:        cfg,
		paths:      p,
		repo:       repo,
		encryption: nil, // Will be set if encryption is enabled
	}
}

// SetEncryption sets the encryption instance
func (s *Syncer) SetEncryption(enc crypto.Encryption) {
	s.encryption = enc
}

// SyncState represents the current sync state
type SyncState struct {
	IsClean          bool
	HasLocalChanges  bool
	HasRemoteChanges bool
	LocalFiles       []FileInfo
	ConflictFiles    []string
	LastSyncTime     time.Time
}

// FileInfo represents information about a file
type FileInfo struct {
	Path       string
	RelPath    string
	Size       int64
	ModTime    time.Time
	Hash       string
	IsNew      bool
	IsModified bool
	IsDeleted  bool
}

// GetState returns the current sync state
func (s *Syncer) GetState() (*SyncState, error) {
	state := &SyncState{
		LocalFiles:    []FileInfo{},
		ConflictFiles: []string{},
	}

	// Check if repo is clean
	isClean, err := s.repo.IsClean()
	if err != nil {
		return nil, fmt.Errorf("failed to check repo status: %w", err)
	}
	state.IsClean = isClean

	// Check for local changes
	hasChanges, err := s.repo.HasChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to check for changes: %w", err)
	}
	state.HasLocalChanges = hasChanges

	// Get file info
	files, err := s.getSyncableFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get syncable files: %w", err)
	}
	state.LocalFiles = files

	return state, nil
}

// CopyToRepo copies OpenCode config files to the sync repository
func (s *Syncer) CopyToRepo() error {
	syncablePaths := s.paths.SyncableOpenCodePaths()

	for _, srcPath := range syncablePaths {
		// Check if path exists
		info, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			continue // Skip non-existent paths
		}
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", srcPath, err)
		}

		var relPath string
		var dstPath string

		if srcPath == s.paths.ClaudeSkillsDir {
			dstPath = filepath.Join(s.paths.SyncRepoDir(), "claude-skills")
		} else {
			relPath, err = filepath.Rel(s.paths.OpenCodeConfigDir, srcPath)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}
			dstPath = filepath.Join(s.paths.SyncRepoDir(), relPath)
		}

		if info.IsDir() {
			// Copy directory recursively
			if err := s.copyDir(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy directory %s: %w", srcPath, err)
			}
		} else {
			// Copy file
			if err := s.copyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", srcPath, err)
			}
		}
	}

	// Handle auth.json if enabled
	if s.cfg.Sync.IncludeAuth {
		if s.encryption == nil {
			return fmt.Errorf("includeAuth requires encryption to be enabled")
		}

		authSrc := s.paths.OpenCodeAuthFile()
		if _, err := os.Stat(authSrc); err == nil {
			authDst := filepath.Join(s.paths.SyncRepoDir(), "auth.json.age")

			if err := s.encryption.EncryptFile(authSrc, authDst); err != nil {
				return fmt.Errorf("failed to encrypt auth.json: %w", err)
			}
		}
	}

	// Handle mcp-auth.json if enabled
	if s.cfg.Sync.IncludeMcpAuth {
		if s.encryption == nil {
			return fmt.Errorf("includeMcpAuth requires encryption to be enabled")
		}

		mcpAuthSrc := s.paths.OpenCodeMcpAuthFile()
		if _, err := os.Stat(mcpAuthSrc); err == nil {
			mcpAuthDst := filepath.Join(s.paths.SyncRepoDir(), "mcp-auth.json.age")

			if err := s.encryption.EncryptFile(mcpAuthSrc, mcpAuthDst); err != nil {
				return fmt.Errorf("failed to encrypt mcp-auth.json: %w", err)
			}
		}
	}

	return nil
}

// CopyFromRepo copies files from sync repository to OpenCode config
func (s *Syncer) CopyFromRepo() error {
	repoDir := s.paths.SyncRepoDir()

	// Walk through repo directory
	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(repoDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip excluded patterns
		if s.shouldExclude(relPath) {
			return nil
		}

		// Determine destination
		var dstPath string
		if strings.HasPrefix(relPath, "claude-skills"+string(filepath.Separator)) || relPath == "claude-skills" {
			relToClaudeSkills, _ := filepath.Rel("claude-skills", relPath)
			if relToClaudeSkills == "." {
				return nil
			}
			dstPath = filepath.Join(s.paths.ClaudeSkillsDir, relToClaudeSkills)
		} else {
			dstPath = filepath.Join(s.paths.OpenCodeConfigDir, relPath)
		}

		// Handle encrypted auth.json
		if relPath == "auth.json.age" && s.cfg.Sync.IncludeAuth {
			if s.encryption == nil {
				return fmt.Errorf("found encrypted auth.json but encryption is not enabled")
			}

			dstPath = s.paths.OpenCodeAuthFile()

			if err := s.encryption.DecryptFile(path, dstPath); err != nil {
				return fmt.Errorf("failed to decrypt auth.json: %w", err)
			}
			return nil
		}

		// Handle encrypted mcp-auth.json
		if relPath == "mcp-auth.json.age" && s.cfg.Sync.IncludeMcpAuth {
			if s.encryption == nil {
				return fmt.Errorf("found encrypted mcp-auth.json but encryption is not enabled")
			}

			dstPath = s.paths.OpenCodeMcpAuthFile()

			if err := s.encryption.DecryptFile(path, dstPath); err != nil {
				return fmt.Errorf("failed to decrypt mcp-auth.json: %w", err)
			}
			return nil
		}

		// Copy file
		if err := s.copyFile(path, dstPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", relPath, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to copy from repo: %w", err)
	}

	return nil
}

// getSyncableFiles returns list of files that should be synced
func (s *Syncer) getSyncableFiles() ([]FileInfo, error) {
	var files []FileInfo

	syncablePaths := s.paths.SyncableOpenCodePaths()

	for _, srcPath := range syncablePaths {
		info, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to stat %s: %w", srcPath, err)
		}

		var relPath string
		if srcPath == s.paths.ClaudeSkillsDir {
			relPath = "claude-skills"
		} else {
			relPath, _ = filepath.Rel(s.paths.OpenCodeConfigDir, srcPath)
		}

		if info.IsDir() {
			// Walk directory
			err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return err
				}

				var fileRelPath string
				if srcPath == s.paths.ClaudeSkillsDir {
					pathRelToClaudeSkills, _ := filepath.Rel(s.paths.ClaudeSkillsDir, path)
					fileRelPath = filepath.Join("claude-skills", pathRelToClaudeSkills)
				} else {
					fileRelPath, _ = filepath.Rel(s.paths.OpenCodeConfigDir, path)
				}

				if s.shouldExclude(fileRelPath) {
					return nil
				}

				hash, err := s.hashFile(path)
				if err != nil {
					return err
				}

				files = append(files, FileInfo{
					Path:    path,
					RelPath: fileRelPath,
					Size:    info.Size(),
					ModTime: info.ModTime(),
					Hash:    hash,
				})

				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			if s.shouldExclude(relPath) {
				continue
			}

			hash, err := s.hashFile(srcPath)
			if err != nil {
				return nil, err
			}

			files = append(files, FileInfo{
				Path:    srcPath,
				RelPath: relPath,
				Size:    info.Size(),
				ModTime: info.ModTime(),
				Hash:    hash,
			})
		}
	}

	return files, nil
}

// shouldExclude checks if a path should be excluded
func (s *Syncer) shouldExclude(path string) bool {
	for _, pattern := range s.cfg.Sync.Exclude {
		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}
		// Also check if pattern matches any part of path
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// copyFile copies a single file
func (s *Syncer) copyFile(src, dst string) error {
	// Create destination directory
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()

	// Copy contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy contents: %w", err)
	}

	// Copy file mode
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set mode: %w", err)
	}

	return nil
}

// copyDir copies a directory recursively
func (s *Syncer) copyDir(src, dst string) error {
	// Get source info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}

	// Read directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := s.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := s.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// hashFile calculates SHA256 hash of a file
func (s *Syncer) hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
