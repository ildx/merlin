package installer

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ildx/merlin/internal/models"
)

// SelectPackages interactively prompts the user to select packages
func SelectPackages(packages []models.BrewPackage, packageType string, input io.Reader, output io.Writer) ([]models.BrewPackage, error) {
	if len(packages) == 0 {
		return packages, nil
	}

	// Display packages with numbers
	fmt.Fprintf(output, "\n%s to install (%d total):\n\n", packageType, len(packages))
	
	for i, pkg := range packages {
		desc := pkg.Description
		if desc == "" {
			desc = "No description"
		}
		fmt.Fprintf(output, "  %2d. %-35s - %s\n", i+1, pkg.Name, desc)
	}

	// Prompt for selection
	fmt.Fprintf(output, "\nSelect packages to install:\n")
	fmt.Fprintf(output, "  • Enter 'all' to install everything\n")
	fmt.Fprintf(output, "  • Enter 'none' to skip\n")
	fmt.Fprintf(output, "  • Enter numbers separated by spaces (e.g., '1 3 5')\n")
	fmt.Fprintf(output, "  • Enter ranges (e.g., '1-5 8 10-12')\n")
	fmt.Fprintf(output, "\nYour choice: ")

	// Read user input
	scanner := bufio.NewScanner(input)
	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read input")
	}

	choice := strings.TrimSpace(scanner.Text())
	
	// Handle special cases
	switch strings.ToLower(choice) {
	case "all":
		return packages, nil
	case "none", "":
		return []models.BrewPackage{}, nil
	}

	// Parse selection
	selected, err := parseSelection(choice, len(packages))
	if err != nil {
		return nil, err
	}

	// Build result list
	result := make([]models.BrewPackage, 0, len(selected))
	for _, idx := range selected {
		result = append(result, packages[idx])
	}

	if len(result) > 0 {
		fmt.Fprintf(output, "\n✓ Selected %d package(s)\n", len(result))
	} else {
		fmt.Fprintf(output, "\n⚠️  No packages selected\n")
	}

	return result, nil
}

// parseSelection parses user input like "1 3 5" or "1-5 8 10-12"
func parseSelection(input string, maxIndex int) ([]int, error) {
	selected := make(map[int]bool)
	parts := strings.Fields(input)

	for _, part := range parts {
		// Check if it's a range (e.g., "1-5")
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid number in range: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid number in range: %s", rangeParts[1])
			}

			if start < 1 || start > maxIndex || end < 1 || end > maxIndex {
				return nil, fmt.Errorf("range %d-%d is out of bounds (1-%d)", start, end, maxIndex)
			}

			if start > end {
				return nil, fmt.Errorf("invalid range %d-%d (start must be <= end)", start, end)
			}

			for i := start; i <= end; i++ {
				selected[i-1] = true // Convert to 0-based index
			}
		} else {
			// Single number
			num, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", part)
			}

			if num < 1 || num > maxIndex {
				return nil, fmt.Errorf("number %d is out of bounds (1-%d)", num, maxIndex)
			}

			selected[num-1] = true // Convert to 0-based index
		}
	}

	// Convert map to sorted slice
	result := make([]int, 0, len(selected))
	for i := 0; i < maxIndex; i++ {
		if selected[i] {
			result = append(result, i)
		}
	}

	return result, nil
}

// ConfirmInstallation asks the user to confirm before installing
func ConfirmInstallation(formulaeCount, casksCount int, input io.Reader, output io.Writer) (bool, error) {
	total := formulaeCount + casksCount
	if total == 0 {
		return false, nil
	}

	fmt.Fprintf(output, "\n")
	fmt.Fprintf(output, "════════════════════════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(output, "Ready to install:\n")
	if formulaeCount > 0 {
		fmt.Fprintf(output, "  • %d formulae\n", formulaeCount)
	}
	if casksCount > 0 {
		fmt.Fprintf(output, "  • %d casks\n", casksCount)
	}
	fmt.Fprintf(output, "════════════════════════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(output, "\nProceed with installation? [y/N]: ")

	scanner := bufio.NewScanner(input)
	if !scanner.Scan() {
		return false, fmt.Errorf("failed to read input")
	}

	response := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return response == "y" || response == "yes", nil
}

// SelectMASApps interactively prompts the user to select Mac App Store apps
func SelectMASApps(apps []models.MASApp, input io.Reader, output io.Writer) ([]models.MASApp, error) {
	if len(apps) == 0 {
		return apps, nil
	}

	// Display apps with numbers
	fmt.Fprintf(output, "\n🍎 Mac App Store apps to install (%d total):\n\n", len(apps))
	
	for i, app := range apps {
		desc := app.Description
		if desc == "" {
			desc = "No description"
		}
		fmt.Fprintf(output, "  %2d. %-35s [%d] - %s\n", i+1, app.Name, app.ID, desc)
	}

	// Prompt for selection
	fmt.Fprintf(output, "\nSelect apps to install:\n")
	fmt.Fprintf(output, "  • Enter 'all' to install everything\n")
	fmt.Fprintf(output, "  • Enter 'none' to skip\n")
	fmt.Fprintf(output, "  • Enter numbers separated by spaces (e.g., '1 3 5')\n")
	fmt.Fprintf(output, "  • Enter ranges (e.g., '1-5 8 10-12')\n")
	fmt.Fprintf(output, "\nYour choice: ")

	// Read user input
	scanner := bufio.NewScanner(input)
	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read input")
	}

	choice := strings.TrimSpace(scanner.Text())
	
	// Handle special cases
	switch strings.ToLower(choice) {
	case "all":
		return apps, nil
	case "none", "":
		return []models.MASApp{}, nil
	}

	// Parse selection
	selected, err := parseSelection(choice, len(apps))
	if err != nil {
		return nil, err
	}

	// Build result list
	result := make([]models.MASApp, 0, len(selected))
	for _, idx := range selected {
		result = append(result, apps[idx])
	}

	if len(result) > 0 {
		fmt.Fprintf(output, "\n✓ Selected %d app(s)\n", len(result))
	} else {
		fmt.Fprintf(output, "\n⚠️  No apps selected\n")
	}

	return result, nil
}

// ConfirmMASInstallation asks the user to confirm before installing MAS apps
func ConfirmMASInstallation(appCount int, input io.Reader, output io.Writer) (bool, error) {
	if appCount == 0 {
		return false, nil
	}

	fmt.Fprintf(output, "\n")
	fmt.Fprintf(output, "════════════════════════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(output, "Ready to install:\n")
	fmt.Fprintf(output, "  • %d Mac App Store app(s)\n", appCount)
	fmt.Fprintf(output, "════════════════════════════════════════════════════════════════════════════════\n")
	fmt.Fprintf(output, "\nProceed with installation? [y/N]: ")

	scanner := bufio.NewScanner(input)
	if !scanner.Scan() {
		return false, fmt.Errorf("failed to read input")
	}

	response := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return response == "y" || response == "yes", nil
}

