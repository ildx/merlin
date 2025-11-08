package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// helper to run a command and capture stdout/stderr
// newTestRoot builds an isolated root command containing only diff.
func newTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "merlin"}
	d := &cobra.Command{Use: "diff", Run: func(c *cobra.Command, args []string) { runDiff(c) }}
	d.Flags().Bool("packages", false, "Include package (brew & mas) differences")
	d.Flags().Bool("configs", false, "Include config/symlink differences")
	d.Flags().Bool("scripts", false, "Include script differences")
	d.Flags().Bool("json", false, "Output JSON instead of human-readable text")
	root.AddCommand(d)
	return root
}

func runRootCommand(args ...string) (string, error) {
	cmd := newTestRoot()
	// Capture stdout/stderr because the diff command prints via fmt.Println directly
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	oldOut := os.Stdout
	oldErr := os.Stderr
	os.Stdout = wOut
	os.Stderr = wErr
	cmd.SetArgs(args)
	err := cmd.Execute()
	_ = wOut.Close()
	_ = wErr.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, rOut)
	_, _ = io.Copy(&buf, rErr)
	_ = rOut.Close()
	_ = rErr.Close()
	os.Stdout = oldOut
	os.Stderr = oldErr
	return buf.String(), err
}

// setupTempRepo creates a minimal repo layout with configs needed by diff.
func setupTempRepo(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	// root merlin.toml
	rootContent := []byte(`metadata = { name = "test" }
[settings]
conflict_strategy = "skip"
auto_commit = false
`)
	if err := os.WriteFile(filepath.Join(tmp, "merlin.toml"), rootContent, 0644); err != nil {
		t.Fatalf("write root: %v", err)
	}
	// brew config
	brewDir := filepath.Join(tmp, "config", "brew", "config")
	if err := os.MkdirAll(brewDir, 0755); err != nil {
		t.Fatalf("mkdir brew: %v", err)
	}
	brewContent := []byte(`[[formulae]]
name = "fzf"
[[casks]]
name = "raycast"
`)
	if err := os.WriteFile(filepath.Join(brewDir, "brew.toml"), brewContent, 0644); err != nil {
		t.Fatalf("write brew: %v", err)
	}
	// mas config
	masDir := filepath.Join(tmp, "config", "mas", "config")
	if err := os.MkdirAll(masDir, 0755); err != nil {
		t.Fatalf("mkdir mas: %v", err)
	}
	masContent := []byte(`[[apps]]
id = 123456789
name = "TestApp"
`)
	if err := os.WriteFile(filepath.Join(masDir, "mas.toml"), masContent, 0644); err != nil {
		t.Fatalf("write mas: %v", err)
	}
	// tool with link & scripts
	toolDir := filepath.Join(tmp, "config", "zsh")
	if err := os.MkdirAll(toolDir, 0755); err != nil {
		t.Fatalf("mkdir tool: %v", err)
	}
	toolMerlin := []byte(`tool = { name = "zsh" }
[[link]]
source = "config/omp.toml"
target = "{config_dir}/zsh/omp.toml"
[scripts]
directory = "scripts"
scripts = [{ file = "setup.sh" }]
`)
	if err := os.WriteFile(filepath.Join(toolDir, "merlin.toml"), toolMerlin, 0644); err != nil {
		t.Fatalf("write tool merlin: %v", err)
	}
	// create referenced config and script dirs
	if err := os.MkdirAll(filepath.Join(toolDir, "config"), 0755); err != nil {
		t.Fatalf("mkdir tool config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "config", "omp.toml"), []byte("theme=\"test\""), 0644); err != nil {
		t.Fatalf("write omp: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(toolDir, "scripts"), 0755); err != nil {
		t.Fatalf("mkdir scripts: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "scripts", "setup.sh"), []byte("echo setup"), 0755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return tmp
}

func TestDiffPackagesFlag(t *testing.T) {
	repo := setupTempRepo(t)
	os.Setenv("MERLIN_DOTFILES", repo)
	out, err := runRootCommand("diff", "--packages")
	if err != nil {
		t.Fatalf("command error: %v", err)
	}
	if !contains(out, "== Brew Formulae ==") || !contains(out, "== Brew Casks ==") || !contains(out, "== MAS Apps ==") {
		t.Fatalf("expected package sections, got: %s", out)
	}
	if contains(out, "== Symlinks ==") || contains(out, "== Scripts ==") {
		t.Fatalf("unexpected non-package sections present: %s", out)
	}
}

func TestDiffConfigsFlag(t *testing.T) {
	repo := setupTempRepo(t)
	os.Setenv("MERLIN_DOTFILES", repo)
	out, err := runRootCommand("diff", "--configs")
	if err != nil {
		t.Fatalf("command error: %v", err)
	}
	if !contains(out, "== Symlinks ==") {
		t.Fatalf("expected symlink section: %s", out)
	}
	if contains(out, "== Brew Formulae ==") || contains(out, "== Scripts ==") {
		t.Fatalf("unexpected sections present: %s", out)
	}
}

func TestDiffScriptsFlag(t *testing.T) {
	repo := setupTempRepo(t)
	os.Setenv("MERLIN_DOTFILES", repo)
	out, err := runRootCommand("diff", "--scripts")
	if err != nil {
		t.Fatalf("command error: %v", err)
	}
	if !contains(out, "== Scripts ==") {
		t.Fatalf("expected scripts section: %s", out)
	}
	if contains(out, "== Brew Formulae ==") || contains(out, "== Symlinks ==") {
		t.Fatalf("unexpected sections present: %s", out)
	}
}

func TestDiffJSONFlag(t *testing.T) {
	repo := setupTempRepo(t)
	os.Setenv("MERLIN_DOTFILES", repo)
	out, err := runRootCommand("diff", "--json")
	if err != nil {
		t.Fatalf("command error: %v", err)
	}
	if len(out) == 0 || out[0] != '{' {
		t.Fatalf("expected JSON output, got: %s", out)
	}
	// Basic schema keys
	for _, key := range []string{"brew_formulae", "brew_casks", "mas_apps", "symlinks", "scripts"} {
		if !contains(out, key) {
			t.Fatalf("expected key %s in JSON output", key)
		}
	}
}

func contains(haystack, needle string) bool { return bytes.Contains([]byte(haystack), []byte(needle)) }
