package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ildx/merlin/internal/cli"
	"github.com/ildx/merlin/internal/config"
	"github.com/ildx/merlin/internal/installer"
	"github.com/ildx/merlin/internal/models"
	"github.com/ildx/merlin/internal/parser"
	"github.com/ildx/merlin/internal/system"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install packages and apps",
	Long: `Install Homebrew packages and Mac App Store applications defined in TOML.

SUBCOMMANDS
	brew   Install Homebrew formulae & casks from brew.toml
	mas    Install Mac App Store apps from mas.toml

BEHAVIOR
	Interactive selector is shown unless --all or --dry-run is used.
	Already-installed items are skipped automatically.

FLAGS (brew)
	--all            Install all formulae & casks without prompting
	--formulae-only  Only install formulae
	--casks-only     Only install casks
	--dry-run        Show what would be installed
	--verbose,-v     More detailed output

FLAGS (mas)
	--all            Install all apps without prompting
	--dry-run        Preview actions only
	--verbose,-v     More detailed output

EXAMPLES
	merlin install brew                 # Interactive picker
	merlin install brew --all           # Install everything
	merlin install brew --formulae-only # Only CLI tools
	merlin install mas                  # Interactive MAS selection
	merlin install mas --all --dry-run  # Preview full install

NOTES
	• For MAS installs you must be signed into the App Store.
	• Use merlin list brew|mas to inspect definitions first.
	• Combine with --dry-run to stage changes safely.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var installBrewCmd = &cobra.Command{
	Use:   "brew",
	Short: "Install Homebrew packages",
	Long: `Install Homebrew formulae and casks from brew.toml

By default, this command will interactively prompt you to select which packages to install.
Use --all to install all packages without prompting.
Use --dry-run to preview what would be installed without actually installing.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInstallBrew(cmd); err != nil {
			cli.Error("%v", err)
			os.Exit(1)
		}
	},
}

