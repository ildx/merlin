package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ildx/merlin/internal/models"
)

// Helper function to create a temporary test file
func createTestFile(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "merlin-test-*.toml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	
	tmpFile.Close()
	return tmpFile.Name()
}

func TestParseBrewTOML(t *testing.T) {
	t.Run("valid brew.toml", func(t *testing.T) {
		content := `
[metadata]
version = "1.0.0"
description = "Test brew config"

[[brew]]
name = "git"
description = "Version control"
category = "development"
dependencies = []

[[cask]]
name = "firefox"
description = "Web browser"
category = "browser"
dependencies = []
`
		path := createTestFile(t, content)
		defer os.Remove(path)
		
		config, err := ParseBrewTOML(path)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		
		if len(config.Formulae) != 1 {
			t.Errorf("expected 1 formula, got %d", len(config.Formulae))
		}
		
		if len(config.Casks) != 1 {
			t.Errorf("expected 1 cask, got %d", len(config.Casks))
		}
		
		if config.Formulae[0].Name != "git" {
			t.Errorf("expected git, got %s", config.Formulae[0].Name)
		}
	})
	
	t.Run("non-existent file", func(t *testing.T) {
		_, err := ParseBrewTOML("/non/existent/path")
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})
	
	t.Run("invalid TOML", func(t *testing.T) {
		content := `invalid toml content [[[`
		path := createTestFile(t, content)
		defer os.Remove(path)
		
		_, err := ParseBrewTOML(path)
		if err == nil {
			t.Error("expected error for invalid TOML")
		}
	})
}

func TestParseMASTOML(t *testing.T) {
	t.Run("valid mas.toml", func(t *testing.T) {
		content := `
[metadata]
version = "1.0.0"
description = "Test MAS config"

[[app]]
name = "Xcode"
id = 497799835
description = "IDE"
category = "development"
dependencies = []
`
		path := createTestFile(t, content)
		defer os.Remove(path)
		
		config, err := ParseMASTOML(path)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		
		if len(config.Apps) != 1 {
			t.Errorf("expected 1 app, got %d", len(config.Apps))
		}
		
		if config.Apps[0].Name != "Xcode" {
			t.Errorf("expected Xcode, got %s", config.Apps[0].Name)
		}
		
		if config.Apps[0].ID != 497799835 {
			t.Errorf("expected ID 497799835, got %d", config.Apps[0].ID)
		}
	})
}

func TestParseRootMerlinTOML(t *testing.T) {
	t.Run("valid root merlin.toml", func(t *testing.T) {
		content := `
[metadata]
name = "test-dotfiles"
version = "1.0.0"
description = "Test dotfiles"

[settings]
auto_link = false
confirm_before_install = true
conflict_strategy = "backup"
home_dir = "~"
config_dir = "{home_dir}/.config"

[preinstall]
tools = ["brew", "git"]

[[profile]]
name = "personal"
default = true
description = "Personal setup"
tools = ["git", "zsh"]
`
		path := createTestFile(t, content)
		defer os.Remove(path)
		
		config, err := ParseRootMerlinTOML(path)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		
		if config.Metadata.Name != "test-dotfiles" {
			t.Errorf("expected test-dotfiles, got %s", config.Metadata.Name)
		}
		
		if config.Settings.ConflictStrategy != "backup" {
			t.Errorf("expected backup strategy, got %s", config.Settings.ConflictStrategy)
		}
		
		if len(config.Preinstall.Tools) != 2 {
			t.Errorf("expected 2 preinstall tools, got %d", len(config.Preinstall.Tools))
		}
		
		if len(config.Profiles) != 1 {
			t.Errorf("expected 1 profile, got %d", len(config.Profiles))
		}
		
		if !config.Profiles[0].Default {
			t.Error("expected profile to be default")
		}
	})
	
	t.Run("sets defaults when not provided", func(t *testing.T) {
		content := `
[metadata]
name = "test-dotfiles"
version = "1.0.0"

[settings]

[preinstall]
tools = []

[[profile]]
name = "minimal"
tools = []
`
		path := createTestFile(t, content)
		defer os.Remove(path)
		
		config, err := ParseRootMerlinTOML(path)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		
		if config.Settings.ConflictStrategy != "interactive" {
			t.Errorf("expected default interactive strategy, got %s", config.Settings.ConflictStrategy)
		}
		
		if config.Settings.HomeDir != "~" {
			t.Errorf("expected default home_dir ~, got %s", config.Settings.HomeDir)
		}
		
		if config.Settings.ConfigDir != "{home_dir}/.config" {
			t.Errorf("expected default config_dir, got %s", config.Settings.ConfigDir)
		}
	})
}

