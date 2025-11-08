package models

import "fmt"

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

// ScriptItem represents a single script with optional tags.
// Backward compatibility: a plain string in the TOML array becomes ScriptItem{File: <string>}.
// Extended form: { file = "script.sh", tags = ["tag1", "tag2"] }
// Alternate key: { name = "script.sh" } is also accepted for convenience.
type ScriptItem struct {
	File string   // Actual script file name (relative to scripts directory)
	Tags []string // Optional tags used for selection/filtering
}

// UnmarshalTOML implements custom decoding to support both string and table entries.
func (s *ScriptItem) UnmarshalTOML(data any) error {
	switch v := data.(type) {
	case string:
		s.File = v
		return nil
	case map[string]any:
		// Accept either "file" or legacy/alternate "name"
		if fileVal, ok := v["file"].(string); ok {
			s.File = fileVal
		} else if nameVal, ok := v["name"].(string); ok {
			s.File = nameVal
		}
		if s.File == "" {
			return fmt.Errorf("script item missing 'file' or 'name' field")
		}
		if rawTags, ok := v["tags"].([]any); ok {
			for _, t := range rawTags {
				if ts, ok := t.(string); ok {
					s.Tags = append(s.Tags, ts)
				}
			}
		}
		return nil
	default:
		return fmt.Errorf("invalid script item type %T", v)
	}
}

// ScriptsSection contains script execution configuration
type ScriptsSection struct {
	Directory string       `toml:"directory"` // Directory containing scripts (relative to tool root)
	Scripts   []ScriptItem `toml:"scripts"`   // Scripts to execute in order
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

// HasScriptTag returns true if any script item includes the specified tag.
func (c *ToolMerlinConfig) HasScriptTag(tag string) bool {
	for _, s := range c.Scripts.Scripts {
		for _, t := range s.Tags {
			if t == tag {
				return true
			}
		}
	}
	return false
}

// FilterScriptsByTag returns scripts that include at least one of the provided tags.
func (c *ToolMerlinConfig) FilterScriptsByTag(tags []string) []ScriptItem {
	if len(tags) == 0 {
		return c.Scripts.Scripts
	}
	tagSet := map[string]struct{}{}
	for _, t := range tags {
		tagSet[t] = struct{}{}
	}
	var filtered []ScriptItem
	for _, s := range c.Scripts.Scripts {
		for _, st := range s.Tags {
			if _, ok := tagSet[st]; ok {
				filtered = append(filtered, s)
				break
			}
		}
	}
	return filtered
}
