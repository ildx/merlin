package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ildx/merlin/internal/cli"
	"github.com/ildx/merlin/internal/config"
	"github.com/ildx/merlin/internal/parser"
	"github.com/ildx/merlin/internal/symlink"
	"github.com/spf13/cobra"
)

var unlinkAll bool

var unlinkCmd = &cobra.Command{
	Use:   "unlink [tool]",
	Short: "Remove symlinks for dotfiles",
	Long: `Remove symlinks previously created by merlin.

SAFETY
	• Only removes symlinks that point back into your dotfiles repo
	• Regular files / foreign symlinks are left untouched

FLAGS
	--all        Unlink all discovered tools
	--dry-run    Preview what would be removed
	--verbose    Show each evaluated path

EXAMPLES
	merlin unlink git            # Remove git links
	merlin unlink zsh --dry-run  # Preview zsh unlinking
	merlin unlink --all          # Remove all links

TIPS
	Run 'merlin link --all' again to restore after a dry run preview.
	Use 'merlin validate' to ensure configs before re-linking.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Find dotfiles repo
		repo, err := config.FindDotfilesRepo()
		if err != nil {
			cli.Error("%v", err)
			os.Exit(1)
		}

		if verbose {
			fmt.Printf("Dotfiles repository: %s\n", repo.Root)
			if dryRun {
				fmt.Println("Mode: Dry run (no changes will be made)")
			}
			fmt.Println()
		}

		// Load root config for variables
		rootConfigPath := repo.GetRootMerlinConfig()
		rootConfig, err := parser.ParseRootMerlinTOML(rootConfigPath)
		if err != nil {
			cli.Error("parsing root config: %v", err)
			os.Exit(1)
		}

		// Get variables
		vars, err := symlink.GetVariablesFromRoot(rootConfig)
		if err != nil {
			cli.Error("getting variables: %v", err)
			os.Exit(1)
		}

		if unlinkAll {
			runUnlinkAll(repo, vars, dryRun, verbose)
		} else if len(args) == 1 {
			runUnlinkTool(repo, args[0], vars, dryRun, verbose)
		} else {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	rootCmd.AddCommand(unlinkCmd)
	unlinkCmd.Flags().BoolVar(&unlinkAll, "all", false, "Unlink all discovered configs")
}

func runUnlinkTool(repo *config.DotfilesRepo, toolName string, vars symlink.Variables, dryRun, verbose bool) {
	// Check if tool exists
	if !repo.ToolExists(toolName) {
		cli.Error("Tool '%s' not found in dotfiles repository", toolName)
		os.Exit(1)
	}

	// Discover tool config
	tool, err := symlink.DiscoverToolConfig(repo, toolName, vars)
	if err != nil {
		cli.Error("discovering tool config: %v", err)
		os.Exit(1)
	}

	if len(tool.Links) == 0 {
		fmt.Printf("No links configured for %s\n", toolName)
		return
	}

	// Display tool info
	fmt.Printf("Unlinking %s", toolName)
	if tool.Description != "" {
		fmt.Printf(" - %s", tool.Description)
	}
	fmt.Println()

	if verbose {
		fmt.Printf("  Links to remove: %d\n", len(tool.Links))
		for i, link := range tool.Links {
			fmt.Printf("  %d. %s\n", i+1, link.Target)
		}
		fmt.Println()
	}

	// Unlink the tool
	results, err := symlink.UnlinkTool(tool, dryRun)
	if err != nil {
		cli.Warning("unlinking tool: %v", err)
	}

	// Display results
	displayUnlinkResults(results, verbose)
}

func runUnlinkAll(repo *config.DotfilesRepo, vars symlink.Variables, dryRun, verbose bool) {
	// Discover all tools
	tools, err := symlink.DiscoverTools(repo, vars)
	if err != nil {
		cli.Error("discovering tools: %v", err)
		os.Exit(1)
	}

	if len(tools) == 0 {
		fmt.Println("No tools found to unlink")
		return
	}

	fmt.Printf("Unlinking %d tools\n\n", len(tools))

	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, tool := range tools {
		if len(tool.Links) == 0 {
			continue
		}

		fmt.Printf("Unlinking %s", tool.Name)
		if tool.Description != "" {
			fmt.Printf(" - %s", tool.Description)
		}
		fmt.Println()

		results, _ := symlink.UnlinkTool(tool, dryRun)

		for _, result := range results {
			switch result.Status {
			case symlink.LinkStatusSuccess:
				successCount++
				if verbose {
					fmt.Printf("  ✓ %s\n", result.Target)
				}
			case symlink.LinkStatusSkipped:
				skipCount++
				if verbose {
					fmt.Printf("  ⊘ %s (%s)\n", result.Target, result.Message)
				}
			case symlink.LinkStatusError:
				errorCount++
				fmt.Printf("  ✗ %s (error: %s)\n", result.Target, result.Message)
			}
		}

		if !verbose {
			// Show summary per tool
			toolSuccess := 0
			toolSkip := 0
			toolError := 0
			for _, r := range results {
				if r.Status == symlink.LinkStatusSuccess {
					toolSuccess++
				} else if r.Status == symlink.LinkStatusSkipped {
					toolSkip++
				} else {
					toolError++
				}
			}
			fmt.Printf("  %d removed, %d skipped, %d errors\n", toolSuccess, toolSkip, toolError)
		}

		fmt.Println()
	}

	// Summary
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("Summary: %d removed, %d skipped, %d errors\n",
		successCount, skipCount, errorCount)

	if dryRun {
		fmt.Println("\nThis was a dry run. No changes were made.")
	}
}

func displayUnlinkResults(results []*symlink.UnlinkResult, verbose bool) {
	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, result := range results {
		switch result.Status {
		case symlink.LinkStatusSuccess:
			successCount++
			if verbose {
				fmt.Printf("  ✓ %s (removed)\n", result.Target)
			} else {
				fmt.Printf("  ✓ %s\n", result.Target)
			}
		case symlink.LinkStatusSkipped:
			skipCount++
			fmt.Printf("  ⊘ %s (%s)\n", result.Target, result.Message)
		case symlink.LinkStatusError:
			errorCount++
			fmt.Printf("  ✗ %s (error: %s)\n", result.Target, result.Message)
		}
	}

	fmt.Println()
	fmt.Printf("Summary: %d removed, %d skipped, %d errors\n",
		successCount, skipCount, errorCount)
}
