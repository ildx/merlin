package installer

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/ildx/merlin/internal/models"
)

// BrewInstaller handles Homebrew package installation
type BrewInstaller struct {
	DryRun  bool
	Verbose bool
}

// InstallResult represents the result of an installation attempt
type InstallResult struct {
	Package       string
	Success       bool
	AlreadyExists bool
	Error         error
	Output        string
}

// NewBrewInstaller creates a new Homebrew installer
func NewBrewInstaller(dryRun, verbose bool) *BrewInstaller {
	return &BrewInstaller{
		DryRun:  dryRun,
		Verbose: verbose,
	}
}

// IsFormulaInstalled checks if a Homebrew formula is installed
func (b *BrewInstaller) IsFormulaInstalled(name string) (bool, error) {
	cmd := exec.Command("brew", "list", "--formula", name)
	err := cmd.Run()
	return err == nil, nil
}

// IsCaskInstalled checks if a Homebrew cask is installed
func (b *BrewInstaller) IsCaskInstalled(name string) (bool, error) {
	cmd := exec.Command("brew", "list", "--cask", name)
	err := cmd.Run()
	return err == nil, nil
}

// InstallFormula installs a single Homebrew formula
func (b *BrewInstaller) InstallFormula(pkg models.BrewPackage, output io.Writer) *InstallResult {
	result := &InstallResult{
		Package: pkg.Name,
		Success: false,
	}

	// Check if already installed
	installed, err := b.IsFormulaInstalled(pkg.Name)
	if err != nil {
		result.Error = fmt.Errorf("failed to check if installed: %w", err)
		return result
	}

	if installed {
		result.AlreadyExists = true
		result.Success = true
		if output != nil {
			fmt.Fprintf(output, "  ⏭  %s (already installed)\n", pkg.Name)
		}
		return result
	}

	// Dry run mode
	if b.DryRun {
		if output != nil {
			fmt.Fprintf(output, "  [DRY RUN] Would install: %s\n", pkg.Name)
		}
		result.Success = true
		return result
	}

	// Install the formula
	if output != nil {
		fmt.Fprintf(output, "  📦 Installing %s...\n", pkg.Name)
	}

	cmd := exec.Command("brew", "install", pkg.Name)
	
	// Stream output if verbose
	if b.Verbose && output != nil {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			result.Error = err
			return result
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			result.Error = err
			return result
		}

		if err := cmd.Start(); err != nil {
			result.Error = err
			return result
		}

		// Stream stdout
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				fmt.Fprintf(output, "     %s\n", scanner.Text())
			}
		}()

		// Stream stderr
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				fmt.Fprintf(output, "     %s\n", scanner.Text())
			}
		}()

		err = cmd.Wait()
		if err != nil {
			result.Error = fmt.Errorf("installation failed: %w", err)
			return result
		}
	} else {
		// Non-verbose: just run and capture output
		outputBytes, err := cmd.CombinedOutput()
		result.Output = string(outputBytes)
		if err != nil {
			result.Error = fmt.Errorf("installation failed: %w", err)
			if output != nil {
				fmt.Fprintf(output, "     Error: %v\n", err)
			}
			return result
		}
	}

	result.Success = true
	if output != nil {
		fmt.Fprintf(output, "  ✓ %s installed successfully\n", pkg.Name)
	}

	return result
}

// InstallCask installs a single Homebrew cask
func (b *BrewInstaller) InstallCask(pkg models.BrewPackage, output io.Writer) *InstallResult {
	result := &InstallResult{
		Package: pkg.Name,
		Success: false,
	}

	// Check if already installed
	installed, err := b.IsCaskInstalled(pkg.Name)
	if err != nil {
		result.Error = fmt.Errorf("failed to check if installed: %w", err)
		return result
	}

	if installed {
		result.AlreadyExists = true
		result.Success = true
		if output != nil {
			fmt.Fprintf(output, "  ⏭  %s (already installed)\n", pkg.Name)
		}
		return result
	}

	// Dry run mode
	if b.DryRun {
		if output != nil {
			fmt.Fprintf(output, "  [DRY RUN] Would install: %s\n", pkg.Name)
		}
		result.Success = true
		return result
	}

	// Install the cask
	if output != nil {
		fmt.Fprintf(output, "  📱 Installing %s...\n", pkg.Name)
	}

	cmd := exec.Command("brew", "install", "--cask", pkg.Name)
	
	// Stream output if verbose
	if b.Verbose && output != nil {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			result.Error = err
			return result
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			result.Error = err
			return result
		}

		if err := cmd.Start(); err != nil {
			result.Error = err
			return result
		}

		// Stream stdout
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				fmt.Fprintf(output, "     %s\n", scanner.Text())
			}
		}()

		// Stream stderr
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				fmt.Fprintf(output, "     %s\n", scanner.Text())
			}
		}()

		err = cmd.Wait()
		if err != nil {
			result.Error = fmt.Errorf("installation failed: %w", err)
			return result
		}
	} else {
		// Non-verbose: just run and capture output
		outputBytes, err := cmd.CombinedOutput()
		result.Output = string(outputBytes)
		if err != nil {
			result.Error = fmt.Errorf("installation failed: %w", err)
			if output != nil {
				fmt.Fprintf(output, "     Error: %v\n", err)
			}
			return result
		}
	}

	result.Success = true
	if output != nil {
		fmt.Fprintf(output, "  ✓ %s installed successfully\n", pkg.Name)
	}

	return result
}

