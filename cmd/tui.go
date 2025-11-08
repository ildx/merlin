package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ildx/merlin/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:     "tui",
	Aliases: []string{"interactive", "ui"},
	Short:   "Launch interactive TUI",
	Long: `Launch the interactive Terminal User Interface (TUI) for Merlin.

FEATURES
	• Browse & install Homebrew packages (formulae & casks)
	• Manage dotfiles (link/unlink configs)
	• Run setup scripts with multi-select and real-time progress
	• System doctor shortcut

NAVIGATION
	Arrow keys / j k   Move
	Space              Select / toggle
	Enter              Confirm
	Esc / q            Back / quit

SCRIPT EXECUTION
	• Select a tool with defined scripts
	• Multi-select scripts to run (supports tags for organization)
	• Watch real-time execution with status indicators
	• Review summary with timing and errors

EXAMPLES
	merlin tui         # Launch interface
	merlin             # Same (default when no subcommand)

NOTES
	All operations respect global --dry-run and --verbose flags when applicable.
	Scripts are logged to ~/.merlin/merlin.log with timing information.

SEE ALSO
	merlin install, merlin link, merlin run, merlin doctor`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runTUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI() error {
	// Create and run main menu
	menu := tui.NewMenuModel()
	p := tea.NewProgram(menu, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Get the final menu model
	menuModel, ok := finalModel.(tui.MenuModel)
	if !ok {
		return nil
	}

	selected := menuModel.GetSelected()

	// Handle the selected action
	switch selected {
	case "install":
		return runTUIInstall()
	case "dotfiles":
		return runTUIDotfiles()
	case "scripts":
		return runTUIScripts()
	case "doctor":
		runDoctor()
		return nil
	case "quit":
		return nil
	default:
		return nil
	}
}

func runTUIInstall() error {
	return tui.LaunchPackageInstaller()
}

func runTUIDotfiles() error {
	return tui.LaunchConfigManager()
}

func runTUIScripts() error {
	return tui.LaunchScriptRunner()
}
