package models

// MASConfig represents the complete mas.toml configuration
type MASConfig struct {
	Metadata Metadata `toml:"metadata"`
	Apps     []MASApp `toml:"app"`
}

// MASApp represents a single Mac App Store application
type MASApp struct {
	Name         string   `toml:"name"`
	ID           int      `toml:"id"`
	Description  string   `toml:"description"`
	Category     string   `toml:"category"`
	Dependencies []string `toml:"dependencies"`
}

// GetByCategory returns all apps in a specific category
func (c *MASConfig) GetByCategory(category string) []MASApp {
	var apps []MASApp
	for _, app := range c.Apps {
		if app.Category == category {
			apps = append(apps, app)
		}
	}
	return apps
}

// GetCategories returns a unique list of all categories
func (c *MASConfig) GetCategories() []string {
	categoryMap := make(map[string]bool)
	for _, app := range c.Apps {
		if app.Category != "" {
			categoryMap[app.Category] = true
		}
	}
	
	categories := make([]string, 0, len(categoryMap))
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	return categories
}

// FindByID finds an app by its App Store ID
func (c *MASConfig) FindByID(id int) *MASApp {
	for _, app := range c.Apps {
		if app.ID == id {
			return &app
		}
	}
	return nil
}

// FindByName finds an app by its name
func (c *MASConfig) FindByName(name string) *MASApp {
	for _, app := range c.Apps {
		if app.Name == name {
			return &app
		}
	}
	return nil
}

