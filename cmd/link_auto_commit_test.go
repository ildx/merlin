package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// executeLink runs `merlin link` (single tool) inside isolated temp repo.
func executeLinkSingle(t *testing.T, repo string, home string, tool string, autoCommit bool) (string, error) {
	t.Helper()
	os.Setenv("MERLIN_DOTFILES", repo)
	os.Setenv("HOME", home)
	writeRootConfig(t, repo, autoCommit)
	ensureToolConfig(t, repo, tool)
	initAndCommitRepo(t, repo)
	return runMerlinCommand(t, repo, []string{"link", tool})
}

// executeLinkAll runs `merlin link --all` inside isolated temp repo with provided tools.
func executeLinkAll(t *testing.T, repo string, home string, tools []string, autoCommit bool) (string, error) {
	t.Helper()
	os.Setenv("MERLIN_DOTFILES", repo)
	os.Setenv("HOME", home)
	writeRootConfig(t, repo, autoCommit)
	for _, tool := range tools {
		ensureToolConfig(t, repo, tool)
	}
	initAndCommitRepo(t, repo)
	return runMerlinCommand(t, repo, []string{"link", "--all"})
}

// writeRootConfig creates a minimal root merlin.toml with auto_commit flag.
func writeRootConfig(t *testing.T, repo string, auto bool) {
	t.Helper()
	if err := os.MkdirAll(repo, 0755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	ac := "false"
	if auto {
		ac = "true"
	}
	root := []byte("metadata = { name = \"test\" }\n[settings]\nauto_commit = " + ac + "\nconflict_strategy = \"skip\"\n")
	if err := os.WriteFile(filepath.Join(repo, "merlin.toml"), root, 0644); err != nil {
		t.Fatalf("write root merlin.toml: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "config"), 0755); err != nil {
		t.Fatalf("mkdir config root: %v", err)
	}
}

// ensureToolConfig sets up minimal tool config directory structure w/ config/<tool>/config + file.
func ensureToolConfig(t *testing.T, repo, tool string) {
	t.Helper()
	dir := filepath.Join(repo, "config", tool, "config")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir tool config: %v", err)
	}
	// create a sample file to link
	sample := filepath.Join(dir, "sample.txt")
	if err := os.WriteFile(sample, []byte("data"), 0644); err != nil {
		t.Fatalf("write sample: %v", err)
	}
}

// initAndCommitRepo initializes git repo & makes initial commit.
func initAndCommitRepo(t *testing.T, repo string) {
	t.Helper()
	runGitOrFail(t, repo, "init")
	runGitOrFail(t, repo, "config", "user.email", "tester@example.com")
	runGitOrFail(t, repo, "config", "user.name", "Tester")
	runGitOrFail(t, repo, "add", ".")
	runGitOrFail(t, repo, "commit", "-m", "chore: init")
}

// runMerlinCommand executes rootCmd with provided args capturing combined stdout+stderr.
func runMerlinCommand(t *testing.T, repo string, args []string) (string, error) {
	t.Helper()
	// Capture IO
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	_ = wOut.Close()
	_ = wErr.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, rOut)
	_, _ = io.Copy(&buf, rErr)
	_ = rOut.Close()
	_ = rErr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	return buf.String(), err
}

// git helpers
func runGitOrFail(t *testing.T, repo string, args ...string) {
	t.Helper()
	full := append([]string{"-C", repo}, args...)
	cmd := exec.Command("git", full...)
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v failed: %v", args, err)
	}
}

func gitOutput(t *testing.T, repo string, args ...string) []byte {
	t.Helper()
	full := append([]string{"-C", repo}, args...)
	cmd := exec.Command("git", full...)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v output failed: %v", args, err)
	}
	return out
}

// Test auto-commit enabled for single tool
func TestLinkAutoCommitSingleTool(t *testing.T) {
	if _, err := exec.Command("git", "--version").Output(); err != nil {
		t.Skip("git not available")
	}
	repo := t.TempDir()
	home := t.TempDir()
	out, err := executeLinkSingle(t, repo, home, "zsh", true)
	if err != nil {
		t.Fatalf("link command failed: %v\nOutput: %s", err, out)
	}
	// Wait briefly for commit
	time.Sleep(50 * time.Millisecond)
	msg := bytes.TrimSpace(gitOutput(t, repo, "log", "-1", "--pretty=%s"))
	if !bytes.Equal(msg, []byte("chore(link): link zsh")) {
		t.Fatalf("unexpected commit message: %s", string(msg))
	}
	// Repo should be clean
	status := gitOutput(t, repo, "status", "--porcelain")
	if len(status) != 0 {
		t.Fatalf("expected clean repo, status: %s", string(status))
	}
}

// Test auto-commit disabled (no new commit beyond init)
func TestLinkAutoCommitDisabled(t *testing.T) {
	if _, err := exec.Command("git", "--version").Output(); err != nil {
		t.Skip("git not available")
	}
	repo := t.TempDir()
	home := t.TempDir()
	out, err := executeLinkSingle(t, repo, home, "zsh", false)
	if err != nil {
		t.Fatalf("link command failed: %v\nOutput: %s", err, out)
	}
	// Latest commit should still be init
	msg := bytes.TrimSpace(gitOutput(t, repo, "log", "-1", "--pretty=%s"))
	if !bytes.Equal(msg, []byte("chore: init")) {
		t.Fatalf("unexpected commit message when disabled: %s", string(msg))
	}
}

// Test multi-tool commit message formatting (>3 tools triggers ellipsis)
func TestLinkAutoCommitMultiToolMessage(t *testing.T) {
	if _, err := exec.Command("git", "--version").Output(); err != nil {
		t.Skip("git not available")
	}
	repo := t.TempDir()
	home := t.TempDir()
	tools := []string{"zsh", "git", "eza", "brew", "mas"}
	out, err := executeLinkAll(t, repo, home, tools, true)
	if err != nil {
		t.Fatalf("link --all failed: %v\nOutput: %s", err, out)
	}
	time.Sleep(50 * time.Millisecond)
	msg := string(bytes.TrimSpace(gitOutput(t, repo, "log", "-1", "--pretty=%s")))
	if !bytes.HasPrefix([]byte(msg), []byte("chore(link): link 5 tools (")) || !bytes.Contains([]byte(msg), []byte("…)")) {
		t.Fatalf("multi-tool message format mismatch: %s", msg)
	}
}
