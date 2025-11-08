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

// executeBackupCreate runs `merlin backup create` in isolated temp repo capturing stdout.
func executeBackupCreate(t *testing.T, repo string, home string, files []string, autoCommit bool) (string, error) {
	t.Helper()
	// Set environment for repo + HOME isolation
	os.Setenv("MERLIN_DOTFILES", repo)
	os.Setenv("HOME", home)
	// Build minimal root merlin.toml
	rootCfg := []byte("metadata = { name = \"test\" }\n[settings]\nauto_commit = ")
	if autoCommit {
		rootCfg = append(rootCfg, []byte("true\nconflict_strategy = \"skip\"\n")...)
	} else {
		rootCfg = append(rootCfg, []byte("false\nconflict_strategy = \"skip\"\n")...)
	}
	if err := os.WriteFile(filepath.Join(repo, "merlin.toml"), rootCfg, 0644); err != nil {
		t.Fatalf("write merlin.toml: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "config"), 0755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}

	// Create dummy files for backup inside HOME
	for _, f := range files {
		abs := filepath.Join(home, f)
		if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
			t.Fatalf("mkdir file parent: %v", err)
		}
		if err := os.WriteFile(abs, []byte("data"), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}

	// Initialize git repo at repo root
	if err := initGitRepo(repo); err != nil {
		t.Fatalf("init git: %v", err)
	}
	// Stage initial files so repo clean pre-command
	if err := runGit(repo, "add", "."); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := runGit(repo, "config", "user.email", "tester@example.com"); err != nil {
		t.Fatalf("git config email: %v", err)
	}
	if err := runGit(repo, "config", "user.name", "Tester"); err != nil {
		t.Fatalf("git config name: %v", err)
	}
	if err := runGit(repo, "add", "."); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := runGit(repo, "commit", "-m", "chore: init"); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	// Capture stdout/stderr during command
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr

	// Build command root with backup subcommands only (reuse existing global rootCmd unsafe; we mimic minimal usage)
	rootCmd.SetArgs(append([]string{"backup", "create"}, buildFileArgs(home, files)...))
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

func buildFileArgs(home string, files []string) []string {
	out := []string{}
	for _, f := range files {
		out = append(out, filepath.Join(home, f))
	}
	return out
}

// Minimal git helpers for test
func initGitRepo(path string) error { return runGit(path, "init") }

func runGit(path string, args ...string) error {
	full := append([]string{"-C", path}, args...)
	cmd := execCommand("git", full...)
	return cmd.Run()
}

// indirection for testability
var execCommand = func(name string, args ...string) *exec.Cmd { return exec.Command(name, args...) }

// Test auto-commit occurs when enabled
func TestBackupAutoCommitEnabled(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	out, err := executeBackupCreate(t, repo, home, []string{"test/a.txt"}, true)
	if err != nil {
		t.Fatalf("command failed: %v\nOutput: %s", err, out)
	}
	// Verify index file exists
	idx := filepath.Join(repo, ".merlin-meta", "backups.json")
	if _, err := os.Stat(idx); err != nil {
		t.Fatalf("index file missing: %v", err)
	}
	// Repo should be clean after commit (no staged/untracked)
	stOut, _ := runGitOutput(repo, "status", "--porcelain")
	if len(stOut) != 0 {
		t.Fatalf("expected clean repo, got status output: %s", string(stOut))
	}
}

// Test no auto-commit when disabled
func TestBackupAutoCommitDisabled(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	out, err := executeBackupCreate(t, repo, home, []string{"test/b.txt"}, false)
	if err != nil {
		t.Fatalf("command failed: %v\nOutput: %s", err, out)
	}
	idx := filepath.Join(repo, ".merlin-meta", "backups.json")
	if _, err := os.Stat(idx); err == nil {
		t.Fatalf("index file should not exist when auto_commit disabled")
	}
}

func runGitOutput(path string, args ...string) ([]byte, error) {
	full := append([]string{"-C", path}, args...)
	cmd := execCommand("git", full...)
	return cmd.Output()
}

// Guard: ensure commit message format applied (best-effort by checking log) if git available
func TestBackupCommitMessageFormat(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	// Skip if git not available
	if _, err := execCommand("git", "--version").Output(); err != nil {
		t.Skip("git not available")
	}
	_, err := executeBackupCreate(t, repo, home, []string{"test/c.txt"}, true)
	if err != nil {
		t.Fatalf("backup create failed: %v", err)
	}
	// Wait briefly to ensure commit written
	time.Sleep(100 * time.Millisecond)
	logOut, err := runGitOutput(repo, "log", "-1", "--pretty=%s")
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}
	if !bytes.HasPrefix(bytes.TrimSpace(logOut), []byte("chore(backup): record")) {
		t.Fatalf("unexpected commit message: %s", string(logOut))
	}
}
