package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ildx/merlin/internal/cli"
	"github.com/ildx/merlin/internal/config"
	"github.com/ildx/merlin/internal/logger"
	"github.com/ildx/merlin/internal/parser"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate TOML configuration files",
	Long: `Validate Merlin TOML configuration files for structural and content issues.

CHECKS PERFORMED
	• TOML syntax errors
	• Duplicate packages/apps/profile names
	• Invalid conflict strategies
	• Missing tool config files
	• Broken or missing link sources
	• Missing or invalid script references

FLAGS
	--strict   Treat warnings as errors (non‑zero exit code)
	--dry-run  (Global) No effect here but accepted for consistency
	--verbose  Show additional internal logging

EXIT STATUS
	0 if no errors (warnings allowed unless --strict)
	Non-zero if errors found or warnings in strict mode

EXAMPLES
	merlin validate             # Standard validation
	merlin validate --strict    # Enforce warnings as errors

TIPS
	Run before linking or installing for early feedback.
	Combine with --verbose to see debug log output (file: ~/.merlin/merlin.log).`,
	Run: func(cmd *cobra.Command, args []string) {
		strict, _ := cmd.Flags().GetBool("strict")

		if err := runValidate(strict); err != nil {
			cli.Error("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().Bool("strict", false, "Treat warnings as errors")
}

type ValidationResult struct {
	File     string
	Errors   []string
	Warnings []string
}

func runValidate(strict bool) error {
	logger.Info("Starting configuration validation")

	// Find dotfiles repository
	repo, err := config.FindDotfilesRepo()
	if err != nil {
		return fmt.Errorf("dotfiles repository not found: %w", err)
	}

	fmt.Printf("\n🔍 Validating Merlin Configuration\n")
	fmt.Printf("Repository: %s\n\n", repo.Root)

	results := make([]ValidationResult, 0)

	// Validate root merlin.toml
	rootResult := validateRootConfig(repo)
	results = append(results, rootResult)

	// Validate brew.toml
	brewResult := validateBrewConfig(repo)
	if brewResult != nil {
		results = append(results, *brewResult)
	}

	// Validate mas.toml
	masResult := validateMASConfig(repo)
	if masResult != nil {
		results = append(results, *masResult)
	}

	// Validate tool configs
	tools, err := repo.ListTools()
	if err != nil {
		logger.Warn("Failed to list tools", "error", err)
	} else {
		for _, tool := range tools {
			toolResult := validateToolConfig(repo, tool)
			if toolResult != nil {
				results = append(results, *toolResult)
			}
		}
	}

	// Print results
	totalErrors := 0
	totalWarnings := 0

	for _, result := range results {
		if len(result.Errors) > 0 || len(result.Warnings) > 0 {
			fmt.Printf("📄 %s\n", result.File)

			for _, err := range result.Errors {
				fmt.Printf("  ✗ Error: %s\n", err)
				totalErrors++
			}

			for _, warn := range result.Warnings {
				fmt.Printf("  ⚠ Warning: %s\n", warn)
				totalWarnings++
			}

			fmt.Println()
		}
	}

	// Summary
	fmt.Println(strings.Repeat("─", 60))

	if totalErrors == 0 && totalWarnings == 0 {
		fmt.Println("✅ All configuration files are valid!")
		logger.Info("Validation completed successfully")
		return nil
	}

	fmt.Printf("Found %d error(s) and %d warning(s)\n", totalErrors, totalWarnings)

	if totalErrors > 0 {
		logger.Error("Validation failed", "errors", totalErrors, "warnings", totalWarnings)
		return fmt.Errorf("validation failed with %d error(s)", totalErrors)
	}

	if strict && totalWarnings > 0 {
		logger.Error("Validation failed (strict mode)", "warnings", totalWarnings)
		return fmt.Errorf("validation failed with %d warning(s) in strict mode", totalWarnings)
	}

	logger.Info("Validation completed with warnings", "warnings", totalWarnings)
	return nil
}

func validateRootConfig(repo *config.DotfilesRepo) ValidationResult {
	result := ValidationResult{
		File: "merlin.toml",
	}

	rootPath := repo.GetRootMerlinConfig()

	// Check if file exists
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		result.Errors = append(result.Errors, "Root merlin.toml not found")
		return result
	}

	// Parse root config
	rootConfig, err := parser.ParseRootMerlinTOML(rootPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse: %v", err))
		return result
	}

	// Validate metadata
	if rootConfig.Metadata.Name == "" {
		result.Warnings = append(result.Warnings, "Metadata name is empty")
	}

	// Validate settings
	if rootConfig.Settings.ConflictStrategy != "" {
		validStrategies := []string{"skip", "backup", "overwrite"}
		valid := false
		for _, s := range validStrategies {
			if rootConfig.Settings.ConflictStrategy == s {
				valid = true
				break
			}
		}
		if !valid {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Invalid conflict_strategy '%s' (must be: skip, backup, or overwrite)",
					rootConfig.Settings.ConflictStrategy))
		}
	}

	// Validate profiles
	profileNames := make(map[string]bool)
	for i, profile := range rootConfig.Profiles {
		if profile.Name == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Profile %d is missing a name", i))
		} else if profileNames[profile.Name] {
			result.Errors = append(result.Errors, fmt.Sprintf("Duplicate profile name: %s", profile.Name))
		} else {
			profileNames[profile.Name] = true
		}

		if len(profile.Tools) == 0 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Profile '%s' has no tools defined", profile.Name))
		}
	}

	return result
}

