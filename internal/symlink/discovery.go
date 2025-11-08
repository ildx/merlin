package symlink

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ildx/merlin/internal/config"
	"github.com/ildx/merlin/internal/models"
	"github.com/ildx/merlin/internal/parser"
)

// ToolConfig represents a tool's symlink configuration
type ToolConfig struct {
	Name         string
	Description  string
	ToolRoot     string // Absolute path to config/TOOL/
	ConfigDir    string // Absolute path to config/TOOL/config/
	Links        []ResolvedLink
	Dependencies []string
	HasMerlinTOML bool
}

// ResolvedLink represents a fully resolved symlink with expanded variables
type ResolvedLink struct {
	Source string // Absolute source path
	Target string // Absolute target path
	IsDir  bool   // True if source is a directory
}

// Variables holds the variable values for expansion
type Variables struct {
	HomeDir   string
	ConfigDir string
}

// DiscoverTools discovers all tools in the dotfiles repository
func DiscoverTools(repo *config.DotfilesRepo, vars Variables) ([]*ToolConfig, error) {
	tools, err := repo.ListTools()
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	toolConfigs := make([]*ToolConfig, 0, len(tools))
	
	for _, toolName := range tools {
		toolConfig, err := DiscoverToolConfig(repo, toolName, vars)
		if err != nil {
			// Skip tools that can't be discovered
			continue
		}
		toolConfigs = append(toolConfigs, toolConfig)
	}

	return toolConfigs, nil
}

// DiscoverToolConfig discovers configuration for a single tool
func DiscoverToolConfig(repo *config.DotfilesRepo, toolName string, vars Variables) (*ToolConfig, error) {
	toolRoot := repo.GetToolRoot(toolName)
	configDir := repo.GetToolConfigDir(toolName)
	merlinPath := repo.GetToolMerlinConfig(toolName)

	toolConfig := &ToolConfig{
		Name:      toolName,
		ToolRoot:  toolRoot,
		ConfigDir: configDir,
		Links:     []ResolvedLink{},
	}

	// Check if tool has a merlin.toml
	if _, err := os.Stat(merlinPath); err == nil {
		toolConfig.HasMerlinTOML = true
		
		// Parse the merlin.toml
		merlinConfig, err := parser.ParseToolMerlinTOML(merlinPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", merlinPath, err)
		}

		toolConfig.Description = merlinConfig.Tool.Description
		toolConfig.Dependencies = merlinConfig.Tool.Dependencies

		// Process links
		for _, link := range merlinConfig.Links {
			resolvedLinks, err := resolveLink(link, toolRoot, configDir, vars)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve link for %s: %w", toolName, err)
			}
			toolConfig.Links = append(toolConfig.Links, resolvedLinks...)
		}
	} else {
		// Use default: config/ → ~/.config/TOOL/
		defaultTarget := filepath.Join(vars.ConfigDir, toolName)
		
		// Check if config directory exists
		if info, err := os.Stat(configDir); err == nil && info.IsDir() {
			toolConfig.Links = []ResolvedLink{
				{
					Source: configDir,
					Target: defaultTarget,
					IsDir:  true,
				},
			}
		}
	}

	return toolConfig, nil
}

// resolveLink resolves a link configuration into actual source/target pairs
func resolveLink(link models.Link, toolRoot, configDir string, vars Variables) ([]ResolvedLink, error) {
	var results []ResolvedLink

	// Expand target variables
	target := expandVariables(link.Target, vars)

	// If there are specific files, handle them
	if len(link.Files) > 0 {
		for _, file := range link.Files {
			// Source is relative to tool root
			source := filepath.Join(toolRoot, file.Source)
			// Target is relative to the link target
			fileTarget := filepath.Join(target, file.Target)

			// Check if source exists
			info, err := os.Stat(source)
			if err != nil {
				continue // Skip non-existent sources
			}

			results = append(results, ResolvedLink{
				Source: source,
				Target: fileTarget,
				IsDir:  info.IsDir(),
			})
		}
		return results, nil
	}

	// Determine source
	var source string
	if link.Source != "" {
		// Explicit source relative to tool root
		source = filepath.Join(toolRoot, link.Source)
	} else {
		// Implicit source: config/ directory
		source = configDir
	}

	// Check if source exists
	info, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("source does not exist: %s", source)
	}

	results = append(results, ResolvedLink{
		Source: source,
		Target: target,
		IsDir:  info.IsDir(),
	})

	return results, nil
}

// expandVariables expands {var} patterns in a string
func expandVariables(s string, vars Variables) string {
	s = strings.ReplaceAll(s, "{home_dir}", vars.HomeDir)
	s = strings.ReplaceAll(s, "{config_dir}", vars.ConfigDir)
	
	// Handle ~ expansion
	if strings.HasPrefix(s, "~/") {
		s = filepath.Join(vars.HomeDir, s[2:])
	} else if s == "~" {
		s = vars.HomeDir
	}

	return s
}

// GetDefaultVariables returns default variable values
func GetDefaultVariables() (Variables, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Variables{}, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config")

	return Variables{
		HomeDir:   homeDir,
		ConfigDir: configDir,
	}, nil
}

// GetVariablesFromRoot gets variables from root merlin.toml, with defaults as fallback
func GetVariablesFromRoot(rootConfig *models.RootMerlinConfig) (Variables, error) {
	vars, err := GetDefaultVariables()
	if err != nil {
		return vars, err
	}

	// Override with values from config if present
	if rootConfig.Settings.HomeDir != "" {
		vars.HomeDir = expandVariables(rootConfig.Settings.HomeDir, vars)
	}
	if rootConfig.Settings.ConfigDir != "" {
		vars.ConfigDir = expandVariables(rootConfig.Settings.ConfigDir, vars)
	}

	return vars, nil
}

// ToolExists checks if a tool directory exists
func ToolExists(repo *config.DotfilesRepo, toolName string) bool {
	return repo.ToolExists(toolName)
}

