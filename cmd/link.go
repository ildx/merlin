package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ildx/merlin/internal/cli"
	"github.com/ildx/merlin/internal/config"
	"github.com/ildx/merlin/internal/git"
	"github.com/ildx/merlin/internal/models"
	"github.com/ildx/merlin/internal/parser"
	"github.com/ildx/merlin/internal/scripts"
	"github.com/ildx/merlin/internal/symlink"
	"github.com/spf13/cobra"
)

// buildLinkCommitMessage crafts a concise commit message for auto-commit after linking.
// Format examples:
//
//	chore(link): link zsh
//	chore(link): link zsh, git (2 tools)
//	chore(link): link 5 tools (zsh, git, eza, mise, zellij)
func buildLinkCommitMessage(tools []string) string {
	if len(tools) == 0 {
		return "chore(link): no tools"
	}
	if len(tools) == 1 {
		return fmt.Sprintf("chore(link): link %s", tools[0])
	}
	joined := strings.Join(tools, ", ")
	if len(tools) <= 3 {
		return fmt.Sprintf("chore(link): link %s (%d tools)", joined, len(tools))
	}
	// For many tools, keep message short and list first 3 + ellipsis
	preview := strings.Join(tools[:3], ", ")
	return fmt.Sprintf("chore(link): link %d tools (%s, …)", len(tools), preview)
}

var (
	linkStrategy     string
	linkAll          bool
	linkRunScripts   bool
	linkProfile      string
	linkNoAutoCommit bool // per-invocation override for auto-commit
)

var linkCmd = &cobra.Command{
	Use:   "link [tool]",
	Short: "Create symlinks for dotfiles",
	Long: `Create symbolic links from your dotfiles repository to target locations.

BEHAVIOR
	• Without flags: link a single tool's configuration.
	• --all links every discovered tool.
	• --profile filters tools by a named profile from root merlin.toml.
	• Variable placeholders in targets (e.g. {home_dir}) are expanded.

CONFLICT STRATEGIES
	skip (default)    Leave existing files untouched
	backup            Move existing file to .backup.<timestamp>
	overwrite         Replace existing file/symlink

FLAGS
	--all             Link all tools
	--strategy <s>    Conflict strategy (skip|backup|overwrite)
	--run-scripts     Run tool scripts after linking (if defined)
	--profile <name>  Filter tools to profile list
	--dry-run         Preview actions only
	--verbose,-v      Detailed per-link output

EXAMPLES
	merlin link git                            # Link git configs
	merlin link zsh --dry-run                  # Preview linking
	merlin link eza --strategy backup          # Backup existing files
	merlin link --all                          # Link everything
	merlin link --all --profile personal       # Profile-filtered batch
	merlin link zellij --run-scripts           # Link + run scripts

SEE ALSO
	merlin unlink   Remove symlinks
	merlin validate Validate configurations
	merlin list     Overview of tools`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Parse strategy
		strategy, err := symlink.ParseStrategy(linkStrategy)
		if err != nil {
			cli.Error("%v", err)
			os.Exit(1)
		}

		// Find dotfiles repo
		repo, err := config.FindDotfilesRepo()
		if err != nil {
			cli.Error("%v", err)
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
			cli.Error("parsing root config: %v", err)
			os.Exit(1)
		}

		// Get variables
		vars, err := symlink.GetVariablesFromRoot(rootConfig)
		if err != nil {
			cli.Error("getting variables: %v", err)
			os.Exit(1)
		}

		processedTools := []string{}
		if linkAll || linkProfile != "" {
			processedTools = runLinkAll(repo, vars, strategy, dryRun, verbose, linkRunScripts, rootConfig)
		} else if len(args) == 1 {
			runLinkTool(repo, args[0], vars, strategy, dryRun, verbose, linkRunScripts)
			processedTools = append(processedTools, args[0])
		} else {
			cmd.Help()
			os.Exit(0)
		}

		// Auto-commit hook (Phase 13 integration + safety) unless overridden
		if rootConfig.Settings.AutoCommit && !linkNoAutoCommit && !dryRun && git.IsGitAvailable() {
			if len(processedTools) > 0 {
				if repoGit, err := git.Open(rootConfigPathDir(repo)); err == nil {
					paths := make([]string, 0, len(processedTools))
					for _, t := range processedTools {
						paths = append(paths, filepath.Join("config", t))
					}
					// Safety: abort if unrelated unstaged/untracked changes outside allowed paths
					if unrelated, uErr := repoGit.HasUnrelatedChanges(paths); uErr == nil && unrelated {
						cli.Warning("auto-commit skipped: unrelated changes detected outside tool directories")
					} else {
						paths = repoGit.FilterPaths(paths)
						msg := buildLinkCommitMessage(processedTools)
						if err := repoGit.Commit(msg, paths); err != nil {
							if strings.Contains(err.Error(), "no staged changes") {
								// Allow empty commit for traceability
								cmd := exec.Command("git", "-C", repoGit.Root, "commit", "--allow-empty", "-m", msg)
								if e2 := cmd.Run(); e2 != nil {
									cli.Warning("auto-commit skipped (no changes): %v", err)
								} else {
									cli.Success("Auto-commit created (%s)", msg)
								}
							} else {
								cli.Warning("auto-commit failed: %v", err)
							}
						} else {
							cli.Success("Auto-commit created (%s)", msg)
						}
					}
				}
			}
		}
	},
}

