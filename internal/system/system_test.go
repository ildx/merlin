package system

import (
	"runtime"
	"testing"
)

func TestCheckCommand(t *testing.T) {
	t.Run("check existing command", func(t *testing.T) {
		// ls should exist on all Unix systems
		check := CheckCommand("ls")
		if !check.Exists {
			t.Error("expected ls to exist")
		}
		if check.Name != "ls" {
			t.Errorf("expected name 'ls', got %s", check.Name)
		}
		if check.Path == "" {
			t.Error("expected path to be set")
		}
		if check.Error != nil {
			t.Errorf("expected no error, got: %v", check.Error)
		}
	})

	t.Run("check non-existent command", func(t *testing.T) {
		check := CheckCommand("this-command-definitely-does-not-exist-12345")
		if check.Exists {
			t.Error("expected command to not exist")
		}
		if check.Error == nil {
			t.Error("expected error for non-existent command")
		}
	})
}

func TestCheckHomebrew(t *testing.T) {
	t.Run("check homebrew", func(t *testing.T) {
		check := CheckHomebrew()
		
		// On macOS with Homebrew installed, this should pass
		// On other systems or without brew, it will fail (which is expected)
		if check.Exists {
			t.Logf("Homebrew found at: %s", check.Path)
			if check.Version != "" {
				t.Logf("Homebrew version: %s", check.Version)
			}
			if check.Path == "" {
				t.Error("expected path to be set when brew exists")
			}
		} else {
			t.Logf("Homebrew not found (this may be expected): %v", check.Error)
		}
	})
}

func TestCheckMAS(t *testing.T) {
	t.Run("check mas-cli", func(t *testing.T) {
		check := CheckMAS()
		
		if check.Exists {
			t.Logf("mas-cli found at: %s", check.Path)
			if check.Version != "" {
				t.Logf("mas version: %s", check.Version)
			}
		} else {
			t.Logf("mas-cli not found (this may be expected): %v", check.Error)
		}
	})
}

func TestIsCommandAvailable(t *testing.T) {
	t.Run("available command", func(t *testing.T) {
		if !IsCommandAvailable("ls") {
			t.Error("expected ls to be available")
		}
	})

	t.Run("unavailable command", func(t *testing.T) {
		if IsCommandAvailable("this-command-does-not-exist-xyz") {
			t.Error("expected command to be unavailable")
		}
	})
}

func TestGetOS(t *testing.T) {
	os := GetOS()
	if os == "" {
		t.Error("expected OS to be non-empty")
	}
	
	// Should match runtime.GOOS
	if os != runtime.GOOS {
		t.Errorf("expected OS to be %s, got %s", runtime.GOOS, os)
	}
	
	t.Logf("Current OS: %s", os)
}

func TestIsMacOS(t *testing.T) {
	isMac := IsMacOS()
	expectedMac := runtime.GOOS == "darwin"
	
	if isMac != expectedMac {
		t.Errorf("expected IsMacOS to be %v, got %v", expectedMac, isMac)
	}
	
	t.Logf("Running on macOS: %v", isMac)
}

func TestGetHostname(t *testing.T) {
	hostname, err := GetHostname()
	if err != nil {
		t.Fatalf("failed to get hostname: %v", err)
	}
	
	if hostname == "" {
		t.Error("expected hostname to be non-empty")
	}
	
	t.Logf("Hostname: %s", hostname)
}

func TestGetSystemInfo(t *testing.T) {
	info, err := GetSystemInfo()
	if err != nil {
		t.Fatalf("failed to get system info: %v", err)
	}
	
	if info.OS == "" {
		t.Error("expected OS to be set")
	}
	if info.Arch == "" {
		t.Error("expected Arch to be set")
	}
	if info.Hostname == "" {
		t.Error("expected Hostname to be set")
	}
	
	t.Logf("System Info: OS=%s, Arch=%s, Hostname=%s", info.OS, info.Arch, info.Hostname)
}

func TestCheckPrerequisites(t *testing.T) {
	err := CheckPrerequisites()
	
	if !IsMacOS() {
		// Should fail on non-macOS systems
		if err == nil {
			t.Error("expected error on non-macOS system")
		}
		t.Logf("Prerequisites check failed as expected on non-macOS: %v", err)
	} else {
		// On macOS, depends on whether Homebrew is installed
		if err != nil {
			t.Logf("Prerequisites check failed (Homebrew may not be installed): %v", err)
		} else {
			t.Log("Prerequisites check passed")
		}
	}
}

func TestCheckAllCommands(t *testing.T) {
	commands := []string{"ls", "echo", "this-does-not-exist"}
	results := CheckAllCommands(commands...)
	
	if len(results) != len(commands) {
		t.Errorf("expected %d results, got %d", len(commands), len(results))
	}
	
	// ls should exist
	if lsCheck, ok := results["ls"]; ok {
		if !lsCheck.Exists {
			t.Error("expected ls to exist")
		}
	} else {
		t.Error("ls check missing from results")
	}
	
	// echo should exist
	if echoCheck, ok := results["echo"]; ok {
		if !echoCheck.Exists {
			t.Error("expected echo to exist")
		}
	} else {
		t.Error("echo check missing from results")
	}
	
	// this-does-not-exist should not exist
	if fakeCheck, ok := results["this-does-not-exist"]; ok {
		if fakeCheck.Exists {
			t.Error("expected fake command to not exist")
		}
	} else {
		t.Error("fake command check missing from results")
	}
}

func TestFormatCommandCheck(t *testing.T) {
	t.Run("format existing command", func(t *testing.T) {
		check := &CommandCheck{
			Name:    "brew",
			Exists:  true,
			Path:    "/usr/local/bin/brew",
			Version: "Homebrew 4.0.0",
		}
		
		formatted := FormatCommandCheck(check)
		if formatted == "" {
			t.Error("expected non-empty formatted string")
		}
		if !contains(formatted, "brew") {
			t.Error("expected formatted string to contain command name")
		}
		if !contains(formatted, "✓") {
			t.Error("expected formatted string to contain success indicator")
		}
		
		t.Logf("Formatted: %s", formatted)
	})
	
	t.Run("format non-existent command", func(t *testing.T) {
		check := &CommandCheck{
			Name:   "nonexistent",
			Exists: false,
			Error:  nil,
		}
		
		formatted := FormatCommandCheck(check)
		if formatted == "" {
			t.Error("expected non-empty formatted string")
		}
		if !contains(formatted, "nonexistent") {
			t.Error("expected formatted string to contain command name")
		}
		if !contains(formatted, "✗") {
			t.Error("expected formatted string to contain failure indicator")
		}
		
		t.Logf("Formatted: %s", formatted)
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		someContains(s, substr)))
}

func someContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

