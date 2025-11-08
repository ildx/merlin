package models

// ToolMerlinConfig represents a per-tool merlin.toml configuration
type ToolMerlinConfig struct {
	Tool    ToolInfo       `toml:"tool"`
	Links   []Link         `toml:"link"`
	Scripts ScriptsSection `toml:"scripts"`
}

// ToolInfo contains basic information about a tool
type ToolInfo struct {
	Name         string   `toml:"name"`
	Description  string   `toml:"description"`
	Dependencies []string `toml:"dependencies"`
}

// Link represents a symlink configuration
type Link struct {
	Source string     `toml:"source"` // Source path relative to tool's config directory
	Target string     `toml:"target"` // Target path (can contain variables like {config_dir})
	Files  []FileLink `toml:"files"`  // Optional: multiple files to same base target
}

// FileLink represents a file to be linked within a base target
type FileLink struct {
	Source string `toml:"source"` // Source file path
	Target string `toml:"target"` // Target file name (relative to parent Link.Target)
}

// ScriptsSection contains script execution configuration
type ScriptsSection struct {
	Directory string   `toml:"directory"` // Directory containing scripts (relative to tool root)
	Scripts   []string `toml:"scripts"`   // Scripts to execute in order
}

// HasScripts returns true if the tool has scripts to execute
func (c *ToolMerlinConfig) HasScripts() bool {
	return len(c.Scripts.Scripts) > 0
}

// HasLinks returns true if the tool has symlinks to create
func (c *ToolMerlinConfig) HasLinks() bool {
	return len(c.Links) > 0
}

// HasDependencies returns true if the tool has dependencies
func (c *ToolMerlinConfig) HasDependencies() bool {
	return len(c.Tool.Dependencies) > 0
}

