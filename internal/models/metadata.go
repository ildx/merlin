package models

// Metadata contains information about a TOML configuration file
type Metadata struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
}