func validateBrewConfig(repo *config.DotfilesRepo) *ValidationResult {
	brewPath := filepath.Join(repo.GetToolConfigDir("brew"), "brew.toml")

	// Skip if file doesn't exist
	if _, err := os.Stat(brewPath); os.IsNotExist(err) {
		return nil
	}

	result := &ValidationResult{
		File: "config/brew/config/brew.toml",
	}

	// Parse brew config
	brewConfig, err := parser.ParseBrewTOML(brewPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse: %v", err))
		return result
	}

	// Check for duplicates
	formulaeNames := make(map[string]bool)
	for _, pkg := range brewConfig.Formulae {
		if pkg.Name == "" {
			result.Errors = append(result.Errors, "Formulae entry with empty name")
		} else if formulaeNames[pkg.Name] {
			result.Errors = append(result.Errors, fmt.Sprintf("Duplicate formulae: %s", pkg.Name))
		} else {
			formulaeNames[pkg.Name] = true
		}
	}

	caskNames := make(map[string]bool)
	for _, pkg := range brewConfig.Casks {
		if pkg.Name == "" {
			result.Errors = append(result.Errors, "Cask entry with empty name")
		} else if caskNames[pkg.Name] {
			result.Errors = append(result.Errors, fmt.Sprintf("Duplicate cask: %s", pkg.Name))
		} else {
			caskNames[pkg.Name] = true
		}
	}

	return result
}

func validateMASConfig(repo *config.DotfilesRepo) *ValidationResult {
	masPath := filepath.Join(repo.GetToolConfigDir("mas"), "mas.toml")

	// Skip if file doesn't exist
	if _, err := os.Stat(masPath); os.IsNotExist(err) {
		return nil
	}

	result := &ValidationResult{
		File: "config/mas/config/mas.toml",
	}

	// Parse MAS config
	masConfig, err := parser.ParseMASTOML(masPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse: %v", err))
		return result
	}

	// Check for duplicates
	appIDs := make(map[int]string)
	appNames := make(map[string]bool)

	for _, app := range masConfig.Apps {
		if app.Name == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("App with ID %d has empty name", app.ID))
		} else if appNames[app.Name] {
			result.Errors = append(result.Errors, fmt.Sprintf("Duplicate app name: %s", app.Name))
		} else {
			appNames[app.Name] = true
		}

		if app.ID == 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("App '%s' has invalid ID (0)", app.Name))
		} else if existingName, exists := appIDs[app.ID]; exists {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Duplicate app ID %d: %s and %s", app.ID, existingName, app.Name))
		} else {
			appIDs[app.ID] = app.Name
		}
	}

	return result
}

func validateToolConfig(repo *config.DotfilesRepo, toolName string) *ValidationResult {
	merlinPath := repo.GetToolMerlinConfig(toolName)

	// Skip if no merlin.toml
	if _, err := os.Stat(merlinPath); os.IsNotExist(err) {
		return nil
	}

	result := &ValidationResult{
		File: fmt.Sprintf("config/%s/merlin.toml", toolName),
	}

	// Parse tool config
	toolConfig, err := parser.ParseToolMerlinTOML(merlinPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse: %v", err))
		return result
	}

	// Validate tool name matches
	if toolConfig.Tool.Name != "" && toolConfig.Tool.Name != toolName {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Tool name '%s' doesn't match directory name '%s'", toolConfig.Tool.Name, toolName))
	}

	// Validate links
	for i, link := range toolConfig.Links {
		if link.Target == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Link %d is missing target", i))
		}

		// Check if source exists (if specified)
		if link.Source != "" {
			sourcePath := filepath.Join(repo.GetToolRoot(toolName), link.Source)
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Link source doesn't exist: %s", link.Source))
			}
		}
	}

	// Validate scripts
	if toolConfig.HasScripts() {
		scriptsDir := filepath.Join(repo.GetToolRoot(toolName), toolConfig.Scripts.Directory)
		for _, script := range toolConfig.Scripts.Scripts {
			scriptPath := filepath.Join(scriptsDir, script)
			if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
				result.Errors = append(result.Errors, fmt.Sprintf("Script doesn't exist: %s", script))
			}
		}
	}

	return result
}
