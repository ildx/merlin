package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ildx/merlin/internal/models"
)

// PackageItem wraps a package with selection state
type PackageItem struct {
	Package  models.BrewPackage
	Selected bool
}

// PackageSelectorModel allows interactive package selection
type PackageSelectorModel struct {
	title      string
	items      []PackageItem
	cursor     int
	selected   map[int]bool
	confirmed  bool
	cancelled  bool
	width      int
	height     int
	viewOffset int
}

// NewPackageSelectorModel creates a new package selector
func NewPackageSelectorModel(title string, packages []models.BrewPackage) PackageSelectorModel {
	items := make([]PackageItem, len(packages))
	for i, pkg := range packages {
		items[i] = PackageItem{Package: pkg, Selected: false}
	}

	return PackageSelectorModel{
		title:    title,
		items:    items,
		selected: make(map[int]bool),
	}
}

func (m PackageSelectorModel) Init() tea.Cmd {
	return nil
}

func (m PackageSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Scroll up if needed
				if m.cursor < m.viewOffset {
					m.viewOffset = m.cursor
				}
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				// Scroll down if needed
				maxVisible := m.height - 10 // Account for header/footer
				if maxVisible < 5 {
					maxVisible = 5
				}
				if m.cursor >= m.viewOffset+maxVisible {
					m.viewOffset = m.cursor - maxVisible + 1
				}
			}

		case " ", "x":
			// Toggle selection
			m.selected[m.cursor] = !m.selected[m.cursor]
			m.items[m.cursor].Selected = m.selected[m.cursor]

		case "a":
			// Select all
			for i := range m.items {
				m.selected[i] = true
				m.items[i].Selected = true
			}

		case "n":
			// Select none
			m.selected = make(map[int]bool)
			for i := range m.items {
				m.items[i].Selected = false
			}

		case "enter":
			m.confirmed = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m PackageSelectorModel) View() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render(m.title) + "\n\n")

	// Group by category
	categorized := m.groupByCategory()
	categories := make([]string, 0, len(categorized))
	for cat := range categorized {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	// Calculate visible range
	maxVisible := m.height - 10
	if maxVisible < 5 {
		maxVisible = 10
	}

	// Flatten items with category headers for display
	displayItems := []struct {
		isCategory bool
		category   string
		itemIdx    int
	}{}

	itemIndex := 0
	for _, cat := range categories {
		displayItems = append(displayItems, struct {
			isCategory bool
			category   string
			itemIdx    int
		}{isCategory: true, category: cat})

		for range categorized[cat] {
			displayItems = append(displayItems, struct {
				isCategory bool
				category   string
				itemIdx    int
			}{isCategory: false, itemIdx: itemIndex})
			itemIndex++
		}
	}

	// Find cursor position in display items
	cursorDisplayIdx := 0
	itemsSeen := 0
	for i, di := range displayItems {
		if !di.isCategory {
			if itemsSeen == m.cursor {
				cursorDisplayIdx = i
				break
			}
			itemsSeen++
		}
	}

	// Calculate view window
	startIdx := 0
	endIdx := len(displayItems)

	if len(displayItems) > maxVisible {
		// Center cursor in view
		startIdx = cursorDisplayIdx - maxVisible/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + maxVisible
		if endIdx > len(displayItems) {
			endIdx = len(displayItems)
			startIdx = endIdx - maxVisible
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	// Render visible items
	itemIndex = 0
	for idx := startIdx; idx < endIdx && idx < len(displayItems); idx++ {
		di := displayItems[idx]

		if di.isCategory {
			// Category header
			catStyle := lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true).
				MarginTop(1)
			s.WriteString(catStyle.Render(fmt.Sprintf("▸ %s", di.category)) + "\n")
		} else {
			// Package item
			realIdx := di.itemIdx
			item := m.items[realIdx]

			checkbox := "☐"
			if m.selected[realIdx] {
				checkbox = "☑"
			}

			cursor := "  "
			style := normalItemStyle

			if realIdx == m.cursor {
				cursor = "▸ "
				style = selectedItemStyle
			}

			line := fmt.Sprintf("%s%s %s", cursor, checkbox, item.Package.Name)
			s.WriteString(style.Render(line) + "\n")

			// Show description for selected cursor item
			if realIdx == m.cursor && item.Package.Description != "" {
				desc := lipgloss.NewStyle().
					Foreground(mutedColor).
					PaddingLeft(6).
					Italic(true).
					Render(item.Package.Description)
				s.WriteString(desc + "\n")
			}
		}
	}

	// Show scroll indicator
	if len(displayItems) > maxVisible {
		scrollInfo := lipgloss.NewStyle().Foreground(mutedColor).Render(
			fmt.Sprintf("  (showing %d-%d of %d)", startIdx+1, endIdx, len(m.items)))
		s.WriteString("\n" + scrollInfo)
	}

	// Stats
	selectedCount := len(m.selected)
	stats := fmt.Sprintf("\nSelected: %d/%d", selectedCount, len(m.items))
	s.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(stats) + "\n")

	// Help
	help := helpStyle.Render("\n↑/↓: navigate • space: toggle • a: all • n: none • enter: confirm • esc: cancel")
	s.WriteString(help)

	return boxStyle.Render(s.String())
}

func (m PackageSelectorModel) groupByCategory() map[string][]PackageItem {
	categorized := make(map[string][]PackageItem)

	for _, item := range m.items {
		category := item.Package.Category
		if category == "" {
			category = "uncategorized"
		}
		categorized[category] = append(categorized[category], item)
	}

	return categorized
}

// GetSelectedPackages returns the packages that were selected
func (m PackageSelectorModel) GetSelectedPackages() []models.BrewPackage {
	var selected []models.BrewPackage
	for idx := range m.selected {
		if m.selected[idx] {
			selected = append(selected, m.items[idx].Package)
		}
	}
	return selected
}

// IsConfirmed returns true if user confirmed selection
func (m PackageSelectorModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns true if user cancelled
func (m PackageSelectorModel) IsCancelled() bool {
	return m.cancelled
}
