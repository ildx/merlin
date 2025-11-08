package cmd

import (
	"fmt"
	"os"

	"github.com/ildx/merlin/internal/config"
	"github.com/ildx/merlin/internal/parser"
	"github.com/ildx/merlin/internal/scripts"
	"github.com/ildx/merlin/internal/symlink"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [tool]",
	Short: "Run setup scripts for a tool",
	Long: `Execute setup scripts defined in a tool's merlin.toml configuration.
	
This command runs the scripts defined in the [scripts] section of a tool's
configuration. Scripts are executed in the order they appear.

Examples:
  merlin run zellij              # Run zellij setup scripts
  merlin run cursor --dry-run    # Preview cursor scripts
  merlin run git --verbose       # Run git scripts with verbose output`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verbose, _ := cmd.Flags().GetBool("verbose")

		toolName := args[0]

		if err := runToolScripts(toolName, dryRun, verbose); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runToolScripts(toolName string, dryRun, verbose bool) error {
	// Find dotfiles repo
	repo, err := config.FindDotfilesRepo()
	if err != nil {
		return fmt.Errorf("dotfiles repository not found: %w", err)
	}

	if verbose {
		fmt.Printf("Dotfiles repository: %s\n", repo.Root)
		if dryRun {
			fmt.Println("Mode: Dry run (no scripts will be executed)")
		}
		fmt.Println()
	}

	// Check if tool exists
	if !repo.ToolExists(toolName) {
		return fmt.Errorf("tool '%s' not found in dotfiles repository", toolName)
	}

	// Parse tool's merlin.toml
	merlinPath := repo.GetToolMerlinConfig(toolName)
	toolConfig, err := parser.ParseToolMerlinTOML(merlinPath)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", merlinPath, err)
	}

	// Check if tool has scripts
	if !toolConfig.HasScripts() {
		fmt.Printf("Tool '%s' has no scripts configured\n", toolName)
		return nil
	}

	// Get environment variables
	rootConfigPath := repo.GetRootMerlinConfig()
	rootConfig, err := parser.ParseRootMerlinTOML(rootConfigPath)
	if err != nil {
		return fmt.Errorf("failed to parse root config: %w", err)
	}

	vars, err := symlink.GetVariablesFromRoot(rootConfig)
	if err != nil {
		return fmt.Errorf("failed to get variables: %w", err)
	}

	// Create environment for scripts
	toolRoot := repo.GetToolRoot(toolName)
	env := scripts.GetDefaultEnvironment(toolRoot, toolName, vars.HomeDir, vars.ConfigDir)

	// Display tool info
	fmt.Printf("Running scripts for %s", toolName)
	if toolConfig.Tool.Description != "" {
		fmt.Printf(" - %s", toolConfig.Tool.Description)
	}
	fmt.Println()

	if verbose {
		fmt.Printf("  Script directory: %s\n", toolConfig.Scripts.Directory)
		fmt.Printf("  Scripts to run: %d\n", len(toolConfig.Scripts.Scripts))
		for i, script := range toolConfig.Scripts.Scripts {
			fmt.Printf("    %d. %s\n", i+1, script)
		}
		fmt.Println()
	}

	// Validate scripts first
	if errors := scripts.ValidateScripts(toolRoot, toolConfig); len(errors) > 0 {
		fmt.Println("\n⚠️  Script validation errors:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("script validation failed")
	}

	// Run scripts
	runner := scripts.NewScriptRunner(toolRoot, env, dryRun, verbose, os.Stdout)
	results, err := runner.RunScripts(toolConfig)
	if err != nil {
		return fmt.Errorf("failed to run scripts: %w", err)
	}

	// Display results
	if !verbose {
		fmt.Println()
		for _, result := range results {
			fmt.Println(scripts.FormatScriptResult(result, verbose))
		}
	}

	// Summary
	successCount := 0
	failureCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	fmt.Println()
	if failureCount > 0 {
		fmt.Printf("Summary: %d succeeded, %d failed\n", successCount, failureCount)
		return fmt.Errorf("some scripts failed")
	} else {
		fmt.Printf("Summary: All %d scripts completed successfully\n", successCount)
	}

	if dryRun {
		fmt.Println("\nThis was a dry run. No scripts were executed.")
	}

	return nil
}

