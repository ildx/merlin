package parser

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/ildx/merlin/internal/models"
)

// ParseBrewTOML parses a brew.toml file
func ParseBrewTOML(path string) (*models.BrewConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read brew.toml: %w", err)
	}

	var config models.BrewConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse brew.toml: %w", err)
	}

	return &config, nil
}

// ParseMASTOML parses a mas.toml file
func ParseMASTOML(path string) (*models.MASConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read mas.toml: %w", err)
	}

	var config models.MASConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse mas.toml: %w", err)
	}

	return &config, nil
}

// ParseRootMerlinTOML parses the root merlin.toml file
func ParseRootMerlinTOML(path string) (*models.RootMerlinConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read root merlin.toml: %w", err)
	}

	var config models.RootMerlinConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse root merlin.toml: %w", err)
	}

	// Set defaults for settings if not provided
	setRootConfigDefaults(&config)

	return &config, nil
}

// ParseToolMerlinTOML parses a per-tool merlin.toml file
func ParseToolMerlinTOML(path string) (*models.ToolMerlinConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read tool merlin.toml: %w", err)
	}

	var config models.ToolMerlinConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse tool merlin.toml: %w", err)
	}

	return &config, nil
}

// setRootConfigDefaults sets default values for root config if not specified
func setRootConfigDefaults(config *models.RootMerlinConfig) {
	// Set default settings if not provided
	if config.Settings.ConflictStrategy == "" {
		config.Settings.ConflictStrategy = "interactive"
	}
	if config.Settings.HomeDir == "" {
		config.Settings.HomeDir = "~"
	}
	if config.Settings.ConfigDir == "" {
		config.Settings.ConfigDir = "{home_dir}/.config"
	}
}

// ValidateBrewConfig validates a BrewConfig
func ValidateBrewConfig(config *models.BrewConfig) error {
	if len(config.Formulae) == 0 && len(config.Casks) == 0 {
		return fmt.Errorf("brew.toml must contain at least one formula or cask")
	}

	// Check for duplicate package names
	seen := make(map[string]bool)
	for _, pkg := range config.GetAllPackages() {
		if seen[pkg.Name] {
			return fmt.Errorf("duplicate package name: %s", pkg.Name)
		}
		seen[pkg.Name] = true
	}

	return nil
}

// ValidateMASConfig validates a MASConfig
func ValidateMASConfig(config *models.MASConfig) error {
	if len(config.Apps) == 0 {
		return fmt.Errorf("mas.toml must contain at least one app")
	}

	// Check for duplicate app IDs
	seenIDs := make(map[int]bool)
	seenNames := make(map[string]bool)
	for _, app := range config.Apps {
		if seenIDs[app.ID] {
			return fmt.Errorf("duplicate app ID: %d", app.ID)
		}
		if seenNames[app.Name] {
			return fmt.Errorf("duplicate app name: %s", app.Name)
		}
		seenIDs[app.ID] = true
		seenNames[app.Name] = true
	}

	return nil
}

// ValidateRootMerlinConfig validates a RootMerlinConfig
func ValidateRootMerlinConfig(config *models.RootMerlinConfig) error {
	// Check for duplicate profile names
	seen := make(map[string]bool)
	defaultCount := 0
	
	for _, profile := range config.Profiles {
		if seen[profile.Name] {
			return fmt.Errorf("duplicate profile name: %s", profile.Name)
		}
		seen[profile.Name] = true
		
		if profile.Default {
			defaultCount++
		}
	}

	if defaultCount > 1 {
		return fmt.Errorf("only one profile can be marked as default")
	}

	// Validate conflict strategy
	validStrategies := map[string]bool{
		"backup":      true,
		"skip":        true,
		"overwrite":   true,
		"interactive": true,
	}
	if !validStrategies[config.Settings.ConflictStrategy] {
		return fmt.Errorf("invalid conflict_strategy: %s (must be: backup, skip, overwrite, or interactive)", 
			config.Settings.ConflictStrategy)
	}

	return nil
}

// ValidateToolMerlinConfig validates a ToolMerlinConfig
func ValidateToolMerlinConfig(config *models.ToolMerlinConfig) error {
	if config.Tool.Name == "" {
		return fmt.Errorf("tool.name is required")
	}

	// Validate links
	for i, link := range config.Links {
		if link.Target == "" {
			return fmt.Errorf("link[%d]: target is required", i)
		}
	}

	return nil
}

