package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ildx/merlin/internal/config"
	"github.com/ildx/merlin/internal/installer"
	"github.com/ildx/merlin/internal/models"
	"github.com/ildx/merlin/internal/parser"
	"github.com/ildx/merlin/internal/symlink"
)

// LaunchPackageInstaller shows package selection and installation
func LaunchPackageInstaller() error {
	// Find dotfiles repo
	repo, err := config.FindDotfilesRepo()
	if err != nil {
		return fmt.Errorf("dotfiles repository not found: %w", err)
	}

	// Parse brew.toml
	brewPath := filepath.Join(repo.GetToolConfigDir("brew"), "brew.toml")
	if _, err := os.Stat(brewPath); os.IsNotExist(err) {
		return fmt.Errorf("brew.toml not found at %s", brewPath)
	}

	brewConfig, err := parser.ParseBrewTOML(brewPath)
	if err != nil {
		return fmt.Errorf("failed to parse brew.toml: %w", err)
	}

	// Show package type selection menu
	typeMenu := NewPackageTypeMenu()
	p := tea.NewProgram(typeMenu, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	typeModel, ok := finalModel.(PackageTypeMenu)
	if !ok || typeModel.cancelled {
		return nil
	}

	selectedType := typeModel.selected

	// Show package selector based on type
	var packages []models.BrewPackage
	var title string

	switch selectedType {
	case "formulae":
		packages = brewConfig.Formulae
		title = "🔧 Select Homebrew Formulae to Install"
	case "casks":
		packages = brewConfig.Casks
		title = "📱 Select Homebrew Casks to Install"
	case "both":
		packages = append(brewConfig.Formulae, brewConfig.Casks...)
		title = "📦 Select Packages to Install"
	default:
		return nil
	}

	if len(packages) == 0 {
		fmt.Println("\nNo packages found.")
		return nil
	}

	// Show package selector
	selector := NewPackageSelectorModel(title, packages)
	p = tea.NewProgram(selector, tea.WithAltScreen())
	finalModel, err = p.Run()
	if err != nil {
		return err
	}

	selectorModel, ok := finalModel.(PackageSelectorModel)
	if !ok || !selectorModel.IsConfirmed() {
		return nil
	}

	selected := selectorModel.GetSelectedPackages()
	if len(selected) == 0 {
		fmt.Println("No packages selected.")
		return nil
	}

	// Separate into formulae and casks
	var formulae, casks []models.BrewPackage
	for _, pkg := range selected {
		// Check if it's a cask (simple heuristic: check if in original casks list)
		isCask := false
		for _, c := range brewConfig.Casks {
			if c.Name == pkg.Name {
				isCask = true
				break
			}
		}
		if isCask {
			casks = append(casks, pkg)
		} else {
			formulae = append(formulae, pkg)
		}
	}

	// Install packages
	fmt.Println("\n📦 Installing selected packages...")
	brewInstaller := installer.NewBrewInstaller(false, true)

	var formulaeResults, caskResults []*installer.InstallResult

	if len(formulae) > 0 {
		formulaeResults = brewInstaller.InstallFormulae(formulae, os.Stdout)
	}

	if len(casks) > 0 {
		caskResults = brewInstaller.InstallCasks(casks, os.Stdout)
	}

	installer.PrintSummary(formulaeResults, caskResults, os.Stdout)

	return nil
}

// LaunchConfigManager shows config selection and link/unlink
func LaunchConfigManager() error {
	// Find dotfiles repo
	repo, err := config.FindDotfilesRepo()
	if err != nil {
		return fmt.Errorf("dotfiles repository not found: %w", err)
	}

	// Load root config for variables
	rootConfigPath := repo.GetRootMerlinConfig()
	rootConfig, err := parser.ParseRootMerlinTOML(rootConfigPath)
	if err != nil {
		return fmt.Errorf("error parsing root config: %w", err)
	}

	vars, err := symlink.GetVariablesFromRoot(rootConfig)
	if err != nil {
		return fmt.Errorf("error getting variables: %w", err)
	}

	// Get all tools
	tools, err := symlink.DiscoverTools(repo, vars)
	if err != nil {
		return fmt.Errorf("failed to discover tools: %w", err)
	}

	if len(tools) == 0 {
		fmt.Println("\nNo config tools found.")
		return nil
	}

	// Convert to ConfigItems
	configItems := make([]ConfigItem, len(tools))
	for i, tool := range tools {
		configItems[i] = ConfigItem{
			Name:        tool.Name,
			Description: tool.Description,
			IsLinked:    false, // TODO: Check actual link status
			HasConflict: false,
			Selected:    false,
		}
	}

	// Show action menu (link or unlink)
	actionMenu := NewConfigActionMenu()
	p := tea.NewProgram(actionMenu, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	actionModel, ok := finalModel.(ConfigActionMenu)
	if !ok || actionModel.cancelled {
		return nil
	}

	action := actionModel.selected

	// Show config selector
	title := "🔗 Select Configs to Link"
	if action == "unlink" {
		title = "🔓 Select Configs to Unlink"
	}

	selector := NewConfigSelectorModel(title, configItems, action)
	p = tea.NewProgram(selector, tea.WithAltScreen())
	finalModel, err = p.Run()
	if err != nil {
		return err
	}

	selectorModel, ok := finalModel.(ConfigSelectorModel)
	if !ok || !selectorModel.IsConfirmed() {
		return nil
	}

	selectedNames := selectorModel.GetSelectedConfigs()
	if len(selectedNames) == 0 {
		fmt.Println("\nNo configs selected.")
		return nil
	}

	// Perform action
	strategy := symlink.StrategySkip // Default strategy

	for _, name := range selectedNames {
		// Find the tool
		var tool *symlink.ToolConfig
		for _, t := range tools {
			if t.Name == name {
				tool = t
				break
			}
		}

		if tool == nil {
			continue
		}

		if action == "link" {
			fmt.Printf("\n🔗 Linking %s...\n", name)
			results, err := symlink.LinkToolWithStrategy(tool, strategy, false)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				printLinkResults(results)
			}
		} else {
			fmt.Printf("\n🔓 Unlinking %s...\n", name)
			results, err := symlink.UnlinkTool(tool, false)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				printUnlinkResults(results)
			}
		}
	}

	fmt.Println("\n✓ Complete!")
	return nil
}

func printLinkResults(results []*symlink.LinkResult) {
	for _, result := range results {
		switch result.Status {
		case symlink.LinkStatusSuccess:
			fmt.Printf("  ✓ %s\n", result.Target)
		case symlink.LinkStatusSkipped:
			fmt.Printf("  ⊘ %s (skipped)\n", result.Target)
		case symlink.LinkStatusAlreadyLinked:
			fmt.Printf("  ✓ %s (already linked)\n", result.Target)
		case symlink.LinkStatusConflict:
			fmt.Printf("  ⚠ %s (conflict)\n", result.Target)
		case symlink.LinkStatusError:
			fmt.Printf("  ✗ %s (error: %s)\n", result.Target, result.Message)
		}
	}
}

func printUnlinkResults(results []*symlink.UnlinkResult) {
	for _, result := range results {
		switch result.Status {
		case symlink.LinkStatusSuccess:
			fmt.Printf("  ✓ %s (removed)\n", result.Target)
		case symlink.LinkStatusSkipped:
			fmt.Printf("  ⊘ %s (skipped)\n", result.Target)
		case symlink.LinkStatusError:
			fmt.Printf("  ✗ %s (error: %s)\n", result.Target, result.Message)
		}
	}
}
