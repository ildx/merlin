package models

// RootMerlinConfig represents the root merlin.toml configuration
type RootMerlinConfig struct {
	Metadata   Metadata           `toml:"metadata"`
	Settings   Settings           `toml:"settings"`
	Preinstall PreinstallSettings `toml:"preinstall"`
	Profiles   []Profile          `toml:"profile"`
}

// Settings contains global configuration settings
type Settings struct {
	AutoLink            bool   `toml:"auto_link"`
	ConfirmBeforeInstall bool   `toml:"confirm_before_install"`
	ConflictStrategy    string `toml:"conflict_strategy"`
	HomeDir             string `toml:"home_dir"`
	ConfigDir           string `toml:"config_dir"`
}

// PreinstallSettings defines system requirements installed before profiles
type PreinstallSettings struct {
	Tools []string `toml:"tools"`
}

// Profile represents a machine-specific configuration profile
type Profile struct {
	Name        string   `toml:"name"`
	Hostname    string   `toml:"hostname"`
	Default     bool     `toml:"default"`
	Description string   `toml:"description"`
	Tools       []string `toml:"tools"`
}

// GetDefaultProfile returns the default profile, or nil if none exists
func (c *RootMerlinConfig) GetDefaultProfile() *Profile {
	for _, profile := range c.Profiles {
		if profile.Default {
			return &profile
		}
	}
	return nil
}

// GetProfileByName returns a profile by name, or nil if not found
func (c *RootMerlinConfig) GetProfileByName(name string) *Profile {
	for _, profile := range c.Profiles {
		if profile.Name == name {
			return &profile
		}
	}
	return nil
}

// GetProfileByHostname returns a profile by hostname, or nil if not found
func (c *RootMerlinConfig) GetProfileByHostname(hostname string) *Profile {
	for _, profile := range c.Profiles {
		if profile.Hostname == hostname {
			return &profile
		}
	}
	return nil
}

