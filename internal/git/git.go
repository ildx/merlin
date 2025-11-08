package git

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repo represents a git repository at a given root path.
type Repo struct {
	Root string
}

// Status holds a simplified view of git status porcelain output.
type Status struct {
	Unstaged   []string // modified but not staged
	Staged     []string // staged changes
	Untracked  []string // untracked files
	Conflicted []string // merge conflicts
	Clean      bool     // true if no changes at all
}

var ErrNotRepo = errors.New("not a git repository")

// Open attempts to open a git repo at path. It checks for .git directory.
func Open(path string) (*Repo, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(filepath.Join(abs, ".git")); err != nil {
		return nil, ErrNotRepo
	}
	return &Repo{Root: abs}, nil
}

// Status returns parsed status information using 'git status --porcelain=v1'.
func (r *Repo) Status() (*Status, error) {
	cmd := exec.Command("git", "-C", r.Root, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	st := &Status{}
	lines := bytes.Split(out, []byte("\n"))
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		code := string(line[:2])
		path := strings.TrimSpace(string(line[3:]))
		switch {
		case code == "??":
			st.Untracked = append(st.Untracked, path)
		case strings.Contains(code, "U"):
			st.Conflicted = append(st.Conflicted, path)
		case code[0] != ' ' && code[0] != '?':
			st.Staged = append(st.Staged, path)
		case code[1] != ' ' && code[1] != '?':
			st.Unstaged = append(st.Unstaged, path)
		}
	}
	st.Clean = len(st.Untracked) == 0 && len(st.Unstaged) == 0 && len(st.Staged) == 0 && len(st.Conflicted) == 0
	return st, nil
}

// HasUnrelatedChanges returns true if there are unstaged or untracked changes outside the allowlist prefixes.
// allowPrefixes should be relative paths (directories) under repo root considered safe to commit.
func (r *Repo) HasUnrelatedChanges(allowPrefixes []string) (bool, error) {
	st, err := r.Status()
	if err != nil {
		return false, err
	}
	// helper to test membership
	inAllowed := func(p string) bool {
		for _, pref := range allowPrefixes {
			if pref == "" {
				continue
			}
			// Normalize: ensure trailing slash for directory semantics
			ap := strings.TrimSuffix(pref, "/") + "/"
			// Also allow exact file match
			if p == pref || strings.HasPrefix(p, ap) {
				return true
			}
		}
		return false
	}
	for _, lists := range [][]string{st.Untracked, st.Unstaged, st.Conflicted} {
		for _, path := range lists {
			if !inAllowed(path) {
				return true, nil
			}
		}
	}
	return false, nil
}

// Commit stages provided paths (relative to repo root) and creates a commit.
// If paths is empty, it commits all staged changes; if none staged returns error.
func (r *Repo) Commit(message string, paths []string) error {
	if len(paths) > 0 {
		// Stage only given paths
		args := append([]string{"-C", r.Root, "add"}, paths...)
		if err := exec.Command("git", args...).Run(); err != nil {
			return err
		}
	}
	// Verify something to commit
	st, err := r.Status()
	if err != nil {
		return err
	}
	if len(st.Staged) == 0 {
		return errors.New("no staged changes to commit")
	}
	cmd := exec.Command("git", "-C", r.Root, "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// IsGitAvailable checks if git binary exists.
func IsGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// FilterPaths returns subset of provided paths that actually exist under repo root.
func (r *Repo) FilterPaths(paths []string) []string {
	var out []string
	for _, p := range paths {
		abs := filepath.Join(r.Root, p)
		if _, err := os.Stat(abs); err == nil {
			out = append(out, p)
		}
	}
	return out
}
