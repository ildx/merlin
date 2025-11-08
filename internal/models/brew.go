package models

// BrewConfig represents the complete brew.toml configuration
type BrewConfig struct {
	Metadata Metadata      `toml:"metadata"`
	Formulae []BrewPackage `toml:"brew"`
	Casks    []BrewPackage `toml:"cask"`
}

// BrewPackage represents a single Homebrew formula or cask
type BrewPackage struct {
	Name         string   `toml:"name"`
	Description  string   `toml:"description"`
	Category     string   `toml:"category"`
	Dependencies []string `toml:"dependencies"`
}

// GetAllPackages returns all formulae and casks combined
func (c *BrewConfig) GetAllPackages() []BrewPackage {
	return append(c.Formulae, c.Casks...)
}

// GetByCategory returns all packages in a specific category
func (c *BrewConfig) GetByCategory(category string) []BrewPackage {
	var packages []BrewPackage
	for _, pkg := range c.GetAllPackages() {
		if pkg.Category == category {
			packages = append(packages, pkg)
		}
	}
	return packages
}

// GetCategories returns a unique list of all categories
func (c *BrewConfig) GetCategories() []string {
	categoryMap := make(map[string]bool)
	for _, pkg := range c.GetAllPackages() {
		if pkg.Category != "" {
			categoryMap[pkg.Category] = true
		}
	}
	
	categories := make([]string, 0, len(categoryMap))
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	return categories
}

