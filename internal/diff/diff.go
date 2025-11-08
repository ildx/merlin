package diff

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ildx/merlin/internal/config"
	"github.com/ildx/merlin/internal/parser"
	"github.com/ildx/merlin/internal/state"
)

// PackageDiff captures differences for brew/mas packages
// Added: present in system but not in repo (extra)
// Missing: present in repo but not installed
// Present: intersection (could be extended with version info later)
// NOTE: Removed vs Added naming kept intuitive to user perspective.
type PackageDiff struct {
	Added   []string `json:"added"`   // installed locally but not declared
	Missing []string `json:"missing"` // declared but not installed
}

// SymlinkDiff captures configuration link differences.
// MissingLinks: declared in tool configs but not present as symlink
// OrphanedLinks: symlinks pointing into repo not declared in any tool config
// BrokenLinks: symlinks whose target does not exist
// DivergentLinks: reserved for future content hashing support
// For now DivergentLinks stays empty.
type SymlinkDiff struct {
	MissingLinks   []string `json:"missing_links"`
	OrphanedLinks  []string `json:"orphaned_links"`
	BrokenLinks    []string `json:"broken_links"`
	DivergentLinks []string `json:"divergent_links"`
}

// DiffResult aggregates all diff categories.
type DiffResult struct {
	BrewFormulae PackageDiff `json:"brew_formulae"`
	BrewCasks    PackageDiff `json:"brew_casks"`
	MASApps      PackageDiff `json:"mas_apps"`
	Symlinks     SymlinkDiff `json:"symlinks"`
	Scripts      PackageDiff `json:"scripts"` // Added/ Missing semantics: file exists vs declared
}

// Compute generates a DiffResult by comparing the repository definitions with a system snapshot.
func Compute(repo *config.DotfilesRepo, snap *state.SystemSnapshot) (*DiffResult, error) {
	result := &DiffResult{}

	// Brew diff
	brewConfig, brewErr := parser.ParseBrewTOML(filepath.Join(repo.ConfigDir, "brew", "config", "brew.toml"))
	if brewErr == nil && brewConfig != nil {
		formulaDeclared := make(map[string]bool)
		caskDeclared := make(map[string]bool)
		for _, f := range brewConfig.Formulae {
			formulaDeclared[f.Name] = true
		}
		for _, c := range brewConfig.Casks {
			caskDeclared[c.Name] = true
		}
		result.BrewFormulae = buildPackageDiff(formulaDeclared, snap.BrewFormulae)
		result.BrewCasks = buildPackageDiff(caskDeclared, snap.BrewCasks)
	}

	// MAS diff
	masConfig, masErr := parser.ParseMASTOML(filepath.Join(repo.ConfigDir, "mas", "config", "mas.toml"))
	if masErr == nil && masConfig != nil {
		appsDeclared := make(map[string]bool)
		for _, a := range masConfig.Apps {
			// MAS IDs are integers in config; snapshot keys are string IDs from `mas list`
			if a.ID > 0 {
				appsDeclared[strconv.Itoa(a.ID)] = true
			}
		}
		result.MASApps = buildPackageDiff(appsDeclared, snap.MASApps)
	}

	// Symlink diff
	symlinkDiff, err := computeSymlinkDiff(repo, snap)
	if err == nil {
		result.Symlinks = *symlinkDiff
	}

	// Script diff (aggregate across tools)
	scriptsDeclared := map[string]bool{}
	scriptsPresent := map[string]bool{}

	tools, tErr := repo.ListTools()
	if tErr == nil {
		for _, tool := range tools {
			cfgPath := repo.GetToolMerlinConfig(tool)
			c, perr := parser.ParseToolMerlinTOML(cfgPath)
			if perr != nil || c == nil || !c.HasScripts() {
				continue
			}
			// Determine script directory (default to "scripts" if unspecified)
			toolRoot := repo.GetToolRoot(tool)
			scriptDir := filepath.Join(toolRoot, c.Scripts.Directory)
			if c.Scripts.Directory == "" {
				scriptDir = filepath.Join(toolRoot, "scripts")
			}
			// Declared scripts
			for _, s := range c.Scripts.Scripts {
				// Key namespaced by tool for uniqueness (tool/script)
				key := fmt.Sprintf("%s/%s", tool, s.File)
				scriptsDeclared[key] = true
				// Presence check
				sp := filepath.Join(scriptDir, s.File)
				if fi, err := os.Stat(sp); err == nil && !fi.IsDir() {
					scriptsPresent[key] = true
				}
			}
			// Orphan (present but undeclared) detection: list directory files
			entries, _ := os.ReadDir(scriptDir)
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				key := fmt.Sprintf("%s/%s", tool, e.Name())
				if !scriptsDeclared[key] {
					scriptsPresent[key] = true // present but undeclared → Added
				}
			}
		}
	}
	result.Scripts = buildPackageDiff(scriptsDeclared, scriptsPresent)

	return result, nil
}

// buildPackageDiff computes Added (installed not declared) and Missing (declared not installed)
func buildPackageDiff(declared map[string]bool, installed map[string]bool) PackageDiff {
	var added []string
	var missing []string

	for name := range installed {
		if !declared[name] {
			added = append(added, name)
		}
	}
	for name := range declared {
		if !installed[name] {
			missing = append(missing, name)
		}
	}
	return PackageDiff{Added: added, Missing: missing}
}

