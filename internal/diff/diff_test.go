package diff

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ildx/merlin/internal/config"
	"github.com/ildx/merlin/internal/state"
)

func TestBuildPackageDiff(t *testing.T) {
	decl := map[string]bool{"a": true, "b": true}
	inst := map[string]bool{"b": true, "c": true}
	d := buildPackageDiff(decl, inst)
	// Added: c (installed not declared)
	if len(d.Added) != 1 || d.Added[0] != "c" {
		t.Errorf("expected added=c, got %#v", d.Added)
	}
	// Missing: a (declared not installed)
	if len(d.Missing) != 1 || d.Missing[0] != "a" {
		t.Errorf("expected missing=a, got %#v", d.Missing)
	}
}

func TestComputeSymlinkDiffBasic(t *testing.T) {
	tmp := t.TempDir()
	// Create minimal fake repo structure with config directory
	repoRoot := filepath.Join(tmp, "repo")
	configDir := filepath.Join(repoRoot, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	repo := &config.DotfilesRepo{Root: repoRoot, ConfigDir: configDir}

	// Snapshot: one symlink pointing into repoRoot (will be orphaned), one broken external
	snap := &state.SystemSnapshot{
		Symlinks: []state.SymlinkEntry{
			{LinkPath: filepath.Join(tmp, "a"), TargetPath: filepath.Join(repoRoot, "config", "tool", "config", "file"), Broken: false},
			{LinkPath: filepath.Join(tmp, "b"), TargetPath: filepath.Join(tmp, "nonexistent"), Broken: true},
		},
	}

	d, err := computeSymlinkDiff(repo, snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(d.OrphanedLinks) != 1 {
		t.Errorf("expected 1 orphaned link, got %d", len(d.OrphanedLinks))
	}
	if len(d.BrokenLinks) != 1 {
		t.Errorf("expected 1 broken link, got %d", len(d.BrokenLinks))
	}
}

func TestSymlinkDivergenceDetection(t *testing.T) {
	tmp := t.TempDir()
	repoRoot := filepath.Join(tmp, "repo")
	configDir := filepath.Join(repoRoot, "config", "tool")
	if err := os.MkdirAll(filepath.Join(configDir, "config"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Declared source file with content A
	srcFile := filepath.Join(configDir, "config", "file.txt")
	if err := os.WriteFile(srcFile, []byte("A"), 0644); err != nil {
		t.Fatalf("write src: %v", err)
	}
	// Create tool merlin.toml declaring link to absolute target
	toolMerlin := filepath.Join(configDir, "merlin.toml")
	targetPath := filepath.Join(tmp, "linked.txt")
	merlinContent := []byte("[[link]]\nsource = \"config/file.txt\"\ntarget = \"" + targetPath + "\"\n")
	if err := os.WriteFile(toolMerlin, merlinContent, 0644); err != nil {
		t.Fatalf("write merlin: %v", err)
	}
	repo := &config.DotfilesRepo{Root: repoRoot, ConfigDir: filepath.Join(repoRoot, "config")}

	// Create a symlink at target path pointing to DIFFERENT file to simulate divergence
	otherFile := filepath.Join(tmp, "other.txt")
	if err := os.WriteFile(otherFile, []byte("B"), 0644); err != nil {
		t.Fatalf("write other: %v", err)
	}
	if err := os.Symlink(otherFile, targetPath); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	snap := &state.SystemSnapshot{Symlinks: []state.SymlinkEntry{{LinkPath: targetPath, TargetPath: otherFile, Broken: false}}}
	d, err := computeSymlinkDiff(repo, snap)
	if err != nil {
		t.Fatalf("diff err: %v", err)
	}
	if len(d.DivergentLinks) != 1 {
		t.Fatalf("expected 1 divergent link, got %d", len(d.DivergentLinks))
	}
}

func TestScriptDiff(t *testing.T) {
	tmp := t.TempDir()
	repoRoot := filepath.Join(tmp, "repo")
	configDir := filepath.Join(repoRoot, "config", "tool")
	scriptDir := filepath.Join(configDir, "scripts")
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Create declared script present
	if err := os.WriteFile(filepath.Join(scriptDir, "present.sh"), []byte("echo hi"), 0755); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Create orphan script (undeclared)
	if err := os.WriteFile(filepath.Join(scriptDir, "extra.sh"), []byte("echo extra"), 0755); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Tool merlin with two declared scripts: present.sh & missing.sh
	toolMerlin := filepath.Join(configDir, "merlin.toml")
	merlinContent := []byte("[scripts]\n directory = \"scripts\"\n scripts = [\n \"present.sh\", \"missing.sh\"\n]\n")
	if err := os.WriteFile(toolMerlin, merlinContent, 0644); err != nil {
		t.Fatalf("write merlin: %v", err)
	}
	repo := &config.DotfilesRepo{Root: repoRoot, ConfigDir: filepath.Join(repoRoot, "config")}
	snap := &state.SystemSnapshot{} // scripts diff does not rely on snapshot currently
	result, err := Compute(repo, snap)
	if err != nil {
		t.Fatalf("compute err: %v", err)
	}
	// Added should include extra.sh
	foundAdded := false
	for _, a := range result.Scripts.Added {
		if strings.Contains(a, "extra.sh") {
			foundAdded = true
			break
		}
	}
	if !foundAdded {
		t.Fatalf("expected extra.sh in Added, got %#v", result.Scripts.Added)
	}
	// Missing should include missing.sh
	foundMissing := false
	for _, m := range result.Scripts.Missing {
		if strings.Contains(m, "missing.sh") {
			foundMissing = true
			break
		}
	}
	if !foundMissing {
		t.Fatalf("expected missing.sh in Missing, got %#v", result.Scripts.Missing)
	}
}
