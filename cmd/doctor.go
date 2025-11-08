package cmd

import (
	"fmt"

	"github.com/ildx/merlin/internal/cli"
	"github.com/ildx/merlin/internal/system"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system prerequisites",
	Long:  "Check if required system tools (Homebrew, mas-cli, optional utilities) are installed and report environment details.\n\nOUTPUT SECTIONS\n  • System information (OS, arch, hostname)\n  • macOS suitability\n  • Required package managers\n  • Optional helper tools (git, curl, jq, yq)\n\nEXIT STATUS\n  Always exits 0; missing prerequisites are reported with suggestions.\n\nEXAMPLES\n  merlin doctor          # Full system check\n  merlin doctor --verbose (global flag for more logging)\n\nTIPS\n  Run this first on a new machine to confirm prerequisites before installs.",
	Run: func(cmd *cobra.Command, args []string) {
		runDoctor()
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor() {
	fmt.Println("\n🔍 Merlin System Check")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Get system info
	sysInfo, err := system.GetSystemInfo()
	if err != nil {
		cli.Error("Failed to get system info: %v", err)
		return
	}

	fmt.Printf("\n📋 System Information:\n")
	fmt.Printf("   OS:       %s\n", sysInfo.OS)
	fmt.Printf("   Arch:     %s\n", sysInfo.Arch)
	fmt.Printf("   Hostname: %s\n", sysInfo.Hostname)

	// Check if running on macOS
	fmt.Printf("\n🍎 Operating System:\n")
	if system.IsMacOS() {
		fmt.Println("   ✓ Running on macOS")
	} else {
		fmt.Printf("   ✗ Not running on macOS (detected: %s)\n", sysInfo.OS)
		fmt.Println("   ⚠️  Merlin is designed for macOS")
	}

	// Check Homebrew
	fmt.Printf("\n📦 Package Managers:\n")
	brewCheck := system.CheckHomebrew()
	fmt.Printf("   %s\n", system.FormatCommandCheck(brewCheck))

	// Check mas-cli
	masCheck := system.CheckMAS()
	fmt.Printf("   %s\n", system.FormatCommandCheck(masCheck))

	// Check other useful commands
	fmt.Printf("\n🔧 Optional Tools:\n")
	optionalTools := []string{"git", "curl", "jq", "yq"}
	checks := system.CheckAllCommands(optionalTools...)

	for _, tool := range optionalTools {
		if check, ok := checks[tool]; ok {
			fmt.Printf("   %s\n", system.FormatCommandCheck(check))
		}
	}

	// Overall status
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if err := system.CheckPrerequisites(); err != nil {
		cli.Error("System check failed: %v", err)
		if !brewCheck.Exists {
			fmt.Println("\n💡 To install Homebrew, run:")
			fmt.Println("   /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"")
		}
	} else {
		cli.Success("All prerequisites satisfied! Merlin is ready to use.")
	}
	fmt.Println()
}
