package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	version = "0.1.0-dev"
	appName = "Merlin"
)

var rootCmd = &cobra.Command{
	Use:   "merlin",
	Short: "A macOS-focused CLI tool for managing dotfiles with style ✨",
	Long: `Merlin is a powerful yet simple dotfiles manager that handles package 
installation, configuration symlinking, and custom setup scripts—all from 
declarative TOML files.

Built with Go and Charm for a beautiful terminal experience.`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, show help
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Show what would be done without doing it")
	
	// Hide the default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

