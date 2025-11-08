package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ildx/merlin/internal/config"
	"github.com/ildx/merlin/internal/parser"
	"github.com/ildx/merlin/internal/scripts"
	"github.com/ildx/merlin/internal/symlink"
	"github.com/spf13/cobra"
)

var (
	linkStrategy  string
	linkAll       bool
	linkRunScripts bool
)

var linkCmd = &cobra.Command{
	Use:   "link [tool]",
	Short: "Create symlinks for dotfiles",
	Long: `Create symbolic links from your dotfiles to their target locations.
	
By default, this command links a specific tool's configuration. You can also
use --all to link all discovered configs.

Examples:
  merlin link git              # Link git config
  merlin link --all            # Link all configs
  merlin link zsh --dry-run    # Preview zsh linking
  merlin link eza --strategy backup  # Link with backup strategy`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Parse strategy
		strategy, err := symlink.ParseStrategy(linkStrategy)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Find dotfiles repo
		repo, err := config.FindDotfilesRepo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if verbose {
			fmt.Printf("Dotfiles repository: %s\n", repo.Root)
			fmt.Printf("Conflict strategy: %s\n", strategy)
			if dryRun {
				fmt.Println("Mode: Dry run (no changes will be made)")
			}
			fmt.Println()
		}

		// Load root config for variables
		rootConfigPath := repo.GetRootMerlinConfig()
		rootConfig, err := parser.ParseRootMerlinTOML(rootConfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing root config: %v\n", err)
			os.Exit(1)
		}

		// Get variables
		vars, err := symlink.GetVariablesFromRoot(rootConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting variables: %v\n", err)
			os.Exit(1)
		}

		if linkAll {
			runLinkAll(repo, vars, strategy, dryRun, verbose, linkRunScripts)
		} else if len(args) == 1 {
			runLinkTool(repo, args[0], vars, strategy, dryRun, verbose, linkRunScripts)
		} else {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	rootCmd.AddCommand(linkCmd)
	linkCmd.Flags().StringVar(&linkStrategy, "strategy", "skip", "Conflict resolution strategy (skip, backup, overwrite)")
	linkCmd.Flags().BoolVar(&linkAll, "all", false, "Link all discovered configs")
	linkCmd.Flags().BoolVar(&linkRunScripts, "run-scripts", false, "Run tool scripts after linking")
}

func runLinkTool(repo *config.DotfilesRepo, toolName string, vars symlink.Variables, strategy symlink.ConflictStrategy, dryRun, verbose, runScripts bool) {
	// Check if tool exists
	if !repo.ToolExists(toolName) {
		fmt.Fprintf(os.Stderr, "Error: Tool '%s' not found in dotfiles repository\n", toolName)
		os.Exit(1)
	}

	// Discover tool config
	tool, err := symlink.DiscoverToolConfig(repo, toolName, vars)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering tool config: %v\n", err)
		os.Exit(1)
	}

	if len(tool.Links) == 0 {
		fmt.Printf("No links configured for %s\n", toolName)
		return
	}

	// Display tool info
	fmt.Printf("Linking %s", toolName)
	if tool.Description != "" {
		fmt.Printf(" - %s", tool.Description)
	}
	fmt.Println()

	if verbose {
		fmt.Printf("  Links to create: %d\n", len(tool.Links))
		for i, link := range tool.Links {
			fmt.Printf("  %d. %s → %s\n", i+1, link.Source, link.Target)
		}
		fmt.Println()
	}

	// Link the tool
	results, err := symlink.LinkToolWithStrategy(tool, strategy, dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error linking tool: %v\n", err)
	}

	// Display results
	displayLinkResults(results, verbose)

	// Run post-link scripts if requested
	if runScripts {
		runPostLinkScripts(repo, toolName, vars, dryRun, verbose)
	}
}