// InstallFormulae installs multiple formulae
func (b *BrewInstaller) InstallFormulae(packages []models.BrewPackage, output io.Writer) []*InstallResult {
	results := make([]*InstallResult, 0, len(packages))
	
	if output != nil {
		fmt.Fprintf(output, "\n🔧 Installing %d formulae...\n\n", len(packages))
	}

	for _, pkg := range packages {
		result := b.InstallFormula(pkg, output)
		results = append(results, result)
	}

	return results
}

// InstallCasks installs multiple casks
func (b *BrewInstaller) InstallCasks(packages []models.BrewPackage, output io.Writer) []*InstallResult {
	results := make([]*InstallResult, 0, len(packages))
	
	if output != nil {
		fmt.Fprintf(output, "\n📱 Installing %d casks...\n\n", len(packages))
	}

	for _, pkg := range packages {
		result := b.InstallCask(pkg, output)
		results = append(results, result)
	}

	return results
}

// InstallAll installs all formulae and casks
func (b *BrewInstaller) InstallAll(config *models.BrewConfig, output io.Writer) ([]*InstallResult, []*InstallResult) {
	var formulaeResults, caskResults []*InstallResult

	if len(config.Formulae) > 0 {
		formulaeResults = b.InstallFormulae(config.Formulae, output)
	}

	if len(config.Casks) > 0 {
		caskResults = b.InstallCasks(config.Casks, output)
	}

	return formulaeResults, caskResults
}

// PrintSummary prints a summary of installation results
func PrintSummary(formulaeResults, caskResults []*InstallResult, output io.Writer) {
	totalFormulae := len(formulaeResults)
	totalCasks := len(caskResults)
	
	if totalFormulae == 0 && totalCasks == 0 {
		return
	}

	fmt.Fprintf(output, "\n")
	fmt.Fprintln(output, strings.Repeat("═", 80))
	fmt.Fprintf(output, "Installation Summary\n")
	fmt.Fprintln(output, strings.Repeat("═", 80))

	// Formulae summary
	if totalFormulae > 0 {
		successCount := 0
		alreadyInstalledCount := 0
		failedCount := 0

		for _, result := range formulaeResults {
			if result.AlreadyExists {
				alreadyInstalledCount++
			} else if result.Success {
				successCount++
			} else {
				failedCount++
			}
		}

		fmt.Fprintf(output, "\n🔧 Formulae (%d total):\n", totalFormulae)
		fmt.Fprintf(output, "   ✓ %d installed\n", successCount)
		fmt.Fprintf(output, "   ⏭  %d already installed\n", alreadyInstalledCount)
		if failedCount > 0 {
			fmt.Fprintf(output, "   ✗ %d failed\n", failedCount)
		}
	}

	// Casks summary
	if totalCasks > 0 {
		successCount := 0
		alreadyInstalledCount := 0
		failedCount := 0

		for _, result := range caskResults {
			if result.AlreadyExists {
				alreadyInstalledCount++
			} else if result.Success {
				successCount++
			} else {
				failedCount++
			}
		}

		fmt.Fprintf(output, "\n📱 Casks (%d total):\n", totalCasks)
		fmt.Fprintf(output, "   ✓ %d installed\n", successCount)
		fmt.Fprintf(output, "   ⏭  %d already installed\n", alreadyInstalledCount)
		if failedCount > 0 {
			fmt.Fprintf(output, "   ✗ %d failed\n", failedCount)
		}
	}

	// List failures if any
	failures := []*InstallResult{}
	for _, result := range formulaeResults {
		if !result.Success && !result.AlreadyExists {
			failures = append(failures, result)
		}
	}
	for _, result := range caskResults {
		if !result.Success && !result.AlreadyExists {
			failures = append(failures, result)
		}
	}

	if len(failures) > 0 {
		fmt.Fprintf(output, "\n❌ Failed installations:\n")
		for _, failure := range failures {
			fmt.Fprintf(output, "   • %s: %v\n", failure.Package, failure.Error)
		}
	}

	fmt.Fprintln(output, strings.Repeat("═", 80))
	fmt.Println()
}