func TestParseToolMerlinTOML(t *testing.T) {
	t.Run("valid tool merlin.toml", func(t *testing.T) {
		content := `
[tool]
name = "git"
description = "Git config"
dependencies = []

[[link]]
target = "{config_dir}/git"

[[link]]
source = "config/.gitconfig"
target = "{home_dir}/.gitconfig"
`
		path := createTestFile(t, content)
		defer os.Remove(path)
		
		config, err := ParseToolMerlinTOML(path)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		
		if config.Tool.Name != "git" {
			t.Errorf("expected git, got %s", config.Tool.Name)
		}
		
		if len(config.Links) != 2 {
			t.Errorf("expected 2 links, got %d", len(config.Links))
		}
		
		if config.Links[1].Source != "config/.gitconfig" {
			t.Errorf("expected config/.gitconfig, got %s", config.Links[1].Source)
		}
	})
	
	t.Run("tool with scripts", func(t *testing.T) {
		content := `
[tool]
name = "cursor"
description = "AI editor"
dependencies = ["brew"]

[[link]]
target = "{config_dir}/cursor"

[scripts]
directory = "scripts"
scripts = ["install_extensions.sh"]
`
		path := createTestFile(t, content)
		defer os.Remove(path)
		
		config, err := ParseToolMerlinTOML(path)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		
		if !config.HasScripts() {
			t.Error("expected config to have scripts")
		}
		
		if config.Scripts.Directory != "scripts" {
			t.Errorf("expected scripts directory, got %s", config.Scripts.Directory)
		}
		
		if len(config.Scripts.Scripts) != 1 {
			t.Errorf("expected 1 script, got %d", len(config.Scripts.Scripts))
		}
	})
}

func TestValidateBrewConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &models.BrewConfig{
			Formulae: []models.BrewPackage{
				{Name: "git", Category: "development"},
			},
		}
		
		err := ValidateBrewConfig(config)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
	
	t.Run("empty config", func(t *testing.T) {
		config := &models.BrewConfig{}
		
		err := ValidateBrewConfig(config)
		if err == nil {
			t.Error("expected error for empty config")
		}
	})
	
	t.Run("duplicate package name", func(t *testing.T) {
		config := &models.BrewConfig{
			Formulae: []models.BrewPackage{
				{Name: "git", Category: "development"},
			},
			Casks: []models.BrewPackage{
				{Name: "git", Category: "development"},
			},
		}
		
		err := ValidateBrewConfig(config)
		if err == nil {
			t.Error("expected error for duplicate package name")
		}
	})
}

func TestValidateRootMerlinConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &models.RootMerlinConfig{
			Settings: models.Settings{
				ConflictStrategy: "backup",
			},
			Profiles: []models.Profile{
				{Name: "personal", Default: true},
				{Name: "work"},
			},
		}
		
		err := ValidateRootMerlinConfig(config)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
	
	t.Run("invalid conflict strategy", func(t *testing.T) {
		config := &models.RootMerlinConfig{
			Settings: models.Settings{
				ConflictStrategy: "invalid",
			},
		}
		
		err := ValidateRootMerlinConfig(config)
		if err == nil {
			t.Error("expected error for invalid conflict strategy")
		}
	})
	
	t.Run("multiple default profiles", func(t *testing.T) {
		config := &models.RootMerlinConfig{
			Settings: models.Settings{
				ConflictStrategy: "backup",
			},
			Profiles: []models.Profile{
				{Name: "personal", Default: true},
				{Name: "work", Default: true},
			},
		}
		
		err := ValidateRootMerlinConfig(config)
		if err == nil {
			t.Error("expected error for multiple default profiles")
		}
	})
}

