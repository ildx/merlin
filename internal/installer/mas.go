package installer

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ildx/merlin/internal/models"
)

// MASInstaller handles Mac App Store app installation
type MASInstaller struct {
	DryRun  bool
	Verbose bool
}

// NewMASInstaller creates a new Mac App Store installer
func NewMASInstaller(dryRun, verbose bool) *MASInstaller {
	return &MASInstaller{
		DryRun:  dryRun,
		Verbose: verbose,
	}
}

// IsAppInstalled checks if a Mac App Store app is installed
func (m *MASInstaller) IsAppInstalled(appID int) (bool, error) {
	cmd := exec.Command("mas", "list")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to list installed apps: %w", err)
	}

	// Parse output to find app ID
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		// Format: "497799835 Xcode (16.0)"
		parts := strings.Fields(line)
		if len(parts) > 0 {
			id, err := strconv.Atoi(parts[0])
			if err == nil && id == appID {
				return true, nil
			}
		}
	}

	return false, nil
}

// CheckMASAccount checks if the user is signed into the Mac App Store
func (m *MASInstaller) CheckMASAccount() (bool, string, error) {
	cmd := exec.Command("mas", "account")
	output, err := cmd.Output()
	
	if err != nil {
		// Exit code 1 usually means not signed in
		return false, "", nil
	}

	account := strings.TrimSpace(string(output))
	if account == "Not signed in" || account == "" {
		return false, "", nil
	}

	return true, account, nil
}

// InstallApp installs a single Mac App Store app
func (m *MASInstaller) InstallApp(app models.MASApp, output io.Writer) *InstallResult {
	result := &InstallResult{
		Package: app.Name,
		Success: false,
	}

	// Check if already installed
	installed, err := m.IsAppInstalled(app.ID)
	if err != nil {
		result.Error = fmt.Errorf("failed to check if installed: %w", err)
		return result
	}

	if installed {
		result.AlreadyExists = true
		result.Success = true
		if output != nil {
			fmt.Fprintf(output, "  ⏭  %s (already installed)\n", app.Name)
		}
		return result
	}

	// Dry run mode
	if m.DryRun {
		if output != nil {
			fmt.Fprintf(output, "  [DRY RUN] Would install: %s (ID: %d)\n", app.Name, app.ID)
		}
		result.Success = true
		return result
	}

	// Install the app
	if output != nil {
		fmt.Fprintf(output, "  🍎 Installing %s (ID: %d)...\n", app.Name, app.ID)
	}

	cmd := exec.Command("mas", "install", strconv.Itoa(app.ID))

	// Stream output if verbose
	if m.Verbose && output != nil {
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
		fmt.Fprintf(output, "  ✓ %s installed successfully\n", app.Name)
	}

	return result
}

// InstallApps installs multiple Mac App Store apps
func (m *MASInstaller) InstallApps(apps []models.MASApp, output io.Writer) []*InstallResult {
	results := make([]*InstallResult, 0, len(apps))

	if output != nil {
		fmt.Fprintf(output, "\n🍎 Installing %d Mac App Store app(s)...\n\n", len(apps))
	}

	for _, app := range apps {
		result := m.InstallApp(app, output)
		results = append(results, result)
	}

	return results
}

// PrintMASSummary prints a summary of Mac App Store installation results
func PrintMASSummary(results []*InstallResult, output io.Writer) {
	if len(results) == 0 {
		return
	}

	successCount := 0
	alreadyInstalledCount := 0
	failedCount := 0

	for _, result := range results {
		if result.AlreadyExists {
			alreadyInstalledCount++
		} else if result.Success {
			successCount++
		} else {
			failedCount++
		}
	}

	fmt.Fprintf(output, "\n")
	fmt.Fprintln(output, strings.Repeat("═", 80))
	fmt.Fprintf(output, "Mac App Store Installation Summary\n")
	fmt.Fprintln(output, strings.Repeat("═", 80))

	fmt.Fprintf(output, "\n🍎 Apps (%d total):\n", len(results))
	fmt.Fprintf(output, "   ✓ %d installed\n", successCount)
	fmt.Fprintf(output, "   ⏭  %d already installed\n", alreadyInstalledCount)
	if failedCount > 0 {
		fmt.Fprintf(output, "   ✗ %d failed\n", failedCount)
	}

	// List failures if any
	failures := []*InstallResult{}
	for _, result := range results {
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