func runPostLinkScripts(repo *config.DotfilesRepo, toolName string, vars symlink.Variables, dryRun, verbose bool) {
	// Parse tool's merlin.toml
	merlinPath := repo.GetToolMerlinConfig(toolName)
	toolConfig, err := parser.ParseToolMerlinTOML(merlinPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: Failed to parse %s: %v\n", merlinPath, err)
		return
	}

	// Check if tool has scripts
	if !toolConfig.HasScripts() {
		return
	}

	fmt.Println("\n📜 Running post-link scripts...")

	// Create environment for scripts
	toolRoot := repo.GetToolRoot(toolName)
	env := scripts.GetDefaultEnvironment(toolRoot, toolName, vars.HomeDir, vars.ConfigDir)

	// Run scripts
	runner := scripts.NewScriptRunner(toolRoot, env, dryRun, verbose, os.Stdout)
	scriptResults, err := runner.RunScripts(toolConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to run scripts: %v\n", err)
		return
	}

	// Display results
	for _, result := range scriptResults {
		fmt.Println(scripts.FormatScriptResult(result, verbose))
	}
}

func runLinkAll(repo *config.DotfilesRepo, vars symlink.Variables, strategy symlink.ConflictStrategy, dryRun, verbose, runScripts bool) {
	// Discover all tools
	tools, err := symlink.DiscoverTools(repo, vars)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering tools: %v\n", err)
		os.Exit(1)
	}

	if len(tools) == 0 {
		fmt.Println("No tools found to link")
		return
	}

	fmt.Printf("Linking %d tools\n\n", len(tools))

	successCount := 0
	skipCount := 0
	errorCount := 0
	conflictCount := 0

	for _, tool := range tools {
		if len(tool.Links) == 0 {
			continue
		}

		fmt.Printf("Linking %s", tool.Name)
		if tool.Description != "" {
			fmt.Printf(" - %s", tool.Description)
		}
		fmt.Println()

		results, _ := symlink.LinkToolWithStrategy(tool, strategy, dryRun)

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
					fmt.Printf("  ⊘ %s (skipped)\n", result.Target)
				}
			case symlink.LinkStatusAlreadyLinked:
				successCount++
				if verbose {
					fmt.Printf("  ✓ %s (already linked)\n", result.Target)
				}
			case symlink.LinkStatusConflict:
				conflictCount++
				if verbose {
					fmt.Printf("  ⚠ %s (conflict: %s)\n", result.Target, result.Message)
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
				if r.Status == symlink.LinkStatusSuccess || r.Status == symlink.LinkStatusAlreadyLinked {
					toolSuccess++
				} else if r.Status == symlink.LinkStatusSkipped || r.Status == symlink.LinkStatusConflict {
					toolSkip++
				} else {
					toolError++
				}
			}
			fmt.Printf("  %d linked, %d skipped, %d errors\n", toolSuccess, toolSkip, toolError)
		}

		fmt.Println()

		// Run post-link scripts if requested
		if runScripts {
			runPostLinkScripts(repo, tool.Name, vars, dryRun, verbose)
		}
	}

	// Summary
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("Summary: %d linked, %d skipped, %d conflicts, %d errors\n",
		successCount, skipCount, conflictCount, errorCount)

	if dryRun {
		fmt.Println("\nThis was a dry run. No changes were made.")
	}
}

func displayLinkResults(results []*symlink.LinkResult, verbose bool) {
	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, result := range results {
		switch result.Status {
		case symlink.LinkStatusSuccess:
			successCount++
			symbol := "✓"
			if verbose {
				fmt.Printf("  %s %s\n", symbol, result.Target)
				fmt.Printf("    → %s\n", result.Source)
			} else {
				fmt.Printf("  %s %s\n", symbol, result.Target)
			}
		case symlink.LinkStatusSkipped:
			skipCount++
			fmt.Printf("  ⊘ %s (skipped)\n", result.Target)
		case symlink.LinkStatusAlreadyLinked:
			successCount++
			fmt.Printf("  ✓ %s (already linked)\n", result.Target)
		case symlink.LinkStatusConflict:
			skipCount++
			fmt.Printf("  ⚠ %s (conflict: %s)\n", result.Target, result.Message)
		case symlink.LinkStatusError:
			errorCount++
			fmt.Printf("  ✗ %s (error: %s)\n", result.Target, result.Message)
		}
	}

	fmt.Println()
	fmt.Printf("Summary: %d linked, %d skipped, %d errors\n",
		successCount, skipCount, errorCount)
}

