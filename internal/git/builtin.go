package git

import (
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// BuiltinGit implements Repository using go-git
type BuiltinGit struct {
	path string
	repo *git.Repository
	auth transport.AuthMethod
}

// NewBuiltinGit creates a new BuiltinGit instance
func NewBuiltinGit(path string) *BuiltinGit {
	return &BuiltinGit{
		path: path,
		auth: nil, // Will be set based on repo URL
	}
}

// SetAuth sets the authentication method
func (g *BuiltinGit) SetAuth(auth transport.AuthMethod) {
	g.auth = auth
}

// SetSSHAuth sets SSH authentication
func (g *BuiltinGit) SetSSHAuth(keyPath string, password string) error {
	auth, err := ssh.NewPublicKeysFromFile("git", keyPath, password)
	if err != nil {
		return fmt.Errorf("failed to load SSH key: %w", err)
	}
	g.auth = auth
	return nil
}

// SetHTTPAuth sets HTTP basic authentication
func (g *BuiltinGit) SetHTTPAuth(username, password string) {
	g.auth = &http.BasicAuth{
		Username: username,
		Password: password,
	}
}

// Clone clones a repository
func (g *BuiltinGit) Clone(url string) error {
	repo, err := git.PlainClone(g.path, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
		Auth:     g.auth,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	g.repo = repo
	return nil
}

// Init initializes a new repository
func (g *BuiltinGit) Init() error {
	repo, err := git.PlainInit(g.path, false)
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	g.repo = repo
	return nil
}

// Open opens an existing repository
func (g *BuiltinGit) Open() error {
	repo, err := git.PlainOpen(g.path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	g.repo = repo
	return nil
}

// AddRemote adds a remote
func (g *BuiltinGit) AddRemote(name, url string) error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	_, err := g.repo.CreateRemote(&config.RemoteConfig{
		Name: name,
		URLs: []string{url},
	})
	if err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	return nil
}

// Status returns repository status
func (g *BuiltinGit) Status() (*Status, error) {
	if g.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	w, err := g.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Get current branch
	head, err := g.repo.Head()
	var branch string
	if err == nil {
		branch = head.Name().Short()
	}

	// Parse status
	result := &Status{
		Branch:         branch,
		IsClean:        status.IsClean(),
		UntrackedFiles: []string{},
		ModifiedFiles:  []string{},
		StagedFiles:    []string{},
	}

	for path, fileStatus := range status {
		switch {
		case fileStatus.Worktree == git.Untracked:
			result.HasUntracked = true
			result.UntrackedFiles = append(result.UntrackedFiles, path)
		case fileStatus.Worktree == git.Modified || fileStatus.Worktree == git.Deleted:
			result.HasModified = true
			result.ModifiedFiles = append(result.ModifiedFiles, path)
		case fileStatus.Staging != git.Unmodified:
			result.HasStaged = true
			result.StagedFiles = append(result.StagedFiles, path)
		}
	}

	return result, nil
}

// Add stages files
func (g *BuiltinGit) Add(paths []string) error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	w, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	for _, path := range paths {
		_, err := w.Add(path)
		if err != nil {
			return fmt.Errorf("failed to add %s: %w", path, err)
		}
	}

	return nil
}

// AddAll stages all changes
func (g *BuiltinGit) AddAll() error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	w, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = w.Add(".")
	if err != nil {
		return fmt.Errorf("failed to add all: %w", err)
	}

	return nil
}

// Commit creates a commit
func (g *BuiltinGit) Commit(message string) error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	w, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get git config for author info
	cfg, err := g.repo.ConfigScoped(config.GlobalScope)
	if err != nil {
		cfg, _ = g.repo.Config()
	}

	author := &object.Signature{
		Name:  cfg.User.Name,
		Email: cfg.User.Email,
		When:  time.Now(),
	}

	if author.Name == "" {
		author.Name = "opencode-sync"
	}
	if author.Email == "" {
		author.Email = "opencode-sync@local"
	}

	_, err = w.Commit(message, &git.CommitOptions{
		Author: author,
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// Push pushes to remote
func (g *BuiltinGit) Push() error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	err := g.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       g.auth,
		Progress:   os.Stdout,
	})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			return nil
		}
		return &AuthError{Remote: "origin", Err: err}
	}

	return nil
}

func (g *BuiltinGit) ForcePush() error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	err := g.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       g.auth,
		Progress:   os.Stdout,
		Force:      true,
	})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			return nil
		}
		return &AuthError{Remote: "origin", Err: err}
	}

	return nil
}

func (g *BuiltinGit) Pull() error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	w, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = w.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth:       g.auth,
		Progress:   os.Stdout,
	})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			return nil
		}
		// Check for conflicts
		if err.Error() == "merge conflicts" {
			return &ConflictError{Files: []string{}}
		}
		return fmt.Errorf("failed to pull: %w", err)
	}

	return nil
}

// Diff returns the diff
func (g *BuiltinGit) Diff() (string, error) {
	if g.repo == nil {
		return "", fmt.Errorf("repository not initialized")
	}

	// Get HEAD commit
	head, err := g.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := g.repo.CommitObject(head.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get tree: %w", err)
	}

	// Get worktree status
	w, err := g.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	// Build simple diff output
	var diff string
	for path, fileStatus := range status {
		if fileStatus.Worktree != git.Unmodified {
			diff += fmt.Sprintf("%s: %s\n", path, fileStatus.Worktree)
		}
	}

	_ = tree // TODO: Implement proper diff using tree

	return diff, nil
}

// GetRemoteURL returns the remote URL
func (g *BuiltinGit) GetRemoteURL(name string) (string, error) {
	if g.repo == nil {
		return "", fmt.Errorf("repository not initialized")
	}

	remote, err := g.repo.Remote(name)
	if err != nil {
		return "", fmt.Errorf("failed to get remote: %w", err)
	}

	cfg := remote.Config()
	if len(cfg.URLs) == 0 {
		return "", fmt.Errorf("no URLs configured for remote %s", name)
	}

	return cfg.URLs[0], nil
}

// HasChanges returns true if there are uncommitted changes
func (g *BuiltinGit) HasChanges() (bool, error) {
	status, err := g.Status()
	if err != nil {
		return false, err
	}

	return !status.IsClean, nil
}

// IsClean returns true if working directory is clean
func (g *BuiltinGit) IsClean() (bool, error) {
	status, err := g.Status()
	if err != nil {
		return false, err
	}

	return status.IsClean, nil
}

// GetLastCommit returns the last commit info
func (g *BuiltinGit) GetLastCommit() (*CommitInfo, error) {
	if g.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	head, err := g.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := g.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	return &CommitInfo{
		Hash:      commit.Hash.String()[:7],
		Author:    commit.Author.Name,
		Email:     commit.Author.Email,
		Message:   commit.Message,
		Timestamp: commit.Author.When,
	}, nil
}

// Fetch fetches from remote without merging
func (g *BuiltinGit) Fetch() error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	err := g.repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Auth:       g.auth,
		Progress:   os.Stdout,
	})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			return nil
		}
		return fmt.Errorf("failed to fetch: %w", err)
	}

	return nil
}

// GetBranch returns the current branch name
func (g *BuiltinGit) GetBranch() (string, error) {
	if g.repo == nil {
		return "", fmt.Errorf("repository not initialized")
	}

	head, err := g.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	return head.Name().Short(), nil
}

// CheckoutBranch checks out a branch
func (g *BuiltinGit) CheckoutBranch(branch string) error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	w, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}
