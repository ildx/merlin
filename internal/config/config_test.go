package config

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestRepo creates a temporary test dotfiles repository
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()
	
	// Create temp directory
	tmpDir := t.TempDir()
	
	// Create merlin.toml
	merlinToml := filepath.Join(tmpDir, RootConfigFile)
	if err := os.WriteFile(merlinToml, []byte("[metadata]\nname = \"test\""), 0644); err != nil {
		t.Fatalf("failed to create merlin.toml: %v", err)
	}
	
	// Create config directory
	configDir := filepath.Join(tmpDir, ConfigDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}
	
	cleanup := func() {
		// t.TempDir() handles cleanup automatically
	}
	
	return tmpDir, cleanup
}

// setupTestRepoWithTools creates a test repo with some tool directories
func setupTestRepoWithTools(t *testing.T, tools []string) (string, func()) {
	t.Helper()
	
	tmpDir, cleanup := setupTestRepo(t)
	
	configDir := filepath.Join(tmpDir, ConfigDir)
	for _, tool := range tools {
		toolDir := filepath.Join(configDir, tool)
		toolConfigDir := filepath.Join(toolDir, ConfigDir)
		if err := os.MkdirAll(toolConfigDir, 0755); err != nil {
			t.Fatalf("failed to create tool directory %s: %v", tool, err)
		}
	}
	
	return tmpDir, cleanup
}

func TestLoadDotfilesRepo(t *testing.T) {
	t.Run("valid repository", func(t *testing.T) {
		tmpDir, cleanup := setupTestRepo(t)
		defer cleanup()
		
		repo, err := LoadDotfilesRepo(tmpDir)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		
		if repo.Root != tmpDir {
			t.Errorf("expected root %s, got %s", tmpDir, repo.Root)
		}
		
		expectedConfigDir := filepath.Join(tmpDir, ConfigDir)
		if repo.ConfigDir != expectedConfigDir {
			t.Errorf("expected config dir %s, got %s", expectedConfigDir, repo.ConfigDir)
		}
	})
	
	t.Run("non-existent path", func(t *testing.T) {
		_, err := LoadDotfilesRepo("/non/existent/path")
		if err != ErrDotfilesNotFound {
			t.Errorf("expected ErrDotfilesNotFound, got: %v", err)
		}
	})
	
	t.Run("path without merlin.toml", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		_, err := LoadDotfilesRepo(tmpDir)
		if err != ErrNotADotfilesRepo {
			t.Errorf("expected ErrNotADotfilesRepo, got: %v", err)
		}
	})
}

func TestIsValidDotfilesRepo(t *testing.T) {
	t.Run("valid repository", func(t *testing.T) {
		tmpDir, cleanup := setupTestRepo(t)
		defer cleanup()
		
		if !IsValidDotfilesRepo(tmpDir) {
			t.Error("expected repository to be valid")
		}
	})
	
	t.Run("invalid repository", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		if IsValidDotfilesRepo(tmpDir) {
			t.Error("expected repository to be invalid")
		}
	})
}

func TestFindDotfilesInPath(t *testing.T) {
	t.Run("finds in current directory", func(t *testing.T) {
		tmpDir, cleanup := setupTestRepo(t)
		defer cleanup()
		
		repo, err := findDotfilesInPath(tmpDir)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		
		if repo.Root != tmpDir {
			t.Errorf("expected root %s, got %s", tmpDir, repo.Root)
		}
	})
	
	t.Run("finds in parent directory", func(t *testing.T) {
		tmpDir, cleanup := setupTestRepo(t)
		defer cleanup()
		
		// Create a subdirectory
		subDir := filepath.Join(tmpDir, "subdir", "nested")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("failed to create subdirectory: %v", err)
		}
		
		repo, err := findDotfilesInPath(subDir)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		
		if repo.Root != tmpDir {
			t.Errorf("expected root %s, got %s", tmpDir, repo.Root)
		}
	})
	
	t.Run("returns error when not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		_, err := findDotfilesInPath(tmpDir)
		if err != ErrDotfilesNotFound {
			t.Errorf("expected ErrDotfilesNotFound, got: %v", err)
		}
	})
}

func TestFindDotfilesRepo_EnvVar(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()
	
	// Set environment variable
	oldEnv := os.Getenv(EnvVarDotfiles)
	defer os.Setenv(EnvVarDotfiles, oldEnv)
	
	os.Setenv(EnvVarDotfiles, tmpDir)
	
	repo, err := FindDotfilesRepo()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	
	if repo.Root != tmpDir {
		t.Errorf("expected root %s, got %s", tmpDir, repo.Root)
	}
}

func TestDotfilesRepo_ToolMethods(t *testing.T) {
	tools := []string{"git", "zsh", "cursor"}
	tmpDir, cleanup := setupTestRepoWithTools(t, tools)
	defer cleanup()
	
	repo, err := LoadDotfilesRepo(tmpDir)
	if err != nil {
		t.Fatalf("failed to load repo: %v", err)
	}
	
	t.Run("GetToolConfigDir", func(t *testing.T) {
		expected := filepath.Join(tmpDir, ConfigDir, "git", ConfigDir)
		got := repo.GetToolConfigDir("git")
		if got != expected {
			t.Errorf("expected %s, got %s", expected, got)
		}
	})
	
	t.Run("GetToolRoot", func(t *testing.T) {
		expected := filepath.Join(tmpDir, ConfigDir, "git")
		got := repo.GetToolRoot("git")
		if got != expected {
			t.Errorf("expected %s, got %s", expected, got)
		}
	})
	
	t.Run("GetToolMerlinConfig", func(t *testing.T) {
		expected := filepath.Join(tmpDir, ConfigDir, "git", RootConfigFile)
		got := repo.GetToolMerlinConfig("git")
		if got != expected {
			t.Errorf("expected %s, got %s", expected, got)
		}
	})
	
	t.Run("ToolExists", func(t *testing.T) {
		if !repo.ToolExists("git") {
			t.Error("expected git to exist")
		}
		
		if repo.ToolExists("nonexistent") {
			t.Error("expected nonexistent tool to not exist")
		}
	})
	
	t.Run("ListTools", func(t *testing.T) {
		foundTools, err := repo.ListTools()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		
		if len(foundTools) != len(tools) {
			t.Errorf("expected %d tools, got %d", len(tools), len(foundTools))
		}
		
		// Check all tools are found
		toolMap := make(map[string]bool)
		for _, tool := range foundTools {
			toolMap[tool] = true
		}
		
		for _, expected := range tools {
			if !toolMap[expected] {
				t.Errorf("expected to find tool %s", expected)
			}
		}
	})
}

