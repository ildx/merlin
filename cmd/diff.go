package cmd

import (
	"fmt"
	"os"

	"github.com/ildx/merlin/internal/cli"
	"github.com/ildx/merlin/internal/config"
	"github.com/ildx/merlin/internal/diff"
	"github.com/ildx/merlin/internal/state"
	"github.com/spf13/cobra"
)

// diffCmd provides a high-level overview of differences between the current
// system state and the declarative repository definitions. This is the main
// entry point for drift detection (Phase 12).
//
// FLAGS
//
//	--packages   Include package (brew/mas) differences
//	--configs    Include symlink/config differences
//	--scripts    Include script differences (placeholder)
//	--json       Output machine-readable JSON instead of text summary
//
// When no category flags are provided, all categories are shown.
//
// EXAMPLES
//
//	merlin diff                     # Full diff
//	merlin diff --packages          # Only package drift
//	merlin diff --configs --json    # Symlink diff as JSON
//	merlin diff --scripts           # (will show placeholder until implemented)
//
// EXIT STATUS
//
//	Exits 0 even when differences are found; non-zero only on internal errors.
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show differences between system state and repo configs",
	Long:  "Compute and display drift between installed packages, symlinked configs, and declared repository state. Useful for auditing and reconciling machines.",
	Run: func(cmd *cobra.Command, args []string) {
		runDiff(cmd)
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().Bool("packages", false, "Include package (brew & mas) differences")
	diffCmd.Flags().Bool("configs", false, "Include config/symlink differences")
	diffCmd.Flags().Bool("scripts", false, "Include script differences")
	diffCmd.Flags().Bool("json", false, "Output JSON instead of human-readable text")
}

func runDiff(cmd *cobra.Command) {
	// Locate repository
	repo, err := config.FindDotfilesRepo()
	if err != nil {
		cli.Error("Dotfiles repository not found: %v", err)
		os.Exit(1)
	}

	// Collect system snapshot (read-only operation)
	snap := state.CollectSnapshot(repo.Root)

	// Compute diff
	result, err := diff.Compute(repo, snap)
	if err != nil {
		cli.Error("Failed to compute diff: %v", err)
		os.Exit(1)
	}

	// Resolve flags
	includePackages, _ := cmd.Flags().GetBool("packages")
	includeConfigs, _ := cmd.Flags().GetBool("configs")
	includeScripts, _ := cmd.Flags().GetBool("scripts")
	asJSON, _ := cmd.Flags().GetBool("json")

	// If no specific categories requested, default to all
	if !includePackages && !includeConfigs && !includeScripts {
		includePackages = true
		includeConfigs = true
		includeScripts = true
	}

	if asJSON {
		jsonStr, jErr := result.ToJSON()
		if jErr != nil {
			cli.Error("Failed to marshal diff to JSON: %v", jErr)
			os.Exit(1)
		}
		fmt.Println(jsonStr)
		return
	}

	// Human readable output
	fmt.Println("\n🧭 Merlin Diff Report")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Repository: %s\n", repo.Root)
	fmt.Println()

	output := result.HumanReadable(includePackages, includeConfigs, includeScripts)
	fmt.Println(output)

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Legend: Added=present but undeclared | Missing=declared but absent")
	fmt.Println("Symlink categories: Missing=not created | Orphaned=points into repo but undeclared | Broken=target missing | Divergent=hash mismatch")
	fmt.Println("Scripts use Added/Missing semantics (namespaced as tool/script).")
	fmt.Println()
	cli.Success("Diff completed")
}
