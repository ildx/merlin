package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ildx/merlin/internal/config"
	"github.com/ildx/merlin/internal/models"
	"github.com/ildx/merlin/internal/parser"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List packages, apps, or configs",
	Long: `List available Homebrew packages, Mac App Store apps, or config tools.

When run without subcommands, shows everything.

Examples:
  merlin list          List everything (brew, mas, configs)
  merlin list brew     List all Homebrew packages (formulae & casks)
  merlin list mas      List all Mac App Store apps
  merlin list configs  List all available config tools`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand provided, show everything
		if err := runListAll(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var listBrewCmd = &cobra.Command{
	Use:   "brew",
	Short: "List Homebrew packages",
	Long:  "List all Homebrew formulae and casks from brew.toml",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runListBrew(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var listMASCmd = &cobra.Command{
	Use:   "mas",
	Short: "List Mac App Store apps",
	Long:  "List all Mac App Store applications from mas.toml",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runListMAS(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var listConfigsCmd = &cobra.Command{
	Use:     "configs",
	Aliases: []string{"tools"},
	Short:   "List available config tools",
	Long:    "List all available configuration tools in the dotfiles repository",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runListConfigs(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(listBrewCmd)
	listCmd.AddCommand(listMASCmd)
	listCmd.AddCommand(listConfigsCmd)
	
	// Flags for filtering/formatting
	listBrewCmd.Flags().StringP("category", "c", "", "Filter by category")
	listBrewCmd.Flags().Bool("formulae-only", false, "Show only formulae")
	listBrewCmd.Flags().Bool("casks-only", false, "Show only casks")
	
	listMASCmd.Flags().StringP("category", "c", "", "Filter by category")
}

func runListAll(cmd *cobra.Command) error {
	// Find dotfiles repository once
	repo, err := config.FindDotfilesRepo()
	if err != nil {
		return fmt.Errorf("dotfiles repository not found: %w", err)
	}
	
	fmt.Printf("\n════════════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("  MERLIN DOTFILES OVERVIEW\n")
	fmt.Printf("  Repository: %s\n", repo.Root)
	fmt.Printf("════════════════════════════════════════════════════════════════════════════════\n")
	
	// List Homebrew packages
	brewPath := filepath.Join(repo.GetToolConfigDir("brew"), "brew.toml")
	if _, err := os.Stat(brewPath); err == nil {
		if err := runListBrew(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "\n⚠️  Failed to list brew packages: %v\n", err)
		}
	} else {
		fmt.Println("\n📦 Homebrew Packages")
		fmt.Println(strings.Repeat("─", 80))
		fmt.Println("  No brew.toml found")
	}
	
	// List Mac App Store apps
	masPath := filepath.Join(repo.GetToolConfigDir("mas"), "mas.toml")
	if _, err := os.Stat(masPath); err == nil {
		if err := runListMAS(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "\n⚠️  Failed to list MAS apps: %v\n", err)
		}
	} else {
		fmt.Println("\n🍎 Mac App Store Applications")
		fmt.Println(strings.Repeat("─", 80))
		fmt.Println("  No mas.toml found")
	}
	
	// List config tools
	if err := runListConfigs(); err != nil {
		fmt.Fprintf(os.Stderr, "\n⚠️  Failed to list config tools: %v\n", err)
	}
	
	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════════════════════════")
	fmt.Printf("  Use 'merlin list <subcommand>' for filtered views\n")
	fmt.Println("════════════════════════════════════════════════════════════════════════════════")
	fmt.Println()
	
	return nil
}

func runListBrew(cmd *cobra.Command) error {
	// Find dotfiles repository
	repo, err := config.FindDotfilesRepo()
	if err != nil {
		return fmt.Errorf("dotfiles repository not found: %w", err)
	}
	
	// Find brew.toml
	brewPath := filepath.Join(repo.GetToolConfigDir("brew"), "brew.toml")
	if _, err := os.Stat(brewPath); os.IsNotExist(err) {
		return fmt.Errorf("brew.toml not found at %s", brewPath)
	}
	
	// Parse brew.toml
	brewConfig, err := parser.ParseBrewTOML(brewPath)
	if err != nil {
		return fmt.Errorf("failed to parse brew.toml: %w", err)
	}
	
	// Get filter flags
	categoryFilter, _ := cmd.Flags().GetString("category")
	formulaeOnly, _ := cmd.Flags().GetBool("formulae-only")
	casksOnly, _ := cmd.Flags().GetBool("casks-only")
	
	// Print header
	fmt.Printf("\n📦 Homebrew Packages\n")
	fmt.Printf("Repository: %s\n\n", repo.Root)
	
	// Print formulae
	if !casksOnly && len(brewConfig.Formulae) > 0 {
		fmt.Printf("🔧 Formulae (%d)\n", len(brewConfig.Formulae))
		fmt.Println(strings.Repeat("─", 80))
		printBrewPackages(brewConfig.Formulae, categoryFilter)
		fmt.Println()
	}
	
	// Print casks
	if !formulaeOnly && len(brewConfig.Casks) > 0 {
		fmt.Printf("📱 Casks (%d)\n", len(brewConfig.Casks))
		fmt.Println(strings.Repeat("─", 80))
		printBrewPackages(brewConfig.Casks, categoryFilter)
		fmt.Println()
	}
	
	// Print categories summary
	categories := brewConfig.GetCategories()
	if len(categories) > 0 {
		sort.Strings(categories)
		fmt.Printf("📂 Categories: %s\n\n", strings.Join(categories, ", "))
	}
	
	return nil
}

func runListMAS(cmd *cobra.Command) error {
	// Find dotfiles repository
	repo, err := config.FindDotfilesRepo()
	if err != nil {
		return fmt.Errorf("dotfiles repository not found: %w", err)
	}
	
	// Find mas.toml
	masPath := filepath.Join(repo.GetToolConfigDir("mas"), "mas.toml")
	if _, err := os.Stat(masPath); os.IsNotExist(err) {
		return fmt.Errorf("mas.toml not found at %s", masPath)
	}
	
	// Parse mas.toml
	masConfig, err := parser.ParseMASTOML(masPath)
	if err != nil {
		return fmt.Errorf("failed to parse mas.toml: %w", err)
	}
	
	// Get filter flags
	categoryFilter, _ := cmd.Flags().GetString("category")
	
	// Print header
	fmt.Printf("\n🍎 Mac App Store Applications\n")
	fmt.Printf("Repository: %s\n\n", repo.Root)
	
	// Filter apps if category specified
	apps := masConfig.Apps
	if categoryFilter != "" {
		apps = masConfig.GetByCategory(categoryFilter)
		if len(apps) == 0 {
			fmt.Printf("No apps found in category: %s\n\n", categoryFilter)
			return nil
		}
	}
	
	// Print apps
	fmt.Printf("Found %d app(s)\n", len(apps))
	fmt.Println(strings.Repeat("─", 80))
	
	for _, app := range apps {
		category := app.Category
		if category == "" {
			category = "uncategorized"
		}
		
		fmt.Printf("%-30s [%d]\n", app.Name, app.ID)
		if app.Description != "" {
			fmt.Printf("  %s\n", app.Description)
		}
		fmt.Printf("  Category: %s\n", category)
		fmt.Println()
	}
	
	// Print categories summary
	categories := masConfig.GetCategories()
	if len(categories) > 0 {
		sort.Strings(categories)
		fmt.Printf("📂 Categories: %s\n\n", strings.Join(categories, ", "))
	}
	
	return nil
}

func runListConfigs() error {
	// Find dotfiles repository
	repo, err := config.FindDotfilesRepo()
	if err != nil {
		return fmt.Errorf("dotfiles repository not found: %w", err)
	}
	
	// Get all tools
	tools, err := repo.ListTools()
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}
	
	if len(tools) == 0 {
		fmt.Println("\nNo config tools found in repository.")
		return nil
	}
	
	// Sort tools alphabetically
	sort.Strings(tools)
	
	// Print header
	fmt.Printf("\n⚙️  Available Config Tools\n")
	fmt.Printf("Repository: %s\n\n", repo.Root)
	fmt.Printf("Found %d tool(s)\n", len(tools))
	fmt.Println(strings.Repeat("─", 80))
	
	// Print each tool with details
	for _, tool := range tools {
		// Check if tool has a merlin.toml
		merlinPath := repo.GetToolMerlinConfig(tool)
		hasMerlinConfig := false
		var toolConfig *models.ToolMerlinConfig
		
		if _, err := os.Stat(merlinPath); err == nil {
			hasMerlinConfig = true
			if cfg, err := parser.ParseToolMerlinTOML(merlinPath); err == nil {
				toolConfig = cfg
			}
		}
		
		// Check if config directory exists
		configDir := repo.GetToolConfigDir(tool)
		hasConfigDir := false
		if info, err := os.Stat(configDir); err == nil && info.IsDir() {
			hasConfigDir = true
		}
		
		// Print tool name
		status := "✓"
		if !hasConfigDir && !hasMerlinConfig {
			status = "⚠"
		}
		
		fmt.Printf("%s %-20s", status, tool)
		
		// Print description if available
		if toolConfig != nil && toolConfig.Tool.Description != "" {
			fmt.Printf(" - %s", toolConfig.Tool.Description)
		}
		fmt.Println()
		
		// Print details
		details := []string{}
		if hasMerlinConfig {
			details = append(details, "has merlin.toml")
			
			if toolConfig != nil {
				if toolConfig.HasLinks() {
					details = append(details, fmt.Sprintf("%d link(s)", len(toolConfig.Links)))
				}
				if toolConfig.HasScripts() {
					details = append(details, fmt.Sprintf("%d script(s)", len(toolConfig.Scripts.Scripts)))
				}
				if toolConfig.HasDependencies() {
					details = append(details, fmt.Sprintf("deps: %s", strings.Join(toolConfig.Tool.Dependencies, ", ")))
				}
			}
		}
		if hasConfigDir {
			details = append(details, "has config/")
		}
		
		if len(details) > 0 {
			fmt.Printf("  %s\n", strings.Join(details, ", "))
		}
		fmt.Println()
	}
	
	return nil
}

func printBrewPackages(packages []models.BrewPackage, categoryFilter string) {
	// Group packages by category
	byCategory := make(map[string][]models.BrewPackage)
	for _, pkg := range packages {
		category := pkg.Category
		if category == "" {
			category = "uncategorized"
		}
		
		// Apply category filter
		if categoryFilter != "" && category != categoryFilter {
			continue
		}
		
		byCategory[category] = append(byCategory[category], pkg)
	}
	
	// If filtering and nothing found
	if categoryFilter != "" && len(byCategory) == 0 {
		fmt.Printf("No packages found in category: %s\n", categoryFilter)
		return
	}
	
	// Sort categories
	categories := make([]string, 0, len(byCategory))
	for cat := range byCategory {
		categories = append(categories, cat)
	}
	sort.Strings(categories)
	
	// Print packages grouped by category
	for i, category := range categories {
		if i > 0 {
			fmt.Println()
		}
		
		packages := byCategory[category]
		fmt.Printf("[%s] (%d)\n", category, len(packages))
		
		// Sort packages within category
		sort.Slice(packages, func(i, j int) bool {
			return packages[i].Name < packages[j].Name
		})
		
		for _, pkg := range packages {
			fmt.Printf("  • %-30s", pkg.Name)
			if pkg.Description != "" {
				fmt.Printf(" - %s", pkg.Description)
			}
			fmt.Println()
		}
	}
}

