package models

import (
	"testing"
)

func TestBrewConfig(t *testing.T) {
	config := BrewConfig{
		Metadata: Metadata{
			Name:        "test-brew",
			Version:     "1.0.0",
			Description: "Test brew config",
		},
		Formulae: []BrewPackage{
			{Name: "git", Category: "development", Description: "Version control"},
			{Name: "wget", Category: "development", Description: "Network downloader"},
		},
		Casks: []BrewPackage{
			{Name: "firefox", Category: "browser", Description: "Web browser"},
			{Name: "chrome", Category: "browser", Description: "Web browser"},
		},
	}

	t.Run("GetAllPackages", func(t *testing.T) {
		all := config.GetAllPackages()
		if len(all) != 4 {
			t.Errorf("expected 4 packages, got %d", len(all))
		}
	})

	t.Run("GetByCategory", func(t *testing.T) {
		devPkgs := config.GetByCategory("development")
		if len(devPkgs) != 2 {
			t.Errorf("expected 2 development packages, got %d", len(devPkgs))
		}

		browserPkgs := config.GetByCategory("browser")
		if len(browserPkgs) != 2 {
			t.Errorf("expected 2 browser packages, got %d", len(browserPkgs))
		}
	})

	t.Run("GetCategories", func(t *testing.T) {
		categories := config.GetCategories()
		if len(categories) != 2 {
			t.Errorf("expected 2 categories, got %d", len(categories))
		}
	})
}

func TestMASConfig(t *testing.T) {
	config := MASConfig{
		Metadata: Metadata{
			Name:        "test-mas",
			Version:     "1.0.0",
			Description: "Test MAS config",
		},
		Apps: []MASApp{
			{Name: "Xcode", ID: 497799835, Category: "development", Description: "IDE"},
			{Name: "Pages", ID: 409201541, Category: "productivity", Description: "Word processor"},
			{Name: "Numbers", ID: 409203825, Category: "productivity", Description: "Spreadsheet"},
		},
	}

	t.Run("GetByCategory", func(t *testing.T) {
		devApps := config.GetByCategory("development")
		if len(devApps) != 1 {
			t.Errorf("expected 1 development app, got %d", len(devApps))
		}

		prodApps := config.GetByCategory("productivity")
		if len(prodApps) != 2 {
			t.Errorf("expected 2 productivity apps, got %d", len(prodApps))
		}
	})

	t.Run("FindByID", func(t *testing.T) {
		app := config.FindByID(497799835)
		if app == nil {
			t.Fatal("expected to find Xcode")
		}
		if app.Name != "Xcode" {
			t.Errorf("expected Xcode, got %s", app.Name)
		}

		missing := config.FindByID(999999999)
		if missing != nil {
			t.Error("expected nil for missing app")
		}
	})

	t.Run("FindByName", func(t *testing.T) {
		app := config.FindByName("Pages")
		if app == nil {
			t.Fatal("expected to find Pages")
		}
		if app.ID != 409201541 {
			t.Errorf("expected ID 409201541, got %d", app.ID)
		}

		missing := config.FindByName("NonExistent")
		if missing != nil {
			t.Error("expected nil for missing app")
		}
	})
}

func TestRootMerlinConfig(t *testing.T) {
	config := RootMerlinConfig{
		Metadata: Metadata{
			Name:        "test-dotfiles",
			Version:     "1.0.0",
			Description: "Test dotfiles",
		},
		Settings: Settings{
			AutoLink:             false,
			ConfirmBeforeInstall: true,
			ConflictStrategy:     "backup",
			HomeDir:              "~",
			ConfigDir:            "~/.config",
		},
		Preinstall: PreinstallSettings{
			Tools: []string{"brew", "git"},
		},
		Profiles: []Profile{
			{Name: "personal", Default: true, Tools: []string{"git", "zsh"}},
			{Name: "work", Hostname: "work-mac", Tools: []string{"git"}},
		},
	}

	t.Run("GetDefaultProfile", func(t *testing.T) {
		profile := config.GetDefaultProfile()
		if profile == nil {
			t.Fatal("expected to find default profile")
		}
		if profile.Name != "personal" {
			t.Errorf("expected personal profile, got %s", profile.Name)
		}
	})

	t.Run("GetProfileByName", func(t *testing.T) {
		profile := config.GetProfileByName("work")
		if profile == nil {
			t.Fatal("expected to find work profile")
		}
		if profile.Hostname != "work-mac" {
			t.Errorf("expected work-mac hostname, got %s", profile.Hostname)
		}

		missing := config.GetProfileByName("nonexistent")
		if missing != nil {
			t.Error("expected nil for missing profile")
		}
	})

	t.Run("GetProfileByHostname", func(t *testing.T) {
		profile := config.GetProfileByHostname("work-mac")
		if profile == nil {
			t.Fatal("expected to find profile by hostname")
		}
		if profile.Name != "work" {
			t.Errorf("expected work profile, got %s", profile.Name)
		}

		missing := config.GetProfileByHostname("unknown-host")
		if missing != nil {
			t.Error("expected nil for missing hostname")
		}
	})
}

func TestToolMerlinConfig(t *testing.T) {
	t.Run("HasScripts", func(t *testing.T) {
		configWithScripts := ToolMerlinConfig{
			Scripts: ScriptsSection{
				Directory: "scripts",
				Scripts:   []ScriptItem{{File: "install.sh"}},
			},
		}

		if !configWithScripts.HasScripts() {
			t.Error("expected HasScripts to be true")
		}

		configWithoutScripts := ToolMerlinConfig{}
		if configWithoutScripts.HasScripts() {
			t.Error("expected HasScripts to be false")
		}

		// Tag helper methods
		configWithTagged := ToolMerlinConfig{Scripts: ScriptsSection{Scripts: []ScriptItem{{File: "a.sh", Tags: []string{"core"}}, {File: "b.sh", Tags: []string{"dev", "slow"}}, {File: "c.sh"}}}}
		if !configWithTagged.HasScriptTag("core") {
			t.Error("expected HasScriptTag core true")
		}
		if configWithTagged.HasScriptTag("missing") {
			t.Error("expected HasScriptTag missing false")
		}
		filtered := configWithTagged.FilterScriptsByTag([]string{"dev"})
		if len(filtered) != 1 || filtered[0].File != "b.sh" {
			t.Errorf("expected filter to return b.sh, got %v", filtered)
		}
	})

	t.Run("HasLinks", func(t *testing.T) {
		configWithLinks := ToolMerlinConfig{
			Links: []Link{
				{Target: "~/.config/tool"},
			},
		}

		if !configWithLinks.HasLinks() {
			t.Error("expected HasLinks to be true")
		}

		configWithoutLinks := ToolMerlinConfig{}
		if configWithoutLinks.HasLinks() {
			t.Error("expected HasLinks to be false")
		}
	})

	t.Run("HasDependencies", func(t *testing.T) {
		configWithDeps := ToolMerlinConfig{
			Tool: ToolInfo{
				Dependencies: []string{"brew"},
			},
		}

		if !configWithDeps.HasDependencies() {
			t.Error("expected HasDependencies to be true")
		}

		configWithoutDeps := ToolMerlinConfig{}
		if configWithoutDeps.HasDependencies() {
			t.Error("expected HasDependencies to be false")
		}
	})
}