// rootConfigPathDir extracts repo root directory from DotfilesRepo
func rootConfigPathDir(repo *config.DotfilesRepo) string { return repo.Root }

func init() {
	rootCmd.AddCommand(linkCmd)
	linkCmd.Flags().StringVar(&linkStrategy, "strategy", "skip", "Conflict resolution strategy (skip, backup, overwrite)")
	linkCmd.Flags().BoolVar(&linkAll, "all", false, "Link all discovered configs")
	linkCmd.Flags().BoolVar(&linkRunScripts, "run-scripts", false, "Run tool scripts after linking")
	linkCmd.Flags().StringVar(&linkProfile, "profile", "", "Use specific profile to filter tools")
	linkCmd.Flags().BoolVar(&linkNoAutoCommit, "no-auto-commit", false, "Disable auto-commit even if enabled in settings")
}

func runLinkTool(repo *config.DotfilesRepo, toolName string, vars symlink.Variables, strategy symlink.ConflictStrategy, dryRun, verbose, runScripts bool) {
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
		cli.Warning("linking tool: %v", err)
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
		cli.Warning("Failed to parse %s: %v", merlinPath, err)
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
		cli.Warning("Failed to run scripts: %v", err)
		return
	}

	// Display results
	for _, result := range scriptResults {
		fmt.Println(scripts.FormatScriptResult(result, verbose))
	}
}

func runLinkAll(repo *config.DotfilesRepo, vars symlink.Variables, strategy symlink.ConflictStrategy, dryRun, verbose, runScripts bool, rootConfig *models.RootMerlinConfig) []string {
	// Discover all tools
	tools, err := symlink.DiscoverTools(repo, vars)
	if err != nil {
		cli.Error("discovering tools: %v", err)
		os.Exit(1)
	}

	if len(tools) == 0 {
		fmt.Println("No tools found to link")
		return []string{}
	}

	// Filter by profile if specified
	if linkProfile != "" {
		profile := rootConfig.GetProfileByName(linkProfile)
		if profile == nil {
			cli.Error("Profile '%s' not found", linkProfile)
			os.Exit(1)
		}

		// Filter tools to only those in profile
		if len(profile.Tools) > 0 {
			filteredTools := make([]*symlink.ToolConfig, 0)
			profileToolSet := make(map[string]bool)
			for _, name := range profile.Tools {
				profileToolSet[name] = true
			}

			for _, tool := range tools {
				if profileToolSet[tool.Name] {
					filteredTools = append(filteredTools, tool)
				}
			}

			tools = filteredTools
			fmt.Printf("Using profile '%s' (%d tools)\n\n", linkProfile, len(tools))
		}
	}

	if len(tools) == 0 {
		fmt.Println("No tools found to link (after profile filtering)")
		return []string{}
	}

	fmt.Printf("Linking %d tools\n\n", len(tools))

	successCount := 0
	skipCount := 0
	errorCount := 0
	conflictCount := 0

	processed := []string{}
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
		processed = append(processed, tool.Name)
	}

	// Summary
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("Summary: %d linked, %d skipped, %d conflicts, %d errors\n",
		successCount, skipCount, conflictCount, errorCount)

	if dryRun {
		fmt.Println("\nThis was a dry run. No changes were made.")
	}
	return processed
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
