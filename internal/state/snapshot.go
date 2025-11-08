package state

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SystemSnapshot represents a point-in-time view of relevant system state
// for diff/export operations.
type SystemSnapshot struct {
	BrewFormulae map[string]bool
	BrewCasks    map[string]bool
	MASApps      map[string]bool
	Symlinks     []SymlinkEntry
}

// SymlinkEntry captures a discovered symlink and its resolution status.
type SymlinkEntry struct {
	LinkPath   string // path of the symlink on disk
	TargetPath string // resolved target (absolute)
	Broken     bool   // true if target does not exist
}

// CollectSnapshot gathers current system state. Individual collectors are
// resilient: failures (e.g., brew not installed) result in empty sets.
func CollectSnapshot(rootDir string) *SystemSnapshot {
	return &SystemSnapshot{
		BrewFormulae: collectBrew("formula"),
		BrewCasks:    collectBrew("cask"),
		MASApps:      collectMAS(),
		Symlinks:     collectSymlinks(rootDir),
	}
}

// collectBrew collects installed brew items of a given type (formula|cask).
func collectBrew(kind string) map[string]bool {
	items := make(map[string]bool)
	// Check if brew exists
	if _, err := exec.LookPath("brew"); err != nil {
		return items
	}

	var cmd *exec.Cmd
	if kind == "formula" {
		cmd = exec.Command("brew", "list", "--formula")
	} else {
		cmd = exec.Command("brew", "list", "--cask")
	}

	out, err := cmd.Output()
	if err != nil {
		return items
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		items[line] = true
	}
	return items
}

// collectMAS collects installed MAS apps by id or name.
// Uses `mas list` output lines like: "123456789 App Name".
func collectMAS() map[string]bool {
	apps := make(map[string]bool)
	if _, err := exec.LookPath("mas"); err != nil {
		return apps
	}

	cmd := exec.Command("mas", "list")
	out, err := cmd.Output()
	if err != nil {
		return apps
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Split first token (id) and remainder (name)
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		id := parts[0]
		apps[id] = true
	}
	return apps
}

// collectSymlinks walks the user's home directory and records symlinks whose
// targets exist or are broken. Scope kept small initially: only symlinks inside
// ~/.config and top-level dotfiles starting with '.'
func collectSymlinks(rootDir string) []SymlinkEntry {
	var entries []SymlinkEntry
	if rootDir == "" {
		home, _ := os.UserHomeDir()
		rootDir = home
	}

	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config")

	// Helper to process a path
	process := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.Type()&os.ModeSymlink == 0 {
			return nil
		}
		// Resolve target
		linkTarget, err := os.Readlink(path)
		broken := false
		abs := linkTarget
		if !filepath.IsAbs(linkTarget) {
			abs = filepath.Join(filepath.Dir(path), linkTarget)
		}
		abs = filepath.Clean(abs)
		if _, statErr := os.Stat(abs); statErr != nil {
			broken = true
		}
		entries = append(entries, SymlinkEntry{LinkPath: path, TargetPath: abs, Broken: broken})
		return nil
	}

	// Walk ~/.config
	filepath.WalkDir(configDir, process)

	// Scan top-level dotfiles in home (e.g., ~/.zshrc, ~/.gitconfig)
	entries = append(entries, scanTopLevelSymlinks(home)...)

	return entries
}

func scanTopLevelSymlinks(home string) []SymlinkEntry {
	patterns := []string{".zshrc", ".bashrc", ".gitconfig", ".tmux.conf", ".wezterm.lua"}
	var out []SymlinkEntry
	for _, name := range patterns {
		p := filepath.Join(home, name)
		info, err := os.Lstat(p)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		linkTarget, err := os.Readlink(p)
		if err != nil {
			continue
		}
		abs := linkTarget
		if !filepath.IsAbs(linkTarget) {
			abs = filepath.Join(filepath.Dir(p), linkTarget)
		}
		abs = filepath.Clean(abs)
		broken := false
		if _, statErr := os.Stat(abs); statErr != nil {
			broken = true
		}
		out = append(out, SymlinkEntry{LinkPath: p, TargetPath: abs, Broken: broken})
	}
	return out
}
