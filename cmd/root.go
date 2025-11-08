package cmd

import (
	"os"

	"github.com/ildx/merlin/internal/cli"
	"github.com/ildx/merlin/internal/logger"
	"github.com/spf13/cobra"
)

const (
	version = "0.1.0-dev"
	appName = "Merlin"
)

var rootCmd = &cobra.Command{
	Use:   "merlin",
	Short: "A macOS-focused CLI tool for managing dotfiles with style ✨",
	Long: `Merlin is a powerful yet simple dotfiles manager that handles:

	• Package installation (Homebrew & Mac App Store)
	• Configuration symlinking with conflict strategies
	• Profile-based environment filtering
	• Custom setup scripts
	• Interactive TUI for a better terminal experience

CONFIG SOURCE
	Declarative TOML files inside your dotfiles repository.

GLOBAL FLAGS
	--dry-run    Preview actions without changing the system
	--verbose,-v More detailed output & debug logging

EXAMPLES
	merlin                 # Launch interactive TUI
	merlin link zsh        # Link one tool
	merlin link --all      # Link everything
	merlin install brew    # Install Homebrew packages
	merlin validate --strict

SEE ALSO
	merlin tui             # Explicitly launch TUI
	docs/USAGE.md          # Detailed usage guide
	merlin doctor          # System prerequisite checks

Built with Go and Charm for a beautiful terminal experience.`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, launch TUI
		if err := runTUI(); err != nil {
			cli.Error("%v", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Initialize logging
	verbose, _ := rootCmd.Flags().GetBool("verbose")
	if err := logger.Init(logger.LevelInfo, verbose); err != nil {
		cli.Warning("Failed to initialize logging: %v", err)
	}

	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed", "error", err)
		cli.Error("%v", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Show what would be done without doing it")

	// Initialize logging early
	cobra.OnInitialize(initLogging)

	// Hide the default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func initLogging() {
	verbose, _ := rootCmd.Flags().GetBool("verbose")
	if err := logger.Init(logger.LevelInfo, verbose); err != nil {
		// Non-fatal - just print warning
		cli.Warning("Failed to initialize logging: %v", err)
	}

	logger.Debug("Merlin starting", "version", version)
}