var installMASCmd = &cobra.Command{
	Use:   "mas",
	Short: "Install Mac App Store apps",
	Long: `Install Mac App Store applications from mas.toml

By default, this command will interactively prompt you to select which apps to install.
Use --all to install all apps without prompting.
Use --dry-run to preview what would be installed without actually installing.

Note: You must be signed into the Mac App Store for installation to work.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInstallMAS(cmd); err != nil {
			cli.Error("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.AddCommand(installBrewCmd)
	installCmd.AddCommand(installMASCmd)

	// Brew flags
	installBrewCmd.Flags().Bool("formulae-only", false, "Install only formulae")
	installBrewCmd.Flags().Bool("casks-only", false, "Install only casks")
	installBrewCmd.Flags().Bool("all", false, "Install all packages without prompting")

	// MAS flags
	installMASCmd.Flags().Bool("all", false, "Install all apps without prompting")
}

func runInstallBrew(cmd *cobra.Command) error {
	// Get flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	verbose, _ := cmd.Flags().GetBool("verbose")
	formulaeOnly, _ := cmd.Flags().GetBool("formulae-only")
	casksOnly, _ := cmd.Flags().GetBool("casks-only")
	installAll, _ := cmd.Flags().GetBool("all")

	// Check prerequisites
	fmt.Println("\n🔍 Checking prerequisites...")
	brewCheck := system.CheckHomebrew()
	if !brewCheck.Exists {
		return fmt.Errorf("Homebrew is not installed. Install it from https://brew.sh")
	}
	fmt.Printf("   ✓ Homebrew found: %s\n", brewCheck.Version)

	// Find dotfiles repository
	fmt.Println("\n📂 Finding dotfiles repository...")
	repo, err := config.FindDotfilesRepo()
	if err != nil {
		return fmt.Errorf("dotfiles repository not found: %w", err)
	}
	fmt.Printf("   ✓ Found: %s\n", repo.Root)

	// Find and parse brew.toml
	fmt.Println("\n📋 Loading package list...")
	brewPath := filepath.Join(repo.GetToolConfigDir("brew"), "brew.toml")
	if _, err := os.Stat(brewPath); os.IsNotExist(err) {
		return fmt.Errorf("brew.toml not found at %s", brewPath)
	}

	brewConfig, err := parser.ParseBrewTOML(brewPath)
	if err != nil {
		return fmt.Errorf("failed to parse brew.toml: %w", err)
	}

	totalPackages := len(brewConfig.Formulae) + len(brewConfig.Casks)
	fmt.Printf("   ✓ Found %d packages (%d formulae, %d casks)\n",
		totalPackages, len(brewConfig.Formulae), len(brewConfig.Casks))

	// Filter packages based on flags
	var formulae, casks []models.BrewPackage
	if !casksOnly {
		formulae = brewConfig.Formulae
	}
	if !formulaeOnly {
		casks = brewConfig.Casks
	}

	if len(formulae) == 0 && len(casks) == 0 {
		fmt.Println("\n⚠️  No packages to install (check your flags)")
		return nil
	}

	// Interactive selection (unless --all is specified or dry-run)
	if !installAll && !dryRun {
		var err error

		// Select formulae
		if len(formulae) > 0 {
			formulae, err = installer.SelectPackages(formulae, "🔧 Formulae", os.Stdin, os.Stdout)
			if err != nil {
				return fmt.Errorf("failed to select formulae: %w", err)
			}
		}

		// Select casks
		if len(casks) > 0 {
			casks, err = installer.SelectPackages(casks, "📱 Casks", os.Stdin, os.Stdout)
			if err != nil {
				return fmt.Errorf("failed to select casks: %w", err)
			}
		}

		// Check if anything was selected
		if len(formulae) == 0 && len(casks) == 0 {
			fmt.Println("\n⚠️  No packages selected. Exiting.")
			return nil
		}

		// Confirm installation
		confirmed, err := installer.ConfirmInstallation(len(formulae), len(casks), os.Stdin, os.Stdout)
		if err != nil {
			return fmt.Errorf("failed to get confirmation: %w", err)
		}
		if !confirmed {
			fmt.Println("\n❌ Installation cancelled.")
			return nil
		}
	}

	// Dry run notification
	if dryRun {
		fmt.Println("\n🔍 DRY RUN MODE - No packages will be installed")
	}

	// Create installer
	brewInstaller := installer.NewBrewInstaller(dryRun, verbose)

	// Install packages
	fmt.Printf("\n%s\n", strings.Repeat("═", 80))
	fmt.Println("Starting Installation")
	fmt.Println(strings.Repeat("═", 80))

	var formulaeResults, caskResults []*installer.InstallResult

	// Install formulae
	if len(formulae) > 0 {
		formulaeResults = brewInstaller.InstallFormulae(formulae, os.Stdout)
	}

	// Install casks
	if len(casks) > 0 {
		caskResults = brewInstaller.InstallCasks(casks, os.Stdout)
	}

	// Print summary
	installer.PrintSummary(formulaeResults, caskResults, os.Stdout)

	return nil
}

func runInstallMAS(cmd *cobra.Command) error {
	// Get flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	verbose, _ := cmd.Flags().GetBool("verbose")
	installAll, _ := cmd.Flags().GetBool("all")

	// Check prerequisites
	fmt.Println("\n🔍 Checking prerequisites...")

	// Check if mas-cli is installed
	masCheck := system.CheckMAS()
	if !masCheck.Exists {
		return fmt.Errorf("mas-cli is not installed. Install it with: brew install mas")
	}
	fmt.Printf("   ✓ mas-cli found: %s\n", masCheck.Version)

	// Check if signed into Mac App Store
	masInstaller := installer.NewMASInstaller(dryRun, verbose)
	signedIn, account, err := masInstaller.CheckMASAccount()
	if err != nil {
		return fmt.Errorf("failed to check Mac App Store account: %w", err)
	}

	if !signedIn {
		fmt.Println("\n❌ You are not signed into the Mac App Store")
		fmt.Println("\n💡 To sign in:")
		fmt.Println("   1. Open the App Store application")
		fmt.Println("   2. Sign in with your Apple ID")
		fmt.Println("   3. Run this command again")
		return fmt.Errorf("not signed into Mac App Store")
	}
	fmt.Printf("   ✓ Signed in as: %s\n", account)

	// Find dotfiles repository
	fmt.Println("\n📂 Finding dotfiles repository...")
	repo, err := config.FindDotfilesRepo()
	if err != nil {
		return fmt.Errorf("dotfiles repository not found: %w", err)
	}
	fmt.Printf("   ✓ Found: %s\n", repo.Root)

	// Find and parse mas.toml
	fmt.Println("\n📋 Loading app list...")
	masPath := filepath.Join(repo.GetToolConfigDir("mas"), "mas.toml")
	if _, err := os.Stat(masPath); os.IsNotExist(err) {
		return fmt.Errorf("mas.toml not found at %s", masPath)
	}

	masConfig, err := parser.ParseMASTOML(masPath)
	if err != nil {
		return fmt.Errorf("failed to parse mas.toml: %w", err)
	}

	if len(masConfig.Apps) == 0 {
		fmt.Println("\n⚠️  No apps found in mas.toml")
		return nil
	}

	fmt.Printf("   ✓ Found %d app(s)\n", len(masConfig.Apps))

	// Get apps list
	apps := masConfig.Apps

	// Interactive selection (unless --all is specified or dry-run)
	if !installAll && !dryRun {
		var err error

		// Select apps
		apps, err = installer.SelectMASApps(apps, os.Stdin, os.Stdout)
		if err != nil {
			return fmt.Errorf("failed to select apps: %w", err)
		}

		// Check if anything was selected
		if len(apps) == 0 {
			fmt.Println("\n⚠️  No apps selected. Exiting.")
			return nil
		}

		// Confirm installation
		confirmed, err := installer.ConfirmMASInstallation(len(apps), os.Stdin, os.Stdout)
		if err != nil {
			return fmt.Errorf("failed to get confirmation: %w", err)
		}
		if !confirmed {
			fmt.Println("\n❌ Installation cancelled.")
			return nil
		}
	}

	// Dry run notification
	if dryRun {
		fmt.Println("\n🔍 DRY RUN MODE - No apps will be installed")
	}

	// Install apps
	fmt.Printf("\n%s\n", strings.Repeat("═", 80))
	fmt.Println("Starting Installation")
	fmt.Println(strings.Repeat("═", 80))

	results := masInstaller.InstallApps(apps, os.Stdout)

	// Print summary
	installer.PrintMASSummary(results, os.Stdout)

	return nil
}
