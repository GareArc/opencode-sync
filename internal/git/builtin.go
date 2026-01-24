package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type BuiltinGit struct {
	path string
	repo *git.Repository
}

func NewBuiltinGit(path string) *BuiltinGit {
	return &BuiltinGit{
		path: path,
	}
}

func (g *BuiltinGit) Clone(url string) error {
	parentDir := filepath.Dir(g.path)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	if err := runGitCommand(parentDir, "clone", "--depth", "1", url, g.path); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	repo, err := git.PlainOpen(g.path)
	if err != nil {
		return fmt.Errorf("failed to open cloned repository: %w", err)
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

func (g *BuiltinGit) Push() error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	if err := runGitCommand(g.path, "push", "origin", "HEAD"); err != nil {
		return &AuthError{Remote: "origin", Err: err}
	}

	return nil
}

func (g *BuiltinGit) ForcePush() error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	if err := runGitCommand(g.path, "push", "--force", "origin", "HEAD"); err != nil {
		return &AuthError{Remote: "origin", Err: err}
	}

	return nil
}

func (g *BuiltinGit) Pull() error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	if err := runGitCommand(g.path, "pull", "origin"); err != nil {
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
			diff += fmt.Sprintf("%s: %c\n", path, fileStatus.Worktree)
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

func (g *BuiltinGit) Fetch() error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	if err := runGitCommand(g.path, "fetch", "origin"); err != nil {
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

func (g *BuiltinGit) GC() error {
	if g.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	if err := runGitCommand(g.path, "gc", "--aggressive", "--prune=now"); err != nil {
		return fmt.Errorf("failed to run git gc: %w", err)
	}

	return nil
}
