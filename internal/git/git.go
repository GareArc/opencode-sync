package git

import (
	"fmt"
	"time"
)

// Repository represents a Git repository interface
type Repository interface {
	// Clone clones a repository from URL to the repo path
	Clone(url string) error

	// Init initializes a new repository
	Init() error

	// AddRemote adds a remote with the given name and URL
	AddRemote(name, url string) error

	// Status returns the current repository status
	Status() (*Status, error)

	// Add stages files for commit
	Add(paths []string) error

	// AddAll stages all changes
	AddAll() error

	// Commit creates a new commit with the given message
	Commit(message string) error

	// Push pushes commits to the remote
	Push() error

	// ForcePush force pushes commits to the remote (overwrites remote)
	ForcePush() error

	// Pull pulls changes from the remote
	Pull() error

	// Diff returns the diff between working directory and HEAD
	Diff() (string, error)

	// GetRemoteURL returns the URL of the given remote
	GetRemoteURL(name string) (string, error)

	// HasChanges returns true if there are uncommitted changes
	HasChanges() (bool, error)

	// IsClean returns true if working directory is clean
	IsClean() (bool, error)

	// GC runs git garbage collection to optimize repository size
	GC() error

	// GetBranch returns the current branch name
	GetBranch() (string, error)

	// Fetch fetches updates from remote without merging
	Fetch() error
}

// Status represents repository status
type Status struct {
	Branch         string
	IsClean        bool
	HasUntracked   bool
	HasModified    bool
	HasStaged      bool
	UntrackedFiles []string
	ModifiedFiles  []string
	StagedFiles    []string
}

// FileChange represents a file change
type FileChange struct {
	Path    string
	Status  ChangeStatus
	OldPath string // For renames
}

// ChangeStatus represents the status of a file
type ChangeStatus int

const (
	StatusUnmodified ChangeStatus = iota
	StatusAdded
	StatusModified
	StatusDeleted
	StatusRenamed
	StatusCopied
	StatusUntracked
)

func (s ChangeStatus) String() string {
	switch s {
	case StatusAdded:
		return "added"
	case StatusModified:
		return "modified"
	case StatusDeleted:
		return "deleted"
	case StatusRenamed:
		return "renamed"
	case StatusCopied:
		return "copied"
	case StatusUntracked:
		return "untracked"
	default:
		return "unmodified"
	}
}

// CommitInfo represents commit metadata
type CommitInfo struct {
	Hash      string
	Author    string
	Email     string
	Message   string
	Timestamp time.Time
}

// ConflictError represents a merge conflict
type ConflictError struct {
	Files []string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("merge conflict in %d file(s)", len(e.Files))
}

// AuthError represents an authentication error
type AuthError struct {
	Remote string
	Err    error
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication failed for remote %s: %v", e.Remote, e.Err)
}

func (e *AuthError) Unwrap() error {
	return e.Err
}