// computeSymlinkDiff walks tool link declarations and compares with system symlink snapshot.
func computeSymlinkDiff(repo *config.DotfilesRepo, snap *state.SystemSnapshot) (*SymlinkDiff, error) {
	declaredTargets := make(map[string]bool)
	// Map of target -> source for declared
	declaredSourceByTarget := make(map[string]string)

	tools, err := repo.ListTools()
	if err != nil {
		return nil, err
	}

	for _, tool := range tools {
		toolConfigPath := repo.GetToolMerlinConfig(tool)
		c, err := parser.ParseToolMerlinTOML(toolConfigPath)
		if err != nil || c == nil {
			continue
		}
		for _, l := range c.Links {
			if len(l.Files) == 0 {
				resolvedTarget := resolveVariables(l.Target, repo)
				declaredTargets[resolvedTarget] = true
				declaredSourceByTarget[resolvedTarget] = buildSourcePath(repo.GetToolRoot(tool), l.Source)
			} else {
				for _, f := range l.Files {
					baseTarget := resolveVariables(l.Target, repo)
					resolvedTarget := filepath.Join(baseTarget, f.Target)
					declaredTargets[resolvedTarget] = true
					declaredSourceByTarget[resolvedTarget] = buildSourcePath(repo.GetToolRoot(tool), f.Source)
				}
			}
		}
	}

	// Build sets from snapshot
	snapshotTargets := make(map[string]state.SymlinkEntry)
	for _, entry := range snap.Symlinks {
		// We treat the symlink path itself as target location
		snapshotTargets[entry.LinkPath] = entry
	}

	var missing []string
	var orphaned []string
	var broken []string
	var divergent []string

	// Declared but not present
	for target := range declaredTargets {
		if _, ok := snapshotTargets[target]; !ok {
			missing = append(missing, target)
		}
	}

	// Orphaned: exists as symlink pointing into repo but not declared
	repoRoot := repo.Root
	for target, entry := range snapshotTargets {
		if !declaredTargets[target] {
			// Check if its target path points into repo root
			if strings.HasPrefix(entry.TargetPath, repoRoot) {
				orphaned = append(orphaned, target)
			}
		} else {
			// Divergence check: declared & present & not broken
			if !entry.Broken {
				src := declaredSourceByTarget[target]
				// Compare file hashes if both exist and are regular files
				if same, err := compareFileContent(src, entry.TargetPath); err == nil && !same {
					divergent = append(divergent, target)
				}
			}
		}
		if entry.Broken {
			broken = append(broken, target)
		}
	}

	return &SymlinkDiff{MissingLinks: missing, OrphanedLinks: orphaned, BrokenLinks: broken, DivergentLinks: divergent}, nil
}

// resolveVariables performs simple placeholder resolution for {home_dir} and {config_dir}
// Future: reuse existing parser variable expansion logic if available.
func resolveVariables(t string, repo *config.DotfilesRepo) string {
	// home_dir
	home, _ := os.UserHomeDir()
	res := strings.ReplaceAll(t, "{home_dir}", home)
	res = strings.ReplaceAll(res, "{config_dir}", filepath.Join(home, ".config"))
	return res
}

// ToJSON marshals the DiffResult into pretty JSON.
func (d *DiffResult) ToJSON() (string, error) {
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// HumanReadable renders a textual summary of the diff respecting filters.
func (d *DiffResult) HumanReadable(includePackages, includeConfigs, includeScripts bool) string {
	var b strings.Builder
	if includePackages {
		b.WriteString("== Brew Formulae ==\n")
		b.WriteString(renderSet("Added", d.BrewFormulae.Added))
		b.WriteString(renderSet("Missing", d.BrewFormulae.Missing))
		b.WriteString("\n== Brew Casks ==\n")
		b.WriteString(renderSet("Added", d.BrewCasks.Added))
		b.WriteString(renderSet("Missing", d.BrewCasks.Missing))
		b.WriteString("\n== MAS Apps ==\n")
		b.WriteString(renderSet("Added", d.MASApps.Added))
		b.WriteString(renderSet("Missing", d.MASApps.Missing))
	}
	if includeConfigs {
		b.WriteString("\n== Symlinks ==\n")
		b.WriteString(renderSet("Missing", d.Symlinks.MissingLinks))
		b.WriteString(renderSet("Orphaned", d.Symlinks.OrphanedLinks))
		b.WriteString(renderSet("Broken", d.Symlinks.BrokenLinks))
		b.WriteString(renderSet("Divergent", d.Symlinks.DivergentLinks))
	}
	if includeScripts {
		b.WriteString("\n== Scripts ==\n")
		b.WriteString(renderSet("Added", d.Scripts.Added))
		b.WriteString(renderSet("Missing", d.Scripts.Missing))
	}
	return b.String()
}

func renderSet(label string, items []string) string {
	if len(items) == 0 {
		return fmt.Sprintf("%s: none\n", label)
	}
	return fmt.Sprintf("%s (%d):\n  - %s\n", label, len(items), strings.Join(items, "\n  - "))
}

// buildSourcePath builds the absolute path to a source entry relative to tool root.
// If the provided source already starts with "config/" we don't duplicate the segment.
func buildSourcePath(toolRoot, src string) string {
	cleaned := filepath.Clean(src)
	if strings.HasPrefix(cleaned, "config/") {
		return filepath.Join(toolRoot, cleaned)
	}
	return filepath.Join(toolRoot, "config", cleaned)
}

// compareFileContent returns true if files have identical SHA256, false if different.
// If either file does not exist or is a directory, returns true (treat as non-divergent).
func compareFileContent(src, dst string) (bool, error) {
	si, serr := os.Stat(src)
	if serr != nil || si.IsDir() {
		return true, nil
	}
	di, derr := os.Stat(dst)
	if derr != nil || di.IsDir() {
		return true, nil
	}
	sh, err := hashFile(src)
	if err != nil {
		return true, err
	}
	dh, err := hashFile(dst)
	if err != nil {
		return true, err
	}
	return sh == dh, nil
}

func hashFile(p string) (string, error) {
	f, err := os.Open(p)
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
