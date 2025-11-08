package system

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// CommandCheck represents the result of checking if a command exists
type CommandCheck struct {
	Name      string
	Exists    bool
	Path      string
	Version   string
	Error     error
}

// CheckCommand checks if a command exists in the system PATH
func CheckCommand(name string) *CommandCheck {
	result := &CommandCheck{
		Name:   name,
		Exists: false,
	}

	// Try to find the command
	path, err := exec.LookPath(name)
	if err != nil {
		result.Error = fmt.Errorf("command not found: %s", name)
		return result
	}

	result.Exists = true
	result.Path = path

	// Try to get version (works for many commands)
	result.Version = getCommandVersion(name)

	return result
}

// CheckHomebrew checks if Homebrew is installed and returns detailed info
func CheckHomebrew() *CommandCheck {
	check := CheckCommand("brew")
	
	if !check.Exists {
		check.Error = fmt.Errorf("Homebrew is not installed. Install it from https://brew.sh")
		return check
	}

	// Get brew version
	if check.Version == "" {
		cmd := exec.Command("brew", "--version")
		if output, err := cmd.Output(); err == nil {
			// Parse "Homebrew X.Y.Z" from output
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 {
				check.Version = strings.TrimSpace(lines[0])
			}
		}
	}

	return check
}

// CheckMAS checks if mas-cli is installed
func CheckMAS() *CommandCheck {
	check := CheckCommand("mas")
	
	if !check.Exists {
		check.Error = fmt.Errorf("mas-cli is not installed. Install it with: brew install mas")
		return check
	}

	return check
}

// getCommandVersion tries to get the version of a command
func getCommandVersion(name string) string {
	// Try common version flags
	versionFlags := []string{"--version", "-v", "version"}
	
	for _, flag := range versionFlags {
		cmd := exec.Command(name, flag)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			// Return first line of output
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 && lines[0] != "" {
				return strings.TrimSpace(lines[0])
			}
		}
	}
	
	return ""
}

// IsCommandAvailable is a simple boolean check for command availability
func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// GetOS returns the current operating system
func GetOS() string {
	return runtime.GOOS
}

// IsMacOS checks if the current OS is macOS
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// GetHostname returns the system hostname
func GetHostname() (string, error) {
	return os.Hostname()
}

// SystemInfo contains information about the system
type SystemInfo struct {
	OS       string
	Arch     string
	Hostname string
}

// GetSystemInfo returns information about the current system
func GetSystemInfo() (*SystemInfo, error) {
	hostname, err := GetHostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	return &SystemInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Hostname: hostname,
	}, nil
}

// CheckPrerequisites checks all system prerequisites for Merlin
func CheckPrerequisites() error {
	// Check if running on macOS
	if !IsMacOS() {
		return fmt.Errorf("Merlin is designed for macOS, but you're running %s", GetOS())
	}

	// Check if Homebrew is installed
	brewCheck := CheckHomebrew()
	if !brewCheck.Exists {
		return brewCheck.Error
	}

	return nil
}

// CheckAllCommands checks multiple commands at once
func CheckAllCommands(commands ...string) map[string]*CommandCheck {
	results := make(map[string]*CommandCheck)
	for _, cmd := range commands {
		results[cmd] = CheckCommand(cmd)
	}
	return results
}

// FormatCommandCheck returns a formatted string for a command check result
func FormatCommandCheck(check *CommandCheck) string {
	if check.Exists {
		status := "✓"
		info := fmt.Sprintf("%s %s", status, check.Name)
		if check.Version != "" {
			info += fmt.Sprintf(" (%s)", check.Version)
		}
		if check.Path != "" {
			info += fmt.Sprintf(" at %s", check.Path)
		}
		return info
	}
	
	return fmt.Sprintf("✗ %s (not found)", check.Name)
}

