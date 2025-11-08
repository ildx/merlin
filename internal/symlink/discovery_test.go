package symlink

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ildx/merlin/internal/config"
	"github.com/ildx/merlin/internal/models"
)

func TestExpandVariables(t *testing.T) {
	vars := Variables{
		HomeDir:   "/Users/test",
		ConfigDir: "/Users/test/.config",
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "home_dir variable",
			input: "{home_dir}/.zshrc",
			want:  "/Users/test/.zshrc",
		},
		{
			name:  "config_dir variable",
			input: "{config_dir}/git",
			want:  "/Users/test/.config/git",
		},
		{
			name:  "tilde expansion",
			input: "~/.bashrc",
			want:  "/Users/test/.bashrc",
		},
		{
			name:  "tilde alone",
			input: "~",
			want:  "/Users/test",
		},
		{
			name:  "multiple variables",
			input: "{home_dir}/test/{config_dir}",
			want:  "/Users/test/test//Users/test/.config",
		},
		{
			name:  "no variables",
			input: "/absolute/path",
			want:  "/absolute/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandVariables(tt.input, vars)
			if got != tt.want {
				t.Errorf("expandVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDefaultVariables(t *testing.T) {
	vars, err := GetDefaultVariables()
	if err != nil {
		t.Fatalf("GetDefaultVariables() error = %v", err)
	}

	if vars.HomeDir == "" {
		t.Error("HomeDir should not be empty")
	}

	if vars.ConfigDir == "" {
		t.Error("ConfigDir should not be empty")
	}

	// ConfigDir should be HomeDir/.config
	expectedConfigDir := filepath.Join(vars.HomeDir, ".config")
	if vars.ConfigDir != expectedConfigDir {
		t.Errorf("ConfigDir = %v, want %v", vars.ConfigDir, expectedConfigDir)
	}

	t.Logf("HomeDir: %s", vars.HomeDir)
	t.Logf("ConfigDir: %s", vars.ConfigDir)
}

func TestGetVariablesFromRoot(t *testing.T) {
	t.Run("with default values", func(t *testing.T) {
		rootConfig := &models.RootMerlinConfig{
			Settings: models.Settings{},
		}

		vars, err := GetVariablesFromRoot(rootConfig)
		if err != nil {
			t.Fatalf("GetVariablesFromRoot() error = %v", err)
		}

		if vars.HomeDir == "" {
			t.Error("HomeDir should not be empty")
		}
		if vars.ConfigDir == "" {
			t.Error("ConfigDir should not be empty")
		}
	})

	t.Run("with custom values", func(t *testing.T) {
		rootConfig := &models.RootMerlinConfig{
			Settings: models.Settings{
				HomeDir:   "~",
				ConfigDir: "{home_dir}/.config",
			},
		}

		vars, err := GetVariablesFromRoot(rootConfig)
		if err != nil {
			t.Fatalf("GetVariablesFromRoot() error = %v", err)
		}

		// Should expand to actual paths
		if vars.HomeDir == "~" {
			t.Error("HomeDir should be expanded from ~")
		}
		if vars.ConfigDir == "{home_dir}/.config" {
			t.Error("ConfigDir should have variables expanded")
		}
	})
}

func TestResolveLink(t *testing.T) {
	// Create a temporary test structure
	tmpDir := t.TempDir()
	toolRoot := filepath.Join(tmpDir, "tool")
	configDir := filepath.Join(toolRoot, "config")
	
	// Create directories
	os.MkdirAll(configDir, 0755)
	
	// Create a test file
	testFile := filepath.Join(configDir, "test.conf")
	os.WriteFile(testFile, []byte("test"), 0644)

	vars := Variables{
		HomeDir:   "/Users/test",
		ConfigDir: "/Users/test/.config",
	}

	t.Run("simple directory link", func(t *testing.T) {
		link := models.Link{
			Target: "{config_dir}/mytool",
		}

		results, err := resolveLink(link, toolRoot, configDir, vars)
		if err != nil {
			t.Fatalf("resolveLink() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}

		if results[0].Source != configDir {
			t.Errorf("Source = %v, want %v", results[0].Source, configDir)
		}

		expectedTarget := "/Users/test/.config/mytool"
		if results[0].Target != expectedTarget {
			t.Errorf("Target = %v, want %v", results[0].Target, expectedTarget)
		}

		if !results[0].IsDir {
			t.Error("IsDir should be true for directory")
		}
	})

	t.Run("explicit source", func(t *testing.T) {
		link := models.Link{
			Source: "config/test.conf",
			Target: "{home_dir}/test.conf",
		}

		results, err := resolveLink(link, toolRoot, configDir, vars)
		if err != nil {
			t.Fatalf("resolveLink() error = %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}

		expectedSource := filepath.Join(toolRoot, "config/test.conf")
		if results[0].Source != expectedSource {
			t.Errorf("Source = %v, want %v", results[0].Source, expectedSource)
		}

		expectedTarget := "/Users/test/test.conf"
		if results[0].Target != expectedTarget {
			t.Errorf("Target = %v, want %v", results[0].Target, expectedTarget)
		}
	})
}

// Test with real Covenant repository if available
func TestDiscoverToolsRealRepo(t *testing.T) {
	covenantPath := "/Users/iivo/Development/personal/covenant"
	
	// Check if covenant repo exists
	if _, err := os.Stat(covenantPath); os.IsNotExist(err) {
		t.Skip("Covenant repository not available")
	}

	// Load the repository
	repo, err := config.LoadDotfilesRepo(covenantPath)
	if err != nil {
		t.Fatalf("failed to load covenant repo: %v", err)
	}

	// Get variables
	vars, err := GetDefaultVariables()
	if err != nil {
		t.Fatalf("failed to get variables: %v", err)
	}

	// Discover tools
	tools, err := DiscoverTools(repo, vars)
	if err != nil {
		t.Fatalf("DiscoverTools() error = %v", err)
	}

	if len(tools) == 0 {
		t.Error("expected to find some tools")
	}

	t.Logf("Discovered %d tools", len(tools))

	// Check specific tools
	for _, tool := range tools {
		t.Logf("Tool: %s", tool.Name)
		t.Logf("  Description: %s", tool.Description)
		t.Logf("  Has merlin.toml: %v", tool.HasMerlinTOML)
		t.Logf("  Links: %d", len(tool.Links))
		
		for i, link := range tool.Links {
			t.Logf("    Link %d:", i+1)
			t.Logf("      Source: %s", link.Source)
			t.Logf("      Target: %s", link.Target)
			t.Logf("      IsDir: %v", link.IsDir)
		}
	}

	// Verify git tool
	var gitTool *ToolConfig
	for _, tool := range tools {
		if tool.Name == "git" {
			gitTool = tool
			break
		}
	}

	if gitTool == nil {
		t.Error("expected to find git tool")
	} else {
		if !gitTool.HasMerlinTOML {
			t.Error("git should have merlin.toml")
		}
		if len(gitTool.Links) < 2 {
			t.Errorf("git should have at least 2 links, got %d", len(gitTool.Links))
		}
	}
}

func TestDiscoverToolConfig(t *testing.T) {
	covenantPath := "/Users/iivo/Development/personal/covenant"
	
	if _, err := os.Stat(covenantPath); os.IsNotExist(err) {
		t.Skip("Covenant repository not available")
	}

	repo, err := config.LoadDotfilesRepo(covenantPath)
	if err != nil {
		t.Fatalf("failed to load covenant repo: %v", err)
	}

	vars, err := GetDefaultVariables()
	if err != nil {
		t.Fatalf("failed to get variables: %v", err)
	}

	t.Run("discover git config", func(t *testing.T) {
		toolConfig, err := DiscoverToolConfig(repo, "git", vars)
		if err != nil {
			t.Fatalf("DiscoverToolConfig() error = %v", err)
		}

		if toolConfig.Name != "git" {
			t.Errorf("Name = %v, want git", toolConfig.Name)
		}

		if !toolConfig.HasMerlinTOML {
			t.Error("git should have merlin.toml")
		}

		if len(toolConfig.Links) == 0 {
			t.Error("git should have links configured")
		}
	})

	t.Run("discover cursor config", func(t *testing.T) {
		toolConfig, err := DiscoverToolConfig(repo, "cursor", vars)
		if err != nil {
			t.Fatalf("DiscoverToolConfig() error = %v", err)
		}

		if toolConfig.Name != "cursor" {
			t.Errorf("Name = %v, want cursor", toolConfig.Name)
		}

		if len(toolConfig.Dependencies) == 0 {
			t.Error("cursor should have dependencies")
		}
	})
}

