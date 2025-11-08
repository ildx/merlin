package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestOpenNotRepo(t *testing.T) {
	tmp := t.TempDir()
	if _, err := Open(tmp); err == nil {
		t.Fatalf("expected error opening non-repo")
	}
}

func TestStatusAndCommit(t *testing.T) {
	if !IsGitAvailable() {
		t.Skip("git not available")
	}
	tmp := t.TempDir()
	// init repo
	cmd := exec.Command("git", "-C", tmp, "init")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, string(out))
	}
	repo, err := Open(tmp)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	// create file
	f := filepath.Join(tmp, "test.txt")
	if err := os.WriteFile(f, []byte("hello"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// status should show untracked
	st, err := repo.Status()
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if len(st.Untracked) == 0 {
		t.Fatalf("expected untracked file")
	}
	// commit
	if err := repo.Commit("chore(test): add test file", []string{"test.txt"}); err != nil {
		t.Fatalf("commit: %v", err)
	}
	// status now clean
	st2, err := repo.Status()
	if err != nil {
		t.Fatalf("status2: %v", err)
	}
	if !st2.Clean {
		t.Fatalf("expected clean repo after commit")
	}
}
