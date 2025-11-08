package config

import (
	"errors"
	"os"
	"path/filepath"
)

var (
	// ErrDotfilesNotFound is returned when the dotfiles repository cannot be located
	ErrDotfilesNotFound = errors.New("dotfiles repository not found")
	
	// ErrNotADotfilesRepo is returned when a directory doesn't contain a valid merlin.toml
	ErrNotADotfilesRepo = errors.New("not a valid dotfiles repository (no merlin.toml found)")
)

const (
	// RootConfigFile is the name of the root configuration file
	RootConfigFile = "merlin.toml"
	
	// ConfigDir is the expected name of the config directory
	ConfigDir = "config"
	
	// EnvVarDotfiles is the environment variable name for the dotfiles path
	EnvVarDotfiles = "MERLIN_DOTFILES"
)

// DotfilesRepo represents a dotfiles repository
type DotfilesRepo struct {
	Root      string // Absolute path to the dotfiles repository root
	ConfigDir string // Absolute path to the config directory
}

// FindDotfilesRepo attempts to locate the dotfiles repository in the following order:
// 1. MERLIN_DOTFILES environment variable
// 2. Current directory (if it contains merlin.toml)
// 3. Parent directories (walking up until merlin.toml is found)
func FindDotfilesRepo() (*DotfilesRepo, error) {
	// Strategy 1: Check environment variable
	if envPath := os.Getenv(EnvVarDotfiles); envPath != "" {
		if repo, err := LoadDotfilesRepo(envPath); err == nil {
			return repo, nil
		}
	}
	
	// Strategy 2 & 3: Check current directory and walk up
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	
	return findDotfilesInPath(cwd)
}

// LoadDotfilesRepo loads a dotfiles repository from a specific path
func LoadDotfilesRepo(path string) (*DotfilesRepo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	
	// Check if the path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, ErrDotfilesNotFound
	}
	
	// Check if merlin.toml exists
	configPath := filepath.Join(absPath, RootConfigFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, ErrNotADotfilesRepo
	}
	
	// Verify this is a root config, not a per-tool config
	// Root repos should have a config/ directory
	configDir := filepath.Join(absPath, ConfigDir)
	if info, err := os.Stat(configDir); err != nil || !info.IsDir() {
		return nil, ErrNotADotfilesRepo
	}
	
	// Build the repo struct
	repo := &DotfilesRepo{
		Root:      absPath,
		ConfigDir: configDir,
	}
	
	return repo, nil
}

// findDotfilesInPath walks up from the given path to find a dotfiles repository
func findDotfilesInPath(startPath string) (*DotfilesRepo, error) {
	currentPath := startPath
	
	// Walk up the directory tree
	for {
		// Skip if we're inside a tool directory (parent is named "config")
		// This prevents matching per-tool merlin.toml files
		parentDir := filepath.Dir(currentPath)
		if filepath.Base(parentDir) != ConfigDir {
			// Try to load from current path
			if repo, err := LoadDotfilesRepo(currentPath); err == nil {
				return repo, nil
			}
		}
		
		// Get parent directory
		parentPath := filepath.Dir(currentPath)
		
		// If we've reached the root, stop
		if parentPath == currentPath {
			break
		}
		
		currentPath = parentPath
	}
	
	return nil, ErrDotfilesNotFound
}

// IsValidDotfilesRepo checks if a path is a valid dotfiles repository
func IsValidDotfilesRepo(path string) bool {
	_, err := LoadDotfilesRepo(path)
	return err == nil
}

// GetToolConfigDir returns the path to a specific tool's config directory
func (r *DotfilesRepo) GetToolConfigDir(toolName string) string {
	return filepath.Join(r.ConfigDir, toolName, ConfigDir)
}

// GetToolRoot returns the path to a specific tool's root directory
func (r *DotfilesRepo) GetToolRoot(toolName string) string {
	return filepath.Join(r.ConfigDir, toolName)
}

// GetToolMerlinConfig returns the path to a tool's merlin.toml file
func (r *DotfilesRepo) GetToolMerlinConfig(toolName string) string {
	return filepath.Join(r.ConfigDir, toolName, RootConfigFile)
}

// GetRootMerlinConfig returns the path to the root merlin.toml file
func (r *DotfilesRepo) GetRootMerlinConfig() string {
	return filepath.Join(r.Root, RootConfigFile)
}

// ToolExists checks if a tool directory exists in the dotfiles repo
func (r *DotfilesRepo) ToolExists(toolName string) bool {
	toolPath := r.GetToolRoot(toolName)
	info, err := os.Stat(toolPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ListTools returns a list of all tool directories in the config directory
func (r *DotfilesRepo) ListTools() ([]string, error) {
	entries, err := os.ReadDir(r.ConfigDir)
	if err != nil {
		return nil, err
	}
	
	var tools []string
	for _, entry := range entries {
		if entry.IsDir() {
			tools = append(tools, entry.Name())
		}
	}
	
	return tools, nil
}