// Test with real Covenant files (if available)
func TestParseRealCovenantFiles(t *testing.T) {
	covenantPath := "/Users/iivo/Development/personal/covenant"
	
	t.Run("parse real brew.toml", func(t *testing.T) {
		brewPath := filepath.Join(covenantPath, "config/brew/config/brew.toml")
		if _, err := os.Stat(brewPath); os.IsNotExist(err) {
			t.Skip("Covenant repo not available")
		}
		
		config, err := ParseBrewTOML(brewPath)
		if err != nil {
			t.Fatalf("failed to parse real brew.toml: %v", err)
		}
		
		if len(config.Formulae) == 0 && len(config.Casks) == 0 {
			t.Error("expected at least one package")
		}
		
		t.Logf("Found %d formulae and %d casks", len(config.Formulae), len(config.Casks))
		
		err = ValidateBrewConfig(config)
		if err != nil {
			t.Errorf("validation failed: %v", err)
		}
	})
	
	t.Run("parse real mas.toml", func(t *testing.T) {
		masPath := filepath.Join(covenantPath, "config/mas/config/mas.toml")
		if _, err := os.Stat(masPath); os.IsNotExist(err) {
			t.Skip("Covenant repo not available")
		}
		
		config, err := ParseMASTOML(masPath)
		if err != nil {
			t.Fatalf("failed to parse real mas.toml: %v", err)
		}
		
		if len(config.Apps) == 0 {
			t.Error("expected at least one app")
		}
		
		t.Logf("Found %d apps", len(config.Apps))
		
		err = ValidateMASConfig(config)
		if err != nil {
			t.Errorf("validation failed: %v", err)
		}
	})
	
	t.Run("parse real root merlin.toml", func(t *testing.T) {
		merlinPath := filepath.Join(covenantPath, "merlin.toml")
		if _, err := os.Stat(merlinPath); os.IsNotExist(err) {
			t.Skip("Covenant repo not available")
		}
		
		config, err := ParseRootMerlinTOML(merlinPath)
		if err != nil {
			t.Fatalf("failed to parse real root merlin.toml: %v", err)
		}
		
		if config.Metadata.Name != "covenant" {
			t.Errorf("expected covenant, got %s", config.Metadata.Name)
		}
		
		t.Logf("Found %d profiles and %d preinstall tools", 
			len(config.Profiles), len(config.Preinstall.Tools))
		
		err = ValidateRootMerlinConfig(config)
		if err != nil {
			t.Errorf("validation failed: %v", err)
		}
	})
	
	t.Run("parse real tool merlin.toml (git)", func(t *testing.T) {
		gitPath := filepath.Join(covenantPath, "config/git/merlin.toml")
		if _, err := os.Stat(gitPath); os.IsNotExist(err) {
			t.Skip("Covenant repo not available")
		}
		
		config, err := ParseToolMerlinTOML(gitPath)
		if err != nil {
			t.Fatalf("failed to parse real git merlin.toml: %v", err)
		}
		
		if config.Tool.Name != "git" {
			t.Errorf("expected git, got %s", config.Tool.Name)
		}
		
		t.Logf("Found %d links", len(config.Links))
		
		err = ValidateToolMerlinConfig(config)
		if err != nil {
			t.Errorf("validation failed: %v", err)
		}
	})
	
	t.Run("parse real tool merlin.toml (cursor)", func(t *testing.T) {
		cursorPath := filepath.Join(covenantPath, "config/cursor/merlin.toml")
		if _, err := os.Stat(cursorPath); os.IsNotExist(err) {
			t.Skip("Covenant repo not available")
		}
		
		config, err := ParseToolMerlinTOML(cursorPath)
		if err != nil {
			t.Fatalf("failed to parse real cursor merlin.toml: %v", err)
		}
		
		if config.Tool.Name != "cursor" {
			t.Errorf("expected cursor, got %s", config.Tool.Name)
		}
		
		if !config.HasScripts() {
			t.Error("expected cursor to have scripts")
		}
		
		t.Logf("Found %d links and %d scripts", len(config.Links), len(config.Scripts.Scripts))
	})
}

